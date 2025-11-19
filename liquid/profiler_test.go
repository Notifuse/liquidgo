package liquid

import (
	"testing"
)

func TestProfilerBasic(t *testing.T) {
	profiler := NewProfiler()
	if profiler == nil {
		t.Fatal("Expected Profiler, got nil")
	}

	if profiler.TotalTime() != 0.0 {
		t.Errorf("Expected total time 0.0, got %f", profiler.TotalTime())
	}

	if profiler.Length() != 0 {
		t.Errorf("Expected length 0, got %d", profiler.Length())
	}
}

func TestProfilerProfile(t *testing.T) {
	profiler := NewProfiler()

	profiler.Profile("test", func() {
		// Do some work
	})

	// Profile() creates a root node, so rootChildren should have 1 element
	if len(profiler.rootChildren) != 1 {
		t.Errorf("Expected rootChildren length 1, got %d", len(profiler.rootChildren))
	}

	// But Children() returns the root node's children if there's only one root node
	// Since the root node has no children, Children() returns empty
	if profiler.Length() != 0 {
		t.Errorf("Expected length 0 (root node has no children), got %d", profiler.Length())
	}

	// Note: Very fast operations may register as 0 time on systems with low timer resolution (e.g., Windows)
	if profiler.TotalTime() < 0 {
		t.Errorf("Expected total time >= 0, got %f", profiler.TotalTime())
	}
}

func TestProfilerProfileNode(t *testing.T) {
	profiler := NewProfiler()

	profiler.Profile("test", func() {
		lineNum := 1
		profiler.ProfileNode("test", "assign x = 1", &lineNum, func() {
			// Do work
		})
	})

	if profiler.Length() != 1 {
		t.Errorf("Expected length 1, got %d", profiler.Length())
	}

	node := profiler.At(0)
	if node == nil {
		t.Fatal("Expected node, got nil")
	}

	if node.Code() != "assign x = 1" {
		t.Errorf("Expected code 'assign x = 1', got %q", node.Code())
	}

	if node.LineNumber() == nil || *node.LineNumber() != 1 {
		t.Errorf("Expected line number 1, got %v", node.LineNumber())
	}
}

func TestProfilerTimingSelfTime(t *testing.T) {
	profiler := NewProfiler()

	profiler.Profile("test", func() {
		profiler.ProfileNode("test", "outer", nil, func() {
			profiler.ProfileNode("test", "inner", nil, func() {
				// Inner work
			})
			// Outer work
		})
	})

	if profiler.Length() != 1 {
		t.Errorf("Expected length 1, got %d", profiler.Length())
	}

	node := profiler.At(0)
	if node == nil {
		t.Fatal("Expected node, got nil")
	}

	if len(node.Children()) != 1 {
		t.Errorf("Expected 1 child, got %d", len(node.Children()))
	}

	// Self time should be total time minus children time
	selfTime := node.SelfTime()
	if selfTime < 0 {
		t.Errorf("Expected self time >= 0, got %f", selfTime)
	}

	if selfTime > node.TotalTime() {
		t.Errorf("Expected self time <= total time, got self=%f total=%f", selfTime, node.TotalTime())
	}
}

func TestProfilerTotalRenderTime(t *testing.T) {
	profiler := NewProfiler()
	
	profiler.Profile("test", func() {
		// Do some work
	})
	
	// TotalRenderTime is alias for TotalTime
	totalRenderTime := profiler.TotalRenderTime()
	totalTime := profiler.TotalTime()
	if totalRenderTime != totalTime {
		t.Errorf("Expected TotalRenderTime to equal TotalTime, got %f != %f", totalRenderTime, totalTime)
	}
}

func TestTimingTemplateName(t *testing.T) {
	profiler := NewProfiler()
	
	profiler.Profile("template_name", func() {
		profiler.ProfileNode("template_name", "code", nil, func() {
			// Do work
		})
	})
	
	if profiler.Length() != 1 {
		t.Fatalf("Expected length 1, got %d", profiler.Length())
	}
	
	node := profiler.At(0)
	if node == nil {
		t.Fatal("Expected node, got nil")
	}
	
	templateName := node.TemplateName()
	if templateName != "template_name" {
		t.Errorf("Expected TemplateName 'template_name', got %q", templateName)
	}
}

func TestTimingPartial(t *testing.T) {
	profiler := NewProfiler()
	
	profiler.Profile("partial_name", func() {
		profiler.ProfileNode("partial_name", "code", nil, func() {
			// Do work
		})
	})
	
	if profiler.Length() != 1 {
		t.Fatalf("Expected length 1, got %d", profiler.Length())
	}
	
	node := profiler.At(0)
	if node == nil {
		t.Fatal("Expected node, got nil")
	}
	
	partial := node.Partial()
	if partial != "partial_name" {
		t.Errorf("Expected Partial 'partial_name', got %q", partial)
	}
	
	// Partial should equal TemplateName
	if partial != node.TemplateName() {
		t.Errorf("Expected Partial to equal TemplateName, got %q != %q", partial, node.TemplateName())
	}
}

func TestTimingRenderTime(t *testing.T) {
	profiler := NewProfiler()
	
	profiler.Profile("test", func() {
		profiler.ProfileNode("test", "code", nil, func() {
			// Do work
		})
	})
	
	if profiler.Length() != 1 {
		t.Fatalf("Expected length 1, got %d", profiler.Length())
	}
	
	node := profiler.At(0)
	if node == nil {
		t.Fatal("Expected node, got nil")
	}
	
	// RenderTime is alias for TotalTime
	renderTime := node.RenderTime()
	totalTime := node.TotalTime()
	if renderTime != totalTime {
		t.Errorf("Expected RenderTime to equal TotalTime, got %f != %f", renderTime, totalTime)
	}
}
