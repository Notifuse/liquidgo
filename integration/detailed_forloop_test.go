package integration

import (
	"testing"

	"github.com/Notifuse/liquidgo/liquid"
	"github.com/Notifuse/liquidgo/liquid/tags"
)

func TestForloopLastDetailed(t *testing.T) {
	tests := []struct {
		name     string
		template string
		want     string
	}{
		{
			name:     "simple forloop.last",
			template: `{% for item in items %}{{ item }}{% if forloop.last %}!{% endif %}{% endfor %}`,
			want:     "123!",
		},
		{
			name:     "forloop.last with trailing comma",
			template: `{% for item in items %}{{ item }},{% endfor %}`,
			want:     "1,2,3,",
		},
		{
			name:     "forloop.last controlling comma",
			template: `{% for item in items %}{{ item }}{% unless forloop.last %},{% endunless %}{% endfor %}`,
			want:     "1,2,3",
		},
		{
			name:     "forloop.last with content then comma",
			template: `{% for item in items %}{{ item }}{% if forloop.last %} last{% endif %},{% endfor %}`,
			want:     "1,2,3 last,",
		},
		{
			name:     "forloop.last debug",
			template: `{% for item in items %}{{ forloop.last }}{% endfor %}`,
			want:     "falsefalsetrue",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := liquid.NewEnvironment()
			tags.RegisterStandardTags(env)
			tmpl := liquid.NewTemplate(&liquid.TemplateOptions{Environment: env})
			
			err := tmpl.Parse(tt.template, nil)
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			vars := map[string]interface{}{"items": []interface{}{1, 2, 3}}
			got := tmpl.Render(vars, nil)

			if got != tt.want {
				t.Errorf("FAILED\nWant: %q\nGot:  %q", tt.want, got)
			} else {
				t.Logf("PASS: %q", got)
			}
		})
	}
}
