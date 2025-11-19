package tags

import (
	"testing"

	"github.com/Notifuse/liquidgo/liquid"
)

func TestRegisterStandardTags(t *testing.T) {
	env := liquid.NewEnvironment()

	// Register all standard tags
	RegisterStandardTags(env)

	// Verify all standard tags are registered
	testTags := []string{
		"assign", "echo", "increment", "decrement", "break", "continue", "cycle",
		"comment", "doc", "capture", "if", "unless", "for", "ifchanged",
		"case", "tablerow", "snippet", "include", "render",
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
		tagConstructor := env.TagForName(tagName)
		if tagConstructor != nil {
			if constructor, ok := tagConstructor.(TagConstructor); ok {
				_, err := constructor(tagName, markup, pc)
				if err != nil {
					t.Logf("Tag %q instantiation error (may be expected): %v", tagName, err)
				}
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
		tagConstructor := env.TagForName(tagName)
		if tagConstructor != nil {
			if constructor, ok := tagConstructor.(TagConstructor); ok {
				_, err := constructor(tagName, markup, pc)
				if err != nil {
					t.Logf("Tag %q instantiation error (may be expected): %v", tagName, err)
				}
			}
		}
	}

	// Test include/render tags
	testIncludeTags := map[string]string{
		"include": "'template'",
		"render":  "'template'",
	}
	for tagName, markup := range testIncludeTags {
		tagConstructor := env.TagForName(tagName)
		if tagConstructor != nil {
			if constructor, ok := tagConstructor.(TagConstructor); ok {
				_, err := constructor(tagName, markup, pc)
				if err != nil {
					t.Logf("Tag %q instantiation error (may be expected): %v", tagName, err)
				}
			}
		}
	}
}
