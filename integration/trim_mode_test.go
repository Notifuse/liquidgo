package integration

import (
	"testing"
)

// TestStandardOutput makes sure the trim isn't applied to standard output
func TestStandardOutput(t *testing.T) {
	text := `
      <div>
        <p>
          {{ 'John' }}
        </p>
      </div>
    `
	expected := `
      <div>
        <p>
          John
        </p>
      </div>
    `
	assertTemplateResult(t, expected, text, nil)
}

// TestVariableOutputWithMultipleBlankLines tests variable output with multiple blank lines
func TestVariableOutputWithMultipleBlankLines(t *testing.T) {
	text := `
      <div>
        <p>


          {{- 'John' -}}


        </p>
      </div>
    `
	expected := `
      <div>
        <p>John</p>
      </div>
    `
	assertTemplateResult(t, expected, text, nil)
}

// TestTagOutputWithMultipleBlankLines tests tag output with multiple blank lines
func TestTagOutputWithMultipleBlankLines(t *testing.T) {
	text := `
      <div>
        <p>


          {%- if true -%}
          yes
          {%- endif -%}


        </p>
      </div>
    `
	expected := `
      <div>
        <p>yes</p>
      </div>
    `
	assertTemplateResult(t, expected, text, nil)
}

// TestStandardTags makes sure the trim isn't applied to standard tags
func TestStandardTags(t *testing.T) {
	whitespace := "          "
	text := `
      <div>
        <p>
          {% if true %}
          yes
          {% endif %}
        </p>
      </div>
    `
	expected := "\n      <div>\n        <p>\n" + whitespace + "\n          yes\n" + whitespace + "\n        </p>\n      </div>\n    "
	assertTemplateResult(t, expected, text, nil)

	text = `
      <div>
        <p>
          {% if false %}
          no
          {% endif %}
        </p>
      </div>
    `
	expected = "\n      <div>\n        <p>\n" + whitespace + "\n        </p>\n      </div>\n    "
	assertTemplateResult(t, expected, text, nil)
}

// TestNoTrimOutput makes sure the trim isn't too aggressive
func TestNoTrimOutput(t *testing.T) {
	text := "<p>{{- 'John' -}}</p>"
	expected := "<p>John</p>"
	assertTemplateResult(t, expected, text, nil)
}

// TestNoTrimTags makes sure the trim isn't too aggressive
func TestNoTrimTags(t *testing.T) {
	text := "<p>{%- if true -%}yes{%- endif -%}</p>"
	expected := "<p>yes</p>"
	assertTemplateResult(t, expected, text, nil)

	text = "<p>{%- if false -%}no{%- endif -%}</p>"
	expected = "<p></p>"
	assertTemplateResult(t, expected, text, nil)
}

// TestSingleLineOuterTag tests single line outer tag
func TestSingleLineOuterTag(t *testing.T) {
	text := "<p> {%- if true %} yes {% endif -%} </p>"
	expected := "<p> yes </p>"
	assertTemplateResult(t, expected, text, nil)

	text = "<p> {%- if false %} no {% endif -%} </p>"
	expected = "<p></p>"
	assertTemplateResult(t, expected, text, nil)
}

// TestSingleLineInnerTag tests single line inner tag
func TestSingleLineInnerTag(t *testing.T) {
	text := "<p> {% if true -%} yes {%- endif %} </p>"
	expected := "<p> yes </p>"
	assertTemplateResult(t, expected, text, nil)

	text = "<p> {% if false -%} no {%- endif %} </p>"
	expected = "<p>  </p>"
	assertTemplateResult(t, expected, text, nil)
}

// TestSingleLinePostTag tests single line post tag
func TestSingleLinePostTag(t *testing.T) {
	text := "<p> {% if true -%} yes {% endif -%} </p>"
	expected := "<p> yes </p>"
	assertTemplateResult(t, expected, text, nil)

	text = "<p> {% if false -%} no {% endif -%} </p>"
	expected = "<p> </p>"
	assertTemplateResult(t, expected, text, nil)
}

// TestSingleLinePreTag tests single line pre tag
func TestSingleLinePreTag(t *testing.T) {
	text := "<p> {%- if true %} yes {%- endif %} </p>"
	expected := "<p> yes </p>"
	assertTemplateResult(t, expected, text, nil)

	text = "<p> {%- if false %} no {%- endif %} </p>"
	expected = "<p> </p>"
	assertTemplateResult(t, expected, text, nil)
}

