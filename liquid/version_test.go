package liquid

import "testing"

func TestVersion(t *testing.T) {
	if Version != "5.11.0" {
		t.Errorf("Expected version to be '5.11.0', got '%s'", Version)
	}
	if PortRevision != 5 {
		t.Errorf("Expected port revision 5, got %d", PortRevision)
	}
}
