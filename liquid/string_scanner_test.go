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

// TestStringScannerEdgeCases tests boundary conditions
func TestStringScannerEdgeCases(t *testing.T) {
	// Test Scan at EOS
	ss := NewStringScanner("hello")
	ss.SetPos(5)
	pattern := regexp.MustCompile(`world`)
	match := ss.Scan(pattern)
	if match != "" {
		t.Errorf("Expected empty match at EOS, got '%s'", match)
	}

	// Test Scan with pattern not at start
	ss2 := NewStringScanner("hello world")
	pattern2 := regexp.MustCompile(`world`)
	match2 := ss2.Scan(pattern2)
	if match2 != "" {
		t.Errorf("Expected empty match when pattern not at start, got '%s'", match2)
	}

	// Test Skip at EOS
	ss3 := NewStringScanner("hello")
	ss3.SetPos(5)
	skipped := ss3.Skip(pattern)
	if skipped != 0 {
		t.Errorf("Expected 0 skipped at EOS, got %d", skipped)
	}

	// Test Skip with pattern not at start
	ss4 := NewStringScanner("hello world")
	skipped2 := ss4.Skip(pattern2)
	if skipped2 != 0 {
		t.Errorf("Expected 0 skipped when pattern not at start, got %d", skipped2)
	}

	// Test SkipUntil at EOS
	ss5 := NewStringScanner("hello")
	ss5.SetPos(5)
	skipped3 := ss5.SkipUntil(pattern)
	if skipped3 != 0 {
		t.Errorf("Expected 0 skipped at EOS, got %d", skipped3)
	}

	// Test SkipUntil when pattern not found
	ss6 := NewStringScanner("hello")
	pattern3 := regexp.MustCompile(`xyz`)
	skipped4 := ss6.SkipUntil(pattern3)
	if skipped4 != 0 {
		t.Errorf("Expected 0 when pattern not found, got %d", skipped4)
	}
	if !ss6.EOS() {
		t.Error("Expected to be at EOS when pattern not found")
	}

	// Test Rest at EOS
	ss7 := NewStringScanner("hello")
	ss7.SetPos(5)
	rest := ss7.Rest()
	if rest != "" {
		t.Errorf("Expected empty rest at EOS, got '%s'", rest)
	}

	// Test Rest beyond EOS
	ss7b := NewStringScanner("hello")
	ss7b.SetPos(10)
	rest2 := ss7b.Rest()
	if rest2 != "" {
		t.Errorf("Expected empty rest beyond EOS, got '%s'", rest2)
	}

	// Test Getch at EOS
	ss8 := NewStringScanner("hello")
	ss8.SetPos(5)
	ch := ss8.Getch()
	if ch != "" {
		t.Errorf("Expected empty char at EOS, got '%s'", ch)
	}

	// Test Getch with UTF-8 multi-byte character
	ss9 := NewStringScanner("hello 世界")
	ss9.SetPos(6)
	ch2 := ss9.Getch()
	if ch2 != "世" {
		t.Errorf("Expected '世', got '%s'", ch2)
	}
	ch3 := ss9.Getch()
	if ch3 != "界" {
		t.Errorf("Expected '界', got '%s'", ch3)
	}

	// Test PeekByte at EOS
	ss10 := NewStringScanner("hello")
	ss10.SetPos(5)
	b := ss10.PeekByte()
	if b != 0 {
		t.Errorf("Expected 0 at EOS, got %d", b)
	}

	// Test ScanByte at EOS
	ss11 := NewStringScanner("hello")
	ss11.SetPos(5)
	b2 := ss11.ScanByte()
	if b2 != 0 {
		t.Errorf("Expected 0 at EOS, got %d", b2)
	}
	if ss11.Pos() != 5 {
		t.Errorf("Expected pos to stay at 5, got %d", ss11.Pos())
	}

	// Test PeekByte beyond EOS
	ss11b := NewStringScanner("hello")
	ss11b.SetPos(10)
	b2b := ss11b.PeekByte()
	if b2b != 0 {
		t.Errorf("Expected 0 beyond EOS, got %d", b2b)
	}

	// Test ScanByte beyond EOS
	ss11c := NewStringScanner("hello")
	ss11c.SetPos(10)
	b2c := ss11c.ScanByte()
	if b2c != 0 {
		t.Errorf("Expected 0 beyond EOS, got %d", b2c)
	}
}

// TestStringScannerBytesliceEdgeCases tests Byteslice boundary conditions
func TestStringScannerBytesliceEdgeCases(t *testing.T) {
	ss := NewStringScanner("hello world")

	// Test with negative start
	slice := ss.Byteslice(-1, 5)
	if slice != "" {
		t.Errorf("Expected empty slice for negative start, got '%s'", slice)
	}

	// Test with start beyond bounds
	slice2 := ss.Byteslice(20, 5)
	if slice2 != "" {
		t.Errorf("Expected empty slice for start beyond bounds, got '%s'", slice2)
	}

	// Test with start at bounds
	slice3 := ss.Byteslice(11, 5)
	if slice3 != "" {
		t.Errorf("Expected empty slice for start at bounds, got '%s'", slice3)
	}

	// Test with length extending beyond bounds
	slice4 := ss.Byteslice(8, 10)
	if slice4 != "rld" {
		t.Errorf("Expected 'rld' when length extends beyond, got '%s'", slice4)
	}

	// Test with zero length
	slice5 := ss.Byteslice(0, 0)
	if slice5 != "" {
		t.Errorf("Expected empty slice for zero length, got '%s'", slice5)
	}
}

// TestStringScannerRuneAtEdgeCases tests runeAt boundary conditions
func TestStringScannerRuneAtEdgeCases(t *testing.T) {
	// Test UTF-8 characters
	s := "hello 世界"
	r, size := runeAt(s, 6)
	if r != '世' || size == 0 {
		t.Errorf("Expected '世' with non-zero size, got %c with size %d", r, size)
	}

	// Test at end of string
	r2, size2 := runeAt(s, len(s))
	if r2 != 0 || size2 != 0 {
		t.Errorf("Expected 0 rune and 0 size at EOS, got %c with size %d", r2, size2)
	}

	// Test beyond end of string
	r3, size3 := runeAt(s, len(s)+5)
	if r3 != 0 || size3 != 0 {
		t.Errorf("Expected 0 rune and 0 size beyond EOS, got %c with size %d", r3, size3)
	}
}
