package tags

import (
	"testing"

	"github.com/Notifuse/liquidgo/liquid"
)

func TestForTag(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewForTag("for", "item in array", pc)
	if err != nil {
		t.Fatalf("NewForTag() error = %v", err)
	}
	if tag == nil {
		t.Fatal("Expected ForTag, got nil")
	}

	if tag.VariableName() != "item" {
		t.Errorf("Expected variable name 'item', got %q", tag.VariableName())
	}
}

func TestForTagSimpleLoop(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewForTag("for", "item in array", pc)
	if err != nil {
		t.Fatalf("NewForTag() error = %v", err)
	}

	// Parse for block
	tokenizer := pc.NewTokenizer("{{item}} {% endfor %}", false, nil, false)
	err = tag.Parse(tokenizer)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	ctx := liquid.NewContext()
	ctx.Set("array", []interface{}{1, 2, 3})
	var output string
	tag.RenderToOutputBuffer(ctx, &output)

	if output != "1 2 3 " {
		t.Errorf("Expected output '1 2 3 ', got %q", output)
	}
}

func TestForTagEmptyCollection(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewForTag("for", "item in array", pc)
	if err != nil {
		t.Fatalf("NewForTag() error = %v", err)
	}

	// Parse for block with else
	tokenizer := pc.NewTokenizer("content {% else %}empty{% endfor %}", false, nil, false)
	err = tag.Parse(tokenizer)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	ctx := liquid.NewContext()
	ctx.Set("array", []interface{}{})
	var output string
	tag.RenderToOutputBuffer(ctx, &output)

	if output != "empty" {
		t.Errorf("Expected output 'empty', got %q", output)
	}
}

func TestForTagReversed(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewForTag("for", "item in array reversed", pc)
	if err != nil {
		t.Fatalf("NewForTag() error = %v", err)
	}

	// Parse for block
	tokenizer := pc.NewTokenizer("{{item}} {% endfor %}", false, nil, false)
	err = tag.Parse(tokenizer)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	ctx := liquid.NewContext()
	ctx.Set("array", []interface{}{1, 2, 3})
	var output string
	tag.RenderToOutputBuffer(ctx, &output)

	if output != "3 2 1 " {
		t.Errorf("Expected output '3 2 1 ', got %q", output)
	}
}

func TestForTagWithLimit(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewForTag("for", "item in array limit:2", pc)
	if err != nil {
		t.Fatalf("NewForTag() error = %v", err)
	}

	// Parse for block
	tokenizer := pc.NewTokenizer("{{item}} {% endfor %}", false, nil, false)
	err = tag.Parse(tokenizer)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	ctx := liquid.NewContext()
	ctx.Set("array", []interface{}{1, 2, 3, 4, 5})
	var output string
	tag.RenderToOutputBuffer(ctx, &output)

	if output != "1 2 " {
		t.Errorf("Expected output '1 2 ', got %q", output)
	}
}

func TestForTagWithOffset(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewForTag("for", "item in array offset:2 limit:2", pc)
	if err != nil {
		t.Fatalf("NewForTag() error = %v", err)
	}

	// Parse for block
	tokenizer := pc.NewTokenizer("{{item}} {% endfor %}", false, nil, false)
	err = tag.Parse(tokenizer)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	ctx := liquid.NewContext()
	ctx.Set("array", []interface{}{1, 2, 3, 4, 5})
	var output string
	tag.RenderToOutputBuffer(ctx, &output)

	if output != "3 4 " {
		t.Errorf("Expected output '3 4 ', got %q", output)
	}
}

func TestForTagCollectionName(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewForTag("for", "item in array", pc)
	if err != nil {
		t.Fatalf("NewForTag() error = %v", err)
	}

	collectionName := tag.CollectionName()
	if collectionName == nil {
		t.Error("Expected CollectionName() to return non-nil expression")
	}
}

func TestForTagLimit(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewForTag("for", "item in array limit:5", pc)
	if err != nil {
		t.Fatalf("NewForTag() error = %v", err)
	}

	limit := tag.Limit()
	if limit == nil {
		t.Error("Expected Limit() to return non-nil expression")
	}
}

