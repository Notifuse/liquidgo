package integration

import (
	"os"
	"strings"
	"testing"

	"github.com/Notifuse/liquidgo/liquid"
	"github.com/Notifuse/liquidgo/liquid/tags"
)

// TestComprehensiveAllTags tests all available Liquid tags with all their parameters
// This ensures all tags work correctly together and produce expected output
func TestComprehensiveAllTags(t *testing.T) {
	env := liquid.NewEnvironment()
	tags.RegisterStandardTags(env)

	// Create a file system for include/render tests
	fs := &mapFileSystem{
		templates: map[string]string{
			"header":        "=== {{ title }} ===",
			"footer":        "--- End ---",
			"product":       "Product: {{ product.name }} - ${{ product.price }}",
			"list_item":     "[{{ forloop.index }}] {{ item }}",
			"priced_item":   "{{ name }} (${{ price }})",
			"user_card":     "User: {{ user }} | Role: {{ role }}",
			"with_product":  "Product: {{ product }}",
			"with_variable": "Value: {{ with_variable }}",
		},
	}
	env.SetFileSystem(fs)

	template := `
{%- comment -%}
COMPREHENSIVE LIQUID TAGS TEST
This template tests all available registered tags with their parameters
{%- endcomment -%}

{%- assign site_name = "Liquid Store" -%}
{%- assign products_count = 10 -%}

{%- capture page_header -%}
{{ site_name | upcase }}
{%- endcapture -%}

{% include 'header' with title: page_header %}

=== ASSIGN & ECHO ===
{% echo "Store: " %}{{ site_name }}

=== FOR LOOP WITH ALL PARAMETERS ===
Full range (1-5):
{%- for num in (1..5) %}
 {{ num }}
{%- endfor %}

With offset=1, limit=3:
{%- for num in (1..10) offset:1 limit:3 %}
 {{ num }}
{%- endfor %}

Reversed:
{%- for num in (1..4) reversed %}
 {{ num }}
{%- endfor %}

For loop variables:
{%- assign items = "A,B,C" | split: "," -%}
{%- for item in items %}
[{{ forloop.index }}:{{ forloop.index0 }}] {{ item }} first={{ forloop.first }} last={{ forloop.last }}
{%- endfor %}

=== BREAK & CONTINUE ===
With break:
{%- for i in (1..10) %}
{%- if i == 4 %}{% break %}{% endif %}
 {{ i }}
{%- endfor %}

With continue:
{%- for i in (1..6) %}
{%- if i == 3 %}{% continue %}{% endif %}
 {{ i }}
{%- endfor %}

=== CASE WITH WHEN ===
{%- assign status = 1 -%}
{%- case status -%}
{%- when 1 %}
Status: One
{% when 2 %}
Status: Two
{% else %}
Status: Other
{% endcase %}

Case with comma-separated:
{%- assign val = 2 -%}
{%- case val -%}
{%- when 1, 2, 3 %}
One Two or Three
{% else %}
Other
{% endcase %}

Case with or:
{%- assign val2 = 5 -%}
{%- case val2 -%}
{%- when 4 or 5 or 6 %}
Four Five or Six
{% else %}
Other
{% endcase %}

=== IF/ELSIF/ELSE ===
{%- assign score = 85 -%}
{%- assign premium = true -%}
{%- if score >= 90 and premium %}
Grade: A+ Premium
{% elsif score >= 80 and premium %}
Grade: A Premium
{% elsif score >= 90 %}
Grade: A
{% else %}
Grade: B
{% endif %}

IF with all operators:
{%- assign x = 10 -%}
{%- assign y = 5 -%}
{%- assign text = "hello world" %}
{% if x > y -%}x > y: true{% endif -%}
{% if x < 20 %} x < 20: true{% endif -%}
{% if y <= 5 %} y <= 5: true{% endif -%}
{% if x != y %} x != y: true{% endif -%}
{% if x == 10 or y == 10 %} or: true{% endif -%}
{% if text contains "world" %} contains: true{% endif %}

=== UNLESS ===
{%- assign stock = 5 -%}
{%- unless stock == 0 %}
In Stock: {{ stock }}
{% endunless %}
{% unless stock > 10 %}
Limited Stock
{% endunless %}

Unless with operators:
{%- assign available = true -%}
{%- assign tags = "sale,new" %}
{% unless stock < 1 or available == false -%}Available{% endunless -%}{% unless tags contains "discontinued" %} Not discontinued{% endunless %}

=== CYCLE WITH GROUPS ===
{% for i in (1..6) -%}
{{ i }}: {% cycle 'g1': 'red', 'green', 'blue' %} / {% cycle 'g2': 'odd', 'even' %}
{% endfor %}

=== IFCHANGED ===
{%- assign sequence = "1,1,2,2,3,3" | split: "," -%}
{% for val in sequence -%}
{% ifchanged %}[{{ val }}]{% endifchanged -%}
{% endfor %}

=== TABLEROW WITH PARAMETERS ===
Standard (cols=3):
{% tablerow num in (1..6) cols:3 -%}
C{{ num }}
{% endtablerow %}

With offset=1, limit=4, cols=2:
{% tablerow num in (1..20) cols:2 offset:1 limit:4 -%}
{{ num }}
{% endtablerow %}

TableRow loop variables:
{% tablerow i in (1..4) cols:2 -%}
[{{ tablerowloop.index }}:{{ tablerowloop.col }}] {{ i }}
{% endtablerow %}

=== INCREMENT & DECREMENT ===
Inc: {% increment c1 %} {% increment c1 %} {% increment c1 %}
Dec: {% decrement c1 %} {% decrement c1 %}
Inc2: {% increment c2 %} {% increment c2 %}

=== RENDER WITH PARAMETERS ===
{% render 'priced_item', name: 'Widget', price: 49.99 %}
{% render 'user_card', user: 'Alice', role: 'Admin' %}

Render with for:
{% assign users = "Bob,Charlie" | split: "," -%}
{% render 'user_card' for users as user, role: 'User' %}

Render with 'with':
{% assign my_product = "Laptop" -%}
{% render 'with_product' with my_product as product %}

=== INCLUDE WITH PARAMETERS ===
{% include 'priced_item', name: 'Gadget', price: 29.99 %}

Include with for:
{% assign products = "Laptop,Mouse" | split: "," -%}
{% include 'list_item' for products as item %}

Include with 'with':
{% assign data = "TestData" -%}
{% include 'with_variable' with data %}

=== NESTED CONTROL FLOW ===
{% for x in (1..3) -%}
Group {{ x }}:
{% for y in (1..2) %}{%- if x == y -%}{% continue -%}{% endif %}
 - {{ x }},{{ y }}
{% endfor %}
{% endfor %}

=== COMPLEX CAPTURE WITH FILTERS ===
{% capture result -%}
{%- assign sum = 0 -%}
{%- for i in (1..5) -%}
{%- assign sum = sum | plus: i -%}
{%- endfor -%}
Sum: {{ sum }} | Doubled: {{ sum | times: 2 }}
{% endcapture %}
{{ result | strip }}

=== RAW TAG (NO PARSING) ===
{% raw %}{{ variable }} {% if true %}not parsed{% endif %}{% endraw %}

{% # This is an inline comment and should not appear in output -%}
=== AFTER INLINE COMMENT ===

=== DOC TAG ===
{% doc -%}
This is documentation that should be blank
{% enddoc -%}
After doc

{% include 'footer' %}
`

	expected := `=== LIQUID STORE ===

=== ASSIGN & ECHO ===
Store: Liquid Store

=== FOR LOOP WITH ALL PARAMETERS ===
Full range (1-5):
 1
 2
 3
 4
 5

With offset=1, limit=3:
 2
 3
 4

Reversed:
 4
 3
 2
 1

For loop variables:
[1:0] A first=true last=false
[2:1] B first=false last=false
[3:2] C first=false last=true

=== BREAK & CONTINUE ===
With break:
 1
 2
 3

With continue:
 1
 2
 4
 5
 6

=== CASE WITH WHEN ===
Status: One


Case with comma-separated:
One Two or Three


Case with or:
Four Five or Six


=== IF/ELSIF/ELSE ===
Grade: A Premium


IF with all operators:
x > y: true x < 20: true y <= 5: true x != y: true or: true contains: true

=== UNLESS ===
In Stock: 5


Limited Stock


Unless with operators:
Available Not discontinued

=== CYCLE WITH GROUPS ===
1: red / odd
2: green / even
3: blue / odd
4: red / even
5: green / odd
6: blue / even


=== IFCHANGED ===[1][2][3]

=== TABLEROW WITH PARAMETERS ===
Standard (cols=3):
<tr class="row1">
<td class="col1">C1
</td><td class="col2">C2
</td><td class="col3">C3
</td></tr>
<tr class="row2"><td class="col1">C4
</td><td class="col2">C5
</td><td class="col3">C6
</td></tr>


With offset=1, limit=4, cols=2:
<tr class="row1">
<td class="col1">2
</td><td class="col2">3
</td></tr>
<tr class="row2"><td class="col1">4
</td><td class="col2">5
</td></tr>


TableRow loop variables:
<tr class="row1">
<td class="col1">[1:1] 1
</td><td class="col2">[2:2] 2
</td></tr>
<tr class="row2"><td class="col1">[3:1] 3
</td><td class="col2">[4:2] 4
</td></tr>


=== INCREMENT & DECREMENT ===
Inc: 0 1 2
Dec: 2 1
Inc2: 0 1

=== RENDER WITH PARAMETERS ===
Widget ($49.99)
User: Alice | Role: Admin

Render with for:
User: Bob | Role: UserUser: Charlie | Role: User

Render with 'with':
Product: Laptop

=== INCLUDE WITH PARAMETERS ===
Gadget ($29.99)

Include with for:
[] Laptop[] Mouse

Include with 'with':
Value: TestData

=== NESTED CONTROL FLOW ===
Group 1:

 - 1,2

Group 2:

 - 2,1

Group 3:

 - 3,1

 - 3,2



=== COMPLEX CAPTURE WITH FILTERS ===

Sum: 15 | Doubled: 30

=== RAW TAG (NO PARSING) ===
{{ variable }} {% if true %}not parsed{% endif %}

=== AFTER INLINE COMMENT ===

=== DOC TAG ===
After doc

--- End ---
`

	tmpl, err := liquid.ParseTemplate(template, &liquid.TemplateOptions{Environment: env})
	if err != nil {
		t.Fatalf("ParseTemplate() error = %v", err)
	}

	result := tmpl.RenderBang(nil, &liquid.RenderOptions{})

	// DEBUG: Write outputs to files for comparison
	_ = os.WriteFile("/tmp/liquidgo_actual.txt", []byte(result), 0644)
	_ = os.WriteFile("/tmp/liquidgo_expected.txt", []byte(expected), 0644)

	if result != expected {
		t.Errorf("Output mismatch\n\n")
		t.Logf("DEBUG: Run 'diff /tmp/liquidgo_expected.txt /tmp/liquidgo_actual.txt' to see differences")
		// Show detailed difference
		expectedLines := strings.Split(expected, "\n")
		resultLines := strings.Split(result, "\n")
		maxLines := len(expectedLines)
		if len(resultLines) > maxLines {
			maxLines = len(resultLines)
		}

		diffCount := 0
		for i := 0; i < maxLines; i++ {
			var expLine, resLine string
			if i < len(expectedLines) {
				expLine = expectedLines[i]
			}
			if i < len(resultLines) {
				resLine = resultLines[i]
			}
			if expLine != resLine {
				diffCount++
				if diffCount <= 10 { // Show first 10 differences
					t.Errorf("Line %d differs:\n  Expected: %q\n  Got:      %q", i+1, expLine, resLine)
				}
			}
		}
		if diffCount > 10 {
			t.Errorf("... and %d more differences", diffCount-10)
		}

		t.Logf("\n=== EXPECTED OUTPUT ===\n%s\n", expected)
		t.Logf("\n=== ACTUAL OUTPUT ===\n%s\n", result)
	}
}

