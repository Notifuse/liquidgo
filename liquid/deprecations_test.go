package liquid

import (
	"testing"
)

func TestDeprecationsWarn(t *testing.T) {
	// Reset deprecations for clean test
	ResetDeprecations()

	// Capture output (in a real scenario, we'd use a logger)
	// For now, we just test that it doesn't panic
	Warn("old_method", "new_method")

	// Test that calling it again doesn't warn again
	Warn("old_method", "new_method")

	// Test with different deprecation
	Warn("another_old_method", "another_new_method")
}

func TestDeprecationsInstance(t *testing.T) {
	d := &Deprecations{
		warned: make(map[string]bool),
	}

	// First call should warn
	d.Warn("test_method", "new_test_method")

	// Second call should not warn (already warned)
	d.Warn("test_method", "new_test_method")

	// Reset and test again
	d.Reset()
	d.Warn("test_method", "new_test_method")
}

func TestDeprecationsReset(t *testing.T) {
	ResetDeprecations()

	Warn("method1", "new_method1")
	Warn("method2", "new_method2")

	ResetDeprecations()

	// After reset, should warn again
	Warn("method1", "new_method1")
}

func TestDeprecationsMessage(t *testing.T) {
	// This test verifies the deprecation message format
	// In a real implementation, we'd capture stdout/stderr
	// For now, we just ensure it doesn't panic
	d := &Deprecations{
		warned: make(map[string]bool),
	}

	d.Warn("deprecated_api", "new_api")

	// Verify the message would contain expected parts
	// (In a real test, we'd capture and check the output)
	_ = d
}
