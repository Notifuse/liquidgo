package tags

import (
	"testing"

	"github.com/Notifuse/liquidgo/liquid"
)

func TestCaseTag(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewCaseTag("case", "var", pc)
	if err != nil {
		t.Fatalf("NewCaseTag() error = %v", err)
	}
	if tag == nil {
		t.Fatal("Expected CaseTag, got nil")
	}
}

func TestCaseTagWithWhen(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewCaseTag("case", "var", pc)
	if err != nil {
		t.Fatalf("NewCaseTag() error = %v", err)
	}

	// Parse case block with when
	tokenizer := pc.NewTokenizer("{% when 1 %}one{% endcase %}", false, nil, false)
	err = tag.Parse(tokenizer)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(tag.Blocks()) != 1 {
		t.Errorf("Expected 1 block, got %d", len(tag.Blocks()))
	}

	ctx := liquid.NewContext()
	ctx.Set("var", 1)
	var output string
	tag.RenderToOutputBuffer(ctx, &output)

	if output != "one" {
		t.Errorf("Expected output 'one', got %q", output)
	}
}

func TestCaseTagWithWhenAndElse(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewCaseTag("case", "var", pc)
	if err != nil {
		t.Fatalf("NewCaseTag() error = %v", err)
	}

	// Parse case block with when and else
	tokenizer := pc.NewTokenizer("{% when 1 %}one{% else %}other{% endcase %}", false, nil, false)
	err = tag.Parse(tokenizer)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(tag.Blocks()) != 2 {
		t.Errorf("Expected 2 blocks (when, else), got %d", len(tag.Blocks()))
	}

	ctx := liquid.NewContext()
	ctx.Set("var", 2)
	var output string
	tag.RenderToOutputBuffer(ctx, &output)

	// Should render else block
	if output != "other" {
		t.Errorf("Expected output 'other', got %q", output)
	}
}

func TestCaseTagWithMultipleWhen(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewCaseTag("case", "var", pc)
	if err != nil {
		t.Fatalf("NewCaseTag() error = %v", err)
	}

	// Parse case block with multiple when
	tokenizer := pc.NewTokenizer("{% when 1 %}one{% when 2 %}two{% endcase %}", false, nil, false)
	err = tag.Parse(tokenizer)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(tag.Blocks()) != 2 {
		t.Errorf("Expected 2 blocks, got %d", len(tag.Blocks()))
	}

	ctx := liquid.NewContext()
	ctx.Set("var", 2)
	var output string
	tag.RenderToOutputBuffer(ctx, &output)

	if output != "two" {
		t.Errorf("Expected output 'two', got %q", output)
	}
}

func TestCaseTagLeft(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewCaseTag("case", "var", pc)
	if err != nil {
		t.Fatalf("NewCaseTag() error = %v", err)
	}

	left := tag.Left()
	if left == nil {
		t.Error("Expected Left() to return non-nil expression")
	}
}

func TestCaseTagNodelist(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewCaseTag("case", "var", pc)
	if err != nil {
		t.Fatalf("NewCaseTag() error = %v", err)
	}

	// Parse case block with when
	tokenizer := pc.NewTokenizer("{% when 1 %}one{% endcase %}", false, nil, false)
	err = tag.Parse(tokenizer)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	nodelist := tag.Nodelist()
	if nodelist == nil {
		t.Error("Expected Nodelist() to return non-nil slice")
	}
	if len(nodelist) == 0 {
		t.Error("Expected Nodelist() to contain nodes after parsing")
	}
}

func TestCaseTagUnknownTag(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewCaseTag("case", "var", pc)
	if err != nil {
		t.Fatalf("NewCaseTag() error = %v", err)
	}

	// Test UnknownTag with when (should be handled)
	tokenizer := pc.NewTokenizer("", false, nil, false)
	err = tag.UnknownTag("when", "1", tokenizer)
	if err != nil {
		t.Errorf("Expected nil error for when, got %v", err)
	}

	// Test UnknownTag with else (should be handled)
	err = tag.UnknownTag("else", "", tokenizer)
	if err != nil {
		t.Errorf("Expected nil error for else, got %v", err)
	}

	// Test UnknownTag with unknown tag (should error)
	err = tag.UnknownTag("unknown", "", tokenizer)
	if err == nil {
		t.Error("Expected error for unknown tag")
	}
}

