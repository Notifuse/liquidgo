package liquid

import (
	"regexp"
)

// StringScanner provides a scanner interface similar to Ruby's StringScanner.
type StringScanner struct {
	source string
	pos    int
}

// NewStringScanner creates a new StringScanner.
func NewStringScanner(source string) *StringScanner {
	return &StringScanner{
		source: source,
		pos:    0,
	}
}

// String returns the source string.
func (s *StringScanner) String() string {
	return s.source
}

// SetString sets the source string and resets position.
func (s *StringScanner) SetString(str string) {
	s.source = str
	s.pos = 0
}

// Pos returns the current position.
func (s *StringScanner) Pos() int {
	return s.pos
}

// SetPos sets the current position.
func (s *StringScanner) SetPos(pos int) {
	s.pos = pos
}

// EOS returns true if we're at the end of the string.
func (s *StringScanner) EOS() bool {
	return s.pos >= len(s.source)
}

// PeekByte returns the byte at the current position without advancing.
func (s *StringScanner) PeekByte() byte {
	if s.pos >= len(s.source) {
		return 0
	}
	return s.source[s.pos]
}

// ScanByte advances and returns the byte at the current position.
func (s *StringScanner) ScanByte() byte {
	if s.pos >= len(s.source) {
		return 0
	}
	b := s.source[s.pos]
	s.pos++
	return b
}

// Scan scans for the given pattern and advances position if matched.
func (s *StringScanner) Scan(pattern *regexp.Regexp) string {
	if s.pos >= len(s.source) {
		return ""
	}
	rest := s.source[s.pos:]
	loc := pattern.FindStringIndex(rest)
	if loc == nil || loc[0] != 0 {
		return ""
	}
	match := rest[loc[0]:loc[1]]
	s.pos += loc[1]
	return match
}

// Skip skips the given pattern.
func (s *StringScanner) Skip(pattern *regexp.Regexp) int {
	if s.pos >= len(s.source) {
		return 0
	}
	rest := s.source[s.pos:]
	loc := pattern.FindStringIndex(rest)
	if loc == nil || loc[0] != 0 {
		return 0
	}
	s.pos += loc[1]
	return loc[1]
}

// SkipUntil skips until the pattern is found.
func (s *StringScanner) SkipUntil(pattern *regexp.Regexp) int {
	if s.pos >= len(s.source) {
		return 0
	}
	rest := s.source[s.pos:]
	loc := pattern.FindStringIndex(rest)
	if loc == nil {
		s.pos = len(s.source)
		return 0
	}
	s.pos += loc[1]
	return loc[1]
}

// Rest returns the rest of the string from current position.
func (s *StringScanner) Rest() string {
	if s.pos >= len(s.source) {
		return ""
	}
	return s.source[s.pos:]
}

// Terminate sets position to end of string.
func (s *StringScanner) Terminate() {
	s.pos = len(s.source)
}

// Getch gets the next character (handles UTF-8).
func (s *StringScanner) Getch() string {
	if s.pos >= len(s.source) {
		return ""
	}
	// Get the next rune
	r, size := runeAt(s.source, s.pos)
	s.pos += size
	return string(r)
}

// Byteslice returns a slice of bytes from start to end.
func (s *StringScanner) Byteslice(start, length int) string {
	if start < 0 || start >= len(s.source) {
		return ""
	}
	end := start + length
	if end > len(s.source) {
		end = len(s.source)
	}
	return s.source[start:end]
}

func runeAt(s string, pos int) (rune, int) {
	if pos >= len(s) {
		return 0, 0
	}
	return []rune(s[pos:])[0], len([]byte(string([]rune(s[pos:])[0])))
}