// TestPreTrimOutput tests pre trim output
func TestPreTrimOutput(t *testing.T) {
	text := `
      <div>
        <p>
          {{- 'John' }}
        </p>
      </div>
    `
	expected := `
      <div>
        <p>John
        </p>
      </div>
    `
	assertTemplateResult(t, expected, text, nil)
}

// TestPreTrimTags tests pre trim tags
func TestPreTrimTags(t *testing.T) {
	text := `
      <div>
        <p>
          {%- if true %}
          yes
          {%- endif %}
        </p>
      </div>
    `
	expected := `
      <div>
        <p>
          yes
        </p>
      </div>
    `
	assertTemplateResult(t, expected, text, nil)

	text = `
      <div>
        <p>
          {%- if false %}
          no
          {%- endif %}
        </p>
      </div>
    `
	expected = `
      <div>
        <p>
        </p>
      </div>
    `
	assertTemplateResult(t, expected, text, nil)
}

// TestPostTrimOutput tests post trim output
func TestPostTrimOutput(t *testing.T) {
	text := `
      <div>
        <p>
          {{ 'John' -}}
        </p>
      </div>
    `
	expected := `
      <div>
        <p>
          John</p>
      </div>
    `
	assertTemplateResult(t, expected, text, nil)
}

// TestPostTrimTags tests post trim tags
func TestPostTrimTags(t *testing.T) {
	text := `
      <div>
        <p>
          {% if true -%}
          yes
          {% endif -%}
        </p>
      </div>
    `
	expected := `
      <div>
        <p>
          yes
          </p>
      </div>
    `
	assertTemplateResult(t, expected, text, nil)

	text = `
      <div>
        <p>
          {% if false -%}
          no
          {% endif -%}
        </p>
      </div>
    `
	expected = `
      <div>
        <p>
          </p>
      </div>
    `
	assertTemplateResult(t, expected, text, nil)
}

// TestPreAndPostTrimTags tests pre and post trim tags
func TestPreAndPostTrimTags(t *testing.T) {
	text := `
      <div>
        <p>
          {%- if true %}
          yes
          {% endif -%}
        </p>
      </div>
    `
	expected := `
      <div>
        <p>
          yes
          </p>
      </div>
    `
	assertTemplateResult(t, expected, text, nil)

	text = `
      <div>
        <p>
          {%- if false %}
          no
          {% endif -%}
        </p>
      </div>
    `
	expected = `
      <div>
        <p></p>
      </div>
    `
	assertTemplateResult(t, expected, text, nil)
}

// TestPostAndPreTrimTags tests post and pre trim tags
func TestPostAndPreTrimTags(t *testing.T) {
	text := `
      <div>
        <p>
          {% if true -%}
          yes
          {%- endif %}
        </p>
      </div>
    `
	expected := `
      <div>
        <p>
          yes
        </p>
      </div>
    `
	assertTemplateResult(t, expected, text, nil)

	whitespace := "          "
	text = `
      <div>
        <p>
          {% if false -%}
          no
          {%- endif %}
        </p>
      </div>
    `
	expected = "\n      <div>\n        <p>\n" + whitespace + "\n        </p>\n      </div>\n    "
	assertTemplateResult(t, expected, text, nil)
}

// TestTrimOutput tests trim output
func TestTrimOutput(t *testing.T) {
	text := `
      <div>
        <p>
          {{- 'John' -}}
        </p>
      </div>
    `
	expected := `
      <div>
        <p>John</p>
      </div>
    `
	assertTemplateResult(t, expected, text, nil)
}

// TestTrimTags tests trim tags
func TestTrimTags(t *testing.T) {
	text := `
      <div>
        <p>
          {%- if true -%}
          yes
          {%- endif -%}
        </p>
      </div>
    `
	expected := `
      <div>
        <p>yes</p>
      </div>
    `
	assertTemplateResult(t, expected, text, nil)

	text = `
      <div>
        <p>
          {%- if false -%}
          no
          {%- endif -%}
        </p>
      </div>
    `
	expected = `
      <div>
        <p></p>
      </div>
    `
	assertTemplateResult(t, expected, text, nil)
}

