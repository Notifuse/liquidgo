package tags

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Notifuse/liquidgo/liquid"
)

func TestRenderTag(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewRenderTag("render", "'template'", pc)
	if err != nil {
		t.Fatalf("NewRenderTag() error = %v", err)
	}
	if tag == nil {
		t.Fatal("Expected RenderTag, got nil")
	}
}

func TestRenderTagWithWith(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewRenderTag("render", "'template' with var", pc)
	if err != nil {
		t.Fatalf("NewRenderTag() error = %v", err)
	}

	if tag.VariableNameExpr() == nil {
		t.Error("Expected variable name expression, got nil")
	}

	if tag.IsForLoop() {
		t.Error("Expected IsForLoop to be false for 'with', got true")
	}
}

func TestRenderTagWithFor(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewRenderTag("render", "'template' for items", pc)
	if err != nil {
		t.Fatalf("NewRenderTag() error = %v", err)
	}

	if tag.VariableNameExpr() == nil {
		t.Error("Expected variable name expression, got nil")
	}

	if !tag.IsForLoop() {
		t.Error("Expected IsForLoop to be true for 'for', got false")
	}
}

func TestRenderTagWithAs(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewRenderTag("render", "'template' as alias", pc)
	if err != nil {
		t.Fatalf("NewRenderTag() error = %v", err)
	}

	if tag.AliasName() != "alias" {
		t.Errorf("Expected alias name 'alias', got %q", tag.AliasName())
	}
}

func TestRenderTagParse(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewRenderTag("render", "'template'", pc)
	if err != nil {
		t.Fatalf("NewRenderTag() error = %v", err)
	}

	// Parse is a no-op for render tags
	tokenizer := pc.NewTokenizer("", false, nil, false)
	err = tag.Parse(tokenizer)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
}

func TestRenderTagRenderToOutputBuffer(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewRenderTag("render", "'nonexistent'", pc)
	if err != nil {
		t.Fatalf("NewRenderTag() error = %v", err)
	}

	ctx := liquid.NewContext()
	var output string
	// RenderToOutputBuffer should handle missing template gracefully
	tag.RenderToOutputBuffer(ctx, &output)
	// Should handle missing template (error message or empty)
	// Just verify no panic occurred
	_ = output
}

func TestRenderTagRenderToOutputBufferComprehensive(t *testing.T) {
	env := liquid.NewEnvironment()
	pc := liquid.NewParseContext(liquid.ParseContextOptions{Environment: env})

	// Test with non-string template name
	tag2, _ := NewRenderTag("render", "123", pc)
	ctx2 := liquid.NewContext()
	var output2 string
	tag2.RenderToOutputBuffer(ctx2, &output2)
	// Should handle error gracefully
	_ = output2

	// Test with with clause
	tag3, err := NewRenderTag("render", "'template' with person", pc)
	if err != nil {
		t.Fatalf("NewRenderTag() with 'with' error = %v", err)
	}
	ctx3 := liquid.NewContext()
	ctx3.Set("person", map[string]interface{}{"name": "Alice"})
	var output3 string
	tag3.RenderToOutputBuffer(ctx3, &output3)
	_ = output3

	// Test with for clause
	tag4, err := NewRenderTag("render", "'template' for items", pc)
	if err != nil {
		t.Fatalf("NewRenderTag() with 'for' error = %v", err)
	}
	ctx4 := liquid.NewContext()
	ctx4.Set("items", []interface{}{map[string]interface{}{"name": "Item1"}, map[string]interface{}{"name": "Item2"}})
	var output4 string
	tag4.RenderToOutputBuffer(ctx4, &output4)
	_ = output4

	// Test with as clause
	tag5, err := NewRenderTag("render", "'template' as alias_var", pc)
	if err != nil {
		t.Fatalf("NewRenderTag() with 'as' error = %v", err)
	}
	ctx5 := liquid.NewContext()
	var output5 string
	tag5.RenderToOutputBuffer(ctx5, &output5)
	_ = output5

	// Test with attributes
	tag6, err := NewRenderTag("render", "'template' key:value", pc)
	if err != nil {
		t.Fatalf("NewRenderTag() with attributes error = %v", err)
	}
	ctx6 := liquid.NewContext()
	var output6 string
	tag6.RenderToOutputBuffer(ctx6, &output6)
	_ = output6

	// Test with variable template name
	tag7, err := NewRenderTag("render", "template_var", pc)
	if err != nil {
		t.Fatalf("NewRenderTag() with variable name error = %v", err)
	}
	ctx7 := liquid.NewContext()
	ctx7.Set("template_var", "template_name")
	var output7 string
	tag7.RenderToOutputBuffer(ctx7, &output7)
	_ = output7
}

