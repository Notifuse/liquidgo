package tags

import (
	"testing"

	"github.com/Notifuse/liquidgo/liquid"
)

func TestSnippetTag(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewSnippetTag("snippet", "my_snippet", pc)
	if err != nil {
		t.Fatalf("NewSnippetTag() error = %v", err)
	}
	if tag == nil {
		t.Fatal("Expected SnippetTag, got nil")
	}

	if tag.To() != "my_snippet" {
		t.Errorf("Expected To 'my_snippet', got %q", tag.To())
	}
}

func TestSnippetTagBlank(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewSnippetTag("snippet", "my_snippet", pc)
	if err != nil {
		t.Fatalf("NewSnippetTag() error = %v", err)
	}

	if !tag.Blank() {
		t.Error("Expected snippet tag to be blank")
	}
}

func TestSnippetTagRenders(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewSnippetTag("snippet", "my_snippet", pc)
	if err != nil {
		t.Fatalf("NewSnippetTag() error = %v", err)
	}

	// Parse snippet block
	tokenizer := pc.NewTokenizer("content {% endsnippet %}", false, nil, false)
	err = tag.Parse(tokenizer)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	ctx := liquid.NewContext()
	ctx.SetTemplateName("test.liquid")
	var output string
	tag.RenderToOutputBuffer(ctx, &output)

	// Snippet tag should not output anything
	if output != "" {
		t.Errorf("Expected empty output, got %q", output)
	}

	// Check that snippet was assigned
	// Check scopes directly to avoid evaluation/conversion
	scopes := ctx.Scopes()
	if len(scopes) == 0 {
		t.Fatal("Expected at least one scope")
	}

	val, ok := scopes[0]["my_snippet"]
	if !ok {
		t.Error("Expected snippet to be assigned to my_snippet")
	}

	// Check that it's a SnippetDrop
	if drop, ok := val.(*liquid.SnippetDrop); ok {
		if drop.Name() != "my_snippet" {
			t.Errorf("Expected snippet name 'my_snippet', got %q", drop.Name())
		}
		if drop.Body() != "content " {
			t.Errorf("Expected snippet body 'content ', got %q", drop.Body())
		}
	} else {
		t.Errorf("Expected SnippetDrop, got %T", val)
	}
}
