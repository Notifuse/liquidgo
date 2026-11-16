package liquid

import (
	gocontext "context"
	"testing"
)

// The output of a render is its own: the next renders of the template, which
// make their buffer at the size of the last output, leave it as it was.
func TestTheOutputOfARenderIsItsOwn(t *testing.T) {
	tmpl, err := ParseTemplate("{{ x }}{{ x }}", nil)
	if err != nil {
		t.Fatal(err)
	}
	first, err := tmpl.Render(gocontext.Background(), map[string]interface{}{"x": "abc"})
	if err != nil {
		t.Fatal(err)
	}
	for _, x := range []string{"xyz", "a", "a much longer value"} {
		if _, err := tmpl.Render(gocontext.Background(), map[string]interface{}{"x": x}); err != nil {
			t.Fatal(err)
		}
	}
	if first.Output != "abcabc" {
		t.Errorf("the first output became %q", first.Output)
	}
}
