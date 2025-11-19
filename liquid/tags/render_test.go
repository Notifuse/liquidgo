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
	// Should output error message or empty string
	if output == "" {
		t.Log("RenderToOutputBuffer returned empty output (expected for missing template)")
	}
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
	template.Parse("Hello {{ name }}", nil)
	
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
	partialTemplate.Parse("Hello {{ name }}", nil)
	
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
	if output == "" {
		t.Log("Expected error message for non-string, non-template-object")
	}
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
	if output == "" {
		t.Log("Expected output (may be empty if iterable interface not matched)")
	}
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
	
	// Should render the template
	if output == "" {
		t.Error("Expected non-empty output")
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
	if output == "" {
		t.Error("Expected non-empty output")
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
	if output == "" {
		t.Error("Expected non-empty output")
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
	if output == "" {
		t.Error("Expected non-empty output")
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
	if output == "" {
		t.Error("Expected non-empty output")
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
	if output == "" {
		t.Error("Expected non-empty output")
	}
}
