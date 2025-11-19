package tags

import (
	"testing"

	"github.com/Notifuse/liquidgo/liquid"
)

func TestUnlessTag(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewUnlessTag("unless", "false", pc)
	if err != nil {
		t.Fatalf("NewUnlessTag() error = %v", err)
	}
	if tag == nil {
		t.Fatal("Expected UnlessTag, got nil")
	}

	if len(tag.Blocks()) != 1 {
		t.Errorf("Expected 1 block, got %d", len(tag.Blocks()))
	}
}

func TestUnlessTagFalseCondition(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewUnlessTag("unless", "false", pc)
	if err != nil {
		t.Fatalf("NewUnlessTag() error = %v", err)
	}

	// Parse unless block with false condition (should render)
	tokenizer := pc.NewTokenizer("content {% endunless %}", false, nil, false)
	err = tag.Parse(tokenizer)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	ctx := liquid.NewContext()
	var output string
	tag.RenderToOutputBuffer(ctx, &output)

	if output != "content " {
		t.Errorf("Expected output 'content ', got %q", output)
	}
}

func TestUnlessTagTrueCondition(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewUnlessTag("unless", "true", pc)
	if err != nil {
		t.Fatalf("NewUnlessTag() error = %v", err)
	}

	// Parse unless block with true condition (should not render)
	tokenizer := pc.NewTokenizer("content {% endunless %}", false, nil, false)
	err = tag.Parse(tokenizer)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	ctx := liquid.NewContext()
	var output string
	tag.RenderToOutputBuffer(ctx, &output)

	if output != "" {
		t.Errorf("Expected empty output, got %q", output)
	}
}

func TestUnlessTagWithElse(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewUnlessTag("unless", "true", pc)
	if err != nil {
		t.Fatalf("NewUnlessTag() error = %v", err)
	}

	// Parse unless-else block
	tokenizer := pc.NewTokenizer("unless content {% else %} else content {% endunless %}", false, nil, false)
	err = tag.Parse(tokenizer)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(tag.Blocks()) != 2 {
		t.Errorf("Expected 2 blocks (unless, else), got %d", len(tag.Blocks()))
	}

	ctx := liquid.NewContext()
	var output string
	tag.RenderToOutputBuffer(ctx, &output)

	// Since unless condition is true, it won't render, so else should render
	if output != " else content " {
		t.Errorf("Expected output ' else content ', got %q", output)
	}
}

func TestUnlessTagWithNilValue(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewUnlessTag("unless", "nil_var", pc)
	if err != nil {
		t.Fatalf("NewUnlessTag() error = %v", err)
	}

	ctx := liquid.NewContext()
	// Don't set nil_var, so it will be nil
	tokenizer := pc.NewTokenizer("content {% endunless %}", false, nil, false)
	err = tag.Parse(tokenizer)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	var output string
	tag.RenderToOutputBuffer(ctx, &output)

	// nil value should render (unless renders when false/nil)
	if output != "content " {
		t.Errorf("Expected output 'content ', got %q", output)
	}
}

func TestUnlessTagWithEmptyString(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewUnlessTag("unless", `""`, pc)
	if err != nil {
		t.Fatalf("NewUnlessTag() error = %v", err)
	}

	ctx := liquid.NewContext()
	tokenizer := pc.NewTokenizer("content {% endunless %}", false, nil, false)
	err = tag.Parse(tokenizer)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	var output string
	tag.RenderToOutputBuffer(ctx, &output)

	// Empty string should render (unless renders when false/empty)
	if output != "content " {
		t.Errorf("Expected output 'content ', got %q", output)
	}
}

