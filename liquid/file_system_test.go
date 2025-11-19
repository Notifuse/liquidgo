package liquid

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBlankFileSystem(t *testing.T) {
	fs := &BlankFileSystem{}
	_, err := fs.ReadTemplateFile("test")
	if err == nil {
		t.Error("Expected error from BlankFileSystem.ReadTemplateFile")
	}
}

func TestLocalFileSystem(t *testing.T) {
	tmpDir := t.TempDir()
	fs := NewLocalFileSystem(tmpDir, "")

	// Create a test template file
	templateName := "test"
	fullPath, err := fs.FullPath(templateName)
	if err != nil {
		t.Fatalf("Failed to get full path: %v", err)
	}

	expectedPath := filepath.Join(tmpDir, "_test.liquid")
	if fullPath != expectedPath {
		t.Errorf("Expected path '%s', got '%s'", expectedPath, fullPath)
	}

	// Create the file
	if err := os.WriteFile(fullPath, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Read it back
	content, err := fs.ReadTemplateFile(templateName)
	if err != nil {
		t.Fatalf("Failed to read template: %v", err)
	}

	if content != "test content" {
		t.Errorf("Expected 'test content', got '%s'", content)
	}
}

func TestLocalFileSystemInvalidName(t *testing.T) {
	tmpDir := t.TempDir()
	fs := NewLocalFileSystem(tmpDir, "")

	_, err := fs.FullPath("../invalid")
	if err == nil {
		t.Error("Expected error for invalid template name")
	}
}

func TestLocalFileSystemCustomPattern(t *testing.T) {
	tmpDir := t.TempDir()
	fs := NewLocalFileSystem(tmpDir, "%s.html")

	fullPath, err := fs.FullPath("index")
	if err != nil {
		t.Fatalf("Failed to get full path: %v", err)
	}

	expectedPath := filepath.Join(tmpDir, "index.html")
	if fullPath != expectedPath {
		t.Errorf("Expected path '%s', got '%s'", expectedPath, fullPath)
	}
}
