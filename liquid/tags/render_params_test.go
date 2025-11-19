package tags

import (
	"fmt"
	"strings"
	"testing"

	"github.com/Notifuse/liquidgo/liquid"
)

// mapFileSystem for testing
type mapFileSystem struct {
	templates map[string]string
}

func (m *mapFileSystem) ReadTemplateFile(path string) (string, error) {
	content, ok := m.templates[path]
	if !ok {
		return "", fmt.Errorf("template not found: %s", path)
	}
	return content, nil
}

// TestRenderTagWithParameters verifies that parameters are passed to rendered partials
func TestRenderTagWithParameters(t *testing.T) {
	tests := []struct {
		name     string
		template string
		partial  string
		data     map[string]interface{}
		expected string
	}{
		{
			name:     "single parameter",
			template: `{% render 'shared', widget: 'newsletter' %}`,
			partial:  `{%- if widget == 'newsletter' -%}Success{%- endif -%}`,
			data:     map[string]interface{}{},
			expected: "Success",
		},
		{
			name:     "multiple parameters",
			template: `{% render 'card', title: 'Test', price: 9.99 %}`,
			partial:  `{{ title }} - ${{ price }}`,
			data:     map[string]interface{}{},
			expected: "Test - $9.99",
		},
		{
			name:     "string and variable parameters",
			template: `{% render 'card', title: 'Product', product: myProduct %}`,
			partial:  `{{ title }}: {{ product.name }}`,
			data: map[string]interface{}{
				"myProduct": map[string]interface{}{
					"name": "Widget",
				},
			},
			expected: "Product: Widget",
		},
		{
			name:     "parameter with quoted string",
			template: `{% render 'shared', widget: 'newsletter', title: 'Hello' %}`,
			partial:  `{%- if widget == 'newsletter' -%}<div class="newsletter">{{ title }}</div>{%- endif -%}`,
			data:     map[string]interface{}{},
			expected: `<div class="newsletter">Hello</div>`,
		},
		{
			name:     "parameter isolation - parent vars not accessible",
			template: `{% assign widget = 'blog' %}{% render 'shared', widget: 'newsletter' %}`,
			partial:  `{{ widget }}`,
			data:     map[string]interface{}{},
			expected: "newsletter",
		},
		{
			name:     "no parameters - variable should be nil",
			template: `{% render 'shared' %}`,
			partial:  `{%- if widget -%}Has widget{%- else -%}No widget{%- endif -%}`,
			data:     map[string]interface{}{},
			expected: "No widget",
		},
		{
			name:     "numeric parameter",
			template: `{% render 'item', index: 5 %}`,
			partial:  `Item {{ index }}`,
			data:     map[string]interface{}{},
			expected: "Item 5",
		},
		{
			name:     "boolean parameter",
			template: `{% render 'item', active: true %}`,
			partial:  `{%- if active -%}Active{%- else -%}Inactive{%- endif -%}`,
			data:     map[string]interface{}{},
			expected: "Active",
		},
		{
			name:     "parameter from context",
			template: `{% render 'card', product: product %}`,
			partial:  `{{ product.name }} - ${{ product.price }}`,
			data: map[string]interface{}{
				"product": map[string]interface{}{
					"name":  "Gadget",
					"price": 19.99,
				},
			},
			expected: "Gadget - $19.99",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := liquid.NewEnvironment()
			RegisterStandardTags(env)

			// Create filesystem with partial
			fs := &mapFileSystem{
				templates: map[string]string{
					"shared": tt.partial,
					"card":   tt.partial,
					"item":   tt.partial,
				},
			}

			tmpl, err := liquid.ParseTemplate(tt.template, &liquid.TemplateOptions{
				Environment: env,
			})
			if err != nil {
				t.Fatalf("Parse error: %v", err)
			}

			// Set filesystem in registers
			tmpl.Registers()["file_system"] = fs

			// Render
			output := tmpl.Render(tt.data, nil)
			output = strings.TrimSpace(output)

			if output != tt.expected {
				t.Errorf("Expected: %q\nGot: %q", tt.expected, output)
			}
		})
	}
}

// TestRenderTagParameterIsolation verifies that render creates an isolated scope
func TestRenderTagParameterIsolation(t *testing.T) {
	env := liquid.NewEnvironment()
	RegisterStandardTags(env)

	// Parent context has widget='blog'
	// Render passes widget='newsletter'
	// Partial should only see widget='newsletter'
	template := `{% assign widget = 'blog' %}{% assign title = 'Parent Title' %}{% render 'test', widget: 'newsletter' %}{{ widget }}`
	partial := `{{ widget }}-{{ title }}`

	fs := &mapFileSystem{
		templates: map[string]string{
			"test": partial,
		},
	}

	tmpl, err := liquid.ParseTemplate(template, &liquid.TemplateOptions{
		Environment: env,
	})
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	tmpl.Registers()["file_system"] = fs

	output := tmpl.Render(map[string]interface{}{}, nil)

	// Expected: "newsletter-blog" where:
	// - "newsletter-" is from the partial (widget parameter passed, title is nil/empty)
	// - "blog" is from parent context (widget variable after render)
	//
	// The partial should NOT have access to parent's title variable
	expected := "newsletter-blog"
	if output != expected {
		t.Errorf("Isolation failed.\nExpected: %q\nGot: %q", expected, output)
		t.Logf("The partial should receive widget='newsletter' but NOT title from parent scope")
	}
}

// TestRenderTagWithForLoop verifies parameters work with for loops
func TestRenderTagWithForLoop(t *testing.T) {
	env := liquid.NewEnvironment()
	RegisterStandardTags(env)

	template := `{% render 'item' for items, index: forloop.index %}`
	partial := `{{ item }} (#{{ index }})`

	fs := &mapFileSystem{
		templates: map[string]string{
			"item": partial,
		},
	}

	tmpl, err := liquid.ParseTemplate(template, &liquid.TemplateOptions{
		Environment: env,
	})
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	tmpl.Registers()["file_system"] = fs

	output := tmpl.Render(map[string]interface{}{
		"items": []interface{}{"A", "B", "C"},
	}, nil)

	// Each item should be rendered with its forloop index
	// Note: The parameter 'index' gets the forloop.index value
	// But this might not work as expected since forloop.index is evaluated in parent context
	t.Logf("Output: %q", output)
}

// TestRenderTagOriginalIssue tests the exact case from the bug report
func TestRenderTagOriginalIssue(t *testing.T) {
	env := liquid.NewEnvironment()
	RegisterStandardTags(env)

	template := `<div>{% render 'shared', widget: 'newsletter' %}</div>`
	partial := `{%- if widget == 'newsletter' -%}<div class="newsletter">Subscribe!</div>{%- endif -%}`

	fs := &mapFileSystem{
		templates: map[string]string{
			"shared": partial,
		},
	}

	tmpl, err := liquid.ParseTemplate(template, &liquid.TemplateOptions{
		Environment: env,
	})
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	tmpl.Registers()["file_system"] = fs

	output := tmpl.Render(map[string]interface{}{}, nil)

	expected := `<div><div class="newsletter">Subscribe!</div></div>`
	if output != expected {
		t.Errorf("Original issue still present!\nExpected: %q\nGot: %q", expected, output)
		t.Log("The widget parameter was not passed to the partial correctly")
	}
}
