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

// TestI18nMessagesAreShopifys checks that the default locale is the English
// locale of Shopify Liquid, embedded in the package: a file named en, or
// liquid/locales/en.yml, in the working directory changes nothing, and a key
// the locale does not have translates to itself instead of panicking.
func TestI18nMessagesAreShopifys(t *testing.T) {
	dir := t.TempDir()
	hostile := "errors:\n  syntax:\n    unknown_tag: \"pwned %{tag}\"\n"
	if err := os.WriteFile(filepath.Join(dir, "en"), []byte(hostile), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "liquid", "locales"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "liquid", "locales", "en.yml"), []byte(hostile), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)

	for _, i18n := range []*I18n{NewI18n(""), NewI18n("en"), NewI18n(DefaultLocalePath), NewParseContext(ParseContextOptions{}).Locale()} {
		if got := i18n.T("errors.syntax.unknown_tag", map[string]interface{}{"tag": "x"}); got != "Unknown tag 'x'" {
			t.Errorf("unknown_tag translates to %q, want Shopify's", got)
		}
		if got := i18n.T("errors.syntax.for", nil); got != "Syntax Error in 'for loop' - Valid syntax: for [item] in [collection]" {
			t.Errorf("for translates to %q, want Shopify's", got)
		}
		if got := i18n.T("errors.syntax.nothing_here", nil); got != "errors.syntax.nothing_here" {
			t.Errorf("a missing key translates to %q, want the key", got)
		}
		if got := i18n.T("errors.syntax", nil); got != "errors.syntax" {
			t.Errorf("a key that names a group translates to %q, want the key", got)
		}
	}

	own := NewI18n(filepath.Join(dir, "en"))
	if got := own.T("errors.syntax.unknown_tag", map[string]interface{}{"tag": "x"}); got != "pwned x" {
		t.Errorf("a locale given by path translates to %q, want its own message", got)
	}
}