// TestComprehensiveFilteredTags tests tag combinations with filters
func TestComprehensiveFilteredTags(t *testing.T) {
	env := liquid.NewEnvironment()
	tags.RegisterStandardTags(env)

	template := `
{%- assign name = "john doe" -%}
{%- assign price = "99.99" -%}
{%- assign words = "hello world test" -%}

Capitalized: {{ name | capitalize }}
Upcase: {{ name | upcase }}
Downcase: {{ "HELLO" | downcase }}
Strip: {{ "  space  " | strip }}
Split and join: {{ words | split: " " | join: "-" }}
Number: {{ 99.99 | plus: 10 }}

{% capture filtered -%}
{{ name | upcase }}
{% endcapture %}
Captured: {{ filtered | strip }}
`

	expected := `Capitalized: John doe
Upcase: JOHN DOE
Downcase: hello
Strip: space
Split and join: hello-world-test
Number: 109.99


Captured: JOHN DOE
`

	tmpl, err := liquid.ParseTemplate(template, &liquid.TemplateOptions{Environment: env})
	if err != nil {
		t.Fatalf("ParseTemplate() error = %v", err)
	}

	result := tmpl.RenderBang(nil, &liquid.RenderOptions{})

	if result != expected {
		t.Errorf("Output mismatch\n\nExpected:\n%s\n\nGot:\n%s", expected, result)
	}
}

