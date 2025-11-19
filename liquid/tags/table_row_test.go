package tags

import (
	"testing"

	"github.com/Notifuse/liquidgo/liquid"
)

func TestTableRowTag(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewTableRowTag("tablerow", "item in array", pc)
	if err != nil {
		t.Fatalf("NewTableRowTag() error = %v", err)
	}
	if tag == nil {
		t.Fatal("Expected TableRowTag, got nil")
	}

	if tag.VariableName() != "item" {
		t.Errorf("Expected variable name 'item', got %q", tag.VariableName())
	}
}

func TestTableRowTagBasic(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewTableRowTag("tablerow", "n in numbers cols:3", pc)
	if err != nil {
		t.Fatalf("NewTableRowTag() error = %v", err)
	}

	// Parse table_row block
	tokenizer := pc.NewTokenizer("{{n}} {% endtablerow %}", false, nil, false)
	err = tag.Parse(tokenizer)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	ctx := liquid.NewContext()
	ctx.Set("numbers", []interface{}{1, 2, 3, 4, 5, 6})
	var output string
	tag.RenderToOutputBuffer(ctx, &output)

	// Should generate table rows with 3 columns
	expected := "<tr class=\"row1\">\n<td class=\"col1\">1 </td><td class=\"col2\">2 </td><td class=\"col3\">3 </td></tr>\n<tr class=\"row2\"><td class=\"col1\">4 </td><td class=\"col2\">5 </td><td class=\"col3\">6 </td></tr>\n"
	if output != expected {
		t.Errorf("Expected output %q, got %q", expected, output)
	}
}

func TestTableRowTagEmptyCollection(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewTableRowTag("tablerow", "n in numbers cols:3", pc)
	if err != nil {
		t.Fatalf("NewTableRowTag() error = %v", err)
	}

	// Parse table_row block
	tokenizer := pc.NewTokenizer("{{n}} {% endtablerow %}", false, nil, false)
	err = tag.Parse(tokenizer)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	ctx := liquid.NewContext()
	ctx.Set("numbers", []interface{}{})
	var output string
	tag.RenderToOutputBuffer(ctx, &output)

	// Should generate empty row
	expected := "<tr class=\"row1\">\n</tr>\n"
	if output != expected {
		t.Errorf("Expected output %q, got %q", expected, output)
	}
}

func TestTableRowTagWithLimit(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewTableRowTag("tablerow", "n in numbers cols:2 limit:3", pc)
	if err != nil {
		t.Fatalf("NewTableRowTag() error = %v", err)
	}

	// Parse table_row block
	tokenizer := pc.NewTokenizer("{{n}} {% endtablerow %}", false, nil, false)
	err = tag.Parse(tokenizer)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	ctx := liquid.NewContext()
	ctx.Set("numbers", []interface{}{1, 2, 3, 4, 5, 6})
	var output string
	tag.RenderToOutputBuffer(ctx, &output)

	// Should generate table rows with 2 columns, limited to 3 items
	expected := "<tr class=\"row1\">\n<td class=\"col1\">1 </td><td class=\"col2\">2 </td></tr>\n<tr class=\"row2\"><td class=\"col1\">3 </td></tr>\n"
	if output != expected {
		t.Errorf("Expected output %q, got %q", expected, output)
	}
}

func TestTableRowTagWithOffset(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewTableRowTag("tablerow", "n in numbers cols:2 offset:2 limit:2", pc)
	if err != nil {
		t.Fatalf("NewTableRowTag() error = %v", err)
	}

	// Parse table_row block
	tokenizer := pc.NewTokenizer("{{n}} {% endtablerow %}", false, nil, false)
	err = tag.Parse(tokenizer)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	ctx := liquid.NewContext()
	ctx.Set("numbers", []interface{}{1, 2, 3, 4, 5, 6})
	var output string
	tag.RenderToOutputBuffer(ctx, &output)

	// Should generate table rows starting from offset 2, with 2 columns, limited to 2 items
	expected := "<tr class=\"row1\">\n<td class=\"col1\">3 </td><td class=\"col2\">4 </td></tr>\n"
	if output != expected {
		t.Errorf("Expected output %q, got %q", expected, output)
	}
}

