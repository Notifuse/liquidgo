package liquid

import "testing"

func TestVersion(t *testing.T) {
	if Version != "5.11.0" {
		t.Errorf("Expected version to be '5.11.0', got '%s'", Version)
	}
	if PortRevision != 10 {
		t.Errorf("Expected port revision 10, got %d", PortRevision)
	}
}
