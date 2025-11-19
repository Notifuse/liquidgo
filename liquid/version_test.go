package liquid

import "testing"

func TestVersion(t *testing.T) {
	if Version != "5.10.0" {
		t.Errorf("Expected version to be '5.10.0', got '%s'", Version)
	}
}
