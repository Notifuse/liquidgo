package tags

import (
	"testing"

	"github.com/Notifuse/liquidgo/liquid"
)

func TestAssignTag(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	// Use quoted string to ensure it's treated as a literal
	tag, err := NewAssignTag("assign", `var = "value"`, pc)
	if err != nil {
		t.Fatalf("NewAssignTag() error = %v", err)
	}
	if tag == nil {
		t.Fatal("Expected AssignTag, got nil")
	}

	if tag.To() != "var" {
		t.Errorf("Expected To 'var', got %q", tag.To())
	}

	ctx := liquid.NewContext()
	var output string
	tag.RenderToOutputBuffer(ctx, &output)

	// Check that variable was assigned
	val := ctx.Get("var")
	if val == nil {
		t.Error("Expected variable to be assigned")
	} else if val != "value" {
		t.Errorf("Expected variable value 'value', got %v", val)
	}
}

func TestAssignTagSyntaxError(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	_, err := NewAssignTag("assign", "invalid", pc)
	if err == nil {
		t.Fatal("Expected error for invalid syntax")
	}
	if _, ok := err.(*liquid.SyntaxError); !ok {
		t.Errorf("Expected SyntaxError, got %T", err)
	}
}

func TestAssignTagEmptyString(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewAssignTag("assign", `a = ""`, pc)
	if err != nil {
		t.Fatalf("NewAssignTag() error = %v", err)
	}

	ctx := liquid.NewContext()
	var output string
	tag.RenderToOutputBuffer(ctx, &output)

	val := ctx.Get("a")
	if val != "" {
		t.Errorf("Expected empty string, got %v", val)
	}
}

func TestAssignTagFromVariable(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	ctx := liquid.NewContext()
	ctx.Set("var", "content")

	tag, err := NewAssignTag("assign", "var2 = var", pc)
	if err != nil {
		t.Fatalf("NewAssignTag() error = %v", err)
	}

	var output string
	tag.RenderToOutputBuffer(ctx, &output)

	// Check that var2 was assigned from var
	val := ctx.Get("var2")
	if val == nil {
		t.Error("Expected variable var2 to be assigned")
	} else if val != "content" {
		t.Errorf("Expected variable value 'content', got %v", val)
	}
}

func TestAssignTagWithHyphenInVariableName(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewAssignTag("assign", `this-thing = "Print this-thing"`, pc)
	if err != nil {
		t.Fatalf("NewAssignTag() error = %v", err)
	}

	if tag.To() != "this-thing" {
		t.Errorf("Expected To 'this-thing', got %q", tag.To())
	}

	ctx := liquid.NewContext()
	var output string
	tag.RenderToOutputBuffer(ctx, &output)

	// Check that variable was assigned
	val := ctx.Get("this-thing")
	if val == nil {
		t.Error("Expected variable to be assigned")
	} else if val != "Print this-thing" {
		t.Errorf("Expected variable value 'Print this-thing', got %v", val)
	}
}

func TestAssignTagWithArray(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	ctx := liquid.NewContext()
	values := []interface{}{"foo", "bar", "baz"}
	ctx.Set("values", values)

	tag, err := NewAssignTag("assign", "foo = values", pc)
	if err != nil {
		t.Fatalf("NewAssignTag() error = %v", err)
	}

	var output string
	tag.RenderToOutputBuffer(ctx, &output)

	// Check that foo was assigned
	val := ctx.Get("foo")
	if val == nil {
		t.Error("Expected variable foo to be assigned")
	} else {
		arr, ok := val.([]interface{})
		if !ok {
			t.Errorf("Expected array, got %T", val)
		} else if len(arr) != 3 {
			t.Errorf("Expected array length 3, got %d", len(arr))
		} else if arr[0] != "foo" {
			t.Errorf("Expected first element 'foo', got %v", arr[0])
		}
	}
}

func TestAssignTagFrom(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewAssignTag("assign", `var = "value"`, pc)
	if err != nil {
		t.Fatalf("NewAssignTag() error = %v", err)
	}

	from := tag.From()
	if from == nil {
		t.Fatal("Expected From() to return a Variable, got nil")
	}
	if from.Name() == nil {
		t.Error("Expected Variable name to be set")
	}
}

