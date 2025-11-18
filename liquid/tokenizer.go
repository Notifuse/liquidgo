package liquid

import (
	"regexp"
	"strings"
)

var (
	tokenizerTagEnd             = regexp.MustCompile(`%\}`)
	tokenizerTagOrVariableStart = regexp.MustCompile(`\{[\{\%]`)
	tokenizerNewline            = regexp.MustCompile(`\n`)
)

const (
	openCurley  = '{'
	closeCurley = '}'
	percentage  = '%'
)

// Tokenizer tokenizes template source into text, tags, and variables.
type Tokenizer struct {
	source       string
	offset       int
	tokens       []string
	lineNumber   *int
	forLiquidTag bool
	ss           *StringScanner
}

// NewTokenizer creates a new tokenizer.
func NewTokenizer(source string, stringScanner *StringScanner, lineNumbers bool, startLineNumber *int, forLiquidTag bool) *Tokenizer {
	t := &Tokenizer{
		source:       source,
		offset:       0,
		tokens:       []string{},
		forLiquidTag: forLiquidTag,
		ss:           stringScanner,
	}

	if startLineNumber != nil {
		t.lineNumber = startLineNumber
	} else if lineNumbers {
		one := 1
		t.lineNumber = &one
	}

	if source != "" {
		if t.ss == nil {
			t.ss = NewStringScanner(source)
		} else {
			t.ss.SetString(source)
		}
		t.tokenize()
	}

	return t
}

// Shift returns the next token and advances the offset.
func (t *Tokenizer) Shift() string {
	if t.offset >= len(t.tokens) {
		return ""
	}

	token := t.tokens[t.offset]
	t.offset++

	if t.lineNumber != nil {
		if t.forLiquidTag {
			*t.lineNumber++
		} else {
			*t.lineNumber += strings.Count(token, "\n")
		}
	}

	return token
}

// LineNumber returns the current line number.
func (t *Tokenizer) LineNumber() *int {
	return t.lineNumber
}

// ForLiquidTag returns whether this tokenizer is for a liquid tag.
func (t *Tokenizer) ForLiquidTag() bool {
	return t.forLiquidTag
}

func (t *Tokenizer) tokenize() {
	if t.forLiquidTag {
		t.tokens = strings.Split(t.source, "\n")
	} else {
		for !t.ss.EOS() {
			token := t.shiftNormal()
			if token == "" {
				// If we get an empty token but we're not at EOS, there might be remaining text
				if !t.ss.EOS() {
					// Get remaining text
					rest := t.ss.Rest()
					if rest != "" {
						t.tokens = append(t.tokens, rest)
						t.ss.Terminate()
					}
				}
				break
			}
			t.tokens = append(t.tokens, token)
		}
	}

	t.source = ""
	t.ss = nil
}

func (t *Tokenizer) shiftNormal() string {
	if t.ss.EOS() {
		return ""
	}
	token := t.nextToken()
	return token
}

func (t *Tokenizer) nextToken() string {
	byteA := t.ss.PeekByte()

	if byteA == openCurley {
		t.ss.ScanByte()

		byteB := t.ss.PeekByte()

		if byteB == percentage {
			t.ss.ScanByte()
			return t.nextTagToken()
		} else if byteB == openCurley {
			t.ss.ScanByte()
			return t.nextVariableToken()
		}

		t.ss.SetPos(t.ss.Pos() - 1)
	}

	return t.nextTextToken()
}

func (t *Tokenizer) nextTextToken() string {
	start := t.ss.Pos()

	// Save rest before SkipUntil (in case there's no match)
	restBeforeSkip := t.ss.Rest()

	skipLen := t.ss.SkipUntil(tokenizerTagOrVariableStart)
	if skipLen == 0 {
		// No match found, return the rest we saved
		t.ss.Terminate()
		return restBeforeSkip
	}

	// Back up 2 characters to get the position before the match
	// (SkipUntil advances to after the match)
	// This also sets the scanner position back so nextToken can detect the tag/variable
	pos := t.ss.Pos() - 2
	if pos < start {
		pos = start
	}
	t.ss.SetPos(pos)
	return t.ss.Byteslice(start, pos-start)
}

func (t *Tokenizer) nextVariableToken() string {
	start := t.ss.Pos() - 2

	byteA := t.ss.ScanByte()
	byteB := byteA

	for byteB != 0 {
		// Scan until we find a closing brace or opening brace
		for byteA != 0 && byteA != closeCurley && byteA != openCurley {
			byteA = t.ss.ScanByte()
		}

		if byteA == 0 {
			break
		}

		if t.ss.EOS() {
			if byteA == closeCurley {
				return t.ss.Byteslice(start, t.ss.Pos()-start)
			}
			return "{{"
		}

		byteB = t.ss.ScanByte()

		if byteA == closeCurley {
			if byteB == closeCurley {
				return t.ss.Byteslice(start, t.ss.Pos()-start)
			} else {
				// Not a closing brace, back up
				t.ss.SetPos(t.ss.Pos() - 1)
				return t.ss.Byteslice(start, t.ss.Pos()-start)
			}
		} else if byteA == openCurley && byteB == percentage {
			return t.nextTagTokenWithStart(start)
		}

		byteA = byteB
	}

	return "{{"
}

func (t *Tokenizer) nextTagToken() string {
	start := t.ss.Pos() - 2
	if len := t.ss.SkipUntil(tokenizerTagEnd); len > 0 {
		return t.ss.Byteslice(start, len+2)
	}
	return "{%"
}

func (t *Tokenizer) nextTagTokenWithStart(start int) string {
	t.ss.SkipUntil(tokenizerTagEnd)
	return t.ss.Byteslice(start, t.ss.Pos()-start)
}
