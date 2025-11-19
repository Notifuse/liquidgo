package liquid

import (
	"os"
	"path/filepath"
	"testing"
)

func TestI18n(t *testing.T) {
	// Create a temporary locale file for testing
	tmpDir := t.TempDir()
	localeFile := filepath.Join(tmpDir, "en.yml")

	localeContent := `en:
  errors:
    syntax:
      unknown_tag: "Unknown tag '%{tag}'"
      unexpected_else: "Unexpected else"
`

	if err := os.WriteFile(localeFile, []byte(localeContent), 0644); err != nil {
		t.Fatalf("Failed to create test locale file: %v", err)
	}

	i18n := NewI18n(localeFile)

	// Test translation
	vars := map[string]interface{}{"tag": "mytag"}
	result := i18n.T("en.errors.syntax.unknown_tag", vars)

	expected := "Unknown tag 'mytag'"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

func TestI18nDefaultPath(t *testing.T) {
	i18n := NewI18n("")
	if i18n.path != DefaultLocalePath {
		t.Errorf("Expected default path '%s', got '%s'", DefaultLocalePath, i18n.path)
	}
}