func TestForTagFrom(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewForTag("for", "item in array offset:2", pc)
	if err != nil {
		t.Fatalf("NewForTag() error = %v", err)
	}

	from := tag.From()
	if from == nil {
		t.Error("Expected From() to return non-nil expression")
	}
}

func TestForTagNodelist(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewForTag("for", "item in array", pc)
	if err != nil {
		t.Fatalf("NewForTag() error = %v", err)
	}

	// Test Nodelist without else block
	nodelist := tag.Nodelist()
	if len(nodelist) != 1 {
		t.Errorf("Expected 1 node (forBlock), got %d", len(nodelist))
	}

	// Test Nodelist with else block
	tag2, _ := NewForTag("for", "item in array", pc)
	tag2.elseBlock = liquid.NewBlockBody()
	nodelist2 := tag2.Nodelist()
	if len(nodelist2) != 2 {
		t.Errorf("Expected 2 nodes (forBlock and elseBlock), got %d", len(nodelist2))
	}
}

func TestForTagUnknownTag(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewForTag("for", "item in array", pc)
	if err != nil {
		t.Fatalf("NewForTag() error = %v", err)
	}

	// Test with else tag (should be handled)
	tokenizer := pc.NewTokenizer("", false, nil, false)
	err = tag.UnknownTag("else", "", tokenizer)
	if err != nil {
		t.Errorf("Expected nil error for else, got %v", err)
	}
	if tag.elseBlock == nil {
		t.Error("Expected elseBlock to be created")
	}

	// Test with unknown tag (should error)
	err = tag.UnknownTag("unknown", "", tokenizer)
	if err == nil {
		t.Error("Expected error for unknown tag")
	}
}

func TestForTagCollectionSegment(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewForTag("for", "item in array", pc)
	if err != nil {
		t.Fatalf("NewForTag() error = %v", err)
	}

	ctx := liquid.NewContext()
	ctx.Set("array", []interface{}{1, 2, 3})

	// Test collectionSegment with simple array
	segment := tag.collectionSegment(ctx)
	if len(segment) != 3 {
		t.Errorf("Expected segment length 3, got %d", len(segment))
	}

	// Test with nil collection
	ctx2 := liquid.NewContext()
	ctx2.Set("array", nil)
	segment2 := tag.collectionSegment(ctx2)
	if len(segment2) != 0 {
		t.Errorf("Expected empty segment for nil collection, got %d", len(segment2))
	}

	// Test with offset:continue
	tag3, _ := NewForTag("for", "item in array offset:continue", pc)
	ctx3 := liquid.NewContext()
	ctx3.Set("array", []interface{}{1, 2, 3})
	// Set offset in registers
	registers := ctx3.Registers()
	forMap := map[string]interface{}{tag3.name: 1}
	registers.Set("for", forMap)
	segment3 := tag3.collectionSegment(ctx3)
	_ = segment3 // Should start from offset 1

	// Test with limit
	tag4, _ := NewForTag("for", "item in array limit:2", pc)
	ctx4 := liquid.NewContext()
	ctx4.Set("array", []interface{}{1, 2, 3, 4, 5})
	segment4 := tag4.collectionSegment(ctx4)
	if len(segment4) != 2 {
		t.Errorf("Expected segment length 2 with limit, got %d", len(segment4))
	}
}

func TestForTagRenderSegment(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewForTag("for", "item in array", pc)
	if err != nil {
		t.Fatalf("NewForTag() error = %v", err)
	}

	// Parse for block
	tokenizer := pc.NewTokenizer("{{item}} {% endfor %}", false, nil, false)
	err = tag.Parse(tokenizer)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	ctx := liquid.NewContext()
	ctx.Set("array", []interface{}{1, 2, 3})
	var output string

	// Test renderSegment
	segment := []interface{}{1, 2, 3}
	tag.renderSegment(ctx, &output, segment)
	if output == "" {
		t.Error("Expected non-empty output from renderSegment")
	}
}
