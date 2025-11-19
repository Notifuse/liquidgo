package tags

import (
	"testing"

	"github.com/Notifuse/liquidgo/liquid"
)

func TestCommentTag(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewCommentTag("comment", "", pc)
	if err != nil {
		t.Fatalf("NewCommentTag() error = %v", err)
	}
	if tag == nil {
		t.Fatal("Expected CommentTag, got nil")
	}

	if !tag.Blank() {
		t.Error("Expected Blank to be true")
	}

	ctx := liquid.NewContext()
	var output string
	tag.RenderToOutputBuffer(ctx, &output)

	if output != "" {
		t.Errorf("Expected empty output, got %q", output)
	}
}

func TestCommentTagParse(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tokenizer := pc.NewTokenizer("some content {% endcomment %}", false, nil, false)

	tag, err := NewCommentTag("comment", "", pc)
	if err != nil {
		t.Fatalf("NewCommentTag() error = %v", err)
	}

	err = tag.Parse(tokenizer)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Comment should be blank
	if !tag.Blank() {
		t.Error("Expected Blank to be true")
	}
}

func TestCommentTagNestedComments(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tokenizer := pc.NewTokenizer("outer {% comment %} inner {% endcomment %} more {% endcomment %}", false, nil, false)

	tag, err := NewCommentTag("comment", "", pc)
	if err != nil {
		t.Fatalf("NewCommentTag() error = %v", err)
	}

	err = tag.Parse(tokenizer)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Should parse successfully with nested comments
	if !tag.Blank() {
		t.Error("Expected Blank to be true")
	}
}

func TestCommentTagUnknownTag(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewCommentTag("comment", "", pc)
	if err != nil {
		t.Fatalf("NewCommentTag() error = %v", err)
	}

	// UnknownTag should return nil (comments ignore unknown tags)
	tokenizer := pc.NewTokenizer("", false, nil, false)
	err = tag.UnknownTag("unknown", "", tokenizer)
	if err != nil {
		t.Errorf("Expected nil error for unknown tag, got %v", err)
	}
}

func TestCommentTagRenderToOutputBuffer(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewCommentTag("comment", "", pc)
	if err != nil {
		t.Fatalf("NewCommentTag() error = %v", err)
	}

	// Parse comment block
	tokenizer := pc.NewTokenizer("some content {% endcomment %}", false, nil, false)
	err = tag.Parse(tokenizer)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	ctx := liquid.NewContext()
	var output string
	// Explicitly test RenderToOutputBuffer
	tag.RenderToOutputBuffer(ctx, &output)

	// Comment should render nothing
	if output != "" {
		t.Errorf("Expected empty output from RenderToOutputBuffer, got %q", output)
	}
}

func TestCommentTagParseRawTagBody(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewCommentTag("comment", "", pc)
	if err != nil {
		t.Fatalf("NewCommentTag() error = %v", err)
	}

	// Test parsing comment with raw tag inside
	tokenizer := pc.NewTokenizer("{% raw %}some content{% endraw %}{% endcomment %}", false, nil, false)
	err = tag.Parse(tokenizer)
	if err != nil {
		t.Fatalf("Parse() with raw tag error = %v", err)
	}

	// Should parse successfully
	if !tag.Blank() {
		t.Error("Expected Blank to be true")
	}
}

func TestCommentTagParseEdgeCases(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})

	// Test with depth limit
	pc2 := liquid.NewParseContext(liquid.ParseContextOptions{})
	for i := 0; i < 100; i++ {
		pc2.IncrementDepth()
	}
	tag2, _ := NewCommentTag("comment", "", pc2)
	tokenizer2 := pc2.NewTokenizer("content {% endcomment %}", false, nil, false)
	err := tag2.Parse(tokenizer2)
	if err == nil {
		t.Error("Expected error for depth limit exceeded")
	}

	// Test with whitespace trimming
	tag3, _ := NewCommentTag("comment", "", pc)
	tokenizer3 := pc.NewTokenizer("content {%- endcomment -%}", false, nil, false)
	err = tag3.Parse(tokenizer3)
	if err != nil {
		t.Fatalf("Parse() with whitespace trimming error = %v", err)
	}

	// Test with for_liquid_tag mode
	tag4, _ := NewCommentTag("comment", "", pc)
	tokenizer4 := pc.NewTokenizer("endcomment", true, nil, true)
	err = tag4.Parse(tokenizer4)
	if err != nil {
		// May error if tag never closed
		_ = err
	}

	// Test with tag never closed
	tag5, _ := NewCommentTag("comment", "", pc)
	tokenizer5 := pc.NewTokenizer("content", false, nil, false)
	err = tag5.Parse(tokenizer5)
	if err == nil {
		t.Error("Expected error for tag never closed")
	}

	// Test with multiple nested comments
	tag6, _ := NewCommentTag("comment", "", pc)
	tokenizer6 := pc.NewTokenizer("outer {% comment %} inner {% comment %} deep {% endcomment %} mid {% endcomment %} outer {% endcomment %}", false, nil, false)
	err = tag6.Parse(tokenizer6)
	if err != nil {
		t.Fatalf("Parse() with multiple nested comments error = %v", err)
	}
}

func TestCommentTagParseRawTagBodyEdgeCases(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, _ := NewCommentTag("comment", "", pc)

	// Test with raw tag never closed
	tokenizer := pc.NewTokenizer("{% raw %}content", false, nil, false)
	err := tag.parseRawTagBody(tokenizer)
	if err == nil {
		t.Error("Expected error for raw tag never closed")
	}

	// Test with endraw with spaces
	tokenizer2 := pc.NewTokenizer("{% endraw %}", false, nil, false)
	err = tag.parseRawTagBody(tokenizer2)
	if err != nil {
		t.Fatalf("parseRawTagBody() with endraw error = %v", err)
	}

	// Test with endraw with extra content
	tokenizer3 := pc.NewTokenizer("{% endraw extra %}", false, nil, false)
	err = tag.parseRawTagBody(tokenizer3)
	if err != nil {
		t.Fatalf("parseRawTagBody() with endraw extra error = %v", err)
	}
}

func TestCommentTagRenderToOutputBufferExplicit(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewCommentTag("comment", "", pc)
	if err != nil {
		t.Fatalf("NewCommentTag() error = %v", err)
	}

	// Parse comment block
	tokenizer := pc.NewTokenizer("some content {% endcomment %}", false, nil, false)
	err = tag.Parse(tokenizer)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	ctx := liquid.NewContext()
	var output string
	// Explicitly test RenderToOutputBuffer (should do nothing)
	tag.RenderToOutputBuffer(ctx, &output)

	// Comment should render nothing
	if output != "" {
		t.Errorf("Expected empty output from RenderToOutputBuffer, got %q", output)
	}
}
