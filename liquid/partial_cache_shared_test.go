package liquid

import (
	"fmt"
	"sync"
	"testing"
)

// mapPartials is a file system of partials held in memory.
type mapPartials map[string]string

func (m mapPartials) ReadTemplateFile(name string) (string, error) {
	if s, ok := m[name]; ok {
		return s, nil
	}
	return "", NewFileSystemError(fmt.Sprintf("No such template '%s'", name))
}

// partialFor loads name as a render with fs would: a context of its own.
func partialFor(t *testing.T, env *Environment, fs FileSystem, name string) *Template {
	t.Helper()
	registers := NewRegisters(map[string]interface{}{"file_system": fs})
	ctx := &mockContextForPartial{registers: registers}
	pc := NewParseContext(ParseContextOptions{Environment: env})
	tmpl, err := LoadPartial(name, ctx, pc)
	if err != nil {
		t.Fatalf("load %s: %v", name, err)
	}
	return tmpl
}

// Renders of one environment share the parse of a partial whose text is the
// same, whatever file system each render reads it from; a partial whose text
// changed is parsed again, and so is the same text under another name.
func TestRendersShareTheParseOfAPartial(t *testing.T) {
	env := NewEnvironment()
	first := partialFor(t, env, mapPartials{"p": "{{ x }}"}, "p")
	second := partialFor(t, env, mapPartials{"p": "{{ x }}"}, "p")
	if first != second {
		t.Error("two renders of the same partial text parsed it twice")
	}
	changed := partialFor(t, env, mapPartials{"p": "{{ y }}"}, "p")
	if changed == first {
		t.Error("a partial whose text changed was served from the cache")
	}
	renamed := partialFor(t, env, mapPartials{"q": "{{ x }}"}, "q")
	if renamed == first || renamed.Name() != "q" {
		t.Errorf("the same text under another name gave %q, the cached partial of p", renamed.Name())
	}
	if other := partialFor(t, NewEnvironment(), mapPartials{"p": "{{ x }}"}, "p"); other == first {
		t.Error("another environment was served the cached partial")
	}
}

// The shared cache keeps at most partialCacheSize partials and
// partialCacheBytes bytes of their text, the least recently used going first.
func TestThePartialCacheIsBounded(t *testing.T) {
	env := NewEnvironment()
	fs := mapPartials{}
	for i := 0; i < partialCacheSize+50; i++ {
		fs[fmt.Sprintf("p%d", i)] = fmt.Sprintf("text %d", i)
		partialFor(t, env, fs, fmt.Sprintf("p%d", i))
	}
	if n := env.partials.len(); n > partialCacheSize {
		t.Errorf("the cache holds %d partials, more than %d", n, partialCacheSize)
	}
	big := mapPartials{}
	for i := 0; i < 5; i++ {
		big[fmt.Sprintf("big%d", i)] = fmt.Sprintf("%d%s", i, string(make([]byte, partialCacheBytes/4)))
		partialFor(t, env, big, fmt.Sprintf("big%d", i))
	}
	if b := env.partials.size(); b > partialCacheBytes {
		t.Errorf("the cache holds %d bytes of text, more than %d", b, partialCacheBytes)
	}
}

// Concurrent renders load partials safely (run with -race).
func TestConcurrentLoadsOfAPartial(t *testing.T) {
	env := NewEnvironment()
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			partialFor(t, env, mapPartials{"p": fmt.Sprintf("{{ x }}%d", i%5)}, "p")
		}(i)
	}
	wg.Wait()
}