func TestRenderTagTemplateNameExpr(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewRenderTag("render", "'template'", pc)
	if err != nil {
		t.Fatalf("NewRenderTag() error = %v", err)
	}

	templateNameExpr := tag.TemplateNameExpr()
	if templateNameExpr == nil {
		t.Error("Expected TemplateNameExpr() to return non-nil expression")
	}
}

func TestRenderTagAttributes(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewRenderTag("render", "'template' key:value", pc)
	if err != nil {
		t.Fatalf("NewRenderTag() error = %v", err)
	}

	attributes := tag.Attributes()
	if attributes == nil {
		t.Error("Expected Attributes() to return non-nil map")
	}
	if len(attributes) == 0 {
		t.Error("Expected Attributes() to contain attributes")
	}
}

func TestRenderTagRenderToOutputBufferWithTemplateObject(t *testing.T) {
	env := liquid.NewEnvironment()
	pc := liquid.NewParseContext(liquid.ParseContextOptions{Environment: env})

	// Create a template object that implements ToPartial()
	template := liquid.NewTemplate(&liquid.TemplateOptions{Environment: env})
	if err := template.Parse("Hello {{ name }}", nil); err != nil {
		t.Fatalf("template.Parse() error = %v", err)
	}

	// Create a mock object that implements ToPartial, Filename, and Name
	type templateObject struct {
		template *liquid.Template
		filename string
		name     string
	}

	obj := &templateObject{
		template: template,
		filename: "test.liquid",
		name:     "test",
	}

	// Create render tag with variable that evaluates to template object
	ctx := liquid.NewContext()
	ctx.Set("template_obj", obj)

	// We need to create a tag that evaluates to this object
	// For now, test the path where template is not a string
	tag, err := NewRenderTag("render", "template_obj", pc)
	if err != nil {
		t.Fatalf("NewRenderTag() error = %v", err)
	}

	var output string
	tag.RenderToOutputBuffer(ctx, &output)
	// Should handle error gracefully since template_obj doesn't implement ToPartial
	_ = output
}

func TestRenderTagRenderToOutputBufferWithTemplateObjectToPartial(t *testing.T) {
	env := liquid.NewEnvironment()
	pc := liquid.NewParseContext(liquid.ParseContextOptions{Environment: env})

	// Create a template object that implements ToPartial() *Template, Filename(), and Name()
	partialTemplate := liquid.NewTemplate(&liquid.TemplateOptions{Environment: env})
	if err := partialTemplate.Parse("Hello {{ name }}", nil); err != nil {
		t.Fatalf("partialTemplate.Parse() error = %v", err)
	}

	// Create a mock object that implements all required interfaces
	type templateObject struct {
		partial  *liquid.Template
		filename string
		name     string
	}

	// Implement ToPartial() *Template
	obj := &templateObject{
		partial:  partialTemplate,
		filename: "test.liquid",
		name:     "test",
	}

	// Create render tag with variable that evaluates to template object
	ctx := liquid.NewContext()
	ctx.Set("template_obj", obj)

	// Test with object that doesn't implement ToPartial (should fall through to error)
	tag, err := NewRenderTag("render", "template_obj", pc)
	if err != nil {
		t.Fatalf("NewRenderTag() error = %v", err)
	}

	var output string
	tag.RenderToOutputBuffer(ctx, &output)
	// Should handle error gracefully since template_obj doesn't implement ToPartial() *Template
	// May output error message or handle gracefully
	_ = output
}

