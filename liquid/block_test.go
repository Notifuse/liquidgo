package liquid

import (
	"strings"
	"testing"
)

func TestBlockBasic(t *testing.T) {
	lineNum := 1
	pc := &mockParseContextForTag{lineNum: &lineNum, env: NewEnvironment()}

	block := NewBlock("if", "condition", pc)
	if block == nil {
		t.Fatal("Expected Block, got nil")
	}
	if block.BlockName() != "if" {
		t.Errorf("Expected block name 'if', got '%s'", block.BlockName())
	}
	if block.BlockDelimiter() != "endif" {
		t.Errorf("Expected delimiter 'endif', got '%s'", block.BlockDelimiter())
	}
}

func TestBlockDelimiter(t *testing.T) {
	lineNum := 1
	pc := &mockParseContextForTag{lineNum: &lineNum, env: NewEnvironment()}

	block := NewBlock("for", "item in items", pc)
	if block.BlockDelimiter() != "endfor" {
		t.Errorf("Expected delimiter 'endfor', got '%s'", block.BlockDelimiter())
	}

	block.SetBlockDelimiter("endloop")
	if block.BlockDelimiter() != "endloop" {
		t.Errorf("Expected delimiter 'endloop', got '%s'", block.BlockDelimiter())
	}
}

func TestBlockBlank(t *testing.T) {
	lineNum := 1
	pc := &mockParseContextForTag{lineNum: &lineNum, env: NewEnvironment()}

	block := NewBlock("if", "condition", pc)
	if !block.Blank() {
		t.Error("Expected block to be blank initially")
	}
}

func TestBlockNodelist(t *testing.T) {
	lineNum := 1
	pc := &mockParseContextForTag{lineNum: &lineNum, env: NewEnvironment()}

	block := NewBlock("if", "condition", pc)
	nodelist := block.Nodelist()
	if nodelist == nil {
		t.Error("Expected nodelist, got nil")
	}
	if len(nodelist) != 0 {
		t.Errorf("Expected empty nodelist, got %d items", len(nodelist))
	}

	// Test with body set
	pc2 := NewParseContext(ParseContextOptions{})
	block2 := NewBlock("if", "condition", pc2)
	block2.body = NewBlockBody()
	block2.body.nodelist = []interface{}{"content"}
	nodelist2 := block2.Nodelist()
	if len(nodelist2) != 1 {
		t.Errorf("Expected nodelist with 1 item, got %d items", len(nodelist2))
	}
}

func TestRaiseUnknownTag(t *testing.T) {
	lineNum := 1
	pc := &mockParseContextForTag{lineNum: &lineNum, env: NewEnvironment()}

	err := RaiseUnknownTag("unknown", "if", "endif", pc)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if _, ok := err.(*SyntaxError); !ok {
		t.Errorf("Expected SyntaxError, got %T", err)
	}
}

func TestRaiseUnknownTagElse(t *testing.T) {
	lineNum := 1
	pc := &mockParseContextForTag{lineNum: &lineNum, env: NewEnvironment()}

	err := RaiseUnknownTag("else", "if", "endif", pc)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if _, ok := err.(*SyntaxError); !ok {
		t.Errorf("Expected SyntaxError, got %T", err)
	}
}

func TestRaiseUnknownTagEnd(t *testing.T) {
	lineNum := 1
	pc := &mockParseContextForTag{lineNum: &lineNum, env: NewEnvironment()}

	err := RaiseUnknownTag("endunless", "if", "endif", pc)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if _, ok := err.(*SyntaxError); !ok {
		t.Errorf("Expected SyntaxError, got %T", err)
	}
}

// TestParseBlock tests the ParseBlock function
func TestParseBlock(t *testing.T) {
	lineNum := 1
	pc := NewParseContext(ParseContextOptions{})
	pc.SetLineNumber(&lineNum)

	source := "content{% endif %}"
	tokenizer := NewTokenizer(source, nil, false, nil, false)

	block, err := ParseBlock("if", "condition", tokenizer, pc)
	if err != nil {
		t.Fatalf("ParseBlock() error = %v", err)
	}
	if block == nil {
		t.Fatal("Expected Block, got nil")
	}
	if block.BlockName() != "if" {
		t.Errorf("Expected block name 'if', got '%s'", block.BlockName())
	}
}

// TestBlockParse tests block parsing with various scenarios
func TestBlockParse(t *testing.T) {
	lineNum := 1
	pc := NewParseContext(ParseContextOptions{})
	pc.SetLineNumber(&lineNum)

	// Test parsing with end tag
	source := "content{% endif %}"
	tokenizer := NewTokenizer(source, nil, false, nil, false)
	block := NewBlock("if", "condition", pc)
	err := block.Parse(tokenizer)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if block.body == nil {
		t.Error("Expected body to be set")
	}

	// Test parsing with nested content
	source2 := "outer{% if inner %}{% endif %}inner{% endif %}"
	tokenizer2 := NewTokenizer(source2, nil, false, nil, false)
	block2 := NewBlock("if", "condition", pc)
	err = block2.Parse(tokenizer2)
	if err != nil {
		t.Fatalf("Parse() with nested content error = %v", err)
	}
}

