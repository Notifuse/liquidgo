package integration

import (
	"strings"
	"testing"
	"time"

	"github.com/Notifuse/liquidgo/v5/liquid"
	"github.com/Notifuse/liquidgo/v5/liquid/tags"
)

func TestWarningsIsNotExponentialTime(t *testing.T) {
	str := "false"
	for i := 0; i < 100; i++ {
		str = "{% if true %}true{% else %}" + str + "{% endif %}"
	}

	env := mustEnv(liquid.EnvironmentConfig{Tags: tags.Standard()})
	tmpl, err := liquid.ParseTemplate(str, &liquid.TemplateOptions{Environment: env})
	if err != nil {
		t.Fatalf("ParseTemplate() error = %v", err)
	}

	// Test that warnings can be retrieved quickly
	done := make(chan bool)
	go func() {
		_ = tmpl.Warnings()
		done <- true
	}()

	select {
	case <-done:
		// Success - warnings retrieved quickly
	case <-time.After(1 * time.Second):
		t.Error("Warnings retrieval took too long (exponential time issue)")
	}

	if len(tmpl.Warnings()) != 0 {
		t.Errorf("Expected no warnings, got %d", len(tmpl.Warnings()))
	}
}

func TestCustomAssignsDoNotPersistOnSameTemplate(t *testing.T) {
	env := mustEnv(liquid.EnvironmentConfig{Tags: tags.Standard()})
	tmpl := liquid.NewTemplate(&liquid.TemplateOptions{Environment: env})

	err := tmpl.Parse(`{{ foo }}`, &liquid.TemplateOptions{Environment: env})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Render with custom assigns
	result := renderRubyBang(tmpl, map[string]interface{}{"foo": "from custom assigns"}, &rubyOptions{})
	if result != "from custom assigns" {
		t.Errorf("Expected 'from custom assigns', got %q", result)
	}

	// Render without custom assigns - should be empty
	result = renderRubyBang(tmpl, nil, &rubyOptions{})
	if result != "" {
		t.Errorf("Expected empty string, got %q", result)
	}
}

func TestCustomAssignsSquashInstanceAssigns(t *testing.T) {
	env := mustEnv(liquid.EnvironmentConfig{Tags: tags.Standard()})
	tmpl := liquid.NewTemplate(&liquid.TemplateOptions{Environment: env})

	err := tmpl.Parse(`{% assign foo = 'from instance assigns' %}{{ foo }}`, &liquid.TemplateOptions{Environment: env})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// First render with instance assigns
	result := renderRubyBang(tmpl, nil, &rubyOptions{})
	if result != "from instance assigns" {
		t.Errorf("Expected 'from instance assigns', got %q", result)
	}

	// Parse again and render with custom assigns, as Ruby's test does: they
	// squash the instance assigns
	if err := tmpl.Parse(`{{ foo }}`, &liquid.TemplateOptions{Environment: env}); err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	result = renderRubyBang(tmpl, map[string]interface{}{"foo": "from custom assigns"}, &rubyOptions{})
	if result != "from custom assigns" {
		t.Errorf("Expected 'from custom assigns', got %q", result)
	}
}

func TestPersistentAssignsSquashInstanceAssigns(t *testing.T) {
	env := mustEnv(liquid.EnvironmentConfig{Tags: tags.Standard()})
	tmpl := liquid.NewTemplate(&liquid.TemplateOptions{Environment: env})

	err := tmpl.Parse(`{% assign foo = 'from instance assigns' %}{{ foo }}`, &liquid.TemplateOptions{Environment: env})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// First render with instance assigns
	result := renderRubyBang(tmpl, nil, &rubyOptions{})
	if result != "from instance assigns" {
		t.Errorf("Expected 'from instance assigns', got %q", result)
	}

	// Set persistent assigns
	tmpl.Assigns()["foo"] = "from persistent assigns"

	// Parse again and render - persistent assigns should squash instance assigns
	err = tmpl.Parse(`{{ foo }}`, &liquid.TemplateOptions{Environment: env})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	result = renderRubyBang(tmpl, nil, &rubyOptions{})
	if result != "from persistent assigns" {
		t.Errorf("Expected 'from persistent assigns', got %q", result)
	}
}