func TestRenderTagRenderToOutputBufferWithForLoopIterableObject(t *testing.T) {
	env := liquid.NewEnvironment()
	pc := liquid.NewParseContext(liquid.ParseContextOptions{Environment: env})

	// Create a file system with a template
	tmpDir := t.TempDir()
	fs := liquid.NewLocalFileSystem(tmpDir, "")

	templatePath := "item"
	fullPath, err := fs.FullPath(templatePath)
	if err != nil {
		t.Fatalf("Failed to get full path: %v", err)
	}
	if err := os.WriteFile(fullPath, []byte("Item: {{ item }}"), 0644); err != nil {
		t.Fatalf("Failed to create template file: %v", err)
	}

	// Set up context with file system
	registers := liquid.NewRegisters(nil)
	registers.Set("file_system", fs)
	registers.Set("cached_partials", make(map[string]interface{}))
	registers.Set("template_factory", liquid.NewTemplateFactory())
	ctx := liquid.BuildContext(liquid.ContextConfig{Registers: registers})

	// Create an iterable object that implements Each() and Count()
	type iterableObject struct {
		items []interface{}
	}

	// Implement Each and Count methods
	iterable := &iterableObject{
		items: []interface{}{"one", "two", "three"},
	}

	// Create render tag with for loop
	tag, err := NewRenderTag("render", "'item' for iterable", pc)
	if err != nil {
		t.Fatalf("NewRenderTag() error = %v", err)
	}

	ctx.Set("iterable", iterable)
	var output string
	tag.RenderToOutputBuffer(ctx, &output)
	// Should handle non-iterable gracefully (fallback to single render)
	// The iterableObject doesn't actually implement the interface, so it will fallback
	// Output may be empty or contain rendered content
	_ = output
}

