package liquid

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
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

func TestLocalFileSystemNestedPath(t *testing.T) {
	tmpDir := t.TempDir()
	fs := NewLocalFileSystem(tmpDir, "")

	// Test nested path: dir/mypartial should become root/dir/_mypartial.liquid
	templatePath := "dir/mypartial"
	fullPath, err := fs.FullPath(templatePath)
	if err != nil {
		t.Fatalf("Failed to get full path: %v", err)
	}

	expectedPath := filepath.Join(tmpDir, "dir", "_mypartial.liquid")
	if fullPath != expectedPath {
		t.Errorf("Expected path '%s', got '%s'", expectedPath, fullPath)
	}

	// Create the directory and file
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}
	if err := os.WriteFile(fullPath, []byte("nested content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Read it back
	content, err := fs.ReadTemplateFile(templatePath)
	if err != nil {
		t.Fatalf("Failed to read template: %v", err)
	}

	if content != "nested content" {
		t.Errorf("Expected 'nested content', got '%s'", content)
	}
}

func TestLocalFileSystemNestedPathWithCustomPattern(t *testing.T) {
	tmpDir := t.TempDir()
	fs := NewLocalFileSystem(tmpDir, "%s.html")

	// Test nested path with custom pattern: dir/mypartial should become root/dir/mypartial.html
	templatePath := "dir/mypartial"
	fullPath, err := fs.FullPath(templatePath)
	if err != nil {
		t.Fatalf("Failed to get full path: %v", err)
	}

	expectedPath := filepath.Join(tmpDir, "dir", "mypartial.html")
	if fullPath != expectedPath {
		t.Errorf("Expected path '%s', got '%s'", expectedPath, fullPath)
	}
}

func TestLocalFileSystemDeeplyNestedPath(t *testing.T) {
	tmpDir := t.TempDir()
	fs := NewLocalFileSystem(tmpDir, "")

	// Test deeply nested path: a/b/c/template should become root/a/b/c/_template.liquid
	templatePath := "a/b/c/template"
	fullPath, err := fs.FullPath(templatePath)
	if err != nil {
		t.Fatalf("Failed to get full path: %v", err)
	}

	expectedPath := filepath.Join(tmpDir, "a", "b", "c", "_template.liquid")
	if fullPath != expectedPath {
		t.Errorf("Expected path '%s', got '%s'", expectedPath, fullPath)
	}
}

func TestLocalFileSystemReadTemplateFileNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	fs := NewLocalFileSystem(tmpDir, "")

	_, err := fs.ReadTemplateFile("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
	if _, ok := err.(*FileSystemError); !ok {
		t.Errorf("Expected FileSystemError, got %T", err)
	}
}

func TestLocalFileSystemFullPathInvalidName(t *testing.T) {
	tmpDir := t.TempDir()
	fs := NewLocalFileSystem(tmpDir, "")

	tests := []struct {
		name         string
		templatePath string
	}{
		{"starts with dot", "./test"},
		{"starts with slash", "/test"},
		{"contains dot dot", "../test"},
		{"empty string", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := fs.FullPath(tt.templatePath)
			if err == nil {
				t.Errorf("Expected error for invalid template path: %s", tt.templatePath)
			}
		})
	}
}

func TestLocalFileSystemFullPathSecurityCheck(t *testing.T) {
	tmpDir := t.TempDir()
	fs := NewLocalFileSystem(tmpDir, "")

	// Create a symlink or try to escape the root directory
	// This test ensures that even if a path resolves outside root, it's caught
	// Note: This is a basic test - actual path traversal attacks would be more complex
	_, err := fs.FullPath("valid/../../etc/passwd")
	if err == nil {
		t.Error("Expected error for path traversal attempt")
	}
}