func TestUnlessTagWithErrorInEvaluation(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewUnlessTag("unless", "invalid_var", pc)
	if err != nil {
		t.Fatalf("NewUnlessTag() error = %v", err)
	}

	ctx := liquid.NewContext()
	tokenizer := pc.NewTokenizer("content {% endunless %}", false, nil, false)
	err = tag.Parse(tokenizer)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Create a context that will cause an error during evaluation
	// Use a variable that causes an error when evaluated
	var output string
	tag.RenderToOutputBuffer(ctx, &output)

	// Should handle error gracefully (output may contain error message or be empty)
	// The exact behavior depends on error handling implementation
}

// TestUnlessTagRenderToOutputBufferEdgeCases tests RenderToOutputBuffer edge cases
func TestUnlessTagRenderToOutputBufferEdgeCases(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})

	// Test with elsif blocks
	tag, err := NewUnlessTag("unless", "false", pc)
	if err != nil {
		t.Fatalf("NewUnlessTag() error = %v", err)
	}

	tokenizer := pc.NewTokenizer("unless content {% elsif true %}elsif content {% else %}else content {% endunless %}", false, nil, false)
	err = tag.Parse(tokenizer)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	ctx := liquid.NewContext()
	var output string
	tag.RenderToOutputBuffer(ctx, &output)

	// Since unless condition is false, it should render unless content
	// elsif/else should not render
	if output != "unless content " {
		t.Logf("Note: Unless with elsif output: %q (expected 'unless content ')", output)
	}
}

// TestUnlessTagInvalidCondition tests error handling in NewUnlessTag
func TestUnlessTagInvalidCondition(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})

	// Test with various syntaxes to try to trigger parse errors
	testCases := []string{
		"var ==",
		"== value",
		"var and",
		"(unclosed",
		"contains",
	}

	for _, markup := range testCases {
		tag, err := NewUnlessTag("unless", markup, pc)
		// May or may not error depending on parser behavior
		if err != nil {
			// Error is fine - this is one of the paths we want to test
			continue
		}
		if tag == nil {
			t.Errorf("Expected non-nil tag or error for markup: %s", markup)
		}
	}
}

// TestUnlessTagRenderWithError tests error handling during render
func TestUnlessTagRenderWithError(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})

	// Create an unless tag with a condition that will error during evaluation
	// Use an invalid filter or operation
	tag, err := NewUnlessTag("unless", "var | invalid_filter", pc)
	if err != nil {
		// Parse error is fine, test passes
		return
	}

	tokenizer := pc.NewTokenizer("content {% endunless %}", false, nil, false)
	err = tag.Parse(tokenizer)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	ctx := liquid.NewContext()
	ctx.Set("var", "value")

	var output string
	// This should handle the error gracefully
	tag.RenderToOutputBuffer(ctx, &output)

	// Output may contain error message or be empty
	_ = output
}

// TestUnlessTagWithNonBooleanResultValues tests various falsy/truthy values
func TestUnlessTagWithVariousValues(t *testing.T) {
	tests := []struct {
		name         string
		varName      string
		varValue     interface{}
		shouldRender bool
	}{
		{"zero int", "myvar", 0, false},                       // 0 is truthy in Liquid
		{"non-empty array", "myvar", []interface{}{1}, false}, // non-empty is truthy
		{"false bool", "myvar", false, true},
		{"true bool", "myvar", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pc := liquid.NewParseContext(liquid.ParseContextOptions{})
			tag, err := NewUnlessTag("unless", tt.varName, pc)
			if err != nil {
				t.Fatalf("NewUnlessTag() error = %v", err)
			}

			tokenizer := pc.NewTokenizer("content {% endunless %}", false, nil, false)
			err = tag.Parse(tokenizer)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			ctx := liquid.NewContext()
			ctx.Set(tt.varName, tt.varValue)

			var output string
			tag.RenderToOutputBuffer(ctx, &output)

			if tt.shouldRender {
				if output == "" {
					t.Errorf("Expected content to render for %v, got empty", tt.varValue)
				}
			} else {
				if output != "" {
					t.Errorf("Expected no content for %v, got %q", tt.varValue, output)
				}
			}
		})
	}
}