func TestResourceLimitsWorksWithCustomLengthMethod(t *testing.T) {
	env := mustEnv(liquid.EnvironmentConfig{Tags: tags.Standard()})
	tmpl, err := liquid.ParseTemplate(`{% assign foo = bar %}`, &liquid.TemplateOptions{Environment: env})
	if err != nil {
		t.Fatalf("ParseTemplate() error = %v", err)
	}

	limit := 42
	setRubyLimits(tmpl, liquid.ResourceLimitsConfig{
		RenderLengthLimit: &limit,
	})
	somethingWithLength := NewSomethingWithLength()
	result := renderRubyBang(tmpl, map[string]interface{}{"bar": somethingWithLength}, &rubyOptions{})
	if result != "" {
		t.Errorf("Expected empty string, got %q", result)
	}
}

func TestResourceLimitsRenderLength(t *testing.T) {
	env := mustEnv(liquid.EnvironmentConfig{Tags: tags.Standard()})
	tmpl, err := liquid.ParseTemplate("0123456789", &liquid.TemplateOptions{Environment: env})
	if err != nil {
		t.Fatalf("ParseTemplate() error = %v", err)
	}

	limit := 9
	setRubyLimits(tmpl, liquid.ResourceLimitsConfig{
		RenderLengthLimit: &limit,
	})
	result := renderRuby(tmpl, nil, &rubyOptions{})
	if !strings.Contains(result, "Memory limits exceeded") {
		t.Errorf("Expected memory limit error, got %q", result)
	}
	if !rubyLimits(tmpl).Reached() {
		t.Error("Expected resource limits to be reached")
	}

	limit = 10
	setRubyLimits(tmpl, liquid.ResourceLimitsConfig{
		RenderLengthLimit: &limit,
	})
	result = renderRubyBang(tmpl, nil, &rubyOptions{})
	if result != "0123456789" {
		t.Errorf("Expected '0123456789', got %q", result)
	}
}

func TestResourceLimitsRenderScore(t *testing.T) {
	env := mustEnv(liquid.EnvironmentConfig{Tags: tags.Standard()})

	// Test nested loops
	tmpl, err := liquid.ParseTemplate(`{% for a in (1..10) %} {% for a in (1..10) %} foo {% endfor %} {% endfor %}`, &liquid.TemplateOptions{Environment: env})
	if err != nil {
		t.Fatalf("ParseTemplate() error = %v", err)
	}

	limit := 50
	setRubyLimits(tmpl, liquid.ResourceLimitsConfig{
		RenderScoreLimit: &limit,
	})
	result := renderRuby(tmpl, nil, &rubyOptions{})
	if !strings.Contains(result, "Memory limits exceeded") {
		t.Errorf("Expected memory limit error, got %q", result)
	}
	if !rubyLimits(tmpl).Reached() {
		t.Error("Expected resource limits to be reached")
	}

	// Test single loop
	tmpl, err = liquid.ParseTemplate(`{% for a in (1..100) %} foo {% endfor %}`, &liquid.TemplateOptions{Environment: env})
	if err != nil {
		t.Fatalf("ParseTemplate() error = %v", err)
	}

	limit = 50
	setRubyLimits(tmpl, liquid.ResourceLimitsConfig{
		RenderScoreLimit: &limit,
	})
	result = renderRuby(tmpl, nil, &rubyOptions{})
	if !strings.Contains(result, "Memory limits exceeded") {
		t.Errorf("Expected memory limit error, got %q", result)
	}
	if !rubyLimits(tmpl).Reached() {
		t.Error("Expected resource limits to be reached")
	}

	// Shopify Liquid 5.14 counts each range item besides its body: 100
	// items and 100 bodies score 201, past the limit of 200 that sufficed in
	// 5.11, and Shopify's test now sets 201
	limit = 200
	setRubyLimits(tmpl, liquid.ResourceLimitsConfig{
		RenderScoreLimit: &limit,
	})
	if result := renderRuby(tmpl, nil, &rubyOptions{}); !strings.Contains(result, "Memory limits exceeded") {
		t.Errorf("Expected memory limit error at 200, got %q", result)
	}

	limit = 201
	setRubyLimits(tmpl, liquid.ResourceLimitsConfig{
		RenderScoreLimit: &limit,
	})
	expected := strings.Repeat(" foo ", 100)
	result = renderRubyBang(tmpl, nil, &rubyOptions{})
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
	if rubyLimits(tmpl).RenderScore() == 0 {
		t.Error("Expected render_score to be set")
	}
}

