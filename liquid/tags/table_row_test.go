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
