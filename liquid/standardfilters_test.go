package liquid

import (
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
	result := sf.Capitalize("hello world")
	if result != "Hello world" {
		t.Errorf("Capitalize() = %v, want 'Hello world'", result)
	}
}

func TestStandardFiltersEscape(t *testing.T) {
	sf := &StandardFilters{}
	result := sf.Escape("<script>alert('xss')</script>")
	if result == "<script>alert('xss')</script>" {
		t.Error("Expected HTML to be escaped")
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
}

func TestStandardFiltersTruncate(t *testing.T) {
	sf := &StandardFilters{}
	result := sf.Truncate("hello world", 5, nil)
	// "hello world" truncated to 5 chars: 5 - 3 (for "...") = 2 chars + "..." = "he..."
	if len(result) != 5 {
		t.Errorf("Truncate() = %v (len=%d), expected length 5", result, len(result))
	}
	if result != "he..." {
		t.Errorf("Truncate() = %v, expected 'he...'", result)
	}
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
	result := sf.Join([]interface{}{"a", "b", "c"}, ",")
	if result != "a,b,c" {
		t.Errorf("Join() = %v, want 'a,b,c'", result)
	}
}

func TestStandardFiltersFirst(t *testing.T) {
	sf := &StandardFilters{}
	result := sf.First([]interface{}{1, 2, 3})
	if result != 1 {
		t.Errorf("First() = %v, want 1", result)
	}
}

func TestStandardFiltersLast(t *testing.T) {
	sf := &StandardFilters{}
	result := sf.Last([]interface{}{1, 2, 3})
	if result != 3 {
		t.Errorf("Last() = %v, want 3", result)
	}
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
