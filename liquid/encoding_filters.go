package liquid

import (
	"encoding/base64"
	"net/url"
	"strings"
	"unicode/utf8"
)

// Escape escapes special characters in HTML.
func (sf *StandardFilters) Escape(input interface{}) string {
	if input == nil {
		return ""
	}
	return htmlEscaper.Replace(ToS(input, nil))
}

// htmlEscaper escapes as Ruby's CGI.escapeHTML: the quote as &quot; and the
// apostrophe as &#39; (Go's html.EscapeString writes &#34;).
var htmlEscaper = strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;", "'", "&#39;")

// H is an alias for Escape.
func (sf *StandardFilters) H(input interface{}) string {
	return sf.Escape(input)
}

// EscapeOnce escapes a string without changing characters that have already been escaped.
// As Ruby's /["><']|&(?!([a-zA-Z]+|(#\d+));)/: an & starting a named entity (&amp;) or a
// decimal one (&#39;) is kept; any other & is escaped.
func (sf *StandardFilters) EscapeOnce(input interface{}) string {
	s := ToS(input, nil)
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); i++ {
		switch c := s[i]; c {
		case '"':
			b.WriteString("&quot;")
		case '\'':
			b.WriteString("&#39;")
		case '<':
			b.WriteString("&lt;")
		case '>':
			b.WriteString("&gt;")
		case '&':
			if isEntityAt(s, i+1) {
				b.WriteByte('&')
			} else {
				b.WriteString("&amp;")
			}
		default:
			b.WriteByte(c)
		}
	}
	return b.String()
}

// isEntityAt reports whether s[i:] starts with letters or #digits followed by ";".
func isEntityAt(s string, i int) bool {
	j := i
	if j < len(s) && s[j] == '#' {
		j++
		for j < len(s) && s[j] >= '0' && s[j] <= '9' {
			j++
		}
		return j > i+1 && j < len(s) && s[j] == ';'
	}
	for j < len(s) && ((s[j] >= 'a' && s[j] <= 'z') || (s[j] >= 'A' && s[j] <= 'Z')) {
		j++
	}
	return j > i && j < len(s) && s[j] == ';'
}

// URLEncode converts URL-unsafe characters to percent-encoded equivalent.
func (sf *StandardFilters) URLEncode(input interface{}) string {
	if input == nil {
		return ""
	}
	return url.QueryEscape(ToS(input, nil))
}

// URLDecode decodes percent-encoded characters.
func (sf *StandardFilters) URLDecode(input interface{}) (string, error) {
	if input == nil {
		return "", nil
	}
	decoded := cgiUnescape(ToS(input, nil))
	if !utf8.ValidString(decoded) {
		return "", NewArgumentError("invalid byte sequence in UTF-8")
	}
	return decoded, nil
}

// cgiUnescape decodes as Ruby's CGI.unescape: + is a space and %XX the byte XX;
// a % that two hexadecimal digits do not follow stays as written. Go's
// url.QueryUnescape fails on it, which lost the whole value.
func cgiUnescape(s string) string {
	if !strings.ContainsAny(s, "+%") {
		return s
	}
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); i++ {
		switch c := s[i]; {
		case c == '+':
			b.WriteByte(' ')
		case c == '%' && i+2 < len(s) && isHexDigit(s[i+1]) && isHexDigit(s[i+2]):
			b.WriteByte(unhex(s[i+1])<<4 | unhex(s[i+2]))
			i += 2
		default:
			b.WriteByte(c)
		}
	}
	return b.String()
}

func isHexDigit(c byte) bool {
	return '0' <= c && c <= '9' || 'a' <= c && c <= 'f' || 'A' <= c && c <= 'F'
}

func unhex(c byte) byte {
	switch {
	case c <= '9':
		return c - '0'
	case c <= 'F':
		return c - 'A' + 10
	default:
		return c - 'a' + 10
	}
}

// Base64Encode encodes a string to Base64 format.
func (sf *StandardFilters) Base64Encode(input interface{}) string {
	return base64.StdEncoding.EncodeToString([]byte(ToS(input, nil)))
}

// Base64Decode decodes a string in Base64 format.
func (sf *StandardFilters) Base64Decode(input interface{}) (string, error) {
	s := ToS(input, nil)
	decoded, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return "", NewArgumentError("invalid base64 provided to base64_decode")
	}
	return string(decoded), nil
}

// Base64URLSafeEncode encodes a string to URL-safe Base64 format.
func (sf *StandardFilters) Base64URLSafeEncode(input interface{}) string {
	return base64.URLEncoding.EncodeToString([]byte(ToS(input, nil)))
}

// Base64URLSafeDecode decodes a string in URL-safe Base64 format.
func (sf *StandardFilters) Base64URLSafeDecode(input interface{}) (string, error) {
	s := ToS(input, nil)
	decoded, err := base64.URLEncoding.DecodeString(s)
	if err != nil {
		return "", NewArgumentError("invalid base64 provided to base64_url_safe_decode")
	}
	return string(decoded), nil
}
