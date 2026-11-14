package liquid

import (
	"math"
	"strings"
	"testing"
)

// TestResourceLimitsOfAnyIntegerType checks that a limit given as an int64,
// a uint, an int32 or a whole float64 is applied, as an int is, and that one
// that is not an integer is left out.
func TestResourceLimitsOfAnyIntegerType(t *testing.T) {
	for _, limit := range []interface{}{int64(5), uint(5), int32(5), uint64(5), int8(5), 5.0} {
		env := NewEnvironment()
		env.SetDefaultResourceLimits(map[string]interface{}{"render_length_limit": limit})
		tmpl, err := ParseTemplate("0123456789", &TemplateOptions{Environment: env})
		if err != nil {
			t.Fatal(err)
		}
		if out := tmpl.Render(nil, nil); !strings.Contains(out, "Memory limits exceeded") {
			t.Errorf("render_length_limit %T(5): got %q, want the limit error", limit, out)
		}
	}
	if limitValue("5") != nil || limitValue(5.5) != nil || limitValue(math.NaN()) != nil || limitValue(nil) != nil {
		t.Error("a limit that is not an integer must be left out")
	}
	// A number decoded from JSON is a float64.
	if got := limitValue(5.0); got == nil || *got != 5 {
		t.Errorf("a whole float64 limit must be applied, got %v", got)
	}
	if got := limitValue(1e300); got == nil || *got != math.MaxInt {
		t.Errorf("a float64 limit past the int range must saturate, got %v", got)
	}
	if got := limitValue(uint64(1) << 63); got == nil || *got <= 0 {
		t.Errorf("a uint64 limit past the int range must saturate, got %v", got)
	}
}

type dropWithArguments struct{ Drop }

func (d *dropWithArguments) AddDate(years, months, days int) string { return "called" }

func (d *dropWithArguments) Name() string { return "drop" }

// TestDropMethodsThatNeedArgumentsAreNotCalled checks that a lookup does not
// call a method that needs arguments, which would panic, and still calls one
// that takes none.
func TestDropMethodsThatNeedArgumentsAreNotCalled(t *testing.T) {
	tmpl, err := ParseTemplate("{{ d.add_date }}|{{ d.AddDate }}|{{ d.name }}", nil)
	if err != nil {
		t.Fatal(err)
	}
	out := tmpl.Render(map[string]interface{}{"d": &dropWithArguments{}}, nil)
	if out != "||drop" {
		t.Errorf("got %q, want %q (errors %v)", out, "||drop", tmpl.Errors())
	}
}

// TestAHashIndexedByNilFindsNothing checks that a key that evaluates to nil
// finds nothing in a hash, as in Ruby, instead of panicking.
func TestAHashIndexedByNilFindsNothing(t *testing.T) {
	tmpl, err := ParseTemplate("a{{ h[nothing] }}b{{ h[] }}c", nil)
	if err != nil {
		t.Fatal(err)
	}
	data := map[string]interface{}{"h": map[string]interface{}{"k": "v"}}
	if out := tmpl.Render(data, nil); out != "abc" || len(tmpl.Errors()) > 0 {
		t.Errorf("got %q (errors %v), want %q", out, tmpl.Errors(), "abc")
	}
}