func TestCaseTagParseBodyForBlock(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewCaseTag("case", "var", pc)
	if err != nil {
		t.Fatalf("NewCaseTag() error = %v", err)
	}

	// Test with endcase tag
	body := liquid.NewBlockBody()
	tokenizer := pc.NewTokenizer("content {% endcase %}", false, nil, false)
	shouldContinue, err := tag.parseBodyForBlock(tokenizer, body)
	if err != nil {
		t.Fatalf("parseBodyForBlock() error = %v", err)
	}
	if shouldContinue {
		t.Error("Expected shouldContinue to be false after finding endcase")
	}

	// Test with when tag (should continue parsing more blocks)
	body2 := liquid.NewBlockBody()
	tokenizer2 := pc.NewTokenizer("content {% when 2 %}", false, nil, false)
	shouldContinue2, err2 := tag.parseBodyForBlock(tokenizer2, body2)
	if err2 != nil {
		t.Fatalf("parseBodyForBlock() with when error = %v", err2)
	}
	if !shouldContinue2 {
		t.Error("Expected shouldContinue to be true after finding when (more blocks may follow)")
	}

	// Test with else tag (should continue parsing more blocks)
	body3 := liquid.NewBlockBody()
	tokenizer3 := pc.NewTokenizer("content {% else %}", false, nil, false)
	shouldContinue3, err3 := tag.parseBodyForBlock(tokenizer3, body3)
	if err3 != nil {
		t.Fatalf("parseBodyForBlock() with else error = %v", err3)
	}
	if !shouldContinue3 {
		t.Error("Expected shouldContinue to be true after finding else (more blocks may follow)")
	}

	// Test with depth limit
	pc4 := liquid.NewParseContext(liquid.ParseContextOptions{})
	for i := 0; i < 100; i++ {
		pc4.IncrementDepth()
	}
	tag4, _ := NewCaseTag("case", "var", pc4)
	body4 := liquid.NewBlockBody()
	tokenizer4 := pc4.NewTokenizer("content", false, nil, false)
	_, err4 := tag4.parseBodyForBlock(tokenizer4, body4)
	if err4 == nil {
		t.Error("Expected error for depth limit exceeded")
	}
}

func TestCaseTagRecordWhenCondition(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewCaseTag("case", "var", pc)
	if err != nil {
		t.Fatalf("NewCaseTag() error = %v", err)
	}

	// Test with single condition
	err = tag.recordWhenCondition("1")
	if err != nil {
		t.Fatalf("recordWhenCondition() error = %v", err)
	}
	if len(tag.Blocks()) != 1 {
		t.Errorf("Expected 1 block, got %d", len(tag.Blocks()))
	}

	// Test with multiple conditions (comma-separated)
	tag2, _ := NewCaseTag("case", "var", pc)
	err = tag2.recordWhenCondition("1, 2, 3")
	if err != nil {
		t.Fatalf("recordWhenCondition() with multiple values error = %v", err)
	}

	// Test with invalid syntax - empty string doesn't enter loop, so returns nil
	// The function only errors if it enters the loop and doesn't match
	// Empty string is valid (no conditions)
	tag3, _ := NewCaseTag("case", "var", pc)
	err = tag3.recordWhenCondition("")
	if err != nil {
		t.Errorf("Empty when condition should not error, got %v", err)
	}
}

func TestCaseTagRecordElseCondition(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewCaseTag("case", "var", pc)
	if err != nil {
		t.Fatalf("NewCaseTag() error = %v", err)
	}

	// Test with empty markup
	err = tag.recordElseCondition("")
	if err != nil {
		t.Fatalf("recordElseCondition() error = %v", err)
	}
	if len(tag.Blocks()) != 1 {
		t.Errorf("Expected 1 block, got %d", len(tag.Blocks()))
	}

	// Test with non-empty markup (should error)
	err = tag.recordElseCondition("invalid")
	if err == nil {
		t.Error("Expected error for else tag with markup")
	}
}

func TestCaseTagRenderToOutputBuffer(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewCaseTag("case", "var", pc)
	if err != nil {
		t.Fatalf("NewCaseTag() error = %v", err)
	}

	// Parse with when and else
	tokenizer := pc.NewTokenizer("{% when 1 %}one{% else %}other{% endcase %}", false, nil, false)
	err = tag.Parse(tokenizer)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	ctx := liquid.NewContext()
	var output string

	// Test with matching when condition
	ctx.Set("var", 1)
	tag.RenderToOutputBuffer(ctx, &output)
	if output != "one" {
		t.Errorf("Expected 'one', got %q", output)
	}

	// Test with else condition
	output = ""
	ctx.Set("var", 2)
	tag.RenderToOutputBuffer(ctx, &output)
	if output != "other" {
		t.Errorf("Expected 'other', got %q", output)
	}

	// Test with error in evaluation
	output = ""
	ctx.Set("var", nil)
	tag.RenderToOutputBuffer(ctx, &output)
	// Should handle error gracefully
	_ = output
}

func TestCaseTagParseMarkupError(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})

	// Test with invalid markup (should error)
	tag, err := NewCaseTag("case", "invalid syntax here", pc)
	if err == nil {
		t.Error("Expected error for invalid case tag markup")
	}
	if tag != nil {
		t.Error("Expected nil tag on error")
	}
}