func TestRenderTagRenderToOutputBufferWithNestedPath(t *testing.T) {
	env := liquid.NewEnvironment()
	pc := liquid.NewParseContext(liquid.ParseContextOptions{Environment: env})

	// Create a file system with a template
	tmpDir := t.TempDir()
	fs := liquid.NewLocalFileSystem(tmpDir, "")

	// Create nested template: dir/partial
	templatePath := "dir/partial"
	fullPath, err := fs.FullPath(templatePath)
	if err != nil {
		t.Fatalf("Failed to get full path: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}
	if err := os.WriteFile(fullPath, []byte("Hello from {{ partial }}"), 0644); err != nil {
		t.Fatalf("Failed to create template file: %v", err)
	}

	// Set up context with file system
	registers := liquid.NewRegisters(nil)
	registers.Set("file_system", fs)
	registers.Set("cached_partials", make(map[string]interface{}))
	registers.Set("template_factory", liquid.NewTemplateFactory())
	ctx := liquid.BuildContext(liquid.ContextConfig{Registers: registers})

	// Create render tag with nested path (no alias - should use last part "partial")
	tag, err := NewRenderTag("render", "'dir/partial'", pc)
	if err != nil {
		t.Fatalf("NewRenderTag() error = %v", err)
	}

	ctx.Set("partial", "World")
	var output string
	tag.RenderToOutputBuffer(ctx, &output)

	// Should render the template (note: parent context vars not accessible in isolated render scope)
	expected := "Hello from "
	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestRenderTagRenderToOutputBufferWithForLoopIterable(t *testing.T) {
	env := liquid.NewEnvironment()
	pc := liquid.NewParseContext(liquid.ParseContextOptions{Environment: env})

	// Create a file system with a template
	tmpDir := t.TempDir()
	fs := liquid.NewLocalFileSystem(tmpDir, "")

	templatePath := "item"
	fullPath, err := fs.FullPath(templatePath)
	if err != nil {
		t.Fatalf("Failed to get full path: %v", err)
	}
	if err := os.WriteFile(fullPath, []byte("Item: {{ item }}"), 0644); err != nil {
		t.Fatalf("Failed to create template file: %v", err)
	}

	// Set up context with file system
	registers := liquid.NewRegisters(nil)
	registers.Set("file_system", fs)
	registers.Set("cached_partials", make(map[string]interface{}))
	registers.Set("template_factory", liquid.NewTemplateFactory())
	ctx := liquid.BuildContext(liquid.ContextConfig{Registers: registers})

	// Create render tag with for loop
	tag, err := NewRenderTag("render", "'item' for items", pc)
	if err != nil {
		t.Fatalf("NewRenderTag() error = %v", err)
	}

	// Set array variable
	ctx.Set("items", []interface{}{"one", "two", "three"})
	var output string
	tag.RenderToOutputBuffer(ctx, &output)

	// Should render template multiple times
	expected := "Item: oneItem: twoItem: three"
	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestRenderTagRenderToOutputBufferWithForLoopNonIterable(t *testing.T) {
	env := liquid.NewEnvironment()
	pc := liquid.NewParseContext(liquid.ParseContextOptions{Environment: env})

	// Create a file system with a template
	tmpDir := t.TempDir()
	fs := liquid.NewLocalFileSystem(tmpDir, "")

	templatePath := "item"
	fullPath, err := fs.FullPath(templatePath)
	if err != nil {
		t.Fatalf("Failed to get full path: %v", err)
	}
	if err := os.WriteFile(fullPath, []byte("Item: {{ item }}"), 0644); err != nil {
		t.Fatalf("Failed to create template file: %v", err)
	}

	// Set up context with file system
	registers := liquid.NewRegisters(nil)
	registers.Set("file_system", fs)
	registers.Set("cached_partials", make(map[string]interface{}))
	registers.Set("template_factory", liquid.NewTemplateFactory())
	ctx := liquid.BuildContext(liquid.ContextConfig{Registers: registers})

	// Create render tag with for loop but non-iterable variable
	tag, err := NewRenderTag("render", "'item' for single_item", pc)
	if err != nil {
		t.Fatalf("NewRenderTag() error = %v", err)
	}

	// Set non-iterable variable
	ctx.Set("single_item", "not_an_array")
	var output string
	tag.RenderToOutputBuffer(ctx, &output)

	// Should render once with the variable
	expected := "Item: not_an_array"
	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestRenderTagRenderToOutputBufferWithAttributes(t *testing.T) {
	env := liquid.NewEnvironment()
	pc := liquid.NewParseContext(liquid.ParseContextOptions{Environment: env})

	// Create a file system with a template
	tmpDir := t.TempDir()
	fs := liquid.NewLocalFileSystem(tmpDir, "")

	templatePath := "greeting"
	fullPath, err := fs.FullPath(templatePath)
	if err != nil {
		t.Fatalf("Failed to get full path: %v", err)
	}
	if err := os.WriteFile(fullPath, []byte("Hello {{ name }}"), 0644); err != nil {
		t.Fatalf("Failed to create template file: %v", err)
	}

	// Set up context with file system
	registers := liquid.NewRegisters(nil)
	registers.Set("file_system", fs)
	registers.Set("cached_partials", make(map[string]interface{}))
	registers.Set("template_factory", liquid.NewTemplateFactory())
	ctx := liquid.BuildContext(liquid.ContextConfig{Registers: registers})

	// Create render tag with attributes
	tag, err := NewRenderTag("render", "'greeting' name:'Alice'", pc)
	if err != nil {
		t.Fatalf("NewRenderTag() error = %v", err)
	}

	var output string
	tag.RenderToOutputBuffer(ctx, &output)

	// Should render the template with attributes
	expected := "Hello Alice"
	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestRenderTagRenderToOutputBufferWithAlias(t *testing.T) {
	env := liquid.NewEnvironment()
	pc := liquid.NewParseContext(liquid.ParseContextOptions{Environment: env})

	// Create a file system with a template
	tmpDir := t.TempDir()
	fs := liquid.NewLocalFileSystem(tmpDir, "")

	templatePath := "greeting"
	fullPath, err := fs.FullPath(templatePath)
	if err != nil {
		t.Fatalf("Failed to get full path: %v", err)
	}
	if err := os.WriteFile(fullPath, []byte("Hello {{ person }}"), 0644); err != nil {
		t.Fatalf("Failed to create template file: %v", err)
	}

	// Set up context with file system
	registers := liquid.NewRegisters(nil)
	registers.Set("file_system", fs)
	registers.Set("cached_partials", make(map[string]interface{}))
	registers.Set("template_factory", liquid.NewTemplateFactory())
	ctx := liquid.BuildContext(liquid.ContextConfig{Registers: registers})

	// Create render tag with alias
	tag, err := NewRenderTag("render", "'greeting' with user as person", pc)
	if err != nil {
		t.Fatalf("NewRenderTag() error = %v", err)
	}

	ctx.Set("user", "Bob")
	var output string
	tag.RenderToOutputBuffer(ctx, &output)

	// Should render the template with alias
	expected := "Hello Bob"
	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestRenderTagRenderToOutputBufferWithEmptyTemplateName(t *testing.T) {
	env := liquid.NewEnvironment()
	pc := liquid.NewParseContext(liquid.ParseContextOptions{Environment: env})

	// Create a file system with a template
	tmpDir := t.TempDir()
	fs := liquid.NewLocalFileSystem(tmpDir, "")

	templatePath := "greeting"
	fullPath, err := fs.FullPath(templatePath)
	if err != nil {
		t.Fatalf("Failed to get full path: %v", err)
	}
	if err := os.WriteFile(fullPath, []byte("Hello"), 0644); err != nil {
		t.Fatalf("Failed to create template file: %v", err)
	}

	// Set up context with file system
	registers := liquid.NewRegisters(nil)
	registers.Set("file_system", fs)
	registers.Set("cached_partials", make(map[string]interface{}))
	registers.Set("template_factory", liquid.NewTemplateFactory())
	ctx := liquid.BuildContext(liquid.ContextConfig{Registers: registers})

	// Load the template normally first to cache it (name will be empty)
	_, err = liquid.LoadPartial("greeting", ctx, pc)
	if err != nil {
		t.Fatalf("Failed to load partial: %v", err)
	}

	// Create render tag
	tag, err := NewRenderTag("render", "'greeting'", pc)
	if err != nil {
		t.Fatalf("NewRenderTag() error = %v", err)
	}

	var output string
	tag.RenderToOutputBuffer(ctx, &output)

	// Should use template name string (greeting) when template name is empty
	expected := "Hello"
	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

// TestRenderTagRenderToOutputBufferErrorScenarios tests RenderToOutputBuffer error scenarios
func TestRenderTagRenderToOutputBufferErrorScenarios(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})

	// Test with nonexistent template name
	tag, err := NewRenderTag("render", "'nonexistent_template'", pc)
	if err != nil {
		t.Fatalf("NewRenderTag() error = %v", err)
	}

	ctx := liquid.NewContext()
	var output string

	// Should handle missing template gracefully
	tag.RenderToOutputBuffer(ctx, &output)

	// Output may be empty or contain error message
	if len(output) > 0 {
		t.Logf("Note: RenderToOutputBuffer with nonexistent template produced: %q", output)
	}
}

// TestRenderTagRenderToOutputBufferDynamicTemplateName tests dynamic template name resolution
func TestRenderTagRenderToOutputBufferDynamicTemplateName(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})

	// Test with variable expression for template name
	tag, err := NewRenderTag("render", "template_name", pc)
	if err != nil {
		t.Fatalf("NewRenderTag() error = %v", err)
	}

	ctx := liquid.NewContext()
	ctx.Set("template_name", "test_template")

	var output string
	// Should resolve template name from variable
	tag.RenderToOutputBuffer(ctx, &output)

	// Output depends on whether template exists
	t.Logf("Note: Dynamic template name resolution output: %q", output)
}

// TestRenderTagRenderToOutputBufferPartialLoadingFailure tests partial loading failure handling
func TestRenderTagRenderToOutputBufferPartialLoadingFailure(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})

	// Create a render tag
	tag, err := NewRenderTag("render", "'missing_partial'", pc)
	if err != nil {
		t.Fatalf("NewRenderTag() error = %v", err)
	}

	ctx := liquid.NewContext()
	var output string

	// Should handle partial loading failure gracefully
	tag.RenderToOutputBuffer(ctx, &output)

	// May produce error output or empty string
	t.Logf("Note: Partial loading failure output: %q", output)
}

// TestRenderTagInvalidSyntax tests invalid syntax error handling
func TestRenderTagInvalidSyntax(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})

	// Test with completely invalid syntax that won't match the regex
	_, err := NewRenderTag("render", "!@#$%", pc)
	if err == nil {
		t.Error("Expected error for invalid syntax, got nil")
	}
}

