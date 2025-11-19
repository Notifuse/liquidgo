package liquid

import (
	"testing"
)

func TestUsageIncrement(t *testing.T) {
	u := &Usage{}
	// Should not panic
	u.Increment("test")
}

func TestIncrementUsage(t *testing.T) {
	// Should not panic
	IncrementUsage("test")
}