func TestResourceLimitsAbortsRenderingAfterFirstError(t *testing.T) {
	env := mustEnv(liquid.EnvironmentConfig{Tags: tags.Standard()})
	tmpl, err := liquid.ParseTemplate(`{% for a in (1..100) %} foo1 {% endfor %} bar {% for a in (1..100) %} foo2 {% endfor %}`, &liquid.TemplateOptions{Environment: env})
	if err != nil {
		t.Fatalf("ParseTemplate() error = %v", err)
	}

	limit := 50
	setRubyLimits(tmpl, liquid.ResourceLimitsConfig{
		RenderScoreLimit: &limit,
	})
	result := renderRuby(tmpl, nil, &rubyOptions{})
	if !strings.Contains(result, "Memory limits exceeded") {
		t.Errorf("Expected memory limit error, got %q", result)
	}
	if !rubyLimits(tmpl).Reached() {
		t.Error("Expected resource limits to be reached")
	}
}

func TestResourceLimitsHashInTemplateGetsUpdatedEvenIfNoLimitsAreSet(t *testing.T) {
	env := mustEnv(liquid.EnvironmentConfig{Tags: tags.Standard()})
	tmpl, err := liquid.ParseTemplate(`{% for a in (1..100) %}x{% assign foo = 1 %} {% endfor %}`, &liquid.TemplateOptions{Environment: env})
	if err != nil {
		t.Fatalf("ParseTemplate() error = %v", err)
	}

	renderRubyBang(tmpl, nil, &rubyOptions{})
	if rubyLimits(tmpl).AssignScore() <= 0 {
		t.Error("Expected assign_score to be greater than 0")
	}
	if rubyLimits(tmpl).RenderScore() <= 0 {
		t.Error("Expected render_score to be greater than 0")
	}
}

func TestRenderLengthPersistsBetweenBlocks(t *testing.T) {
	env := mustEnv(liquid.EnvironmentConfig{Tags: tags.Standard()})

	// Single block
	tmpl, err := liquid.ParseTemplate(`{% if true %}aaaa{% endif %}`, &liquid.TemplateOptions{Environment: env})
	if err != nil {
		t.Fatalf("ParseTemplate() error = %v", err)
	}

	limit := 3
	setRubyLimits(tmpl, liquid.ResourceLimitsConfig{
		RenderLengthLimit: &limit,
	})
	result := renderRuby(tmpl, nil, &rubyOptions{})
	if !strings.Contains(result, "Memory limits exceeded") {
		t.Errorf("Expected memory limit error, got %q", result)
	}

	limit = 4
	setRubyLimits(tmpl, liquid.ResourceLimitsConfig{
		RenderLengthLimit: &limit,
	})
	result = renderRuby(tmpl, nil, &rubyOptions{})
	if result != "aaaa" {
		t.Errorf("Expected 'aaaa', got %q", result)
	}

	// Multiple blocks
	tmpl, err = liquid.ParseTemplate(`{% if true %}aaaa{% endif %}{% if true %}bbb{% endif %}`, &liquid.TemplateOptions{Environment: env})
	if err != nil {
		t.Fatalf("ParseTemplate() error = %v", err)
	}

	limit = 6
	setRubyLimits(tmpl, liquid.ResourceLimitsConfig{
		RenderLengthLimit: &limit,
	})
	result = renderRuby(tmpl, nil, &rubyOptions{})
	if !strings.Contains(result, "Memory limits exceeded") {
		t.Errorf("Expected memory limit error, got %q", result)
	}

	limit = 7
	setRubyLimits(tmpl, liquid.ResourceLimitsConfig{
		RenderLengthLimit: &limit,
	})
	result = renderRuby(tmpl, nil, &rubyOptions{})
	if result != "aaaabbb" {
		t.Errorf("Expected 'aaaabbb', got %q", result)
	}

	// Many blocks
	tmpl, err = liquid.ParseTemplate(`{% if true %}a{% endif %}{% if true %}b{% endif %}{% if true %}a{% endif %}{% if true %}b{% endif %}{% if true %}a{% endif %}{% if true %}b{% endif %}`, &liquid.TemplateOptions{Environment: env})
	if err != nil {
		t.Fatalf("ParseTemplate() error = %v", err)
	}

	limit = 5
	setRubyLimits(tmpl, liquid.ResourceLimitsConfig{
		RenderLengthLimit: &limit,
	})
	result = renderRuby(tmpl, nil, &rubyOptions{})
	if !strings.Contains(result, "Memory limits exceeded") {
		t.Errorf("Expected memory limit error, got %q", result)
	}

	limit = 6
	setRubyLimits(tmpl, liquid.ResourceLimitsConfig{
		RenderLengthLimit: &limit,
	})
	result = renderRuby(tmpl, nil, &rubyOptions{})
	if result != "ababab" {
		t.Errorf("Expected 'ababab', got %q", result)
	}
}