// TestBlockRender tests block rendering
func TestBlockRender(t *testing.T) {
	lineNum := 1
	pc := NewParseContext(ParseContextOptions{})
	pc.SetLineNumber(&lineNum)

	block := NewBlock("if", "condition", pc)
	block.body = NewBlockBody()
	block.body.nodelist = []interface{}{"hello", " ", "world"}

	ctx := NewContext()
	result := block.Render(ctx)
	if result != "hello world" {
		t.Errorf("Expected 'hello world', got %q", result)
	}

	// Test with nil body
	block2 := NewBlock("if", "condition", pc)
	result2 := block2.Render(ctx)
	if result2 != "" {
		t.Errorf("Expected empty string for nil body, got %q", result2)
	}
}

// TestBlockRenderToOutputBuffer tests buffer rendering
func TestBlockRenderToOutputBuffer(t *testing.T) {
	lineNum := 1
	pc := NewParseContext(ParseContextOptions{})
	pc.SetLineNumber(&lineNum)

	block := NewBlock("if", "condition", pc)
	block.body = NewBlockBody()
	block.body.nodelist = []interface{}{"test"}

	ctx := NewContext()
	output := ""
	block.RenderToOutputBuffer(ctx, &output)
	if output != "test" {
		t.Errorf("Expected 'test', got %q", output)
	}

	// Test with nil body
	block2 := NewBlock("if", "condition", pc)
	output2 := ""
	block2.RenderToOutputBuffer(ctx, &output2)
	if output2 != "" {
		t.Errorf("Expected empty output for nil body, got %q", output2)
	}
}

// TestBlockUnknownTag tests UnknownTag method with various scenarios
func TestBlockUnknownTag(t *testing.T) {
	lineNum := 1
	pc := &mockParseContextForTag{lineNum: &lineNum, env: NewEnvironment()}

	block := NewBlock("if", "condition", pc)

	// Test with else tag
	err := block.UnknownTag("else", "", nil)
	if err == nil {
		t.Fatal("Expected error for else tag")
	}
	if _, ok := err.(*SyntaxError); !ok {
		t.Errorf("Expected SyntaxError, got %T", err)
	}

	// Test with invalid delimiter
	err = block.UnknownTag("endunless", "", nil)
	if err == nil {
		t.Fatal("Expected error for invalid delimiter")
	}

	// Test with unknown tag
	err = block.UnknownTag("unknown", "", nil)
	if err == nil {
		t.Fatal("Expected error for unknown tag")
	}
}

// TestBlockRaiseTagNeverClosed tests error when tag is never closed
func TestBlockRaiseTagNeverClosed(t *testing.T) {
	lineNum := 1
	pc := &mockParseContextForTag{lineNum: &lineNum, env: NewEnvironment()}

	block := NewBlock("if", "condition", pc)
	err := block.RaiseTagNeverClosed()
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if _, ok := err.(*SyntaxError); !ok {
		t.Errorf("Expected SyntaxError, got %T", err)
	}
	if !strings.Contains(err.Error(), "was never closed") {
		t.Errorf("Expected 'was never closed' in error message, got %q", err.Error())
	}
}

// TestBlockParseBody tests parseBody with various scenarios
func TestBlockParseBody(t *testing.T) {
	lineNum := 1
	pc := NewParseContext(ParseContextOptions{})
	pc.SetLineNumber(&lineNum)

	block := NewBlock("if", "condition", pc)
	block.body = NewBlockBody()

	// Test with end tag
	source := "content{% endif %}"
	tokenizer := NewTokenizer(source, nil, false, nil, false)
	shouldContinue, err := block.parseBody(tokenizer)
	if err != nil {
		t.Fatalf("parseBody() error = %v", err)
	}
	if shouldContinue {
		t.Error("Expected shouldContinue to be false after finding end tag")
	}

	// Test with depth limit
	pc2 := NewParseContext(ParseContextOptions{})
	pc2.SetLineNumber(&lineNum)
	for i := 0; i < blockMaxDepth; i++ {
		pc2.IncrementDepth()
	}
	block2 := NewBlock("if", "condition", pc2)
	block2.body = NewBlockBody()
	source2 := "content"
	tokenizer2 := NewTokenizer(source2, nil, false, nil, false)
	_, err = block2.parseBody(tokenizer2)
	if err == nil {
		t.Error("Expected error for depth limit exceeded")
	}
	if _, ok := err.(*StackLevelError); !ok {
		t.Errorf("Expected StackLevelError, got %T", err)
	}
}

// TestBlockNesting tests nested blocks and depth checking
func TestBlockNesting(t *testing.T) {
	lineNum := 1
	pc := NewParseContext(ParseContextOptions{})
	pc.SetLineNumber(&lineNum)

	// Test nested if blocks
	source := "outer{% if inner %}{% endif %}inner{% endif %}"
	tokenizer := NewTokenizer(source, nil, false, nil, false)
	block := NewBlock("if", "condition", pc)
	err := block.Parse(tokenizer)
	if err != nil {
		t.Fatalf("Parse() with nested blocks error = %v", err)
	}

	// Verify body was parsed
	if block.body == nil {
		t.Error("Expected body to be set")
	}
}
