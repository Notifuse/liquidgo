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

// TestResourceLimitsIncrementWriteScore tests IncrementWriteScore
func TestResourceLimitsIncrementWriteScore(t *testing.T) {
	limit := 10
	config := ResourceLimitsConfig{
		RenderLengthLimit: &limit,
	}
	rl := NewResourceLimits(config)

	// Test without capture (should check render length limit)
	rl.IncrementWriteScore("short")
	if rl.Reached() {
		t.Error("Expected not reached for short output")
	}

	// Test with long output (should trigger limit)
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic when exceeding render length limit")
			}
		}()
		rl.IncrementWriteScore("this is a very long string that exceeds the limit")
	}()
}

// TestResourceLimitsIncrementWriteScoreWithCapture tests IncrementWriteScore with capture
func TestResourceLimitsIncrementWriteScoreWithCapture(t *testing.T) {
	limit := 100
	config := ResourceLimitsConfig{
		AssignScoreLimit: &limit,
	}
	rl := NewResourceLimits(config)

	// Test with capture
	rl.WithCapture(func() {
		rl.IncrementWriteScore("first")
		if rl.AssignScore() == 0 {
			t.Error("Expected assign score to be incremented")
		}

		rl.IncrementWriteScore("first second")
		// Should increment by difference in length
		score := rl.AssignScore()
		if score <= 5 {
			t.Errorf("Expected assign score > 5, got %d", score)
		}
	})

	// After capture, should reset lastCaptureLength
	rl.IncrementWriteScore("test")
	// Should check render length limit, not assign score
	if rl.Reached() {
		t.Error("Expected not reached after capture")
	}
}

// TestResourceLimitsIncrementWriteScoreWithCaptureLimit tests IncrementWriteScore exceeding assign limit in capture
func TestResourceLimitsIncrementWriteScoreWithCaptureLimit(t *testing.T) {
	limit := 5
	config := ResourceLimitsConfig{
		AssignScoreLimit: &limit,
	}
	rl := NewResourceLimits(config)

	// Test with capture exceeding assign limit
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic when exceeding assign score limit in capture")
			}
		}()
		rl.WithCapture(func() {
			rl.IncrementWriteScore("this is a very long string")
		})
	}()
}

// TestResourceLimitsIncrementWriteScoreEmptyString tests IncrementWriteScore with empty string
func TestResourceLimitsIncrementWriteScoreEmptyString(t *testing.T) {
	config := ResourceLimitsConfig{}
	rl := NewResourceLimits(config)

	// Should not panic with empty string
	rl.IncrementWriteScore("")
	if rl.Reached() {
		t.Error("Expected not reached for empty string")
	}
}

// TestResourceLimitsIncrementWriteScoreByteLength tests IncrementWriteScore uses byte length
func TestResourceLimitsIncrementWriteScoreByteLength(t *testing.T) {
	limit := 5
	config := ResourceLimitsConfig{
		RenderLengthLimit: &limit,
	}
	rl := NewResourceLimits(config)

	// Test with string that has byte length > rune length
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic for string exceeding byte limit")
			}
		}()
		// Use a string with multi-byte characters
		rl.IncrementWriteScore("测试测试测试") // 6 Chinese characters = 18 bytes
	}()
}