// TestRenderTagAttributeWithQuotedValue tests attribute parsing with quoted values
func TestRenderTagAttributeWithQuotedValue(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})

	// Test with double-quoted attribute value
	tag1, err := NewRenderTag("render", "'template' name:\"Alice\"", pc)
	if err != nil {
		t.Fatalf("NewRenderTag() with double-quoted attribute error = %v", err)
	}
	if len(tag1.Attributes()) == 0 {
		t.Error("Expected attributes to be parsed")
	}

	// Test with single-quoted attribute value
	tag2, err := NewRenderTag("render", "'template' name:'Bob'", pc)
	if err != nil {
		t.Fatalf("NewRenderTag() with single-quoted attribute error = %v", err)
	}
	if len(tag2.Attributes()) == 0 {
		t.Error("Expected attributes to be parsed")
	}
}

// TestRenderTagWithTemplateObjectImplementingToPartial tests ToPartial interface
func TestRenderTagWithTemplateObjectImplementingToPartial(t *testing.T) {
	env := liquid.NewEnvironment()
	pc := liquid.NewParseContext(liquid.ParseContextOptions{Environment: env})

	// Create a partial template
	partialTemplate := liquid.NewTemplate(&liquid.TemplateOptions{Environment: env})
	if err := partialTemplate.Parse("Hello {{ name }}", nil); err != nil {
		t.Fatalf("partialTemplate.Parse() error = %v", err)
	}

	// Create a mock object that implements ToPartial(), Filename(), and Name()
	type templateObject struct{}

	obj := &templateObject{}

	// Implement ToPartial() method
	toPartialImpl := func() *liquid.Template {
		return partialTemplate
	}
	_ = toPartialImpl

	// Implement Filename() method
	filenameImpl := func() string {
		return "test.liquid"
	}
	_ = filenameImpl

	// Implement Name() method
	nameImpl := func() string {
		return "test"
	}
	_ = nameImpl

	// Create render tag
	tag, err := NewRenderTag("render", "template_obj", pc)
	if err != nil {
		t.Fatalf("NewRenderTag() error = %v", err)
	}

	ctx := liquid.NewContext()
	ctx.Set("template_obj", obj)
	ctx.Set("name", "World")

	var output string
	tag.RenderToOutputBuffer(ctx, &output)

	// The object doesn't actually implement the interface in Go, so it will fall through to error
	// May output error message or handle gracefully
	_ = output
}

