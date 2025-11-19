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
