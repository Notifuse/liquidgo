package integration

import (
	gocontext "context"
	"reflect"
	"testing"

	"github.com/Notifuse/liquidgo/v5/liquid"
	"github.com/Notifuse/liquidgo/v5/liquid/tags"
)

func renderGo(t *testing.T, src string, data map[string]interface{}) liquid.Result {
	t.Helper()
	env := mustEnv(liquid.EnvironmentConfig{Tags: tags.Standard()})
	tmpl, err := liquid.ParseTemplate(src, &liquid.TemplateOptions{Environment: env})
	if err != nil {
		t.Fatalf("parse %q: %v", src, err)
	}
	res, err := tmpl.Render(gocontext.Background(), data)
	if err != nil {
		t.Fatalf("render %q: %v", src, err)
	}
	return res
}

type goSecret struct{ Token string }

// A reflect.Value writes its String, never the fields of what it holds.
func TestAReflectValueWritesNoField(t *testing.T) {
	rv := reflect.ValueOf(goSecret{Token: "tok-S3CR3T"})
	res := renderGo(t, `{{ rv }}|{{ list }}|{{ h }}`, map[string]interface{}{"rv": rv, "list": []interface{}{rv}, "h": map[string]interface{}{"rv": rv}})
	if containsAny(res.Output, "S3CR3T") {
		t.Errorf("got %q: a reflect.Value wrote the fields it holds", res.Output)
	}
}

// A channel or a func a template cannot call writes nothing, not its address.
func TestAChannelOrAFuncWritesNoAddress(t *testing.T) {
	ch := make(chan int)
	fn := func(int) int { return 0 }
	res := renderGo(t, `[{{ ch }}|{{ fn }}|{{ list }}|{{ h }}]`, map[string]interface{}{"ch": ch, "fn": fn, "list": []interface{}{ch, fn}, "h": map[string]interface{}{"c": ch}})
	if containsAny(res.Output, "0x") {
		t.Errorf("got %q: an address was written", res.Output)
	}
	if res.Output != `[|||{"c"=>nil}]` {
		t.Errorf("got %q", res.Output)
	}
}

// A key a Go map cannot hash finds nothing, as a key it does not have.
func TestAnUnhashableKeyFindsNothing(t *testing.T) {
	y := map[interface{}]interface{}{"k": "v"}
	res := renderGo(t, `[{{ y[y] }}|{{ y[list] }}|{{ y.k }}]`, map[string]interface{}{"y": y, "list": []interface{}{1}})
	if res.Output != "[||v]" || len(res.Errors) != 0 {
		t.Errorf("got %q %v", res.Output, res.Errors)
	}
}

// A nil pointer is nil in a condition as in the output.
func TestANilPointerIsFalse(t *testing.T) {
	var np *goSecret
	res := renderGo(t, `[{{ np }}]{% if np %}yes{% else %}no{% endif %}{% unless np %}!{% endunless %}{% if np == nil %}=nil{% endif %}`, map[string]interface{}{"np": np})
	if res.Output != "[]no!=nil" {
		t.Errorf("got %q", res.Output)
	}
}