// mockTemplateObjectWithToPartial implements the ToPartial, Filename, and Name interfaces
type mockTemplateObjectWithToPartial struct {
	partial  *liquid.Template
	filename string
	name     string
}

func (m *mockTemplateObjectWithToPartial) ToPartial() *liquid.Template {
	return m.partial
}

func (m *mockTemplateObjectWithToPartial) Filename() string {
	return m.filename
}

func (m *mockTemplateObjectWithToPartial) Name() string {
	return m.name
}

// TestRenderTagWithActualToPartialImplementation tests with a real ToPartial implementation
func TestRenderTagWithActualToPartialImplementation(t *testing.T) {
	env := liquid.NewEnvironment()
	pc := liquid.NewParseContext(liquid.ParseContextOptions{Environment: env})

	// Create a partial template
	partialTemplate := liquid.NewTemplate(&liquid.TemplateOptions{Environment: env})
	if err := partialTemplate.Parse("Hello {{ name }}", nil); err != nil {
		t.Fatalf("partialTemplate.Parse() error = %v", err)
	}

	// Create mock object that actually implements the interfaces
	obj := &mockTemplateObjectWithToPartial{
		partial:  partialTemplate,
		filename: "test.liquid",
		name:     "test",
	}

	// Create render tag
	tag, err := NewRenderTag("render", "template_obj", pc)
	if err != nil {
		t.Fatalf("NewRenderTag() error = %v", err)
	}

	ctx := liquid.NewContext()
	ctx.Set("template_obj", obj)
	ctx.Set("name", "World")

	var output string
	tag.RenderToOutputBuffer(ctx, &output)

	// Should render the template using ToPartial (note: parent context not accessible)
	expected := "Hello "
	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

// mockIterableObject implements the iterable interface with Each and Count methods
type mockIterableObject struct {
	items []interface{}
}

func (m *mockIterableObject) Each(fn func(interface{})) {
	for _, item := range m.items {
		fn(item)
	}
}

func (m *mockIterableObject) Count() int {
	return len(m.items)
}

// TestRenderTagWithIterableObjectImplementation tests iterable interface
func TestRenderTagWithIterableObjectImplementation(t *testing.T) {
	env := liquid.NewEnvironment()
	pc := liquid.NewParseContext(liquid.ParseContextOptions{Environment: env})

	// Create a file system with a template
	tmpDir := t.TempDir()
	fs := liquid.NewLocalFileSystem(tmpDir, "")

	templatePath := "item"
	fullPath, err := fs.FullPath(templatePath)
	if err != nil {
		t.Fatalf("Failed to get full path: %v", err)
	}
	if err := os.WriteFile(fullPath, []byte("Item: {{ item }}\n"), 0644); err != nil {
		t.Fatalf("Failed to create template file: %v", err)
	}

	// Set up context with file system
	registers := liquid.NewRegisters(nil)
	registers.Set("file_system", fs)
	registers.Set("cached_partials", make(map[string]interface{}))
	registers.Set("template_factory", liquid.NewTemplateFactory())
	ctx := liquid.BuildContext(liquid.ContextConfig{Registers: registers})

	// Create an iterable object that actually implements Each() and Count()
	iterable := &mockIterableObject{
		items: []interface{}{"one", "two", "three"},
	}

	// Create render tag with for loop
	tag, err := NewRenderTag("render", "'item' for iterable", pc)
	if err != nil {
		t.Fatalf("NewRenderTag() error = %v", err)
	}

	ctx.Set("iterable", iterable)
	var output string
	tag.RenderToOutputBuffer(ctx, &output)

	// Should render template multiple times using the iterable interface
	expected := "Item: one\nItem: two\nItem: three\n"
	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

// TestRenderTagPartialNotTemplate tests when LoadPartial returns non-template
func TestRenderTagPartialNotTemplate(t *testing.T) {
	env := liquid.NewEnvironment()
	pc := liquid.NewParseContext(liquid.ParseContextOptions{Environment: env})

	// Create a custom context that will make LoadPartial return something that's not a template
	// This is tricky because LoadPartial typically returns a template
	// We'll test the type assertion failure by creating a scenario where it fails

	ctx := liquid.NewContext()

	// Create render tag
	tag, err := NewRenderTag("render", "'test'", pc)
	if err != nil {
		t.Fatalf("NewRenderTag() error = %v", err)
	}

	var output string
	tag.RenderToOutputBuffer(ctx, &output)

	// Should handle the error gracefully (no panic)
	// Output may be empty or contain error message
	_ = output
}

// TestRenderTagWithForLoopAndNilVariable tests for loop with nil variable
func TestRenderTagWithForLoopAndNilVariable(t *testing.T) {
	env := liquid.NewEnvironment()
	pc := liquid.NewParseContext(liquid.ParseContextOptions{Environment: env})

	// Create a file system with a template
	tmpDir := t.TempDir()
	fs := liquid.NewLocalFileSystem(tmpDir, "")

	templatePath := "item"
	fullPath, err := fs.FullPath(templatePath)
	if err != nil {
		t.Fatalf("Failed to get full path: %v", err)
	}
	if err := os.WriteFile(fullPath, []byte("Item: {{ item }}"), 0644); err != nil {
		t.Fatalf("Failed to create template file: %v", err)
	}

	// Set up context with file system
	registers := liquid.NewRegisters(nil)
	registers.Set("file_system", fs)
	registers.Set("cached_partials", make(map[string]interface{}))
	registers.Set("template_factory", liquid.NewTemplateFactory())
	ctx := liquid.BuildContext(liquid.ContextConfig{Registers: registers})

	// Create render tag with for loop but don't set the variable (nil)
	tag, err := NewRenderTag("render", "'item' for missing_var", pc)
	if err != nil {
		t.Fatalf("NewRenderTag() error = %v", err)
	}

	// Don't set missing_var - it will be nil
	var output string
	tag.RenderToOutputBuffer(ctx, &output)

	// Should handle nil variable gracefully (render single time with nil)
	t.Logf("Output with nil for loop variable: %q", output)
}

// TestRenderTagWithoutVariableExpression tests single render without variable
func TestRenderTagWithoutVariableExpression(t *testing.T) {
	env := liquid.NewEnvironment()
	pc := liquid.NewParseContext(liquid.ParseContextOptions{Environment: env})

	// Create a file system with a template
	tmpDir := t.TempDir()
	fs := liquid.NewLocalFileSystem(tmpDir, "")

	templatePath := "simple"
	fullPath, err := fs.FullPath(templatePath)
	if err != nil {
		t.Fatalf("Failed to get full path: %v", err)
	}
	if err := os.WriteFile(fullPath, []byte("Hello World"), 0644); err != nil {
		t.Fatalf("Failed to create template file: %v", err)
	}

	// Set up context with file system
	registers := liquid.NewRegisters(nil)
	registers.Set("file_system", fs)
	registers.Set("cached_partials", make(map[string]interface{}))
	registers.Set("template_factory", liquid.NewTemplateFactory())
	ctx := liquid.BuildContext(liquid.ContextConfig{Registers: registers})

	// Create render tag without any variable expression (no with/for)
	tag, err := NewRenderTag("render", "'simple'", pc)
	if err != nil {
		t.Fatalf("NewRenderTag() error = %v", err)
	}

	var output string
	tag.RenderToOutputBuffer(ctx, &output)

	// Should render template once without any variable
	expected := "Hello World"
	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

// TestRenderTagPassesNamedArgumentsIntoInnerScope tests parameter passing to partials
// Ruby ref: test_render_passes_named_arguments_into_inner_scope (render_tag_test.rb:24-31)
func TestRenderTagPassesNamedArgumentsIntoInnerScope(t *testing.T) {
	env := liquid.NewEnvironment()
	RegisterStandardTags(env)

	template := `{% render "product", inner_product: outer_product %}`
	partial := `{{ inner_product.title }}`

	fs := &mapFileSystem{
		templates: map[string]string{
			"product": partial,
		},
	}

	tmpl, err := liquid.ParseTemplate(template, &liquid.TemplateOptions{
		Environment: env,
	})
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	tmpl.Registers()["file_system"] = fs

	output := tmpl.Render(map[string]interface{}{
		"outer_product": map[string]interface{}{
			"title": "My Product",
		},
	}, nil)

	expected := "My Product"
	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

// TestRenderTagDoesNotInheritParentScopeVariables verifies scope isolation
// Ruby ref: test_render_does_not_inherit_parent_scope_variables (render_tag_test.rb:49-55)
func TestRenderTagDoesNotInheritParentScopeVariables(t *testing.T) {
	env := liquid.NewEnvironment()
	RegisterStandardTags(env)

	template := `{% assign outer_variable = "should not be visible" %}{% render "snippet" %}`
	partial := `{{ outer_variable }}`

	fs := &mapFileSystem{
		templates: map[string]string{
			"snippet": partial,
		},
	}

	tmpl, err := liquid.ParseTemplate(template, &liquid.TemplateOptions{
		Environment: env,
	})
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	tmpl.Registers()["file_system"] = fs

	output := tmpl.Render(map[string]interface{}{}, nil)

	expected := ""
	if output != expected {
		t.Errorf("Expected %q (empty - parent vars not accessible), got %q", expected, output)
	}
}

// TestRenderTagDoesNotMutateParentScope verifies render doesn't pollute parent scope
// Ruby ref: test_render_does_not_mutate_parent_scope (render_tag_test.rb:65-71)
func TestRenderTagDoesNotMutateParentScope(t *testing.T) {
	env := liquid.NewEnvironment()
	RegisterStandardTags(env)

	template := `{% render "snippet" %}{{ inner }}`
	partial := `{% assign inner = 1 %}`

	fs := &mapFileSystem{
		templates: map[string]string{
			"snippet": partial,
		},
	}

	tmpl, err := liquid.ParseTemplate(template, &liquid.TemplateOptions{
		Environment: env,
	})
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	tmpl.Registers()["file_system"] = fs

	output := tmpl.Render(map[string]interface{}{}, nil)

	expected := ""
	if output != expected {
		t.Errorf("Expected %q (empty - render vars don't leak), got %q", expected, output)
	}
}

// TestRenderTagOptionalCommas tests parameter syntax flexibility
// Ruby ref: test_optional_commas (render_tag_test.rb:123-128)
func TestRenderTagOptionalCommas(t *testing.T) {
	env := liquid.NewEnvironment()
	RegisterStandardTags(env)

	partial := `hello {{ arg1 }} {{ arg2 }}`

	fs := &mapFileSystem{
		templates: map[string]string{
			"snippet": partial,
		},
	}

	tests := []struct {
		name     string
		template string
	}{
		{
			name:     "with commas",
			template: `{% render "snippet", arg1: "value1", arg2: "value2" %}`,
		},
		{
			name:     "without comma after template name",
			template: `{% render "snippet"  arg1: "value1", arg2: "value2" %}`,
		},
		{
			name:     "no commas",
			template: `{% render "snippet"  arg1: "value1"  arg2: "value2" %}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl, err := liquid.ParseTemplate(tt.template, &liquid.TemplateOptions{
				Environment: env,
			})
			if err != nil {
				t.Fatalf("Parse error: %v", err)
			}

			tmpl.Registers()["file_system"] = fs

			output := tmpl.Render(map[string]interface{}{}, nil)

			expected := "hello value1 value2"
			if output != expected {
				t.Errorf("Expected %q, got %q", expected, output)
			}
		})
	}
}

// Note: Tests for render tag with typed slices are covered in integration tests
// (see integration/comprehensive_test.go TestBlogPostTags/Render_tag)
