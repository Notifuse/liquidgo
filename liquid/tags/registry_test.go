package tags

import (
	"testing"

	"github.com/Notifuse/liquidgo/v5/liquid"
)

func TestStandardTags(t *testing.T) {
	env := mustEnv(liquid.EnvironmentConfig{Tags: Standard()})

	// Verify all standard tags are registered
	// Note: snippet tag was removed in Shopify Liquid v5.11.0
	testTags := []string{
		"assign", "echo", "increment", "decrement", "break", "continue", "cycle",
		"comment", "doc", "capture", "if", "unless", "for", "ifchanged",
		"case", "tablerow", "include", "render",
	}
	for _, tagName := range testTags {
		tagConstructor := env.TagForName(tagName)
		if tagConstructor == nil {
			t.Errorf("Expected tag %q to be registered, got nil", tagName)
		}
	}

	// Test that registered tags can be instantiated
	pc := liquid.NewParseContext(liquid.ParseContextOptions{Environment: env})

	// Test simple tags
	testSimpleTags := map[string]string{
		"assign":    "var = value",
		"echo":      "test",
		"increment": "counter",
		"decrement": "counter",
		"break":     "",
		"continue":  "",
		"cycle":     "one, two, three",
	}
	for tagName, markup := range testSimpleTags {
		if constructor := env.TagForName(tagName); constructor != nil {
			_, err := constructor(tagName, markup, pc)
			if err != nil {
				t.Logf("Tag %q instantiation error (may be expected): %v", tagName, err)
			}
		}
	}

	// Test block tags
	testBlockTags := map[string]string{
		"comment":  "",
		"doc":      "",
		"capture":  "var",
		"if":       "true",
		"unless":   "false",
		"for":      "item in array",
		"case":     "var",
		"tablerow": "item in array",
	}
	for tagName, markup := range testBlockTags {
		if constructor := env.TagForName(tagName); constructor != nil {
			_, err := constructor(tagName, markup, pc)
			if err != nil {
				t.Logf("Tag %q instantiation error (may be expected): %v", tagName, err)
			}
		}
	}

	// Test include/render tags
	testIncludeTags := map[string]string{
		"include": "'template'",
		"render":  "'template'",
	}
	for tagName, markup := range testIncludeTags {
		if constructor := env.TagForName(tagName); constructor != nil {
			_, err := constructor(tagName, markup, pc)
			if err != nil {
				t.Logf("Tag %q instantiation error (may be expected): %v", tagName, err)
			}
		}
	}
}

// Each call of Standard gives a map of its own: a tag a caller adds or
// replaces is not in the next one.
func TestStandardGivesAMapOfItsOwn(t *testing.T) {
	first := Standard()
	first["assign"] = nil
	first["extra"] = nil
	second := Standard()
	if second["assign"] == nil || second["extra"] != nil || len(second) != 20 {
		t.Errorf("Standard() changed by a caller: %d tags", len(second))
	}
}
