package liquid

import (
	"regexp"
	"testing"
)

func TestStringScannerBasic(t *testing.T) {
	ss := NewStringScanner("hello world")
	if ss.String() != "hello world" {
		t.Errorf("Expected 'hello world', got '%s'", ss.String())
	}
	if ss.Pos() != 0 {
		t.Errorf("Expected pos 0, got %d", ss.Pos())
	}
	if ss.EOS() {
		t.Error("Expected not at EOS")
	}
}

func TestStringScannerPeekByte(t *testing.T) {
	ss := NewStringScanner("hello")
	if ss.PeekByte() != 'h' {
		t.Errorf("Expected 'h', got %c", ss.PeekByte())
	}
	if ss.Pos() != 0 {
		t.Error("PeekByte should not advance position")
	}
}

func TestStringScannerScanByte(t *testing.T) {
	ss := NewStringScanner("hello")
	b := ss.ScanByte()
	if b != 'h' {
		t.Errorf("Expected 'h', got %c", b)
	}
	if ss.Pos() != 1 {
		t.Errorf("Expected pos 1, got %d", ss.Pos())
	}
}

func TestStringScannerScan(t *testing.T) {
	ss := NewStringScanner("hello world")
	pattern := regexp.MustCompile(`hello`)
	match := ss.Scan(pattern)
	if match != "hello" {
		t.Errorf("Expected 'hello', got '%s'", match)
	}
	if ss.Pos() != 5 {
		t.Errorf("Expected pos 5, got %d", ss.Pos())
	}
}

func TestStringScannerSkip(t *testing.T) {
	ss := NewStringScanner("hello world")
	pattern := regexp.MustCompile(`hello`)
	skipped := ss.Skip(pattern)
	if skipped != 5 {
		t.Errorf("Expected skipped 5, got %d", skipped)
	}
	if ss.Pos() != 5 {
		t.Errorf("Expected pos 5, got %d", ss.Pos())
	}
}

func TestStringScannerSkipUntil(t *testing.T) {
	ss := NewStringScanner("hello world")
	pattern := regexp.MustCompile(`world`)
	skipped := ss.SkipUntil(pattern)
	if skipped == 0 {
		t.Error("Expected to skip some characters")
	}
	if ss.Pos() < 5 {
		t.Errorf("Expected pos >= 5, got %d", ss.Pos())
	}
}

func TestStringScannerRest(t *testing.T) {
	ss := NewStringScanner("hello world")
	ss.SetPos(6)
	rest := ss.Rest()
	if rest != "world" {
		t.Errorf("Expected 'world', got '%s'", rest)
	}
}

func TestStringScannerTerminate(t *testing.T) {
	ss := NewStringScanner("hello world")
	ss.Terminate()
	if !ss.EOS() {
		t.Error("Expected to be at EOS after Terminate()")
	}
}

func TestStringScannerGetch(t *testing.T) {
	ss := NewStringScanner("hello")
	ch := ss.Getch()
	if ch != "h" {
		t.Errorf("Expected 'h', got '%s'", ch)
	}
	if ss.Pos() != 1 {
		t.Errorf("Expected pos 1, got %d", ss.Pos())
	}
}

func TestStringScannerByteslice(t *testing.T) {
	ss := NewStringScanner("hello world")
	slice := ss.Byteslice(0, 5)
	if slice != "hello" {
		t.Errorf("Expected 'hello', got '%s'", slice)
	}
}

func TestStringScannerSetString(t *testing.T) {
	ss := NewStringScanner("hello")
	ss.SetPos(5)
	ss.SetString("world")
	if ss.String() != "world" {
		t.Errorf("Expected 'world', got '%s'", ss.String())
	}
	if ss.Pos() != 0 {
		t.Error("Expected pos to reset to 0")
	}
}

func TestStringScannerEOS(t *testing.T) {
	ss := NewStringScanner("hello")
	ss.SetPos(5)
	if !ss.EOS() {
		t.Error("Expected to be at EOS")
	}
}

func TestStringScannerSetPos(t *testing.T) {
	ss := NewStringScanner("hello world")
	ss.SetPos(6)
	if ss.Pos() != 6 {
		t.Errorf("Expected pos 6, got %d", ss.Pos())
	}
	if ss.PeekByte() != 'w' {
		t.Errorf("Expected 'w' at pos 6, got %c", ss.PeekByte())
	}
}

