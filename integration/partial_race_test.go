package integration

import (
	"fmt"
	"sync"
	"testing"

	"github.com/Notifuse/liquidgo/v5/liquid"
	"github.com/Notifuse/liquidgo/v5/liquid/tags"
)

// Concurrent renders of a template with render and include tags load their
// partials without writing to the tags' shared parse state (detected under
// -race), and each gets its own output.
func TestConcurrentRendersOfPartials(t *testing.T) {
	env := mustEnv(liquid.EnvironmentConfig{Tags: tags.Standard(), FileSystem: &mapFileSystem{templates: map[string]string{
		"p": "{% assign z = v %}{{ v }}{{ z }}",
		"q": "{% increment c %}{{ v }}",
	}}})
	tpl, err := liquid.ParseTemplate(`{% render 'p', v: v %}|{% include 'q' %}|{% assign y = v %}{{ y }}`, &liquid.TemplateOptions{Environment: env})
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	results := make([]string, 100)
	for i := range results {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			results[i] = renderRuby(tpl, map[string]interface{}{"v": i}, nil)
		}(i)
	}
	wg.Wait()
	for i, got := range results {
		if want := fmt.Sprintf("%d%d|0%d|%d", i, i, i, i); got != want {
			t.Errorf("render %d = %q, want %q", i, got, want)
		}
	}
}