// TestWhitespaceTrimOutput tests whitespace trim output
func TestWhitespaceTrimOutput(t *testing.T) {
	text := `
      <div>
        <p>
          {{- 'John' -}},
          {{- '30' -}}
        </p>
      </div>
    `
	expected := `
      <div>
        <p>John,30</p>
      </div>
    `
	assertTemplateResult(t, expected, text, nil)
}

// TestWhitespaceTrimTags tests whitespace trim tags
func TestWhitespaceTrimTags(t *testing.T) {
	text := `
      <div>
        <p>
          {%- if true -%}
          yes
          {%- endif -%}
        </p>
      </div>
    `
	expected := `
      <div>
        <p>yes</p>
      </div>
    `
	assertTemplateResult(t, expected, text, nil)

	text = `
      <div>
        <p>
          {%- if false -%}
          no
          {%- endif -%}
        </p>
      </div>
    `
	expected = `
      <div>
        <p></p>
      </div>
    `
	assertTemplateResult(t, expected, text, nil)
}

// TestComplexTrimOutput tests complex trim output
func TestComplexTrimOutput(t *testing.T) {
	text := `
      <div>
        <p>
          {{- 'John' -}}
          {{- '30' -}}
        </p>
        <b>
          {{ 'John' -}}
          {{- '30' }}
        </b>
        <i>
          {{- 'John' }}
          {{ '30' -}}
        </i>
      </div>
    `
	expected := `
      <div>
        <p>John30</p>
        <b>
          John30
        </b>
        <i>John
          30</i>
      </div>
    `
	assertTemplateResult(t, expected, text, nil)
}

// TestComplexTrim tests complex trim
func TestComplexTrim(t *testing.T) {
	text := `
      <div>
        {%- if true -%}
          {%- if true -%}
            <p>
              {{- 'John' -}}
            </p>
          {%- endif -%}
        {%- endif -%}
      </div>
    `
	expected := `
      <div><p>John</p></div>
    `
	assertTemplateResult(t, expected, text, nil)
}

// TestRightTrimFollowedByTag tests right trim followed by tag
func TestRightTrimFollowedByTag(t *testing.T) {
	assertTemplateResult(t, "ab c", `{{ "a" -}}{{ "b" }} c`, nil)
}

// TestRawOutput tests raw output
func TestRawOutput(t *testing.T) {
	whitespace := "        "
	text := `
      <div>
        {% raw %}
          {%- if true -%}
            <p>
              {{- 'John' -}}
            </p>
          {%- endif -%}
        {% endraw %}
      </div>
    `
	expected := "\n      <div>\n" + whitespace + "\n          {%- if true -%}\n            <p>\n              {{- 'John' -}}\n            </p>\n          {%- endif -%}\n" + whitespace + "\n      </div>\n    "
	assertTemplateResult(t, expected, text, nil)
}

// TestPreTrimBlankPrecedingText tests pre trim blank preceding text
func TestPreTrimBlankPrecedingText(t *testing.T) {
	assertTemplateResult(t, "", "\n{%- raw %}{% endraw %}", nil)
	assertTemplateResult(t, "", "\n{%- if true %}{% endif %}", nil)
	assertTemplateResult(t, "BC", "{{ 'B' }} \n{%- if true %}C{% endif %}", nil)
}

// TestTrimBlank tests trim blank
func TestTrimBlank(t *testing.T) {
	assertTemplateResult(t, "foobar", "foo {{-}} bar", nil)
}

// TestLayoutWithRender tests layout pattern with render tag
func TestLayoutWithRender(t *testing.T) {
	// Template that uses render tag to include a partial
	template := `<div class="layout">
  <header>{% render "header" %}</header>
  <main>{{ content_for_layout }}</main>
  <footer>{% render "footer" %}</footer>
</div>`

	// Expected output with content_for_layout set
	expected := `<div class="layout">
  <header>Welcome</header>
  <main>Page Content</main>
  <footer>Copyright 2024</footer>
</div>`

	// Partials for render tag
	partials := map[string]string{
		"header": "Welcome",
		"footer": "Copyright 2024",
	}

	// Assigns including content_for_layout
	assigns := map[string]interface{}{
		"content_for_layout": "Page Content",
	}

	assertTemplateResult(t, expected, template, assigns, TemplateResultOptions{
		Partials: partials,
	})
}