func TestRenderLengthUsesNumberOfBytesNotCharacters(t *testing.T) {
	env := mustEnv(liquid.EnvironmentConfig{Tags: tags.Standard()})
	tmpl, err := liquid.ParseTemplate(`{% if true %}すごい{% endif %}`, &liquid.TemplateOptions{Environment: env})
	if err != nil {
		t.Fatalf("ParseTemplate() error = %v", err)
	}

	// すごい is 9 bytes in UTF-8
	limit := 8
	setRubyLimits(tmpl, liquid.ResourceLimitsConfig{
		RenderLengthLimit: &limit,
	})
	result := renderRuby(tmpl, nil, &rubyOptions{})
	if !strings.Contains(result, "Memory limits exceeded") {
		t.Errorf("Expected memory limit error, got %q", result)
	}

	limit = 9
	setRubyLimits(tmpl, liquid.ResourceLimitsConfig{
		RenderLengthLimit: &limit,
	})
	result = renderRuby(tmpl, nil, &rubyOptions{})
	if result != "すごい" {
		t.Errorf("Expected 'すごい', got %q", result)
	}
}

func TestDefaultResourceLimitsUnaffectedByRenderWithContext(t *testing.T) {
	env := mustEnv(liquid.EnvironmentConfig{Tags: tags.Standard()})
	ctx := &rubyContext{Environment: env}
	tmpl, err := liquid.ParseTemplate(`{% for a in (1..100) %}x{% assign foo = 1 %} {% endfor %}`, &liquid.TemplateOptions{Environment: env})
	if err != nil {
		t.Fatalf("ParseTemplate() error = %v", err)
	}

	renderRubyBang(tmpl, ctx, &rubyOptions{})
	if ctx.ResourceLimits().AssignScore() <= 0 {
		t.Error("Expected assign_score to be greater than 0")
	}
	if ctx.ResourceLimits().RenderScore() <= 0 {
		t.Error("Expected render_score to be greater than 0")
	}
}

// Ruby renders a drop as the root of the variables; liquidgo takes a map,
// so the drop goes under the key d and the templates read d.foo.
func TestCanUseDropAsContext(t *testing.T) {
	env := mustEnv(liquid.EnvironmentConfig{Tags: tags.Standard()})
	tmpl := liquid.NewTemplate(&liquid.TemplateOptions{Environment: env})
	tmpl.Registers()["lulz"] = "haha"

	drop := NewTemplateContextDrop()

	err := tmpl.Parse(`{{d.foo}}`, &liquid.TemplateOptions{Environment: env})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	result := renderRubyBang(tmpl, map[string]interface{}{"d": drop}, &rubyOptions{})
	if result != "fizzbuzz" {
		t.Errorf("Expected 'fizzbuzz', got %q", result)
	}

	err = tmpl.Parse(`{{d.bar}}`, &liquid.TemplateOptions{Environment: env})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	result = renderRubyBang(tmpl, map[string]interface{}{"d": drop}, &rubyOptions{})
	if result != "bar" {
		t.Errorf("Expected 'bar', got %q", result)
	}

	// baz reads the template's registers, as in Ruby's test, which renders
	// the drop again: a drop given to one render is not kept for the next.
	// The data is a map: the drop goes under a key, where Ruby renders it as
	// the root of the variables
	err = tmpl.Parse(`{{d.baz}}`, &liquid.TemplateOptions{Environment: env})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	result = renderRubyBang(tmpl, map[string]interface{}{"d": drop}, &rubyOptions{})
	if result != "haha" {
		t.Errorf("Expected 'haha', got %q", result)
	}
}

func TestUsingRangeLiteralWorksAsExpected(t *testing.T) {
	source := `{% assign foo = (x..y) %}{{ foo }}`
	assertTemplateResult(t, "1..5", source, map[string]interface{}{"x": 1, "y": 5})

	source = `{% assign nums = (x..y) %}{% for num in nums %}{{ num }}{% endfor %}`
	assertTemplateResult(t, "12345", source, map[string]interface{}{"x": 1, "y": 5})
}

