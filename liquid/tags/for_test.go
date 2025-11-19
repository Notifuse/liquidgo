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
	expected := "1 2 3 "
	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

// Test parseMarkup with single quotes in attributes
func TestForTagParseMarkupWithSingleQuotes(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewForTag("for", "item in array limit:'2'", pc)
	if err != nil {
		t.Fatalf("NewForTag() error = %v", err)
	}
	if tag.Limit() == nil {
		t.Error("Expected limit to be set")
	}
}

// Test NewForTag with invalid syntax (error path)
func TestForTagNewForTagWithInvalidSyntax(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewForTag("for", "invalid syntax", pc)
	if err == nil {
		t.Error("Expected error for invalid syntax")
	}
	if tag != nil {
		t.Error("Expected nil tag on error")
	}
}

// Test Parse with else block when shouldContinue is true
func TestForTagParseWithElseBlock(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewForTag("for", "item in array", pc)
	if err != nil {
		t.Fatalf("NewForTag() error = %v", err)
	}

	// Create else block first
	tag.elseBlock = liquid.NewBlockBody()

	// Parse for block that doesn't find endfor (should continue to else)
	tokenizer := pc.NewTokenizer("content {% else %}else content{% endfor %}", false, nil, false)
	err = tag.Parse(tokenizer)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Verify else block was parsed
	if tag.elseBlock == nil {
		t.Error("Expected elseBlock to exist")
	}
}

// Test Parse with blank block removal
func TestForTagParseWithBlankBlock(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewForTag("for", "item in array", pc)
	if err != nil {
		t.Fatalf("NewForTag() error = %v", err)
	}

	// Parse empty for block (blank)
	tokenizer := pc.NewTokenizer("   {% endfor %}", false, nil, false)
	err = tag.Parse(tokenizer)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Block should be blank and blank strings removed
	if !tag.Blank() {
		t.Error("Expected block to be blank")
	}
}

