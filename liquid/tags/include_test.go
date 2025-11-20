package tags

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Notifuse/liquidgo/liquid"
)

func TestIncludeTag(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewIncludeTag("include", "'template'", pc)
	if err != nil {
		t.Fatalf("NewIncludeTag() error = %v", err)
	}
	if tag == nil {
		t.Fatal("Expected IncludeTag, got nil")
	}
}

func TestIncludeTagWithAttributes(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewIncludeTag("include", "'template' key:value", pc)
	if err != nil {
		t.Fatalf("NewIncludeTag() error = %v", err)
	}

	if len(tag.Attributes()) != 1 {
		t.Errorf("Expected 1 attribute, got %d", len(tag.Attributes()))
	}
}

func TestIncludeTagWithWith(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewIncludeTag("include", "'template' with var", pc)
	if err != nil {
		t.Fatalf("NewIncludeTag() error = %v", err)
	}

	if tag.VariableNameExpr() == nil {
		t.Error("Expected variable name expression, got nil")
	}
}

func TestIncludeTagWithAs(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewIncludeTag("include", "'template' as alias", pc)
	if err != nil {
		t.Fatalf("NewIncludeTag() error = %v", err)
	}

	if tag.AliasName() != "alias" {
		t.Errorf("Expected alias name 'alias', got %q", tag.AliasName())
	}
}

func TestIncludeTagParse(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewIncludeTag("include", "'template'", pc)
	if err != nil {
		t.Fatalf("NewIncludeTag() error = %v", err)
	}

	// Parse is a no-op for include tags
	tokenizer := pc.NewTokenizer("", false, nil, false)
	err = tag.Parse(tokenizer)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
}

func TestIncludeTagRenderToOutputBuffer(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewIncludeTag("include", "'nonexistent'", pc)
	if err != nil {
		t.Fatalf("NewIncludeTag() error = %v", err)
	}

	ctx := liquid.NewContext()
	var output string
	// RenderToOutputBuffer should handle missing template gracefully
	tag.RenderToOutputBuffer(ctx, &output)
	// Should handle missing template (error message or empty)
	// Just verify no panic occurred
	_ = output
}

func TestIncludeTagRenderToOutputBufferComprehensive(t *testing.T) {
	env := liquid.NewEnvironment()
	pc := liquid.NewParseContext(liquid.ParseContextOptions{Environment: env})

	// Test with non-string template name
	tag2, _ := NewIncludeTag("include", "123", pc)
	ctx2 := liquid.NewContext()
	var output2 string
	tag2.RenderToOutputBuffer(ctx2, &output2)
	// Should handle error gracefully
	_ = output2

	// Test with with clause
	tag3, err := NewIncludeTag("include", "'greeting' with person", pc)
	if err != nil {
		t.Fatalf("NewIncludeTag() with 'with' error = %v", err)
	}
	ctx3 := liquid.NewContext()
	ctx3.Set("person", map[string]interface{}{"name": "Alice"})
	var output3 string
	tag3.RenderToOutputBuffer(ctx3, &output3)
	_ = output3

	// Test with for clause
	tag4, err := NewIncludeTag("include", "'greeting' for person", pc)
	if err != nil {
		t.Fatalf("NewIncludeTag() with 'for' error = %v", err)
	}
	ctx4 := liquid.NewContext()
	ctx4.Set("person", map[string]interface{}{"name": "Bob"})
	var output4 string
	tag4.RenderToOutputBuffer(ctx4, &output4)
	_ = output4

	// Test with as clause
	tag5, err := NewIncludeTag("include", "'greeting' as greeting_var", pc)
	if err != nil {
		t.Fatalf("NewIncludeTag() with 'as' error = %v", err)
	}
	ctx5 := liquid.NewContext()
	ctx5.Set("name", "Charlie")
	var output5 string
	tag5.RenderToOutputBuffer(ctx5, &output5)
	_ = output5

	// Test with array variable
	tag6, err := NewIncludeTag("include", "'greeting' for items", pc)
	if err != nil {
		t.Fatalf("NewIncludeTag() with array error = %v", err)
	}
	ctx6 := liquid.NewContext()
	ctx6.Set("items", []interface{}{map[string]interface{}{"name": "Item1"}, map[string]interface{}{"name": "Item2"}})
	var output6 string
	tag6.RenderToOutputBuffer(ctx6, &output6)
	_ = output6
}

