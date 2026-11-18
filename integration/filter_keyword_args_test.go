package integration

import (
	gocontext "context"
	"strings"
	"testing"

	"github.com/Notifuse/liquidgo/v5/liquid"
	"github.com/Notifuse/liquidgo/v5/liquid/tags"
)

// Keyword arguments reach the filter as a trailing map, as in Ruby, and a call with too
// many arguments renders nothing instead of panicking or skipping the filter, or, with
// RaiseErrors, ends the render with Ruby's error.
func TestFilterKeywordArguments(t *testing.T) {
	assigns := map[string]interface{}{"flag": true, "f": false}
	tests := []struct {
		template string
		expected string
		raises   string // the error RaiseErrors (Ruby's render!) ends the render with
	}{
		{`{{ false | default: "x", allow_false: true }}`, "false", ""},
		{`{{ f | default: "x", allow_false: flag }}`, "false", ""},
		{`{{ nil | default: "x", allow_false: true }}`, "x", ""},
		{`{{ false | default: "x", allow_false: false }}`, "x", ""},
		{`{{ false | default: "x" }}`, "x", ""},
		{`{{ "" | default: "x", allow_false: true }}`, "x", ""},
		{`{% assign v = false | default: "x", allow_false: true %}{{ v }}`, "false", ""},
		{`{{ "a" | upcase: 1 }}`, "", "wrong number of arguments"},
		{`{{ "a" | upcase: foo: 1 }}`, "", "wrong number of arguments"},
		{`{{ "<b>" | escape: 1 }}`, "", "wrong number of arguments"},
		{`{{ "abc" | upcase: locale: "fr" }}`, "", "wrong number of arguments"},
		{`{{ "<b>" | escape: 1 | append: "!" }}`, "!", "wrong number of arguments"},
		// as in Ruby, keyword arguments fill the next positional parameter as a hash
		{`{{ "abcdefgh" | truncate: 5, omission: "!" }}`, `{"omission"=>"!"}`, ""},
	}
	for _, tt := range tests {
		for _, mode := range []liquid.ErrorMode{"lax", "strict"} {
			t.Run(tt.template+"/"+string(mode), func(t *testing.T) {
				if tt.raises == "" {
					assertTemplateResult(t, tt.expected, tt.template, assigns, TemplateResultOptions{ErrorMode: mode})
					return
				}
				// Without RaiseErrors the error is recorded and writes nothing;
				// with it, the render ends with the error, as Ruby's render!.
				assertTemplateResult(t, tt.expected, tt.template, assigns, TemplateResultOptions{ErrorMode: mode, RenderErrors: true})
				env := mustEnv(liquid.EnvironmentConfig{Tags: tags.Standard(), ErrorMode: mode})
				tmpl, err := liquid.ParseTemplate(tt.template, &liquid.TemplateOptions{Environment: env})
				if err != nil {
					t.Fatalf("parse: %v", err)
				}
				if _, err := tmpl.Render(gocontext.Background(), assigns, liquid.RaiseErrors()); err == nil || !strings.Contains(err.Error(), tt.raises) {
					t.Errorf("with RaiseErrors, err = %v; want one that contains %q", err, tt.raises)
				}
			})
		}
	}
}