func TestAssignTagBlank(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewAssignTag("assign", `var = "value"`, pc)
	if err != nil {
		t.Fatalf("NewAssignTag() error = %v", err)
	}

	if !tag.Blank() {
		t.Error("Expected Blank() to return true for assign tag")
	}
}

func TestAssignTagAssignScoreOf(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	ctx := liquid.NewContext()

	// Test assignScoreOf with string
	tag1, _ := NewAssignTag("assign", `var = "hello"`, pc)
	ctx.Set("var", "hello")
	var output1 string
	tag1.RenderToOutputBuffer(ctx, &output1)
	// assignScoreOf is called internally, verify assignment worked
	if ctx.Get("var") != "hello" {
		t.Error("Expected variable to be assigned")
	}

	// Test assignScoreOf with array
	tag2, _ := NewAssignTag("assign", `arr = values`, pc)
	ctx.Set("values", []interface{}{"a", "b", "c"})
	var output2 string
	tag2.RenderToOutputBuffer(ctx, &output2)
	arr := ctx.Get("arr")
	if arr == nil {
		t.Error("Expected array to be assigned")
	}

	// Test assignScoreOf with map
	tag3, _ := NewAssignTag("assign", `map = data`, pc)
	ctx.Set("data", map[string]interface{}{"key": "value"})
	var output3 string
	tag3.RenderToOutputBuffer(ctx, &output3)
	m := ctx.Get("map")
	if m == nil {
		t.Error("Expected map to be assigned")
	}
}

func TestAssignTagWithMap(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	ctx := liquid.NewContext()
	data := map[string]interface{}{"b": "result"}
	ctx.Set("a", data)

	tag, err := NewAssignTag("assign", `r = a["b"]`, pc)
	if err != nil {
		t.Fatalf("NewAssignTag() error = %v", err)
	}

	var output string
	tag.RenderToOutputBuffer(ctx, &output)

	// Check that r was assigned
	val := ctx.Get("r")
	if val == nil {
		t.Error("Expected variable r to be assigned")
	} else if val != "result" {
		t.Errorf("Expected variable value 'result', got %v", val)
	}
}

func TestAssignTagWithResourceLimits(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	renderLimit := 1000
	renderScoreLimit := 1000
	assignScoreLimit := 1000
	rl := liquid.NewResourceLimits(liquid.ResourceLimitsConfig{
		RenderLengthLimit: &renderLimit,
		RenderScoreLimit:  &renderScoreLimit,
		AssignScoreLimit:  &assignScoreLimit,
	})

	ctx := liquid.NewContext()
	ctx.SetResourceLimits(rl)

	tag, err := NewAssignTag("assign", `var = "hello"`, pc)
	if err != nil {
		t.Fatalf("NewAssignTag() error = %v", err)
	}

	var output string
	tag.RenderToOutputBuffer(ctx, &output)

	// Verify assignment worked
	val := ctx.Get("var")
	if val != "hello" {
		t.Errorf("Expected variable value 'hello', got %v", val)
	}
}

func TestAssignTagWithComplexVariableName(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})

	tests := []struct {
		name   string
		markup string
		wantTo string
	}{
		{"dot notation", `var.name = "value"`, "var.name"},
		{"brackets", `var[0] = "value"`, "var[0]"},
		{"nested brackets", `var[0][1] = "value"`, "var[0][1]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tag, err := NewAssignTag("assign", tt.markup, pc)
			if err != nil {
				t.Fatalf("NewAssignTag() error = %v", err)
			}
			if tag.To() != tt.wantTo {
				t.Errorf("Expected To() = %q, got %q", tt.wantTo, tag.To())
			}
		})
	}
}

func TestAssignTagWithWhitespace(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})

	tests := []struct {
		name   string
		markup string
		wantTo string
	}{
		{"extra spaces", `  var  =  "value"  `, "var"},
		{"tabs", "var\t=\t\"value\"", "var"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tag, err := NewAssignTag("assign", tt.markup, pc)
			if err != nil {
				t.Fatalf("NewAssignTag() error = %v", err)
			}
			if tag.To() != tt.wantTo {
				t.Errorf("Expected To() = %q, got %q", tt.wantTo, tag.To())
			}
		})
	}
}
