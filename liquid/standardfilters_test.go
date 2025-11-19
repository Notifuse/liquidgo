package liquid

import (
	"strings"
	"testing"
)

func TestStandardFiltersSize(t *testing.T) {
	sf := &StandardFilters{}

	tests := []struct {
		name  string
		input interface{}
		want  interface{}
	}{
		{"string", "hello", 5},
		{"array", []interface{}{1, 2, 3}, 3},
		{"nil", nil, 0},
		{"map", map[string]interface{}{"a": 1, "b": 2}, 2},
		{"empty string", "", 0},
		{"empty array", []interface{}{}, 0},
		{"empty map", map[string]interface{}{}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sf.Size(tt.input)
			if got != tt.want {
				t.Errorf("Size() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStandardFiltersDowncase(t *testing.T) {
	sf := &StandardFilters{}
	result := sf.Downcase("HELLO")
	if result != "hello" {
		t.Errorf("Downcase() = %v, want hello", result)
	}
}

func TestStandardFiltersUpcase(t *testing.T) {
	sf := &StandardFilters{}
	result := sf.Upcase("hello")
	if result != "HELLO" {
		t.Errorf("Upcase() = %v, want HELLO", result)
	}
}

func TestStandardFiltersCapitalize(t *testing.T) {
	sf := &StandardFilters{}

	tests := []struct {
		name  string
		input interface{}
		want  string
	}{
		{"normal string", "hello world", "Hello world"},
		{"empty string", "", ""},
		{"single char", "h", "H"},
		{"already capitalized", "Hello", "Hello"},
		{"all caps", "HELLO", "Hello"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sf.Capitalize(tt.input)
			if result != tt.want {
				t.Errorf("Capitalize() = %v, want %v", result, tt.want)
			}
		})
	}
}

func TestStandardFiltersH(t *testing.T) {
	sf := &StandardFilters{}
	result := sf.H("<script>alert('xss')</script>")
	if result == "<script>alert('xss')</script>" {
		t.Error("H() should escape HTML")
	}
	// H is alias for Escape, so should escape
	if !strings.Contains(result, "&lt;") {
		t.Errorf("H() should escape <, got %q", result)
	}
}

func TestStandardFiltersEscapeOnce(t *testing.T) {
	sf := &StandardFilters{}
	result := sf.EscapeOnce("<script>")
	if result == "<script>" {
		t.Error("EscapeOnce() should escape HTML")
	}
	if !strings.Contains(result, "&lt;") {
		t.Errorf("EscapeOnce() should escape <, got %q", result)
	}
}

func TestStandardFiltersURLEncode(t *testing.T) {
	sf := &StandardFilters{}
	result := sf.URLEncode("hello world")
	if result != "hello+world" && result != "hello%20world" {
		t.Errorf("URLEncode() = %q, want URL-encoded string", result)
	}

	// Test with nil
	result2 := sf.URLEncode(nil)
	if result2 != "" {
		t.Errorf("URLEncode(nil) = %q, want empty string", result2)
	}
}

func TestStandardFiltersURLDecode(t *testing.T) {
	sf := &StandardFilters{}
	result, err := sf.URLDecode("hello+world")
	if err != nil {
		t.Fatalf("URLDecode() error = %v", err)
	}
	if result != "hello world" {
		t.Errorf("URLDecode() = %q, want 'hello world'", result)
	}

	// Test with nil
	result2, err2 := sf.URLDecode(nil)
	if err2 != nil {
		t.Fatalf("URLDecode(nil) error = %v", err2)
	}
	if result2 != "" {
		t.Errorf("URLDecode(nil) = %q, want empty string", result2)
	}

	// Test with invalid encoding
	_, err3 := sf.URLDecode("%invalid")
	if err3 == nil {
		t.Error("Expected error for invalid URL encoding")
	}
}

func TestStandardFiltersBase64Encode(t *testing.T) {
	sf := &StandardFilters{}
	result := sf.Base64Encode("hello")
	expected := "aGVsbG8="
	if result != expected {
		t.Errorf("Base64Encode() = %q, want %q", result, expected)
	}
}

func TestStandardFiltersBase64Decode(t *testing.T) {
	sf := &StandardFilters{}
	encoded := sf.Base64Encode("hello")
	result, err := sf.Base64Decode(encoded)
	if err != nil {
		t.Fatalf("Base64Decode() error = %v", err)
	}
	if result != "hello" {
		t.Errorf("Base64Decode() = %q, want 'hello'", result)
	}

	// Test with invalid base64
	_, err2 := sf.Base64Decode("invalid!")
	if err2 == nil {
		t.Error("Expected error for invalid base64")
	}
}

func TestStandardFiltersBase64URLSafeEncode(t *testing.T) {
	sf := &StandardFilters{}
	result := sf.Base64URLSafeEncode("hello")
	expected := "aGVsbG8="
	if result != expected {
		t.Errorf("Base64URLSafeEncode() = %q, want %q", result, expected)
	}
}

func TestStandardFiltersBase64URLSafeDecode(t *testing.T) {
	sf := &StandardFilters{}
	encoded := sf.Base64URLSafeEncode("hello")
	result, err := sf.Base64URLSafeDecode(encoded)
	if err != nil {
		t.Fatalf("Base64URLSafeDecode() error = %v", err)
	}
	if result != "hello" {
		t.Errorf("Base64URLSafeDecode() = %q, want 'hello'", result)
	}

	// Test with invalid base64
	_, err2 := sf.Base64URLSafeDecode("invalid!")
	if err2 == nil {
		t.Error("Expected error for invalid base64")
	}
}

func TestStandardFiltersStrip(t *testing.T) {
	sf := &StandardFilters{}
	result := sf.Strip("  hello  ")
	if result != "hello" {
		t.Errorf("Strip() = %q, want 'hello'", result)
	}
}

func TestStandardFiltersLstrip(t *testing.T) {
	sf := &StandardFilters{}
	result := sf.Lstrip("  hello")
	if result != "hello" {
		t.Errorf("Lstrip() = %q, want 'hello'", result)
	}
}

func TestStandardFiltersRstrip(t *testing.T) {
	sf := &StandardFilters{}
	result := sf.Rstrip("hello  ")
	if result != "hello" {
		t.Errorf("Rstrip() = %q, want 'hello'", result)
	}
}

func TestStandardFiltersTruncateWords(t *testing.T) {
	sf := &StandardFilters{}

	t.Run("basic truncate words", func(t *testing.T) {
		result := sf.TruncateWords("hello world test", 2, nil)
		expected := "hello world..."
		if result != expected {
			t.Errorf("TruncateWords() = %q, want %q", result, expected)
		}
	})

	t.Run("nil input", func(t *testing.T) {
		result := sf.TruncateWords(nil, 5, nil)
		if result != "" {
			t.Errorf("TruncateWords(nil) = %v, want empty string", result)
		}
	})

	t.Run("custom truncate string", func(t *testing.T) {
		result := sf.TruncateWords("hello world test", 2, "---")
		if result != "hello world---" {
			t.Errorf("TruncateWords() with custom string = %v, want 'hello world---'", result)
		}
	})

	t.Run("words zero or negative", func(t *testing.T) {
		result := sf.TruncateWords("hello world test", 0, nil)
		if !strings.Contains(result, "hello") {
			t.Errorf("TruncateWords() with words 0 = %v, should contain 'hello'", result)
		}
	})

	t.Run("fewer words than limit", func(t *testing.T) {
		result := sf.TruncateWords("hello world", 5, nil)
		if result != "hello world" {
			t.Errorf("TruncateWords() with fewer words = %v, want 'hello world'", result)
		}
	})

	t.Run("single word", func(t *testing.T) {
		result := sf.TruncateWords("hello", 1, nil)
		if result != "hello" {
			t.Errorf("TruncateWords() with single word = %v, want 'hello'", result)
		}
	})
}

func TestStandardFiltersEscape(t *testing.T) {
	sf := &StandardFilters{}

	tests := []struct {
		name  string
		input interface{}
	}{
		{"html tags", "<script>alert('xss')</script>"},
		{"nil input", nil},
		{"empty string", ""},
		{"special chars", "<>&\"'"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sf.Escape(tt.input)
			if tt.input != nil {
				// Empty string doesn't need escaping, so result can equal input
				if tt.input != "" && result == tt.input {
					t.Errorf("Escape() should escape HTML, got %q", result)
				}
				// For empty string, result should also be empty
				if tt.input == "" && result != "" {
					t.Errorf("Escape(\"\") = %q, want empty string", result)
				}
			}
			if tt.input == nil && result != "" {
				t.Errorf("Escape(nil) = %q, want empty string", result)
			}
		})
	}
}

func TestStandardFiltersSlice(t *testing.T) {
	sf := &StandardFilters{}

	t.Run("string", func(t *testing.T) {
		got := sf.Slice("hello", 1, 3)
		if got != "ell" {
			t.Errorf("Slice() = %v, want 'ell'", got)
		}
	})

	t.Run("array", func(t *testing.T) {
		got := sf.Slice([]interface{}{1, 2, 3, 4}, 1, 2)
		gotArr, ok := got.([]interface{})
		if !ok {
			t.Fatalf("Slice() = %T, want []interface{}", got)
		}
		if len(gotArr) != 2 {
			t.Errorf("Slice() returned %d elements, want 2", len(gotArr))
		}
		if gotArr[0] != 2 || gotArr[1] != 3 {
			t.Errorf("Slice() = %v, want [2 3]", gotArr)
		}
	})

	t.Run("string negative offset", func(t *testing.T) {
		got := sf.Slice("hello", -1, 3)
		if got != "" {
			t.Errorf("Slice() with negative offset = %v, want empty string", got)
		}
	})

	t.Run("string offset too large", func(t *testing.T) {
		got := sf.Slice("hello", 10, 3)
		if got != "" {
			t.Errorf("Slice() with offset too large = %v, want empty string", got)
		}
	})

	t.Run("string end exceeds length", func(t *testing.T) {
		got := sf.Slice("hello", 1, 10)
		if got != "ello" {
			t.Errorf("Slice() with end exceeding length = %v, want 'ello'", got)
		}
	})

	t.Run("array negative offset", func(t *testing.T) {
		got := sf.Slice([]interface{}{1, 2, 3}, -1, 2)
		gotArr, ok := got.([]interface{})
		if !ok || len(gotArr) != 0 {
			t.Errorf("Slice() with negative offset = %v, want empty array", got)
		}
	})

	t.Run("array offset too large", func(t *testing.T) {
		got := sf.Slice([]interface{}{1, 2, 3}, 10, 2)
		gotArr, ok := got.([]interface{})
		if !ok || len(gotArr) != 0 {
			t.Errorf("Slice() with offset too large = %v, want empty array", got)
		}
	})

	t.Run("array end exceeds length", func(t *testing.T) {
		got := sf.Slice([]interface{}{1, 2, 3}, 1, 10)
		gotArr, ok := got.([]interface{})
		if !ok || len(gotArr) != 2 {
			t.Errorf("Slice() with end exceeding length = %v, want [2 3]", gotArr)
		}
	})

	t.Run("non-string non-array", func(t *testing.T) {
		got := sf.Slice(42, 1, 2)
		if got != "" {
			t.Errorf("Slice() with non-string non-array = %v, want empty string", got)
		}
	})

	t.Run("nil length", func(t *testing.T) {
		got := sf.Slice("hello", 1, nil)
		if got != "e" {
			t.Errorf("Slice() with nil length = %v, want 'e'", got)
		}
	})
}

func TestStandardFiltersTruncate(t *testing.T) {
	sf := &StandardFilters{}

	t.Run("basic truncate", func(t *testing.T) {
		result := sf.Truncate("hello world", 5, nil)
		// "hello world" truncated to 5 chars: 5 - 3 (for "...") = 2 chars + "..." = "he..."
		if len(result) != 5 {
			t.Errorf("Truncate() = %v (len=%d), expected length 5", result, len(result))
		}
		if result != "he..." {
			t.Errorf("Truncate() = %v, expected 'he...'", result)
		}
	})

	t.Run("nil input", func(t *testing.T) {
		result := sf.Truncate(nil, 10, nil)
		if result != "" {
			t.Errorf("Truncate(nil) = %v, want empty string", result)
		}
	})

	t.Run("custom truncate string", func(t *testing.T) {
		result := sf.Truncate("hello world", 8, "---")
		if result != "hello---" {
			t.Errorf("Truncate() with custom string = %v, want 'hello---'", result)
		}
	})

	t.Run("length zero or negative", func(t *testing.T) {
		// With length 0, it defaults to 50, but "hello world" is shorter, so returns full string
		result := sf.Truncate("hello world", 0, nil)
		if result != "hello world" {
			t.Errorf("Truncate() with length 0 = %v, want 'hello world' (string shorter than default 50)", result)
		}

		// Test with a longer string to see default length behavior
		longStr := strings.Repeat("a", 100)
		result2 := sf.Truncate(longStr, 0, nil)
		if len(result2) != 50 {
			t.Errorf("Truncate() with length 0 and long string = %v (len=%d), expected length 50", result2, len(result2))
		}
	})

	t.Run("string shorter than length", func(t *testing.T) {
		result := sf.Truncate("hi", 10, nil)
		if result != "hi" {
			t.Errorf("Truncate() with short string = %v, want 'hi'", result)
		}
	})

	t.Run("truncate string longer than length", func(t *testing.T) {
		// When truncate string is longer than length, l becomes 0, so result is just truncate string
		result := sf.Truncate("hello", 3, "very long truncate string")
		if result != "very long truncate string" {
			t.Errorf("Truncate() with long truncate string = %v, want 'very long truncate string'", result)
		}
	})
}

func TestStandardFiltersSplit(t *testing.T) {
	sf := &StandardFilters{}
	result := sf.Split("a,b,c", ",")
	if len(result) != 3 {
		t.Errorf("Split() returned %d elements, want 3", len(result))
	}
}

func TestStandardFiltersJoin(t *testing.T) {
	sf := &StandardFilters{}

	t.Run("array with separator", func(t *testing.T) {
		result := sf.Join([]interface{}{"a", "b", "c"}, ",")
		if result != "a,b,c" {
			t.Errorf("Join() = %v, want 'a,b,c'", result)
		}
	})

	t.Run("array with default separator", func(t *testing.T) {
		result := sf.Join([]interface{}{"a", "b", "c"}, nil)
		if result != "a, b, c" {
			t.Errorf("Join() with nil separator = %v, want 'a, b, c'", result)
		}
	})

	t.Run("non-array input", func(t *testing.T) {
		result := sf.Join("not an array", ",")
		if result != "not an array" {
			t.Errorf("Join() with non-array = %v, want 'not an array'", result)
		}
	})

	t.Run("empty array", func(t *testing.T) {
		result := sf.Join([]interface{}{}, ",")
		if result != "" {
			t.Errorf("Join() with empty array = %v, want empty string", result)
		}
	})

	t.Run("array with numbers", func(t *testing.T) {
		result := sf.Join([]interface{}{1, 2, 3}, "-")
		if result != "1-2-3" {
			t.Errorf("Join() with numbers = %v, want '1-2-3'", result)
		}
	})
}

func TestStandardFiltersFirst(t *testing.T) {
	sf := &StandardFilters{}

	t.Run("non-empty array", func(t *testing.T) {
		result := sf.First([]interface{}{1, 2, 3})
		if result != 1 {
			t.Errorf("First() = %v, want 1", result)
		}
	})

	t.Run("empty array", func(t *testing.T) {
		result := sf.First([]interface{}{})
		if result != nil {
			t.Errorf("First() with empty array = %v, want nil", result)
		}
	})

	t.Run("non-array", func(t *testing.T) {
		result := sf.First("not an array")
		if result != nil {
			t.Errorf("First() with non-array = %v, want nil", result)
		}
	})

	t.Run("nil input", func(t *testing.T) {
		result := sf.First(nil)
		if result != nil {
			t.Errorf("First(nil) = %v, want nil", result)
		}
	})
}

func TestStandardFiltersLast(t *testing.T) {
	sf := &StandardFilters{}

	t.Run("non-empty array", func(t *testing.T) {
		result := sf.Last([]interface{}{1, 2, 3})
		if result != 3 {
			t.Errorf("Last() = %v, want 3", result)
		}
	})

	t.Run("empty array", func(t *testing.T) {
		result := sf.Last([]interface{}{})
		if result != nil {
			t.Errorf("Last() with empty array = %v, want nil", result)
		}
	})

	t.Run("non-array", func(t *testing.T) {
		result := sf.Last("not an array")
		if result != nil {
			t.Errorf("Last() with non-array = %v, want nil", result)
		}
	})

	t.Run("nil input", func(t *testing.T) {
		result := sf.Last(nil)
		if result != nil {
			t.Errorf("Last(nil) = %v, want nil", result)
		}
	})
}

// TestStandardFiltersStripHTML tests HTML stripping functionality
func TestStandardFiltersStripHTML(t *testing.T) {
	sf := &StandardFilters{}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple tags",
			input:    "<p>Hello</p>",
			expected: "Hello",
		},
		{
			name:     "nested tags",
			input:    "<div><p>Hello</p></div>",
			expected: "Hello",
		},
		{
			name:     "with attributes",
			input:    `<p class="test">Hello</p>`,
			expected: "Hello",
		},
		{
			name:     "script tags",
			input:    "<script>alert('xss')</script>Hello",
			expected: "Hello",
		},
		{
			name:     "style tags",
			input:    "<style>body { color: red; }</style>Hello",
			expected: "Hello",
		},
		{
			name:     "comments",
			input:    "<!-- comment -->Hello",
			expected: "Hello",
		},
		{
			name:     "mixed content",
			input:    "<p>Hello</p> <script>alert('xss')</script> <span>World</span>",
			expected: "Hello  World",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sf.StripHTML(tt.input)
			if result != tt.expected {
				t.Errorf("StripHTML() = %q, want %q", result, tt.expected)
			}
		})
	}
}
