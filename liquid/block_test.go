package liquid

import (
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