func TestAllowsNonStringValuesAsSource(t *testing.T) {
	env := mustEnv(liquid.EnvironmentConfig{Tags: tags.Standard()})

	// Test nil
	tmpl, err := liquid.ParseTemplate("", &liquid.TemplateOptions{Environment: env})
	if err != nil {
		t.Fatalf("ParseTemplate() error = %v", err)
	}
	result := renderRuby(tmpl, nil, &rubyOptions{})
	if result != "" {
		t.Errorf("Expected empty string, got %q", result)
	}

	// Test integer (converted to string)
	tmpl, err = liquid.ParseTemplate("1", &liquid.TemplateOptions{Environment: env})
	if err != nil {
		t.Fatalf("ParseTemplate() error = %v", err)
	}
	result = renderRuby(tmpl, nil, &rubyOptions{})
	if result != "1" {
		t.Errorf("Expected '1', got %q", result)
	}

	// Test boolean (converted to string)
	tmpl, err = liquid.ParseTemplate("true", &liquid.TemplateOptions{Environment: env})
	if err != nil {
		t.Fatalf("ParseTemplate() error = %v", err)
	}
	result = renderRuby(tmpl, nil, &rubyOptions{})
	if result != "true" {
		t.Errorf("Expected 'true', got %q", result)
	}
}

// TestVariableLookupBlogPostArrayIndex tests array index access with []BlogPost.
// An index reaches the post in a typed slice as in []interface{}, but a
// BlogPost is a plain Go struct, not a drop, and exposes no property, as in
// Ruby, where only a Liquid::Drop has its methods invoked.
func TestVariableLookupBlogPostArrayIndex(t *testing.T) {
	posts := SampleBlogPosts()

	tests := []struct {
		name     string
		template string
		posts    interface{}
		expected string
	}{
		{
			name:     "typed slice - first element",
			template: `{{ posts[0].title }}`,
			posts:    posts,
			expected: "",
		},
		{
			name:     "typed slice - second element",
			template: `{{ posts[1].title }}`,
			posts:    posts,
			expected: "",
		},
		{
			name:     "typed slice - last element",
			template: `{{ posts[4].title }}`,
			posts:    posts,
			expected: "",
		},
		{
			name:     "typed slice - out of bounds",
			template: `{{ posts[10].title }}`,
			posts:    posts,
			expected: "",
		},
		{
			name:     "baseline - []interface{} - first element",
			template: `{{ posts[0].title }}`,
			posts:    BlogPostsToInterfaceSlice(posts),
			expected: "",
		},
		{
			name:     "typed slice - access author property",
			template: `{{ posts[0].author }}`,
			posts:    posts,
			expected: "",
		},
		{
			name:     "typed slice - access comments_count property",
			template: `{{ posts[0].comments_count }}`,
			posts:    posts,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assigns := map[string]interface{}{
				"posts": tt.posts,
			}
			assertTemplateResult(t, tt.expected, tt.template, assigns)
		})
	}
}

// TestConditionBlogPostEmpty tests the .empty method literal with []BlogPost.
// This tests the issue in liquid/condition.go lines 206-212 where .empty only
// checks []interface{} for emptiness.
func TestConditionBlogPostEmpty(t *testing.T) {
	posts := SampleBlogPosts()
	emptyPosts := []BlogPost{}

	tests := []struct {
		name     string
		template string
		posts    interface{}
		expected string
	}{
		{
			name:     "typed slice - non-empty",
			template: `{% if posts.empty %}empty{% else %}not empty{% endif %}`,
			posts:    posts,
			expected: "not empty",
		},
		{
			name:     "typed slice - empty",
			template: `{% if posts.empty %}empty{% else %}not empty{% endif %}`,
			posts:    emptyPosts,
			expected: "empty",
		},
		{
			name:     "baseline - []interface{} - non-empty",
			template: `{% if posts.empty %}empty{% else %}not empty{% endif %}`,
			posts:    BlogPostsToInterfaceSlice(posts),
			expected: "not empty",
		},
		{
			name:     "baseline - []interface{} - empty",
			template: `{% if posts.empty %}empty{% else %}not empty{% endif %}`,
			posts:    []interface{}{},
			expected: "empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assigns := map[string]interface{}{
				"posts": tt.posts,
			}
			assertTemplateResult(t, tt.expected, tt.template, assigns)
		})
	}
}

