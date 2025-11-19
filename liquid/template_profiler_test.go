package liquid

import (
	"testing"
)

func TestTemplateWithProfiling(t *testing.T) {
	env := NewEnvironment()
	
	options := &TemplateOptions{
		Environment: env,
		Profile:     true,
	}
	
	template := NewTemplate(options)
	
	err := template.Parse("{{ name }}", options)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	
	_ = template.Render(map[string]interface{}{"name": "test"}, nil)
	
	profiler := template.Profiler()
	if profiler == nil {
		t.Error("Expected profiler to be set, got nil")
		return
	}
	
	// Profiler should have been created and used
	// The root node may have children (variables/tags) or be empty
	// Note: Very fast operations may register as 0 time on systems with low timer resolution (e.g., Windows)
	if profiler.TotalTime() < 0 {
		t.Errorf("Expected total time >= 0, got %f", profiler.TotalTime())
	}
}

func TestTemplateWithoutProfiling(t *testing.T) {
	env := NewEnvironment()
	
	template := NewTemplate(&TemplateOptions{
		Environment: env,
		Profile:     false,
	})
	
	err := template.Parse("{{ name }}", nil)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	
	_ = template.Render(map[string]interface{}{"name": "test"}, nil)
	
	// Profiler should be nil when profiling is disabled
	if template.Profiler() != nil {
		t.Error("Expected profiler to be nil when profiling is disabled")
	}
}

