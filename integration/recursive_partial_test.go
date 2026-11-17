package integration

import (
	"strings"
	"testing"

	"github.com/Notifuse/liquidgo/v5/liquid"
	"github.com/Notifuse/liquidgo/v5/liquid/tags"
)

// recursivePartials is a file system whose partials render themselves.
type recursivePartials map[string]string

func (f recursivePartials) ReadTemplateFile(path string) (string, error) { return f[path], nil }

// TestRecursivePartialsAreBounded renders partials that render themselves, with
// render and include, alone and inside an if: each must stop with "Nesting too
// deep" about 100 levels down, keep what it wrote before, and never overflow the
// Go stack, which no recover catches and which would end the process.
func TestRecursivePartialsAreBounded(t *testing.T) {
	partials := recursivePartials{
		"a": `x{% if true %}{% render 'a' %}{% endif %}`,
		"b": `x{% if true %}{% include 'b' %}{% endif %}`,
		"c": `x{% render 'c' %}`,
		"d": `x{% include 'd' %}`,
	}
	for _, tc := range []struct{ name, template string }{
		{"render inside an if", `{% if true %}{% render 'a' %}{% endif %}`},
		{"include inside an if", `{% if true %}{% include 'b' %}{% endif %}`},
		{"render", `{% render 'c' %}`},
		{"include", `{% include 'd' %}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			env := mustEnv(liquid.EnvironmentConfig{Tags: tags.Standard()})
			tmpl, err := liquid.ParseTemplate(tc.template, &liquid.TemplateOptions{Environment: env})
			if err != nil {
				t.Fatal(err)
			}
			tmpl.Registers()["file_system"] = partials
			out := renderRuby(tmpl, map[string]interface{}{}, nil)
			if !strings.HasSuffix(out, "Liquid error: Nesting too deep") {
				t.Errorf("the recursion does not end with a nesting error: %q", out)
			}
			if levels := strings.Count(out, "x"); levels < 90 || levels > 110 {
				t.Errorf("the recursion stopped after %d levels, want about 100", levels)
			}
		})
	}
}