// TestNestedIf tests that nested if statements work correctly inside if-elsif-else blocks.
// This matches the Ruby test_nested_if from reference-liquid/test/integration/tags/if_else_tag_test.rb.
func TestNestedIf(t *testing.T) {
	// Test basic nested if statements (matching Ruby test_nested_if lines 111-120)
	assertTemplateResult(t, "", "{% if false %}{% if false %} NO {% endif %}{% endif %}", nil)
	assertTemplateResult(t, "", "{% if false %}{% if true %} NO {% endif %}{% endif %}", nil)
	assertTemplateResult(t, "", "{% if true %}{% if false %} NO {% endif %}{% endif %}", nil)
	assertTemplateResult(t, " YES ", "{% if true %}{% if true %} YES {% endif %}{% endif %}", nil)

	assertTemplateResult(t, " YES ", "{% if true %}{% if true %} YES {% else %} NO {% endif %}{% else %} NO {% endif %}", nil)
	assertTemplateResult(t, " YES ", "{% if true %}{% if false %} NO {% else %} YES {% endif %}{% else %} NO {% endif %}", nil)
	assertTemplateResult(t, " YES ", "{% if false %}{% if true %} NO {% else %} NONO {% endif %}{% else %} YES {% endif %}", nil)
}

// TestNestedIfWithOrOperator tests nested if statements with OR operators in the outer condition.
// This is a real-world use case from meta tag generation.
func TestNestedIfWithOrOperator(t *testing.T) {
	template := `{% if post.seo.og_title or post.title %}

    <meta property="og:title" content="{% if post.seo.og_title %}{{ post.seo.og_title }}{% else %}{{ post.title }}{% endif %}">

{% elsif category.seo.og_title or category.name %}

    <meta property="og:title" content="{% if category.seo.og_title %}{{ category.seo.og_title }}{% else %}{{ category.name }}{% endif %}">

{% else %}

    <meta property="og:title" content="{{ workspace.name }}">

{% endif %}`

	// Test case 1: post.seo.og_title exists (nested if should use og_title)
	assigns1 := map[string]interface{}{
		"post": map[string]interface{}{
			"seo": map[string]interface{}{
				"og_title": "Post OG Title",
			},
			"title": "Post Title",
		},
	}
	expected1 := "\n\n    <meta property=\"og:title\" content=\"Post OG Title\">\n\n"
	assertTemplateResult(t, expected1, template, assigns1)

	// Test case 2: post.seo.og_title doesn't exist, but post.title does (nested if should use title)
	assigns2 := map[string]interface{}{
		"post": map[string]interface{}{
			"seo":   map[string]interface{}{},
			"title": "Post Title Only",
		},
	}
	expected2 := "\n\n    <meta property=\"og:title\" content=\"Post Title Only\">\n\n"
	assertTemplateResult(t, expected2, template, assigns2)

	// Test case 3: category.seo.og_title exists (elsif branch)
	assigns3 := map[string]interface{}{
		"post": map[string]interface{}{
			"seo": map[string]interface{}{},
		},
		"category": map[string]interface{}{
			"seo": map[string]interface{}{
				"og_title": "Category OG Title",
			},
			"name": "Category Name",
		},
	}
	expected3 := "\n\n    <meta property=\"og:title\" content=\"Category OG Title\">\n\n"
	assertTemplateResult(t, expected3, template, assigns3)

	// Test case 4: category.seo.og_title doesn't exist, but category.name does (elsif branch, nested if)
	assigns4 := map[string]interface{}{
		"post": map[string]interface{}{
			"seo": map[string]interface{}{},
		},
		"category": map[string]interface{}{
			"seo":  map[string]interface{}{},
			"name": "Category Name Only",
		},
	}
	expected4 := "\n\n    <meta property=\"og:title\" content=\"Category Name Only\">\n\n"
	assertTemplateResult(t, expected4, template, assigns4)

	// Test case 5: else branch - use workspace.name
	assigns5 := map[string]interface{}{
		"post": map[string]interface{}{
			"seo": map[string]interface{}{},
		},
		"workspace": map[string]interface{}{
			"name": "Workspace Name",
		},
	}
	expected5 := "\n\n    <meta property=\"og:title\" content=\"Workspace Name\">\n\n"
	assertTemplateResult(t, expected5, template, assigns5)
}
