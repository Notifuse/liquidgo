package integration

// Local filters of a render (audit of 2026-11-15, finding 6, first found in
// the strainer cache that is gone): each render gets its own filters, also
// when renders run concurrently, and the environment's filters apply
// alongside local ones.

import (
	"fmt"
	"sync"
	"testing"

	"github.com/Notifuse/liquidgo/v5/integration/internal/samename/a/filters"
	bfilters "github.com/Notifuse/liquidgo/v5/integration/internal/samename/b/filters"
	"github.com/Notifuse/liquidgo/v5/internal/methodfilters"
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

func localFiltersTemplate(t *testing.T, envFilters []liquid.FilterDef, src string) *liquid.Template {
	t.Helper()
	env := mustEnv(liquid.EnvironmentConfig{Tags: tags.Standard(), Filters: envFilters})
	tpl, err := liquid.ParseTemplate(src, &liquid.TemplateOptions{Environment: env})
	if err != nil {
		t.Fatal(err)
	}
	return tpl
}

func withFilters(filters ...interface{}) *rubyOptions {
	return &rubyOptions{Filters: filters}
}

// Each render uses its own filter instance, also on a cache hit.
func TestEachRenderUsesItsLocalFilters(t *testing.T) {
	tpl := localFiltersTemplate(t, nil, `{{ "a" | tenant }}`)
	for _, tag := range []string{"A", "B", "C"} {
		if got := renderRuby(tpl, map[string]interface{}{}, withFilters(&tenantFilters{tag})); got != "t:"+tag {
			t.Errorf("render for %s = %q, want %q", tag, got, "t:"+tag)
		}
	}
}

// Concurrent renders with local filters share the cache without a data race
// (detected under -race), and each gets its own filter instance.
func TestConcurrentRendersWithLocalFilters(t *testing.T) {
	tpl := localFiltersTemplate(t, nil, `{{ "a" | tenant }}`)
	var wg sync.WaitGroup
	results := make([]string, 100)
	for i := range results {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			results[i] = renderRuby(tpl, map[string]interface{}{}, withFilters(&tenantFilters{fmt.Sprint(i)}))
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
func TestEnvironmentFiltersApplyWithLocalOnes(t *testing.T) {
	tpl := localFiltersTemplate(t, methodfilters.Of(&envFilters{}), `{{ "x" | env_only }}|{{ "x" | tenant }}`)
	if got := renderRuby(tpl, map[string]interface{}{}, nil); got != "env|env-tenant" {
		t.Errorf("without local filters = %q, want %q", got, "env|env-tenant")
	}
	for i := 0; i < 2; i++ {
		if got := renderRuby(tpl, map[string]interface{}{}, withFilters(&tenantFilters{"L"})); got != "env|t:L" {
			t.Errorf("with local filters, render %d = %q, want %q", i, got, "env|t:L")
		}
	}
}

// Two filter types with the same name in two packages get their own methods.
func TestFilterStructsOfTheSameNameStayApart(t *testing.T) {
	tpl := localFiltersTemplate(t, nil, `{{ "x" | a_only }}|{{ "x" | b_only }}`)
	if got := renderRuby(tpl, map[string]interface{}{}, withFilters(&filters.Filters{})); got != "a|x" {
		t.Errorf("with a.Filters = %q, want %q", got, "a|x")
	}
	if got := renderRuby(tpl, map[string]interface{}{}, withFilters(&bfilters.Filters{})); got != "x|b" {
		t.Errorf("with b.Filters = %q, want %q", got, "x|b")
	}
}
