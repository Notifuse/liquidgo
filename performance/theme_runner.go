package performance

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/Notifuse/liquidgo/liquid"
	"github.com/Notifuse/liquidgo/liquid/tags"
	"github.com/Notifuse/liquidgo/performance/shopify"
)

// FileSystem implements liquid.FileSystem for template loading
type FileSystem struct {
	path string
}

// NewFileSystem creates a new FileSystem
func NewFileSystem(path string) *FileSystem {
	return &FileSystem{path: path}
}

// ReadTemplateFile reads a template file
func (fs *FileSystem) ReadTemplateFile(templatePath string) (string, error) {
	data, err := os.ReadFile(filepath.Join(fs.path, templatePath+".liquid"))
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// testInfo holds information about a test template
type testInfo struct {
	liquid       string
	layout       string
	templateName string
}

// compiledTest holds a pre-compiled template
type compiledTest struct {
	tmpl    *liquid.Template
	layout  *liquid.Template
	assigns map[string]interface{}
}

// ThemeRunner simulates Shopify's template rendering
type ThemeRunner struct {
	tests         []*testInfo
	compiledTests []*compiledTest
	env           *liquid.Environment
}

// NewThemeRunner creates a new ThemeRunner instance
// Will load all templates into memory to avoid profiling IO
func NewThemeRunner() (*ThemeRunner, error) {
	// Create environment with standard tags
	env := liquid.NewEnvironment()
	tags.RegisterStandardTags(env)
	
	// Register Shopify custom tags and filters
	shopify.RegisterAll(env)

	runner := &ThemeRunner{
		env:   env,
		tests: []*testInfo{},
	}

	// Load all test templates
	testsDir := filepath.Join("..", "reference-liquid", "performance", "tests")
	testDirs, err := os.ReadDir(testsDir)
	if err != nil {
		return nil, err
	}

	for _, testDir := range testDirs {
		if !testDir.IsDir() {
			continue
		}

		themeDir := filepath.Join(testsDir, testDir.Name())
		files, err := os.ReadDir(themeDir)
		if err != nil {
			continue
		}

		// Load theme layout if it exists
		var layoutContent string
		themePath := filepath.Join(themeDir, "theme.liquid")
		if data, err := os.ReadFile(themePath); err == nil {
			layoutContent = string(data)
		}

		// Load each template file
		for _, file := range files {
			if file.IsDir() || file.Name() == "theme.liquid" {
				continue
			}

			if !strings.HasSuffix(file.Name(), ".liquid") {
				continue
			}

			templatePath := filepath.Join(themeDir, file.Name())
			data, err := os.ReadFile(templatePath)
			if err != nil {
				continue
			}

			runner.tests = append(runner.tests, &testInfo{
				liquid:       string(data),
				layout:       layoutContent,
				templateName: templatePath,
			})
		}
	}

	// Pre-compile all tests
	if err := runner.compileAllTests(); err != nil {
		return nil, err
	}

	return runner, nil
}

// Compile benchmarks just the compilation portion
func (tr *ThemeRunner) Compile() error {
	for _, test := range tr.tests {
		_, err := liquid.ParseTemplate(test.liquid, &liquid.TemplateOptions{
			Environment: tr.env,
		})
		if err != nil {
			return err
		}

		if test.layout != "" {
			_, err = liquid.ParseTemplate(test.layout, &liquid.TemplateOptions{
				Environment: tr.env,
			})
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// Tokenize benchmarks just the tokenization portion
func (tr *ThemeRunner) Tokenize() error {
	ss := liquid.NewStringScanner("")
	for _, test := range tr.tests {
		tokenizer := liquid.NewTokenizer(test.liquid, ss, true, nil, false)
		for tokenizer.Shift() != "" {
		}

		if test.layout != "" {
			tokenizer = liquid.NewTokenizer(test.layout, ss, true, nil, false)
			for tokenizer.Shift() != "" {
			}
		}
	}
	return nil
}

// Run benchmarks rendering and compiling at the same time
func (tr *ThemeRunner) Run() error {
	assigns, err := shopify.Tables()
	if err != nil {
		return err
	}

	for _, test := range tr.tests {
		if err := tr.compileAndRender(test.liquid, test.layout, assigns, test.templateName); err != nil {
			return err
		}
	}
	return nil
}

// Render benchmarks just the render portion
func (tr *ThemeRunner) Render() error {
	for _, test := range tr.compiledTests {
		if test.layout != nil {
			// Render template first
			output := test.tmpl.Render(test.assigns, nil)
			
			// Set content_for_layout
			layoutAssigns := make(map[string]interface{})
			for k, v := range test.assigns {
				layoutAssigns[k] = v
			}
			layoutAssigns["content_for_layout"] = output
			
			// Render layout
			test.layout.Render(layoutAssigns, nil)
		} else {
			test.tmpl.Render(test.assigns, nil)
		}
	}
	return nil
}

// compileAllTests pre-compiles all tests for render-only benchmarks
func (tr *ThemeRunner) compileAllTests() error {
	assigns, err := shopify.Tables()
	if err != nil {
		return err
	}

	tr.compiledTests = make([]*compiledTest, 0, len(tr.tests))
	
	for _, test := range tr.tests {
		compiled, err := tr.compileTest(test.liquid, test.layout, assigns, test.templateName)
		if err != nil {
			return err
		}
		tr.compiledTests = append(tr.compiledTests, compiled)
	}
	
	return nil
}

// compileTest compiles a single test
func (tr *ThemeRunner) compileTest(templateSource, layoutSource string, assigns map[string]interface{}, templateFile string) (*compiledTest, error) {
	// Get page template name
	pageTemplate := strings.TrimSuffix(filepath.Base(templateFile), filepath.Ext(templateFile))

	// Create template with assigns and registers
	tmpl, err := tr.initTemplate(pageTemplate, templateFile)
	if err != nil {
		return nil, err
	}

	// Parse template
	err = tmpl.Parse(templateSource, &liquid.TemplateOptions{
		Environment: tr.env,
	})
	if err != nil {
		return nil, err
	}

	result := &compiledTest{
		tmpl:    tmpl,
		assigns: assigns,
	}

	// Parse layout if present
	if layoutSource != "" {
		layoutTmpl, err := tr.initTemplate(pageTemplate, templateFile)
		if err != nil {
			return nil, err
		}

		err = layoutTmpl.Parse(layoutSource, &liquid.TemplateOptions{
			Environment: tr.env,
		})
		if err != nil {
			return nil, err
		}

		result.layout = layoutTmpl
	}

	return result, nil
}

// compileAndRender compiles and renders a template
func (tr *ThemeRunner) compileAndRender(templateSource, layoutSource string, assigns map[string]interface{}, templateFile string) error {
	compiled, err := tr.compileTest(templateSource, layoutSource, assigns, templateFile)
	if err != nil {
		return err
	}

	if compiled.layout != nil {
		// Render template
		output := compiled.tmpl.Render(compiled.assigns, nil)
		
		// Set content_for_layout
		layoutAssigns := make(map[string]interface{})
		for k, v := range compiled.assigns {
			layoutAssigns[k] = v
		}
		layoutAssigns["content_for_layout"] = output
		
		// Render layout
		compiled.layout.Render(layoutAssigns, nil)
	} else {
		compiled.tmpl.Render(compiled.assigns, nil)
	}

	return nil
}

// initTemplate sets up a new template with necessary assigns and registers
func (tr *ThemeRunner) initTemplate(pageTemplate, templateFile string) (*liquid.Template, error) {
	tmpl := liquid.NewTemplate(&liquid.TemplateOptions{
		Environment: tr.env,
	})

	// Set template assigns
	tmpl.Assigns()["page_title"] = "Page title"
	tmpl.Assigns()["template"] = pageTemplate
	tmpl.Assigns()["content_for_header"] = "" // Empty content_for_header

	// Set file system in registers
	fileSystem := NewFileSystem(filepath.Dir(templateFile))
	tmpl.Registers()["file_system"] = fileSystem

	return tmpl, nil
}

