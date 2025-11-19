package liquid

import (
	"testing"
)

func TestResourceLimitsBasic(t *testing.T) {
	config := ResourceLimitsConfig{}
	rl := NewResourceLimits(config)
	if rl == nil {
		t.Fatal("Expected ResourceLimits, got nil")
	}
}

func TestResourceLimitsIncrementRenderScore(t *testing.T) {
	limit := 100
	config := ResourceLimitsConfig{
		RenderScoreLimit: &limit,
	}
	rl := NewResourceLimits(config)

	rl.IncrementRenderScore(50)
	if rl.RenderScore() != 50 {
		t.Errorf("Expected render score 50, got %d", rl.RenderScore())
	}

	rl.IncrementRenderScore(40)
	if rl.RenderScore() != 90 {
		t.Errorf("Expected render score 90, got %d", rl.RenderScore())
	}
}

func TestResourceLimitsIncrementAssignScore(t *testing.T) {
	limit := 100
	config := ResourceLimitsConfig{
		AssignScoreLimit: &limit,
	}
	rl := NewResourceLimits(config)

	rl.IncrementAssignScore(30)
	if rl.AssignScore() != 30 {
		t.Errorf("Expected assign score 30, got %d", rl.AssignScore())
	}
}

func TestResourceLimitsReset(t *testing.T) {
	config := ResourceLimitsConfig{}
	rl := NewResourceLimits(config)

	rl.IncrementRenderScore(50)
	rl.IncrementAssignScore(30)
	rl.Reset()

	if rl.RenderScore() != 0 {
		t.Errorf("Expected render score 0 after reset, got %d", rl.RenderScore())
	}
	if rl.AssignScore() != 0 {
		t.Errorf("Expected assign score 0 after reset, got %d", rl.AssignScore())
	}
	if rl.Reached() {
		t.Error("Expected not reached after reset")
	}
}

func TestResourceLimitsWithCapture(t *testing.T) {
	config := ResourceLimitsConfig{}
	rl := NewResourceLimits(config)

	called := false
	rl.WithCapture(func() {
		called = true
	})

	if !called {
		t.Error("Expected function to be called")
	}
}

func TestResourceLimitsReached(t *testing.T) {
	limit := 10
	config := ResourceLimitsConfig{
		RenderScoreLimit: &limit,
	}
	rl := NewResourceLimits(config)

	// Increment to limit
	rl.IncrementRenderScore(10)
	if rl.Reached() {
		t.Error("Expected not reached at limit")
	}

	// Increment past limit - should panic
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic when exceeding limit")
			}
		}()
		rl.IncrementRenderScore(1)
	}()
}