func TestLocalFileSystemReadTemplateFileError(t *testing.T) {
	tmpDir := t.TempDir()
	fs := NewLocalFileSystem(tmpDir, "")

	// Create a directory with the template name (not a file)
	templateName := "test"
	fullPath, err := fs.FullPath(templateName)
	if err != nil {
		t.Fatalf("Failed to get full path: %v", err)
	}

	// Create directory instead of file
	if err := os.MkdirAll(fullPath, 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	// Try to read it as a file - should fail
	_, err = fs.ReadTemplateFile(templateName)
	if err == nil {
		t.Error("Expected error when trying to read directory as file")
	}
}

func TestLocalFileSystemFullPathWithNestedInvalidName(t *testing.T) {
	tmpDir := t.TempDir()
	fs := NewLocalFileSystem(tmpDir, "")

	// Test that nested paths with invalid names are caught
	_, err := fs.FullPath("valid/../invalid")
	if err == nil {
		t.Error("Expected error for nested path with invalid name")
	}
}

func TestLocalFileSystemFullPathAbsolutePath(t *testing.T) {
	tmpDir := t.TempDir()
	fs := NewLocalFileSystem(tmpDir, "")

	// Test that absolute paths work correctly
	templatePath := "test"
	fullPath, err := fs.FullPath(templatePath)
	if err != nil {
		t.Fatalf("Failed to get full path: %v", err)
	}

	// Verify it's an absolute path
	if !filepath.IsAbs(fullPath) {
		t.Errorf("Expected absolute path, got relative path: %s", fullPath)
	}

	// Verify it's within the root directory
	absRoot, err := filepath.Abs(tmpDir)
	if err != nil {
		t.Fatalf("Failed to get absolute root: %v", err)
	}

	if !strings.HasPrefix(fullPath, absRoot) {
		t.Errorf("Path %s is not within root %s", fullPath, absRoot)
	}
}

// A LocalFileSystem reads only files under its root: no name reaches a file
// out of it, not even through a symbolic link under the root, and a
// prefix of the root is no root (a sibling directory named root-other).
func TestLocalFileSystemReadsOnlyUnderItsRoot(t *testing.T) {
	base := t.TempDir()
	root := filepath.Join(base, "root")
	outside := filepath.Join(base, "root-other")
	for _, dir := range []string{root, outside} {
		if err := os.Mkdir(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(outside, "_secret.liquid"), []byte("SECRET"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(root, "evil")); err != nil {
		t.Skip("no symbolic links here:", err)
	}
	fs := NewLocalFileSystem(root, "")
	for _, name := range []string{
		"evil/secret", "/etc/passwd", "../root-other/secret", "a/../../root-other/secret",
		"..", "a\\..\\b", "a%2e%2e/b", "a\x00b", "", "évil",
	} {
		content, err := fs.ReadTemplateFile(name)
		var fsErr *FileSystemError
		if err == nil || !errors.As(err, &fsErr) || strings.Contains(content, "SECRET") {
			t.Errorf("%q read %q (%v), want a FileSystemError", name, content, err)
		}
	}
}

// The errors of a LocalFileSystem name the template, never a path of the
// host: a render writes them in place of the include.
func TestLocalFileSystemErrorsNameNoPath(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "_dir.liquid"), 0o755); err != nil {
		t.Fatal(err)
	}
	fs := NewLocalFileSystem(root, "")
	for _, name := range []string{"missing", "dir"} {
		_, err := fs.ReadTemplateFile(name)
		if err == nil {
			t.Fatalf("%q: no error", name)
		}
		if strings.Contains(err.Error(), root) || strings.Contains(err.Error(), string(filepath.Separator)+"_") {
			t.Errorf("%q: the error %q shows a path of the host", name, err.Error())
		}
		if !strings.Contains(err.Error(), "'"+name+"'") {
			t.Errorf("%q: the error %q does not name the template", name, err.Error())
		}
	}
}

// A symbolic link that loops is a template the file system cannot read, and
// the error names the template, not a path of the host.
func TestLocalFileSystemALoopingLink(t *testing.T) {
	root := t.TempDir()
	if err := os.Symlink(filepath.Join(root, "_loop.liquid"), filepath.Join(root, "_loop.liquid")); err != nil {
		t.Skip("no symbolic links here:", err)
	}
	_, err := NewLocalFileSystem(root, "").ReadTemplateFile("loop")
	var fsErr *FileSystemError
	if !errors.As(err, &fsErr) || strings.Contains(err.Error(), root) || !strings.Contains(err.Error(), "'loop'") {
		t.Errorf("error %v, want a FileSystemError that names loop and no path", err)
	}
}
