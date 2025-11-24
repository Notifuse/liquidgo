package integration

import (
	"strings"
	"testing"
	"time"

	"github.com/Notifuse/liquidgo/liquid"
	"github.com/Notifuse/liquidgo/liquid/tags"
)

func TestInstanceAssignsPersistOnSameTemplateObjectBetweenParses(t *testing.T) {
	env := liquid.NewEnvironment()
	tags.RegisterStandardTags(env)
	tmpl := liquid.NewTemplate(&liquid.TemplateOptions{Environment: env})

	// First parse and render
	err := tmpl.Parse(`{% assign foo = 'from instance assigns' %}{{ foo }}`, &liquid.TemplateOptions{Environment: env})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	result := tmpl.RenderBang(nil, &liquid.RenderOptions{})
	if result != "from instance assigns" {
		t.Errorf("Expected 'from instance assigns', got %q", result)
	}

	// Second parse and render - instance assigns should persist
	err = tmpl.Parse(`{{ foo }}`, &liquid.TemplateOptions{Environment: env})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	result = tmpl.RenderBang(nil, &liquid.RenderOptions{})
	if result != "from instance assigns" {
		t.Errorf("Expected 'from instance assigns', got %q", result)
	}
}

func TestWarningsIsNotExponentialTime(t *testing.T) {
	str := "false"
	for i := 0; i < 100; i++ {
		str = "{% if true %}true{% else %}" + str + "{% endif %}"
	}

	env := liquid.NewEnvironment()
	tags.RegisterStandardTags(env)
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

func TestInstanceAssignsPersistOnSameTemplateParsingBetweenRenders(t *testing.T) {
	env := liquid.NewEnvironment()
	tags.RegisterStandardTags(env)
	tmpl := liquid.NewTemplate(&liquid.TemplateOptions{Environment: env})

	err := tmpl.Parse(`{{ foo }}{% assign foo = 'foo' %}{{ foo }}`, &liquid.TemplateOptions{Environment: env})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// First render
	result := tmpl.RenderBang(nil, &liquid.RenderOptions{})
	if result != "foo" {
		t.Errorf("Expected 'foo', got %q", result)
	}

	// Second render - should include both
	result = tmpl.RenderBang(nil, &liquid.RenderOptions{})
	if result != "foofoo" {
		t.Errorf("Expected 'foofoo', got %q", result)
	}
}

func TestCustomAssignsDoNotPersistOnSameTemplate(t *testing.T) {
	env := liquid.NewEnvironment()
	tags.RegisterStandardTags(env)
	tmpl := liquid.NewTemplate(&liquid.TemplateOptions{Environment: env})

	err := tmpl.Parse(`{{ foo }}`, &liquid.TemplateOptions{Environment: env})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Render with custom assigns
	result := tmpl.RenderBang(map[string]interface{}{"foo": "from custom assigns"}, &liquid.RenderOptions{})
	if result != "from custom assigns" {
		t.Errorf("Expected 'from custom assigns', got %q", result)
	}

	// Render without custom assigns - should be empty
	result = tmpl.RenderBang(nil, &liquid.RenderOptions{})
	if result != "" {
		t.Errorf("Expected empty string, got %q", result)
	}
}

func TestCustomAssignsSquashInstanceAssigns(t *testing.T) {
	env := liquid.NewEnvironment()
	tags.RegisterStandardTags(env)
	tmpl := liquid.NewTemplate(&liquid.TemplateOptions{Environment: env})

	err := tmpl.Parse(`{% assign foo = 'from instance assigns' %}{{ foo }}`, &liquid.TemplateOptions{Environment: env})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// First render with instance assigns
	result := tmpl.RenderBang(nil, &liquid.RenderOptions{})
	if result != "from instance assigns" {
		t.Errorf("Expected 'from instance assigns', got %q", result)
	}

	// Render with custom assigns - should squash instance assigns
	result = tmpl.RenderBang(map[string]interface{}{"foo": "from custom assigns"}, &liquid.RenderOptions{})
	if result != "from custom assigns" {
		t.Errorf("Expected 'from custom assigns', got %q", result)
	}
}

func TestPersistentAssignsSquashInstanceAssigns(t *testing.T) {
	env := liquid.NewEnvironment()
	tags.RegisterStandardTags(env)
	tmpl := liquid.NewTemplate(&liquid.TemplateOptions{Environment: env})

	err := tmpl.Parse(`{% assign foo = 'from instance assigns' %}{{ foo }}`, &liquid.TemplateOptions{Environment: env})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// First render with instance assigns
	result := tmpl.RenderBang(nil, &liquid.RenderOptions{})
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
	result = tmpl.RenderBang(nil, &liquid.RenderOptions{})
	if result != "from persistent assigns" {
		t.Errorf("Expected 'from persistent assigns', got %q", result)
	}
}

func TestResourceLimitsWorksWithCustomLengthMethod(t *testing.T) {
	env := liquid.NewEnvironment()
	tags.RegisterStandardTags(env)
	tmpl, err := liquid.ParseTemplate(`{% assign foo = bar %}`, &liquid.TemplateOptions{Environment: env})
	if err != nil {
		t.Fatalf("ParseTemplate() error = %v", err)
	}

	limit := 42
	tmpl.SetResourceLimits(liquid.NewResourceLimits(liquid.ResourceLimitsConfig{
		RenderLengthLimit: &limit,
	}))
	somethingWithLength := NewSomethingWithLength()
	result := tmpl.RenderBang(map[string]interface{}{"bar": somethingWithLength}, &liquid.RenderOptions{})
	if result != "" {
		t.Errorf("Expected empty string, got %q", result)
	}
}

func TestResourceLimitsRenderLength(t *testing.T) {
	env := liquid.NewEnvironment()
	tags.RegisterStandardTags(env)
	tmpl, err := liquid.ParseTemplate("0123456789", &liquid.TemplateOptions{Environment: env})
	if err != nil {
		t.Fatalf("ParseTemplate() error = %v", err)
	}

	limit := 9
	tmpl.SetResourceLimits(liquid.NewResourceLimits(liquid.ResourceLimitsConfig{
		RenderLengthLimit: &limit,
	}))
	result := tmpl.Render(nil, &liquid.RenderOptions{})
	if !strings.Contains(result, "Memory limits exceeded") {
		t.Errorf("Expected memory limit error, got %q", result)
	}
	if !tmpl.ResourceLimits().Reached() {
		t.Error("Expected resource limits to be reached")
	}

	limit = 10
	tmpl.SetResourceLimits(liquid.NewResourceLimits(liquid.ResourceLimitsConfig{
		RenderLengthLimit: &limit,
	}))
	result = tmpl.RenderBang(nil, &liquid.RenderOptions{})
	if result != "0123456789" {
		t.Errorf("Expected '0123456789', got %q", result)
	}
}

func TestResourceLimitsRenderScore(t *testing.T) {
	env := liquid.NewEnvironment()
	tags.RegisterStandardTags(env)

	// Test nested loops
	tmpl, err := liquid.ParseTemplate(`{% for a in (1..10) %} {% for a in (1..10) %} foo {% endfor %} {% endfor %}`, &liquid.TemplateOptions{Environment: env})
	if err != nil {
		t.Fatalf("ParseTemplate() error = %v", err)
	}

	limit := 50
	tmpl.SetResourceLimits(liquid.NewResourceLimits(liquid.ResourceLimitsConfig{
		RenderScoreLimit: &limit,
	}))
	result := tmpl.Render(nil, &liquid.RenderOptions{})
	if !strings.Contains(result, "Memory limits exceeded") {
		t.Errorf("Expected memory limit error, got %q", result)
	}
	if !tmpl.ResourceLimits().Reached() {
		t.Error("Expected resource limits to be reached")
	}

	// Test single loop
	tmpl, err = liquid.ParseTemplate(`{% for a in (1..100) %} foo {% endfor %}`, &liquid.TemplateOptions{Environment: env})
	if err != nil {
		t.Fatalf("ParseTemplate() error = %v", err)
	}

	limit = 50
	tmpl.SetResourceLimits(liquid.NewResourceLimits(liquid.ResourceLimitsConfig{
		RenderScoreLimit: &limit,
	}))
	result = tmpl.Render(nil, &liquid.RenderOptions{})
	if !strings.Contains(result, "Memory limits exceeded") {
		t.Errorf("Expected memory limit error, got %q", result)
	}
	if !tmpl.ResourceLimits().Reached() {
		t.Error("Expected resource limits to be reached")
	}

	limit = 200
	tmpl.SetResourceLimits(liquid.NewResourceLimits(liquid.ResourceLimitsConfig{
		RenderScoreLimit: &limit,
	}))
	expected := strings.Repeat(" foo ", 100)
	result = tmpl.RenderBang(nil, &liquid.RenderOptions{})
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
	if tmpl.ResourceLimits().RenderScore() == 0 {
		t.Error("Expected render_score to be set")
	}
}

func TestResourceLimitsAbortsRenderingAfterFirstError(t *testing.T) {
	env := liquid.NewEnvironment()
	tags.RegisterStandardTags(env)
	tmpl, err := liquid.ParseTemplate(`{% for a in (1..100) %} foo1 {% endfor %} bar {% for a in (1..100) %} foo2 {% endfor %}`, &liquid.TemplateOptions{Environment: env})
	if err != nil {
		t.Fatalf("ParseTemplate() error = %v", err)
	}

	limit := 50
	tmpl.SetResourceLimits(liquid.NewResourceLimits(liquid.ResourceLimitsConfig{
		RenderScoreLimit: &limit,
	}))
	result := tmpl.Render(nil, &liquid.RenderOptions{})
	if !strings.Contains(result, "Memory limits exceeded") {
		t.Errorf("Expected memory limit error, got %q", result)
	}
	if !tmpl.ResourceLimits().Reached() {
		t.Error("Expected resource limits to be reached")
	}
}

func TestResourceLimitsHashInTemplateGetsUpdatedEvenIfNoLimitsAreSet(t *testing.T) {
	env := liquid.NewEnvironment()
	tags.RegisterStandardTags(env)
	tmpl, err := liquid.ParseTemplate(`{% for a in (1..100) %}x{% assign foo = 1 %} {% endfor %}`, &liquid.TemplateOptions{Environment: env})
	if err != nil {
		t.Fatalf("ParseTemplate() error = %v", err)
	}

	tmpl.RenderBang(nil, &liquid.RenderOptions{})
	if tmpl.ResourceLimits().AssignScore() <= 0 {
		t.Error("Expected assign_score to be greater than 0")
	}
	if tmpl.ResourceLimits().RenderScore() <= 0 {
		t.Error("Expected render_score to be greater than 0")
	}
}

func TestRenderLengthPersistsBetweenBlocks(t *testing.T) {
	env := liquid.NewEnvironment()
	tags.RegisterStandardTags(env)

	// Single block
	tmpl, err := liquid.ParseTemplate(`{% if true %}aaaa{% endif %}`, &liquid.TemplateOptions{Environment: env})
	if err != nil {
		t.Fatalf("ParseTemplate() error = %v", err)
	}

	limit := 3
	tmpl.SetResourceLimits(liquid.NewResourceLimits(liquid.ResourceLimitsConfig{
		RenderLengthLimit: &limit,
	}))
	result := tmpl.Render(nil, &liquid.RenderOptions{})
	if !strings.Contains(result, "Memory limits exceeded") {
		t.Errorf("Expected memory limit error, got %q", result)
	}

	limit = 4
	tmpl.SetResourceLimits(liquid.NewResourceLimits(liquid.ResourceLimitsConfig{
		RenderLengthLimit: &limit,
	}))
	result = tmpl.Render(nil, &liquid.RenderOptions{})
	if result != "aaaa" {
		t.Errorf("Expected 'aaaa', got %q", result)
	}

	// Multiple blocks
	tmpl, err = liquid.ParseTemplate(`{% if true %}aaaa{% endif %}{% if true %}bbb{% endif %}`, &liquid.TemplateOptions{Environment: env})
	if err != nil {
		t.Fatalf("ParseTemplate() error = %v", err)
	}

	limit = 6
	tmpl.SetResourceLimits(liquid.NewResourceLimits(liquid.ResourceLimitsConfig{
		RenderLengthLimit: &limit,
	}))
	result = tmpl.Render(nil, &liquid.RenderOptions{})
	if !strings.Contains(result, "Memory limits exceeded") {
		t.Errorf("Expected memory limit error, got %q", result)
	}

	limit = 7
	tmpl.SetResourceLimits(liquid.NewResourceLimits(liquid.ResourceLimitsConfig{
		RenderLengthLimit: &limit,
	}))
	result = tmpl.Render(nil, &liquid.RenderOptions{})
	if result != "aaaabbb" {
		t.Errorf("Expected 'aaaabbb', got %q", result)
	}

	// Many blocks
	tmpl, err = liquid.ParseTemplate(`{% if true %}a{% endif %}{% if true %}b{% endif %}{% if true %}a{% endif %}{% if true %}b{% endif %}{% if true %}a{% endif %}{% if true %}b{% endif %}`, &liquid.TemplateOptions{Environment: env})
	if err != nil {
		t.Fatalf("ParseTemplate() error = %v", err)
	}

	limit = 5
	tmpl.SetResourceLimits(liquid.NewResourceLimits(liquid.ResourceLimitsConfig{
		RenderLengthLimit: &limit,
	}))
	result = tmpl.Render(nil, &liquid.RenderOptions{})
	if !strings.Contains(result, "Memory limits exceeded") {
		t.Errorf("Expected memory limit error, got %q", result)
	}

	limit = 6
	tmpl.SetResourceLimits(liquid.NewResourceLimits(liquid.ResourceLimitsConfig{
		RenderLengthLimit: &limit,
	}))
	result = tmpl.Render(nil, &liquid.RenderOptions{})
	if result != "ababab" {
		t.Errorf("Expected 'ababab', got %q", result)
	}
}

func TestRenderLengthUsesNumberOfBytesNotCharacters(t *testing.T) {
	env := liquid.NewEnvironment()
	tags.RegisterStandardTags(env)
	tmpl, err := liquid.ParseTemplate(`{% if true %}すごい{% endif %}`, &liquid.TemplateOptions{Environment: env})
	if err != nil {
		t.Fatalf("ParseTemplate() error = %v", err)
	}

	// すごい is 9 bytes in UTF-8
	limit := 8
	tmpl.SetResourceLimits(liquid.NewResourceLimits(liquid.ResourceLimitsConfig{
		RenderLengthLimit: &limit,
	}))
	result := tmpl.Render(nil, &liquid.RenderOptions{})
	if !strings.Contains(result, "Memory limits exceeded") {
		t.Errorf("Expected memory limit error, got %q", result)
	}

	limit = 9
	tmpl.SetResourceLimits(liquid.NewResourceLimits(liquid.ResourceLimitsConfig{
		RenderLengthLimit: &limit,
	}))
	result = tmpl.Render(nil, &liquid.RenderOptions{})
	if result != "すごい" {
		t.Errorf("Expected 'すごい', got %q", result)
	}
}

func TestDefaultResourceLimitsUnaffectedByRenderWithContext(t *testing.T) {
	env := liquid.NewEnvironment()
	tags.RegisterStandardTags(env)
	ctx := liquid.BuildContext(liquid.ContextConfig{Environment: env})
	tmpl, err := liquid.ParseTemplate(`{% for a in (1..100) %}x{% assign foo = 1 %} {% endfor %}`, &liquid.TemplateOptions{Environment: env})
	if err != nil {
		t.Fatalf("ParseTemplate() error = %v", err)
	}

	tmpl.RenderBang(ctx, &liquid.RenderOptions{})
	if ctx.ResourceLimits().AssignScore() <= 0 {
		t.Error("Expected assign_score to be greater than 0")
	}
	if ctx.ResourceLimits().RenderScore() <= 0 {
		t.Error("Expected render_score to be greater than 0")
	}
}

func TestCanUseDropAsContext(t *testing.T) {
	env := liquid.NewEnvironment()
	tags.RegisterStandardTags(env)
	tmpl := liquid.NewTemplate(&liquid.TemplateOptions{Environment: env})
	tmpl.Registers()["lulz"] = "haha"

	drop := NewTemplateContextDrop()
	drop.SetContext(liquid.BuildContext(liquid.ContextConfig{Environment: env}))

	err := tmpl.Parse(`{{foo}}`, &liquid.TemplateOptions{Environment: env})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	result := tmpl.RenderBang(drop, &liquid.RenderOptions{})
	if result != "fizzbuzz" {
		t.Errorf("Expected 'fizzbuzz', got %q", result)
	}

	err = tmpl.Parse(`{{bar}}`, &liquid.TemplateOptions{Environment: env})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	result = tmpl.RenderBang(drop, &liquid.RenderOptions{})
	if result != "bar" {
		t.Errorf("Expected 'bar', got %q", result)
	}

	// Set registers on context for baz test
	ctx := liquid.BuildContext(liquid.ContextConfig{
		Environment: env,
		Registers:   liquid.NewRegisters(map[string]interface{}{"lulz": "haha"}),
	})
	drop.SetContext(ctx)

	err = tmpl.Parse(`{{baz}}`, &liquid.TemplateOptions{Environment: env})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	result = tmpl.RenderBang(ctx, &liquid.RenderOptions{})
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
	env := liquid.NewEnvironment()
	tags.RegisterStandardTags(env)

	// Test nil
	tmpl, err := liquid.ParseTemplate("", &liquid.TemplateOptions{Environment: env})
	if err != nil {
		t.Fatalf("ParseTemplate() error = %v", err)
	}
	result := tmpl.Render(nil, &liquid.RenderOptions{})
	if result != "" {
		t.Errorf("Expected empty string, got %q", result)
	}

	// Test integer (converted to string)
	tmpl, err = liquid.ParseTemplate("1", &liquid.TemplateOptions{Environment: env})
	if err != nil {
		t.Fatalf("ParseTemplate() error = %v", err)
	}
	result = tmpl.Render(nil, &liquid.RenderOptions{})
	if result != "1" {
		t.Errorf("Expected '1', got %q", result)
	}

	// Test boolean (converted to string)
	tmpl, err = liquid.ParseTemplate("true", &liquid.TemplateOptions{Environment: env})
	if err != nil {
		t.Fatalf("ParseTemplate() error = %v", err)
	}
	result = tmpl.Render(nil, &liquid.RenderOptions{})
	if result != "true" {
		t.Errorf("Expected 'true', got %q", result)
	}
}

// TestVariableLookupBlogPostArrayIndex tests array index access with []BlogPost.
// This tests the issue in liquid/variable_lookup.go lines 189-197 where array index
// access only works with []interface{} but not with typed slices.
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
			expected: "Getting Started with Go",
		},
		{
			name:     "typed slice - second element",
			template: `{{ posts[1].title }}`,
			posts:    posts,
			expected: "Advanced Liquid Templates",
		},
		{
			name:     "typed slice - last element",
			template: `{{ posts[4].title }}`,
			posts:    posts,
			expected: "Performance Optimization",
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
			expected: "Getting Started with Go",
		},
		{
			name:     "typed slice - access author property",
			template: `{{ posts[0].author }}`,
			posts:    posts,
			expected: "Alice",
		},
		{
			name:     "typed slice - access comments_count property",
			template: `{{ posts[0].comments_count }}`,
			posts:    posts,
			expected: "5",
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

// TestConditionBlogPostContains tests the contains operator with []BlogPost.
// This tests the issue in liquid/condition.go lines 274-280 where contains
// only works with []interface{}.
func TestConditionBlogPostContains(t *testing.T) {
	posts := SampleBlogPosts()

	tests := []struct {
		name     string
		template string
		posts    interface{}
		expected string
	}{
		{
			name:     "typed slice - contains title",
			template: `{% if posts contains "Getting Started with Go" %}found{% else %}not found{% endif %}`,
			posts:    posts,
			expected: "found",
		},
		{
			name:     "typed slice - contains author",
			template: `{% if posts contains "Alice" %}found{% else %}not found{% endif %}`,
			posts:    posts,
			expected: "found",
		},
		{
			name:     "typed slice - does not contain",
			template: `{% if posts contains "Non-existent Post" %}found{% else %}not found{% endif %}`,
			posts:    posts,
			expected: "not found",
		},
		{
			name:     "baseline - []interface{} - contains title",
			template: `{% if posts contains "Getting Started with Go" %}found{% else %}not found{% endif %}`,
			posts:    BlogPostsToInterfaceSlice(posts),
			expected: "found",
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
	env := liquid.NewEnvironment()
	tags.RegisterStandardTags(env)

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

			result := tmpl.Render(tt.data, &liquid.RenderOptions{})
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}
