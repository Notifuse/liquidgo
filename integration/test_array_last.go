package integration

import (
	"testing"

	"github.com/Notifuse/liquidgo/liquid"
	"github.com/Notifuse/liquidgo/liquid/tags"
)

// TestArrayCommandMethods tests array command methods like .last, .first, and .size.
func TestArrayCommandMethods(t *testing.T) {
	env := liquid.NewEnvironment()
	tags.RegisterStandardTags(env)

	tests := []struct {
		name     string
		template string
		vars     map[string]interface{}
		want     string
	}{
		{
			name:     "array.last",
			template: `{{ items.last }}`,
			vars:     map[string]interface{}{"items": []interface{}{1, 2, 3}},
			want:     "3",
		},
		{
			name:     "array.first",
			template: `{{ items.first }}`,
			vars:     map[string]interface{}{"items": []interface{}{1, 2, 3}},
			want:     "1",
		},
		{
			name:     "array.size",
			template: `{{ items.size }}`,
			vars:     map[string]interface{}{"items": []interface{}{1, 2, 3}},
			want:     "3",
		},
		{
			name:     "string.size",
			template: `{{ str.size }}`,
			vars:     map[string]interface{}{"str": "hello"},
			want:     "5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl := liquid.NewTemplate(&liquid.TemplateOptions{Environment: env})
			err := tmpl.Parse(tt.template, nil)
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			got := tmpl.Render(tt.vars, nil)
			if got != tt.want {
				t.Errorf("FAILED\nTemplate: %s\nWant: %q\nGot:  %q", tt.template, tt.want, got)
			} else {
				t.Logf("PASS: %q", got)
			}
		})
	}
}
