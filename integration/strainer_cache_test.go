package integration

// Regression tests for the strainer class cache of an Environment (audit of
// 2026-11-15, finding 6): renders with local filters read and write the cache
// concurrently, and a cache hit must still give the render its filters.

import (
	"sync"
	"testing"

	"github.com/Notifuse/liquidgo/v5/liquid"
	"github.com/Notifuse/liquidgo/v5/liquid/tags"
)

type strainerCacheFilters struct{}

func (f *strainerCacheFilters) Shout(input interface{}) string { return "!" }

func strainerCacheTemplate(t *testing.T) *liquid.Template {
	t.Helper()
	env := liquid.NewEnvironment()
	tags.RegisterStandardTags(env)
	tpl, err := liquid.ParseTemplate(`{{ "a" | shout }}`, &liquid.TemplateOptions{Environment: env})
	if err != nil {
		t.Fatal(err)
	}
	return tpl
}

// The local filters of a render apply on every render, not only on the one
// that filled the cache.
func TestStrainerCacheHitKeepsLocalFilters(t *testing.T) {
	tpl := strainerCacheTemplate(t)
	for i := 0; i < 3; i++ {
		got := tpl.Render(map[string]interface{}{}, &liquid.RenderOptions{Filters: []interface{}{&strainerCacheFilters{}}})
		if got != "!" {
			t.Errorf("render %d = %q, want %q", i, got, "!")
		}
	}
}

// Concurrent renders with local filters share the cache without a data race
// (detected under -race) and all apply their filters.
func TestStrainerCacheConcurrentRenders(t *testing.T) {
	tpl := strainerCacheTemplate(t)
	var wg sync.WaitGroup
	results := make([]string, 50)
	for i := range results {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			results[i] = tpl.Render(map[string]interface{}{}, &liquid.RenderOptions{Filters: []interface{}{&strainerCacheFilters{}}})
		}(i)
	}
	wg.Wait()
	for i, got := range results {
		if got != "!" {
			t.Errorf("render %d = %q, want %q", i, got, "!")
		}
	}
}