// mapFileSystem implements liquid.FileSystem for testing
type mapFileSystem struct {
	templates map[string]string
}

func (m *mapFileSystem) ReadTemplateFile(name string) (string, error) {
	if content, ok := m.templates[name]; ok {
		return content, nil
	}
	return "", liquid.NewFileSystemError("Template not found: " + name)
}

func (m *mapFileSystem) FullPath(name string) (string, error) {
	if _, ok := m.templates[name]; ok {
		return name, nil
	}
	return "", liquid.NewFileSystemError("Template not found: " + name)
}

// TestBlogPostSliceTypeCompatibility tests all tags and filters that process slices
// using a []BlogPost struct slice to verify slice type compatibility issues.
// This tests the issues identified in SLICE_TYPE_COMPATIBILITY_REPORT.md.

// TestBlogPostStandardFilters tests standard filters with []BlogPost.
func TestBlogPostStandardFilters(t *testing.T) {
	posts := SampleBlogPosts()
	emptyPosts := []BlogPost{}

	// Size filter
	t.Run("Size filter", func(t *testing.T) {
		assertTemplateResult(t, "5", `{{ posts | size }}`, map[string]interface{}{"posts": posts})
		assertTemplateResult(t, "0", `{{ posts | size }}`, map[string]interface{}{"posts": emptyPosts})
		assertTemplateResult(t, "5", `{{ posts | size }}`, map[string]interface{}{"posts": BlogPostsToInterfaceSlice(posts)})
	})

	// First filter
	t.Run("First filter", func(t *testing.T) {
		assertTemplateResult(t, "Getting Started with Go", `{{ posts | first | title }}`, map[string]interface{}{"posts": posts})
		assertTemplateResult(t, "Getting Started with Go", `{{ posts | first | title }}`, map[string]interface{}{"posts": BlogPostsToInterfaceSlice(posts)})
	})

	// Last filter
	t.Run("Last filter", func(t *testing.T) {
		assertTemplateResult(t, "Performance Optimization", `{{ posts | last | title }}`, map[string]interface{}{"posts": posts})
		assertTemplateResult(t, "Performance Optimization", `{{ posts | last | title }}`, map[string]interface{}{"posts": BlogPostsToInterfaceSlice(posts)})
	})

	// Join filter
	t.Run("Join filter", func(t *testing.T) {
		expected := "Getting Started with Go, Advanced Liquid Templates, Understanding Reflection, Testing Best Practices, Performance Optimization"
		assertTemplateResult(t, expected, `{{ posts | map: 'title' | join: ', ' }}`, map[string]interface{}{"posts": posts})
		assertTemplateResult(t, expected, `{{ posts | map: 'title' | join: ', ' }}`, map[string]interface{}{"posts": BlogPostsToInterfaceSlice(posts)})
	})

	// Slice filter
	t.Run("Slice filter", func(t *testing.T) {
		assertTemplateResult(t, "Getting Started with Go, Advanced Liquid Templates", `{{ posts | slice: 0, 2 | map: 'title' | join: ', ' }}`, map[string]interface{}{"posts": posts})
		assertTemplateResult(t, "Getting Started with Go, Advanced Liquid Templates", `{{ posts | slice: 0, 2 | map: 'title' | join: ', ' }}`, map[string]interface{}{"posts": BlogPostsToInterfaceSlice(posts)})
	})

	// Concat filter
	t.Run("Concat filter", func(t *testing.T) {
		posts1 := posts[:2]
		posts2 := posts[2:]
		assertTemplateResult(t, "5", `{{ posts1 | concat: posts2 | size }}`, map[string]interface{}{"posts1": posts1, "posts2": posts2})
		assertTemplateResult(t, "5", `{{ posts1 | concat: posts2 | size }}`, map[string]interface{}{"posts1": BlogPostsToInterfaceSlice(posts1), "posts2": BlogPostsToInterfaceSlice(posts2)})
	})

	// Reverse filter
	t.Run("Reverse filter", func(t *testing.T) {
		assertTemplateResult(t, "Performance Optimization", `{{ posts | reverse | first | title }}`, map[string]interface{}{"posts": posts})
		assertTemplateResult(t, "Performance Optimization", `{{ posts | reverse | first | title }}`, map[string]interface{}{"posts": BlogPostsToInterfaceSlice(posts)})
	})

	// Sort filter
	t.Run("Sort filter", func(t *testing.T) {
		assertTemplateResult(t, "Advanced Liquid Templates", `{{ posts | sort: 'title' | first | title }}`, map[string]interface{}{"posts": posts})
		assertTemplateResult(t, "Advanced Liquid Templates", `{{ posts | sort: 'title' | first | title }}`, map[string]interface{}{"posts": BlogPostsToInterfaceSlice(posts)})
	})

	// SortNatural filter
	t.Run("SortNatural filter", func(t *testing.T) {
		assertTemplateResult(t, "Advanced Liquid Templates", `{{ posts | sort_natural: 'title' | first | title }}`, map[string]interface{}{"posts": posts})
		assertTemplateResult(t, "Advanced Liquid Templates", `{{ posts | sort_natural: 'title' | first | title }}`, map[string]interface{}{"posts": BlogPostsToInterfaceSlice(posts)})
	})

	// Uniq filter
	t.Run("Uniq filter", func(t *testing.T) {
		duplicatePosts := []BlogPost{posts[0], posts[1], posts[0]}
		assertTemplateResult(t, "2", `{{ posts | uniq: 'author' | size }}`, map[string]interface{}{"posts": duplicatePosts})
		assertTemplateResult(t, "2", `{{ posts | uniq: 'author' | size }}`, map[string]interface{}{"posts": BlogPostsToInterfaceSlice(duplicatePosts)})
	})

	// Compact filter
	t.Run("Compact filter", func(t *testing.T) {
		assertTemplateResult(t, "5", `{{ posts | compact | size }}`, map[string]interface{}{"posts": posts})
		assertTemplateResult(t, "5", `{{ posts | compact | size }}`, map[string]interface{}{"posts": BlogPostsToInterfaceSlice(posts)})
	})

	// Map filter
	t.Run("Map filter", func(t *testing.T) {
		expected := "Getting Started with Go, Advanced Liquid Templates, Understanding Reflection, Testing Best Practices, Performance Optimization"
		assertTemplateResult(t, expected, `{{ posts | map: 'title' | join: ', ' }}`, map[string]interface{}{"posts": posts})
		assertTemplateResult(t, "Alice, Bob, Charlie, Alice, Bob", `{{ posts | map: 'author' | join: ', ' }}`, map[string]interface{}{"posts": posts})
		assertTemplateResult(t, expected, `{{ posts | map: 'title' | join: ', ' }}`, map[string]interface{}{"posts": BlogPostsToInterfaceSlice(posts)})
	})

	// Where filter
	t.Run("Where filter", func(t *testing.T) {
		assertTemplateResult(t, "4", `{{ posts | where: 'published', true | size }}`, map[string]interface{}{"posts": posts})
		assertTemplateResult(t, "2", `{{ posts | where: 'author', 'Alice' | size }}`, map[string]interface{}{"posts": posts})
		assertTemplateResult(t, "4", `{{ posts | where: 'published', true | size }}`, map[string]interface{}{"posts": BlogPostsToInterfaceSlice(posts)})
	})

	// Reject filter
	t.Run("Reject filter", func(t *testing.T) {
		assertTemplateResult(t, "1", `{{ posts | reject: 'published', true | size }}`, map[string]interface{}{"posts": posts})
		assertTemplateResult(t, "1", `{{ posts | reject: 'published', true | size }}`, map[string]interface{}{"posts": BlogPostsToInterfaceSlice(posts)})
	})

	// Has filter
	t.Run("Has filter", func(t *testing.T) {
		assertTemplateResult(t, "found", `{% if posts | has: 'author', 'Alice' %}found{% else %}not found{% endif %}`, map[string]interface{}{"posts": posts})
		assertTemplateResult(t, "not found", `{% if posts | has: 'author', 'David' %}found{% else %}not found{% endif %}`, map[string]interface{}{"posts": posts})
		assertTemplateResult(t, "found", `{% if posts | has: 'author', 'Alice' %}found{% else %}not found{% endif %}`, map[string]interface{}{"posts": BlogPostsToInterfaceSlice(posts)})
	})

	// Find filter
	t.Run("Find filter", func(t *testing.T) {
		assertTemplateResult(t, "Getting Started with Go", `{{ posts | find: 'author', 'Alice' | title }}`, map[string]interface{}{"posts": posts})
		assertTemplateResult(t, "Getting Started with Go", `{{ posts | find: 'author', 'Alice' | title }}`, map[string]interface{}{"posts": BlogPostsToInterfaceSlice(posts)})
	})

	// FindIndex filter
	t.Run("FindIndex filter", func(t *testing.T) {
		assertTemplateResult(t, "1", `{{ posts | find_index: 'author', 'Bob' }}`, map[string]interface{}{"posts": posts})
		assertTemplateResult(t, "1", `{{ posts | find_index: 'author', 'Bob' }}`, map[string]interface{}{"posts": BlogPostsToInterfaceSlice(posts)})
	})

	// Sum filter
	t.Run("Sum filter", func(t *testing.T) {
		assertTemplateResult(t, "43", `{{ posts | sum: 'comments_count' }}`, map[string]interface{}{"posts": posts})
		assertTemplateResult(t, "43", `{{ posts | sum: 'comments_count' }}`, map[string]interface{}{"posts": BlogPostsToInterfaceSlice(posts)})
	})
}