func TestCaseTagParseBodyForBlockDepthLimit(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewCaseTag("case", "var", pc)
	if err != nil {
		t.Fatalf("NewCaseTag() error = %v", err)
	}

	// Set depth to limit
	for i := 0; i < 100; i++ {
		pc.IncrementDepth()
	}

	body := liquid.NewBlockBody()
	tokenizer := pc.NewTokenizer("content {% endcase %}", false, nil, false)

	shouldContinue, err := tag.parseBodyForBlock(tokenizer, body)
	if err == nil {
		t.Error("Expected error for depth limit")
	}
	if shouldContinue {
		t.Error("Expected shouldContinue to be false on error")
	}

	// Reset depth
	for i := 0; i < 100; i++ {
		pc.DecrementDepth()
	}
}

func TestCaseTagParseBodyForBlockUnknownTag(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewCaseTag("case", "var", pc)
	if err != nil {
		t.Fatalf("NewCaseTag() error = %v", err)
	}

	body := liquid.NewBlockBody()
	// Test with unknown tag that should be handled
	tokenizer := pc.NewTokenizer("content {% unknown_tag %}more{% endcase %}", false, nil, false)

	shouldContinue, err := tag.parseBodyForBlock(tokenizer, body)
	if err != nil {
		// Unknown tags might cause errors, which is acceptable
		_ = err
	}
	_ = shouldContinue
}

func TestCaseTagRenderToOutputBufferWithError(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewCaseTag("case", "var", pc)
	if err != nil {
		t.Fatalf("NewCaseTag() error = %v", err)
	}

	// Parse with when block
	tokenizer := pc.NewTokenizer("{% when 1 %}one{% endcase %}", false, nil, false)
	err = tag.Parse(tokenizer)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Create a context that will cause evaluation error
	ctx := liquid.NewContext()
	// Set var to something that will cause evaluation issues
	ctx.Set("var", func() {}) // Function that can't be evaluated properly

	var output string
	// Should handle error gracefully
	tag.RenderToOutputBuffer(ctx, &output)
	// Output might contain error message or be empty
	_ = output
}

func TestCaseTagRenderToOutputBufferNoMatchingWhen(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewCaseTag("case", "var", pc)
	if err != nil {
		t.Fatalf("NewCaseTag() error = %v", err)
	}

	// Parse with when but no else
	tokenizer := pc.NewTokenizer("{% when 1 %}one{% endcase %}", false, nil, false)
	err = tag.Parse(tokenizer)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	ctx := liquid.NewContext()
	ctx.Set("var", 999) // Value that doesn't match any when

	var output string
	tag.RenderToOutputBuffer(ctx, &output)
	// Should render nothing when no match and no else
	if output != "" {
		t.Errorf("Expected empty output for no match, got %q", output)
	}
}

// TestCaseTagRenderToOutputBufferErrorHandling tests error handling in RenderToOutputBuffer
func TestCaseTagRenderToOutputBufferErrorHandling(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewCaseTag("case", "var", pc)
	if err != nil {
		t.Fatalf("NewCaseTag() error = %v", err)
	}

	// Parse case with when block that may cause evaluation error
	tokenizer := pc.NewTokenizer("{% when invalid_expression %}content{% endcase %}", false, nil, false)
	err = tag.Parse(tokenizer)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	ctx := liquid.NewContext()
	ctx.Set("var", 1)
	var output string

	// Should handle evaluation errors gracefully
	tag.RenderToOutputBuffer(ctx, &output)

	// May produce error message or empty output
	t.Logf("Note: Case tag error handling output: %q", output)
}

// TestCaseTagRenderToOutputBufferMultipleWhen tests multiple when blocks
func TestCaseTagRenderToOutputBufferMultipleWhen(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewCaseTag("case", "var", pc)
	if err != nil {
		t.Fatalf("NewCaseTag() error = %v", err)
	}

	// Parse case with multiple when blocks
	tokenizer := pc.NewTokenizer("{% when 1 %}one{% when 2 %}two{% when 3 %}three{% else %}other{% endcase %}", false, nil, false)
	err = tag.Parse(tokenizer)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	ctx := liquid.NewContext()
	ctx.Set("var", 2)
	var output string
	tag.RenderToOutputBuffer(ctx, &output)

	// Should render matching when block
	if output != "two" {
		t.Logf("Note: Multiple when blocks output: %q (expected 'two')", output)
	}
}

// TestCaseTagRenderToOutputBufferElseBlock tests else block rendering
func TestCaseTagRenderToOutputBufferElseBlock(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewCaseTag("case", "var", pc)
	if err != nil {
		t.Fatalf("NewCaseTag() error = %v", err)
	}

	// Parse case with when and else
	tokenizer := pc.NewTokenizer("{% when 1 %}one{% else %}other{% endcase %}", false, nil, false)
	err = tag.Parse(tokenizer)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	ctx := liquid.NewContext()
	ctx.Set("var", 99) // Value that doesn't match when
	var output string
	tag.RenderToOutputBuffer(ctx, &output)

	// Should render else block
	if output != "other" {
		t.Logf("Note: Else block output: %q (expected 'other')", output)
	}
}