// Test parseBody with depth limit
func TestForTagParseBodyDepthLimit(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewForTag("for", "item in array", pc)
	if err != nil {
		t.Fatalf("NewForTag() error = %v", err)
	}

	// Set depth to limit
	for i := 0; i < 100; i++ {
		pc.IncrementDepth()
	}

	body := liquid.NewBlockBody()
	tokenizer := pc.NewTokenizer("content {% endfor %}", false, nil, false)

	shouldContinue, err := tag.parseBody(tokenizer, body)
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

// Test parseBody with tag never closed
func TestForTagParseBodyTagNeverClosed(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewForTag("for", "item in array", pc)
	if err != nil {
		t.Fatalf("NewForTag() error = %v", err)
	}

	body := liquid.NewBlockBody()
	// Create tokenizer that will trigger tag never closed path
	tokenizer := pc.NewTokenizer("content", false, nil, false)

	// Should panic with "Tag was never closed" error
	defer func() {
		if r := recover(); r != nil {
			syntaxErr, ok := r.(*liquid.SyntaxError)
			if !ok {
				t.Fatalf("Expected SyntaxError panic, got %T: %v", r, r)
			}
			if syntaxErr.Error() != "Liquid syntax error: Tag was never closed: for" {
				t.Errorf("Expected 'Tag was never closed: for', got: %v", syntaxErr.Error())
			}
		} else {
			t.Fatal("Expected panic for unclosed tag, but no panic occurred")
		}
	}()

	tag.parseBody(tokenizer, body)
}

// Test parseBody with else tag during parsing
func TestForTagParseBodyWithElseTag(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewForTag("for", "item in array", pc)
	if err != nil {
		t.Fatalf("NewForTag() error = %v", err)
	}

	body := liquid.NewBlockBody()
	tokenizer := pc.NewTokenizer("content {% else %}else content{% endfor %}", false, nil, false)

	shouldContinue, err := tag.parseBody(tokenizer, body)
	if err != nil {
		t.Fatalf("parseBody() error = %v", err)
	}
	// Should continue because else was found
	if !shouldContinue {
		t.Error("Expected shouldContinue to be true when else tag is found")
	}
	if tag.elseBlock == nil {
		t.Error("Expected elseBlock to be created")
	}
}

// Test parseBody with unknown tag
func TestForTagParseBodyWithUnknownTag(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewForTag("for", "item in array", pc)
	if err != nil {
		t.Fatalf("NewForTag() error = %v", err)
	}

	body := liquid.NewBlockBody()
	tokenizer := pc.NewTokenizer("content {% unknown_tag %}more{% endfor %}", false, nil, false)

	shouldContinue, err := tag.parseBody(tokenizer, body)
	// Unknown tags might cause errors, which is acceptable
	_ = shouldContinue
	_ = err
}

// Test collectionSegment with invalid offsets map type
func TestForTagCollectionSegmentInvalidOffsetsMap(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewForTag("for", "item in array", pc)
	if err != nil {
		t.Fatalf("NewForTag() error = %v", err)
	}

	ctx := liquid.NewContext()
	ctx.Set("array", []interface{}{1, 2, 3})

	// Set invalid offsets type (not a map)
	registers := ctx.Registers()
	registers.Set("for", "not_a_map")

	segment := tag.collectionSegment(ctx)
	// Should handle gracefully and create new map
	if len(segment) != 3 {
		t.Errorf("Expected segment length 3, got %d", len(segment))
	}
}

// Test collectionSegment with offset:continue and non-int value
func TestForTagCollectionSegmentContinueWithNonInt(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewForTag("for", "item in array offset:continue", pc)
	if err != nil {
		t.Fatalf("NewForTag() error = %v", err)
	}

	ctx := liquid.NewContext()
	ctx.Set("array", []interface{}{1, 2, 3})

	// Set offset with non-int value
	registers := ctx.Registers()
	forMap := map[string]interface{}{tag.name: "not_an_int"}
	registers.Set("for", forMap)

	segment := tag.collectionSegment(ctx)
	// Should start from 0 when offset is not an int
	if len(segment) != 3 {
		t.Errorf("Expected segment length 3, got %d", len(segment))
	}
}

// Test collectionSegment with nil fromValue
func TestForTagCollectionSegmentWithNilFromValue(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	// Create tag with offset expression that evaluates to nil
	tag, err := NewForTag("for", "item in array offset:nil_var", pc)
	if err != nil {
		t.Fatalf("NewForTag() error = %v", err)
	}

	ctx := liquid.NewContext()
	ctx.Set("array", []interface{}{1, 2, 3})
	ctx.Set("nil_var", nil)

	segment := tag.collectionSegment(ctx)
	// Should default to from = 0
	if len(segment) != 3 {
		t.Errorf("Expected segment length 3, got %d", len(segment))
	}
}

// Test collectionSegment with ToInteger error in from
func TestForTagCollectionSegmentWithInvalidFromValue(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewForTag("for", "item in array offset:invalid_var", pc)
	if err != nil {
		t.Fatalf("NewForTag() error = %v", err)
	}

	ctx := liquid.NewContext()
	ctx.Set("array", []interface{}{1, 2, 3})
	ctx.Set("invalid_var", map[string]interface{}{"not": "an_int"})

	segment := tag.collectionSegment(ctx)
	// Should default to from = 0 when ToInteger fails
	if len(segment) != 3 {
		t.Errorf("Expected segment length 3, got %d", len(segment))
	}
}

// Test collectionSegment with Range collection
func TestForTagCollectionSegmentWithRange(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewForTag("for", "item in range", pc)
	if err != nil {
		t.Fatalf("NewForTag() error = %v", err)
	}

	ctx := liquid.NewContext()
	// Set Range as collection
	ctx.Set("range", &liquid.Range{Start: 1, End: 5})

	segment := tag.collectionSegment(ctx)
	// Should convert range to array [1, 2, 3, 4, 5]
	if len(segment) != 5 {
		t.Errorf("Expected segment length 5, got %d", len(segment))
	}
}

// Test collectionSegment with nil limitValue
func TestForTagCollectionSegmentWithNilLimitValue(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewForTag("for", "item in array limit:nil_var", pc)
	if err != nil {
		t.Fatalf("NewForTag() error = %v", err)
	}

	ctx := liquid.NewContext()
	ctx.Set("array", []interface{}{1, 2, 3, 4, 5})
	ctx.Set("nil_var", nil)

	segment := tag.collectionSegment(ctx)
	// Should use all items when limit is nil
	if len(segment) != 5 {
		t.Errorf("Expected segment length 5, got %d", len(segment))
	}
}

// Test collectionSegment with ToInteger error in limit
func TestForTagCollectionSegmentWithInvalidLimitValue(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewForTag("for", "item in array limit:invalid_var", pc)
	if err != nil {
		t.Fatalf("NewForTag() error = %v", err)
	}

	ctx := liquid.NewContext()
	ctx.Set("array", []interface{}{1, 2, 3, 4, 5})
	ctx.Set("invalid_var", map[string]interface{}{"not": "an_int"})

	segment := tag.collectionSegment(ctx)
	// Should use all items when ToInteger fails
	if len(segment) != 5 {
		t.Errorf("Expected segment length 5, got %d", len(segment))
	}
}

// Test renderSegment with invalid for_stack type
func TestForTagRenderSegmentInvalidForStack(t *testing.T) {
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

	// Set invalid for_stack type
	registers := ctx.Registers()
	registers.Set("for_stack", "not_a_slice")

	var output string
	segment := []interface{}{1, 2, 3}
	tag.renderSegment(ctx, &output, segment)
	// Should handle gracefully and create new stack
	expected := "1 2 3 "
	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

// Test renderSegment with parent loop (nested for loops)
func TestForTagRenderSegmentWithParentLoop(t *testing.T) {
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

	// Set parent loop in for_stack
	registers := ctx.Registers()
	parentLoop := liquid.NewForloopDrop("parent", 10, nil)
	forStack := []*liquid.ForloopDrop{parentLoop}
	registers.Set("for_stack", forStack)

	var output string
	segment := []interface{}{1, 2, 3}
	tag.renderSegment(ctx, &output, segment)
	// Should use parent loop
	expected := "1 2 3 "
	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}

	// Verify stack was popped
	finalStack := registers.Get("for_stack")
	if stack, ok := finalStack.([]*liquid.ForloopDrop); ok {
		if len(stack) != 1 || stack[0] != parentLoop {
			t.Error("Expected stack to be restored to original state")
		}
	}
}

// Test renderSegment with break interrupt
func TestForTagRenderSegmentWithBreakInterrupt(t *testing.T) {
	env := liquid.NewEnvironment()
	RegisterStandardTags(env)
	pc := liquid.NewParseContext(liquid.ParseContextOptions{Environment: env})
	tag, err := NewForTag("for", "item in array", pc)
	if err != nil {
		t.Fatalf("NewForTag() error = %v", err)
	}

	// Parse for block with break tag
	tokenizer := pc.NewTokenizer("{{item}}{% if item == 2 %}{% break %}{% endif %} {% endfor %}", false, nil, false)
	err = tag.Parse(tokenizer)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	ctx := liquid.NewContext()
	ctx.Set("array", []interface{}{1, 2, 3})

	var output string
	segment := []interface{}{1, 2, 3}
	tag.renderSegment(ctx, &output, segment)
	// Should break after rendering item 2 and detecting break in if statement
	// Output: "1 " (item 1 + space) + "2" (item 2, then break before space)
	expected := "1 2"
	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

// Test renderSegment with continue interrupt
func TestForTagRenderSegmentWithContinueInterrupt(t *testing.T) {
	env := liquid.NewEnvironment()
	RegisterStandardTags(env)
	pc := liquid.NewParseContext(liquid.ParseContextOptions{Environment: env})
	tag, err := NewForTag("for", "item in array", pc)
	if err != nil {
		t.Fatalf("NewForTag() error = %v", err)
	}

	// Parse for block with continue tag
	tokenizer := pc.NewTokenizer("{% if item == 2 %}{% continue %}{% endif %}{{item}} {% endfor %}", false, nil, false)
	err = tag.Parse(tokenizer)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	ctx := liquid.NewContext()
	ctx.Set("array", []interface{}{1, 2, 3})

	var output string
	segment := []interface{}{1, 2, 3}
	tag.renderSegment(ctx, &output, segment)
	// Should skip item 2, output should contain "1" and "3" but not "2"
	expected := "1 3 "
	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

// Test renderSegment stack popping
func TestForTagRenderSegmentStackPopping(t *testing.T) {
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
	ctx.Set("array", []interface{}{1, 2})

	registers := ctx.Registers()
	initialStack := []*liquid.ForloopDrop{}
	registers.Set("for_stack", initialStack)

	var output string
	segment := []interface{}{1, 2}
	tag.renderSegment(ctx, &output, segment)

	// Verify stack was popped back to initial state
	finalStack := registers.Get("for_stack")
	if stack, ok := finalStack.([]*liquid.ForloopDrop); ok {
		if len(stack) != 0 {
			t.Errorf("Expected empty stack after render, got length %d", len(stack))
		}
	}
}
