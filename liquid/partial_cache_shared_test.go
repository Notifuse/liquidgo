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
	env := mustEnv(EnvironmentConfig{})
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
	if other := partialFor(t, newEnvironment(), mapPartials{"p": "{{ x }}"}, "p"); other == first {
		t.Error("another environment was served the cached partial")
	}
}

// The shared cache keeps at most partialCacheSize partials and
// partialCacheBytes bytes of their text, the least recently used going first;
// the index of the cache keeps no evicted entry.
func TestThePartialCacheIsBounded(t *testing.T) {
	env := mustEnv(EnvironmentConfig{})
	fs := mapPartials{}
	for i := 0; i < partialCacheSize+50; i++ {
		fs[fmt.Sprintf("p%d", i)] = fmt.Sprintf("text %d", i)
		partialFor(t, env, fs, fmt.Sprintf("p%d", i))
	}
	if n, m := env.partials.len(), len(env.partials.entries); n > partialCacheSize || m != n {
		t.Errorf("the cache holds %d partials and indexes %d, at most %d", n, m, partialCacheSize)
	}
	big := mapPartials{}
	for i := 0; i < 20; i++ {
		big[fmt.Sprintf("big%d", i)] = fmt.Sprintf("%d%s", i, string(make([]byte, partialCacheBytes/10)))
		partialFor(t, env, big, fmt.Sprintf("big%d", i))
	}
	if b := env.partials.size(); b > partialCacheBytes {
		t.Errorf("the cache holds %d bytes of text, more than %d", b, partialCacheBytes)
	}
	if len(env.partials.entries) != env.partials.len() {
		t.Errorf("the cache indexes %d partials and holds %d", len(env.partials.entries), env.partials.len())
	}
}

// A partial larger than an eighth of the byte bound is parsed for its render
// and not kept: one partial cannot fill the cache.
func TestAPartialTooLargeIsNotCached(t *testing.T) {
	env := mustEnv(EnvironmentConfig{})
	fs := mapPartials{"big": string(make([]byte, partialCacheBytes/8+1)), "small": "x"}
	first := partialFor(t, env, fs, "big")
	if again := partialFor(t, env, fs, "big"); again == first || env.partials.len() != 0 {
		t.Errorf("a partial of %d bytes was cached", len(fs["big"]))
	}
	small := partialFor(t, env, fs, "small")
	if again := partialFor(t, env, fs, "small"); again != small {
		t.Error("a small partial was not cached")
	}
}

// The bounds of the cache are set by the config; 0 turns the cache off.
func TestThePartialCacheLimitsCanBeSet(t *testing.T) {
	env := mustEnv(EnvironmentConfig{PartialCache: &PartialCacheLimits{Partials: 2, Bytes: 1 << 20}})
	fs := mapPartials{"a": "a", "b": "b", "c": "c"}
	a := partialFor(t, env, fs, "a")
	partialFor(t, env, fs, "b")
	partialFor(t, env, fs, "a") // a is now the most recently used
	partialFor(t, env, fs, "c") // b goes
	if env.partials.len() != 2 || partialFor(t, env, fs, "a") != a {
		t.Errorf("the least recently used partial was not the one to go: %d cached", env.partials.len())
	}
	env = mustEnv(EnvironmentConfig{PartialCache: &PartialCacheLimits{}})
	off := partialFor(t, env, fs, "a")
	if again := partialFor(t, env, fs, "a"); env.partials.len() != 0 || again == off {
		t.Error("a cache turned off kept a partial")
	}
}

// Two parses of one text that differ by an option that changes the parse
// (line numbers) are two entries. The error mode is the environment's, which
// never changes, so an environment's cache never holds two.
func TestThePartialCacheKeysTheParseOptions(t *testing.T) {
	env := mustEnv(EnvironmentConfig{})
	fs := mapPartials{"p": "{{ x }}"}
	load := func(lineNumbers bool) *Template {
		opts := ParseContextOptions{Environment: env, TemplateOptions: map[string]interface{}{"line_numbers": lineNumbers}}
		tmpl, err := LoadPartial("p", &mockContextForPartial{registers: NewRegisters(map[string]interface{}{"file_system": fs})}, NewParseContext(opts))
		if err != nil {
			t.Fatal(err)
		}
		return tmpl
	}
	plain := load(false)
	numbered := load(true)
	if plain == numbered {
		t.Error("the parse with line numbers shared the entry of the parse without")
	}
}

// A render reads a partial it includes several times once.
func TestARenderReadsAPartialOnce(t *testing.T) {
	env := mustEnv(EnvironmentConfig{})
	fs := &countingPartials{texts: map[string]string{"p": "{{ x }}"}, reads: map[string]int{}}
	registers := NewRegisters(map[string]interface{}{"file_system": fs})
	ctx := &mockContextForPartial{registers: registers}
	pc := NewParseContext(ParseContextOptions{Environment: env})
	for i := 0; i < 3; i++ {
		if _, err := LoadPartial("p", ctx, pc); err != nil {
			t.Fatal(err)
		}
	}
	if fs.reads["p"] != 1 {
		t.Errorf("the partial was read %d times in one render", fs.reads["p"])
	}
}

// countingPartials counts the reads of each partial.
type countingPartials struct {
	texts map[string]string
	reads map[string]int
}

func (c *countingPartials) ReadTemplateFile(name string) (string, error) {
	c.reads[name]++
	return c.texts[name], nil
}

// Concurrent renders load partials safely (run with -race).
func TestConcurrentLoadsOfAPartial(t *testing.T) {
	env := mustEnv(EnvironmentConfig{})
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