// TestBlogPostTags tests tags that process slices with []BlogPost.
func TestBlogPostTags(t *testing.T) {
	posts := SampleBlogPosts()

	// For tag
	t.Run("For tag", func(t *testing.T) {
		expected := "Getting Started with Go by AliceAdvanced Liquid Templates by BobUnderstanding Reflection by CharlieTesting Best Practices by AlicePerformance Optimization by Bob"
		assertTemplateResult(t, expected, `{% for post in posts %}{{ post.title }} by {{ post.author }}{% endfor %}`, map[string]interface{}{"posts": posts})
		assertTemplateResult(t, expected, `{% for post in posts %}{{ post.title }} by {{ post.author }}{% endfor %}`, map[string]interface{}{"posts": BlogPostsToInterfaceSlice(posts)})

		expected = "[1] Getting Started with Go[2] Advanced Liquid Templates[3] Understanding Reflection[4] Testing Best Practices[5] Performance Optimization"
		assertTemplateResult(t, expected, `{% for post in posts %}[{{ forloop.index }}] {{ post.title }}{% endfor %}`, map[string]interface{}{"posts": posts})
		assertTemplateResult(t, expected, `{% for post in posts %}[{{ forloop.index }}] {{ post.title }}{% endfor %}`, map[string]interface{}{"posts": BlogPostsToInterfaceSlice(posts)})

		expected = "FIRST: Getting Started with GoAdvanced Liquid TemplatesUnderstanding ReflectionTesting Best PracticesPerformance Optimization :LAST"
		assertTemplateResult(t, expected, `{% for post in posts %}{% if forloop.first %}FIRST: {% endif %}{{ post.title }}{% if forloop.last %} :LAST{% endif %}{% endfor %}`, map[string]interface{}{"posts": posts})

		assertTemplateResult(t, "Advanced Liquid TemplatesUnderstanding Reflection", `{% for post in posts offset:1 limit:2 %}{{ post.title }}{% endfor %}`, map[string]interface{}{"posts": posts})

		expected = "Performance OptimizationTesting Best PracticesUnderstanding ReflectionAdvanced Liquid TemplatesGetting Started with Go"
		assertTemplateResult(t, expected, `{% for post in posts reversed %}{{ post.title }}{% endfor %}`, map[string]interface{}{"posts": posts})

		assertTemplateResult(t, "EMPTY", `{% for post in posts %}{{ post.title }}{% endfor %}{% if posts.size == 0 %}EMPTY{% endif %}`, map[string]interface{}{"posts": []BlogPost{}})
	})

	// Render tag
	t.Run("Render tag", func(t *testing.T) {
		partials := map[string]string{
			"post": "{{ post.title }} by {{ post.author }}",
		}
		expected := "Getting Started with Go by AliceAdvanced Liquid Templates by BobUnderstanding Reflection by CharlieTesting Best Practices by AlicePerformance Optimization by Bob"
		assertTemplateResult(t, expected, `{% render "post" for posts %}`, map[string]interface{}{"posts": posts}, TemplateResultOptions{Partials: partials})
		assertTemplateResult(t, expected, `{% render "post" for posts %}`, map[string]interface{}{"posts": BlogPostsToInterfaceSlice(posts)}, TemplateResultOptions{Partials: partials})
		assertTemplateResult(t, "Getting Started with Go by Alice", `{% render "post" with posts[0] %}`, map[string]interface{}{"posts": posts}, TemplateResultOptions{Partials: partials})
	})

	// Include tag
	t.Run("Include tag", func(t *testing.T) {
		partials := map[string]string{
			"post_item": "{{ post_item.title }} by {{ post_item.author }}",
		}
		expected := "Getting Started with Go by AliceAdvanced Liquid Templates by BobUnderstanding Reflection by CharlieTesting Best Practices by AlicePerformance Optimization by Bob"
		assertTemplateResult(t, expected, `{% include "post_item" with posts %}`, map[string]interface{}{"posts": posts}, TemplateResultOptions{Partials: partials})
		assertTemplateResult(t, expected, `{% include "post_item" with posts %}`, map[string]interface{}{"posts": BlogPostsToInterfaceSlice(posts)}, TemplateResultOptions{Partials: partials})
		assertTemplateResult(t, "Getting Started with Go by Alice", `{% include "post_item" with posts[0] %}`, map[string]interface{}{"posts": posts}, TemplateResultOptions{Partials: partials})
	})

	// Cycle tag
	t.Run("Cycle tag", func(t *testing.T) {
		// When cycling with the same expression repeated, each iteration evaluates the expression
		// in the current loop context, so we get each post's title
		expected := "Getting Started with GoAdvanced Liquid TemplatesUnderstanding ReflectionTesting Best PracticesPerformance Optimization"
		assertTemplateResult(t, expected, `{% for post in posts %}{% cycle post.title, post.title, post.title %}{% endfor %}`, map[string]interface{}{"posts": posts})
		assertTemplateResult(t, expected, `{% for post in posts %}{% cycle post.title, post.title, post.title %}{% endfor %}`, map[string]interface{}{"posts": BlogPostsToInterfaceSlice(posts)})

		// With 2 cycle values and 5 posts, the cycle alternates between evaluating variables[0] and variables[1]
		// Since both are post.title, each evaluates to the current post's title
		expected = "Getting Started with GoAdvanced Liquid TemplatesUnderstanding ReflectionTesting Best PracticesPerformance Optimization"
		assertTemplateResult(t, expected, `{% for post in posts %}{% cycle 'colors': post.title, post.title %}{% endfor %}`, map[string]interface{}{"posts": posts})
	})
}
