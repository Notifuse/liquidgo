package liquid

import "testing"

func TestVersion(t *testing.T) {
	if Version != "5.14.0" {
		t.Errorf("Expected version to be '5.14.0', got '%s'", Version)
	}
	if PortRevision != 6 {
		t.Errorf("Expected port revision 6, got %d", PortRevision)
	}
}