// contains on a list compares elements for equality, as Ruby's Array#include?: a post
// is not equal to its title, so the title is found with map first.
// A BlogPost is a plain Go struct, not a drop: map reads none of its
// properties, so no title or author is found.
func TestConditionBlogPostContains(t *testing.T) {
	posts := SampleBlogPosts()

	tests := []struct {
		name     string
		template string
		posts    interface{}
		expected string
	}{
		{
			name:     "typed slice - a title is not a post",
			template: `{% if posts contains "Getting Started with Go" %}found{% else %}not found{% endif %}`,
			posts:    posts,
			expected: "not found",
		},
		{
			name:     "typed slice - contains title after map",
			template: `{% assign titles = posts | map: "title" %}{% if titles contains "Getting Started with Go" %}found{% else %}not found{% endif %}`,
			posts:    posts,
			expected: "not found",
		},
		{
			name:     "typed slice - contains author after map",
			template: `{% assign authors = posts | map: "author" %}{% if authors contains "Alice" %}found{% else %}not found{% endif %}`,
			posts:    posts,
			expected: "not found",
		},
		{
			name:     "typed slice - no substring of a title",
			template: `{% assign titles = posts | map: "title" %}{% if titles contains "Getting Started" %}found{% else %}not found{% endif %}`,
			posts:    posts,
			expected: "not found",
		},
		{
			name:     "[]interface{} - contains title after map",
			template: `{% assign titles = posts | map: "title" %}{% if titles contains "Getting Started with Go" %}found{% else %}not found{% endif %}`,
			posts:    BlogPostsToInterfaceSlice(posts),
			expected: "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assigns := map[string]interface{}{
				"posts": tt.posts,
			}
			assertTemplateResult(t, tt.expected, tt.template, assigns)
		})
	}
}

// TestNilComparisons tests nil comparisons matching Shopify Liquid behavior
// Reference: reference-liquid/test/integration/tags/if_else_tag_test.rb:122-132
func TestNilComparisons(t *testing.T) {
	env := mustEnv(liquid.EnvironmentConfig{Tags: tags.Standard()})

	tests := []struct {
		name     string
		template string
		expected string
		data     map[string]interface{}
	}{
		{
			name:     "null less than number",
			template: "{% if null < 10 %} NO {% endif %}",
			expected: "",
			data:     nil,
		},
		{
			name:     "null less than or equal number",
			template: "{% if null <= 10 %} NO {% endif %}",
			expected: "",
			data:     nil,
		},
		{
			name:     "null greater than or equal number",
			template: "{% if null >= 10 %} NO {% endif %}",
			expected: "",
			data:     nil,
		},
		{
			name:     "null greater than number",
			template: "{% if null > 10 %} NO {% endif %}",
			expected: "",
			data:     nil,
		},
		{
			name:     "number less than null",
			template: "{% if 10 < null %} NO {% endif %}",
			expected: "",
			data:     nil,
		},
		{
			name:     "number less than or equal null",
			template: "{% if 10 <= null %} NO {% endif %}",
			expected: "",
			data:     nil,
		},
		{
			name:     "number greater than or equal null",
			template: "{% if 10 >= null %} NO {% endif %}",
			expected: "",
			data:     nil,
		},
		{
			name:     "number greater than null",
			template: "{% if 10 > null %} NO {% endif %}",
			expected: "",
			data:     nil,
		},
		{
			name:     "null less than or equal zero",
			template: "{% if null <= 0 %} true {% else %} false {% endif %}",
			expected: " false ",
			data:     nil,
		},
		{
			name:     "zero less than or equal null",
			template: "{% if 0 <= null %} true {% else %} false {% endif %}",
			expected: " false ",
			data:     nil,
		},
		{
			name:     "nil variable comparison",
			template: "{% if pagination.total_pages > 1 %} YES {% else %} NO {% endif %}",
			expected: " NO ",
			data:     map[string]interface{}{"pagination": nil},
		},
		{
			name:     "nil nested property comparison",
			template: "{% if pagination.total_pages > 1 %} YES {% else %} NO {% endif %}",
			expected: " NO ",
			data:     map[string]interface{}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl, err := liquid.ParseTemplate(tt.template, &liquid.TemplateOptions{Environment: env})
			if err != nil {
				t.Fatalf("ParseTemplate() error = %v", err)
			}

			result := renderRuby(tmpl, tt.data, &rubyOptions{})
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}
