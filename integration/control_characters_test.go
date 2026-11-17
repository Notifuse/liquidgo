package integration

import (
	gocontext "context"
	"errors"
	"strings"
	"testing"

	"github.com/Notifuse/liquidgo/v5/liquid"
	"github.com/Notifuse/liquidgo/v5/liquid/tags"
)

// Security category 8, invalid input: control characters and invalid UTF-8,
// in a template and in its data. The control characters inside tags are
// compared with Shopify Liquid in the differential corpus (control-01 to 05).

// A template that is not valid UTF-8 is refused while parsing, as Ruby's
// TemplateEncodingError.
func TestATemplateOfInvalidUTF8IsRefused(t *testing.T) {
	_, err := liquid.ParseTemplate("a\xffb{{ x }}", nil)
	var encoding *liquid.TemplateEncodingError
	if !errors.As(err, &encoding) {
		t.Errorf("got %v (%T), want a TemplateEncodingError", err, err)
	}
}

// Control characters in the text of a template are written as they are.
func TestControlCharactersInTheTextAreKept(t *testing.T) {
	text := "\x00\x01\x1b[31m\x7f\u2028\ufeff\r\n"
	env := mustEnv(liquid.EnvironmentConfig{Tags: tags.Standard()})
	tmpl, err := liquid.ParseTemplate(text+"{{ x }}"+text, &liquid.TemplateOptions{Environment: env})
	if err != nil {
		t.Fatal(err)
	}
	res, err := tmpl.Render(gocontext.Background(), map[string]interface{}{"x": "X"})
	if err != nil || res.Output != text+"X"+text {
		t.Errorf("got %q (%v), want %q", res.Output, err, text+"X"+text)
	}
}

// Data with invalid UTF-8 and control characters goes through every standard
// filter without a panic, and the text after it is written.
func TestInvalidDataGoesThroughEveryFilter(t *testing.T) {
	env := mustEnv(liquid.EnvironmentConfig{Tags: tags.Standard()})
	data := map[string]interface{}{"s": "a\xff\x00\x1bb\u2028c", "n": "\xff1", "list": []interface{}{"\xff", "\x00"}}
	for _, name := range env.FilterNames() {
		src := "[{{ s | " + name + " }}][{{ n | " + name + ": 1 }}][{{ list | " + name + " }}]end"
		tmpl, err := liquid.ParseTemplate(src, &liquid.TemplateOptions{Environment: env})
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		res, err := tmpl.Render(gocontext.Background(), data)
		if err != nil || !strings.HasSuffix(res.Output, "end") {
			t.Errorf("%s: %q, %v", name, res.Output, err)
		}
	}
}
