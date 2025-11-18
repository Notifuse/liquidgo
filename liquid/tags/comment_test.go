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