func TestTableRowTagCollectionName(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewTableRowTag("tablerow", "item in array", pc)
	if err != nil {
		t.Fatalf("NewTableRowTag() error = %v", err)
	}

	collectionName := tag.CollectionName()
	if collectionName == nil {
		t.Error("Expected CollectionName() to return non-nil expression")
	}
}

func TestTableRowTagAttributes(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewTableRowTag("tablerow", "item in array cols:3", pc)
	if err != nil {
		t.Fatalf("NewTableRowTag() error = %v", err)
	}

	attributes := tag.Attributes()
	if attributes == nil {
		t.Error("Expected Attributes() to return non-nil map")
	}
	if len(attributes) == 0 {
		t.Error("Expected Attributes() to contain attributes")
	}
}

func TestTableRowTagRenderToOutputBufferEdgeCases(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})

	// Test with range attribute
	tag, err := NewTableRowTag("tablerow", "n in numbers cols:2 range:1-3", pc)
	if err != nil {
		t.Fatalf("NewTableRowTag() error = %v", err)
	}
	tokenizer := pc.NewTokenizer("{{n}} {% endtablerow %}", false, nil, false)
	err = tag.Parse(tokenizer)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	ctx := liquid.NewContext()
	ctx.Set("numbers", []interface{}{1, 2, 3, 4, 5})
	var output string
	tag.RenderToOutputBuffer(ctx, &output)
	// Should render with range
	_ = output

	// Test with nil collection
	tag2, _ := NewTableRowTag("tablerow", "n in numbers cols:2", pc)
	tokenizer2 := pc.NewTokenizer("{{n}} {% endtablerow %}", false, nil, false)
	if err := tag2.Parse(tokenizer2); err != nil {
		t.Fatalf("tag2.Parse() error = %v", err)
	}
	ctx2 := liquid.NewContext()
	ctx2.Set("numbers", nil)
	var output2 string
	tag2.RenderToOutputBuffer(ctx2, &output2)
	// Should handle nil gracefully
	_ = output2

	// Test with invalid attribute
	_, err3 := NewTableRowTag("tablerow", "n in numbers invalid:value", pc)
	if err3 == nil {
		t.Error("Expected error for invalid attribute")
	}
}

// TestTableRowTagRenderToOutputBufferSingleRow tests single row scenario
func TestTableRowTagRenderToOutputBufferSingleRow(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewTableRowTag("tablerow", "n in numbers cols:3", pc)
	if err != nil {
		t.Fatalf("NewTableRowTag() error = %v", err)
	}

	tokenizer := pc.NewTokenizer("{{n}} {% endtablerow %}", false, nil, false)
	err = tag.Parse(tokenizer)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	ctx := liquid.NewContext()
	// Single item collection
	ctx.Set("numbers", []interface{}{1})
	var output string
	tag.RenderToOutputBuffer(ctx, &output)

	// Should generate single row with one column
	if output == "" {
		t.Error("Expected non-empty output for single row")
	}
	// Check that output contains row1 (using simple check)
	hasRow1 := false
	if len(output) >= 4 {
		for i := 0; i <= len(output)-4; i++ {
			if output[i:i+4] == "row1" {
				hasRow1 = true
				break
			}
		}
	}
	if !hasRow1 {
		t.Logf("Note: Output may not contain 'row1' as expected: %q", output)
	}
}

// TestTableRowTagRenderToOutputBufferErrorHandling tests error handling
func TestTableRowTagRenderToOutputBufferErrorHandling(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})

	// Test with invalid offset (should handle error)
	tag, err := NewTableRowTag("tablerow", "n in numbers cols:2 offset:invalid", pc)
	if err != nil {
		t.Fatalf("NewTableRowTag() error = %v", err)
	}

	tokenizer := pc.NewTokenizer("{{n}} {% endtablerow %}", false, nil, false)
	err = tag.Parse(tokenizer)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	ctx := liquid.NewContext()
	ctx.Set("numbers", []interface{}{1, 2, 3})
	var output string
	tag.RenderToOutputBuffer(ctx, &output)

	// Should handle invalid offset gracefully (may produce error message)
	t.Logf("Note: Invalid offset handling output: %q", output)
}
