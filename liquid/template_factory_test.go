package liquid

import (
	"testing"
)

func TestTemplateFactoryBasic(t *testing.T) {
	tf := NewTemplateFactory()
	if tf == nil {
		t.Fatal("Expected TemplateFactory, got nil")
	}
}

func TestTemplateFactoryFor(t *testing.T) {
	tf := NewTemplateFactory()
	result := tf.For("test_template")
	// Should return a Template instance
	if result == nil {
		t.Error("Expected Template instance, got nil")
	}
	if _, ok := result.(*Template); !ok {
		t.Errorf("Expected *Template, got %T", result)
	}
}

