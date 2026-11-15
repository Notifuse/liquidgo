package integration

// Regression tests for the strainer class cache of an Environment (audit of
// 2026-11-15, finding 6): renders with local filters read and write the cache
// concurrently, a cache hit must still give the render its own filter
// instances, and the environment's filters apply alongside local ones.

import (
	"fmt"
	"sync"
	"testing"

	"github.com/Notifuse/liquidgo/v5/integration/internal/samename/a/filters"
	bfilters "github.com/Notifuse/liquidgo/v5/integration/internal/samename/b/filters"
	"github.com/Notifuse/liquidgo/v5/liquid"
	"github.com/Notifuse/liquidgo/v5/liquid/tags"
)

// tenantFilters carries the state of one render, such as a tenant's settings.
type tenantFilters struct{ tag string }

func (f *tenantFilters) Tenant(input interface{}) string { return "t:" + f.tag }

type envFilters struct{}

func (f *envFilters) EnvOnly(input interface{}) string { return "env" }

// Tenant is overridden by a local filter of the same name.
func (f *envFilters) Tenant(input interface{}) string { return "env-tenant" }

func strainerCacheTemplate(t *testing.T, env *liquid.Environment, src string) *liquid.Template {
	t.Helper()
	if env == nil {
		env = liquid.NewEnvironment()
	}
	tags.RegisterStandardTags(env)
	tpl, err := liquid.ParseTemplate(src, &liquid.TemplateOptions{Environment: env})
	if err != nil {
		t.Fatal(err)
	}
	return tpl
}

func withFilters(filters ...interface{}) *liquid.RenderOptions {
	return &liquid.RenderOptions{Filters: filters}
}

// Each render uses its own filter instance, also on a cache hit.
func TestStrainerCacheHitUsesTheRendersFilters(t *testing.T) {
	tpl := strainerCacheTemplate(t, nil, `{{ "a" | tenant }}`)
	for _, tag := range []string{"A", "B", "C"} {
		if got := tpl.Render(map[string]interface{}{}, withFilters(&tenantFilters{tag})); got != "t:"+tag {
			t.Errorf("render for %s = %q, want %q", tag, got, "t:"+tag)
		}
	}
}

// Concurrent renders with local filters share the cache without a data race
// (detected under -race), and each gets its own filter instance.
func TestStrainerCacheConcurrentRenders(t *testing.T) {
	tpl := strainerCacheTemplate(t, nil, `{{ "a" | tenant }}`)
	var wg sync.WaitGroup
	results := make([]string, 100)
	for i := range results {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			results[i] = tpl.Render(map[string]interface{}{}, withFilters(&tenantFilters{fmt.Sprint(i)}))
		}(i)
	}
	wg.Wait()
	for i, got := range results {
		if want := "t:" + fmt.Sprint(i); got != want {
			t.Errorf("render %d = %q, want %q", i, got, want)
		}
	}
}

// The filters registered on the environment apply when a render also passes
// local filters, which take precedence for a name both define.
func TestStrainerKeepsEnvironmentFiltersWithLocalOnes(t *testing.T) {
	env := liquid.NewEnvironment()
	if err := env.RegisterFilter(&envFilters{}); err != nil {
		t.Fatal(err)
	}
	tpl := strainerCacheTemplate(t, env, `{{ "x" | env_only }}|{{ "x" | tenant }}`)
	if got := tpl.Render(map[string]interface{}{}, nil); got != "env|env-tenant" {
		t.Errorf("without local filters = %q, want %q", got, "env|env-tenant")
	}
	for i := 0; i < 2; i++ {
		if got := tpl.Render(map[string]interface{}{}, withFilters(&tenantFilters{"L"})); got != "env|t:L" {
			t.Errorf("with local filters, render %d = %q, want %q", i, got, "env|t:L")
		}
	}
}

// Two filter types with the same name in two packages get their own methods.
func TestStrainerCacheTellsSameNamedTypesApart(t *testing.T) {
	tpl := strainerCacheTemplate(t, nil, `{{ "x" | a_only }}|{{ "x" | b_only }}`)
	if got := tpl.Render(map[string]interface{}{}, withFilters(&filters.Filters{})); got != "a|x" {
		t.Errorf("with a.Filters = %q, want %q", got, "a|x")
	}
	if got := tpl.Render(map[string]interface{}{}, withFilters(&bfilters.Filters{})); got != "x|b" {
		t.Errorf("with b.Filters = %q, want %q", got, "x|b")
	}
}
