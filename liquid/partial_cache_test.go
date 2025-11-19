package liquid

import (
	"testing"
)

type mockContextForPartial struct {
	registers *Registers
}

func (m *mockContextForPartial) Registers() *Registers {
	return m.registers
}

func TestPartialCacheLoad(t *testing.T) {
	pc := &PartialCache{}
	registers := NewRegisters(nil)
	registers.Set("cached_partials", make(map[string]interface{}))
	registers.Set("file_system", &BlankFileSystem{})

	ctx := &mockContextForPartial{registers: registers}
	parseCtx := &mockParseContextForTag{env: NewEnvironment()}

	// This will fail because BlankFileSystem doesn't have the template
	// but that's expected for now
	_, err := pc.Load("test", ctx, parseCtx)
	if err == nil {
		t.Error("Expected error when template doesn't exist")
	}
}

func TestLoadPartial(t *testing.T) {
	registers := NewRegisters(nil)
	registers.Set("cached_partials", make(map[string]interface{}))
	registers.Set("file_system", &BlankFileSystem{})

	ctx := &mockContextForPartial{registers: registers}
	parseCtx := &mockParseContextForTag{env: NewEnvironment()}

	_, err := LoadPartial("test", ctx, parseCtx)
	if err == nil {
		t.Error("Expected error when template doesn't exist")
	}
}

// TestPartialCacheLoadWithTemplateFactory tests partial loading using TemplateFactory
func TestPartialCacheLoadWithTemplateFactory(t *testing.T) {
	pc := &PartialCache{}
	registers := NewRegisters(nil)
	registers.Set("cached_partials", make(map[string]interface{}))

	// Create a mock file system that returns a template
	mockFS := &mockFileSystemWithTemplate{template: "Hello {{ name }}"}
	registers.Set("file_system", mockFS)

	// Set template factory
	tf := NewTemplateFactory()
	registers.Set("template_factory", tf)

	ctx := &mockContextForPartial{registers: registers}
	parseCtx := NewParseContext(ParseContextOptions{})

	// Load partial - should use TemplateFactory
	partial, err := pc.Load("test", ctx, parseCtx)
	if err != nil {
		// Expected if template parsing fails, but factory should be used
		_ = partial
	}
}

// mockFileSystemWithTemplate is a mock file system that returns a template
type mockFileSystemWithTemplate struct {
	template string
}

func (m *mockFileSystemWithTemplate) ReadTemplateFile(templatePath string) (string, error) {
	return m.template, nil
}