func TestIncludeTagTemplateNameExpr(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewIncludeTag("include", "'template'", pc)
	if err != nil {
		t.Fatalf("NewIncludeTag() error = %v", err)
	}

	templateNameExpr := tag.TemplateNameExpr()
	if templateNameExpr == nil {
		t.Error("Expected TemplateNameExpr() to return non-nil expression")
	}
}

func TestIncludeTagRenderToOutputBufferWithNestedPath(t *testing.T) {
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

	// Create include tag with nested path (no alias - should use last part "partial")
	tag, err := NewIncludeTag("include", "'dir/partial'", pc)
	if err != nil {
		t.Fatalf("NewIncludeTag() error = %v", err)
	}

	ctx.Set("partial", "World")
	var output string
	tag.RenderToOutputBuffer(ctx, &output)

	// Should render the template (note: variable 'partial' set in parent, but template also defines 'partial' from path)
	expected := "Hello from "
	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestIncludeTagRenderToOutputBufferWithVariableFromContext(t *testing.T) {
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
	if err := os.WriteFile(fullPath, []byte("Hello {{ greeting }}"), 0644); err != nil {
		t.Fatalf("Failed to create template file: %v", err)
	}

	// Set up context with file system
	registers := liquid.NewRegisters(nil)
	registers.Set("file_system", fs)
	registers.Set("cached_partials", make(map[string]interface{}))
	registers.Set("template_factory", liquid.NewTemplateFactory())
	ctx := liquid.BuildContext(liquid.ContextConfig{Registers: registers})

	// Create include tag without variableNameExpr - should find variable by template name
	tag, err := NewIncludeTag("include", "'greeting'", pc)
	if err != nil {
		t.Fatalf("NewIncludeTag() error = %v", err)
	}

	// Set variable with same name as template
	ctx.Set("greeting", "World")
	var output string
	tag.RenderToOutputBuffer(ctx, &output)

	// Should render the template
	expected := "Hello World"
	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestIncludeTagRenderToOutputBufferWithArrayVariable(t *testing.T) {
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

	// Create include tag with array variable (for clause)
	tag, err := NewIncludeTag("include", "'item' for items", pc)
	if err != nil {
		t.Fatalf("NewIncludeTag() error = %v", err)
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

func TestIncludeTagRenderToOutputBufferWithAttributes(t *testing.T) {
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

	// Create include tag with attributes
	tag, err := NewIncludeTag("include", "'greeting' name:'Alice'", pc)
	if err != nil {
		t.Fatalf("NewIncludeTag() error = %v", err)
	}

	var output string
	tag.RenderToOutputBuffer(ctx, &output)

	// Should render the template with attributes
	expected := "Hello Alice"
	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestIncludeTagRenderToOutputBufferWithAlias(t *testing.T) {
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

	// Create include tag with alias
	tag, err := NewIncludeTag("include", "'greeting' with user as person", pc)
	if err != nil {
		t.Fatalf("NewIncludeTag() error = %v", err)
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

func TestIncludeTagRenderToOutputBufferWithLocaleError(t *testing.T) {
	env := liquid.NewEnvironment()
	locale := liquid.NewI18n("en")
	pc := liquid.NewParseContext(liquid.ParseContextOptions{Environment: env, Locale: locale})

	// Create include tag with non-string template name
	tag, err := NewIncludeTag("include", "123", pc)
	if err != nil {
		t.Fatalf("NewIncludeTag() error = %v", err)
	}

	ctx := liquid.NewContext()
	var output string
	tag.RenderToOutputBuffer(ctx, &output)

	// Should output error message (Liquid error format)
	if !strings.Contains(output, "Liquid") && output == "" {
		t.Errorf("Expected Liquid error message, got %q", output)
	}
}

func TestIncludeTagRenderToOutputBufferWithPartialName(t *testing.T) {
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

	// Load the template normally first to cache it
	_, err = liquid.LoadPartial("greeting", ctx, pc)
	if err != nil {
		t.Fatalf("Failed to load partial: %v", err)
	}

	// Get the cached template and set its name
	cache := registers.Get("cached_partials").(map[string]interface{})
	if cached, ok := cache["greeting:lax"]; ok {
		if template, ok := cached.(*liquid.Template); ok {
			template.SetName("custom_name")
		}
	}

	// Create include tag
	tag, err := NewIncludeTag("include", "'greeting'", pc)
	if err != nil {
		t.Fatalf("NewIncludeTag() error = %v", err)
	}

	var output string
	tag.RenderToOutputBuffer(ctx, &output)

	// Should use template name
	expected := "Hello"
	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

// Note: Tests for include tag with typed slices are covered in integration tests
// (see integration/comprehensive_test.go TestBlogPostTags/Include_tag)
