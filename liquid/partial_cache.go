package liquid

import (
	"container/list"
	"sync"
)

// partialCacheSize is the most partials an environment keeps parsed.
const partialCacheSize = 1024

// partialCacheBytes is the most bytes of partial text an environment keeps
// parsed: the cache is keyed by the text.
const partialCacheBytes = 32 << 20

// partialKey identifies the parse of a partial: its name, which the template
// keeps for its errors, its text, and the options that change the parse.
type partialKey struct {
	name        string
	source      string
	errorMode   string
	lineNumbers bool
}

// partialEntry is a partial the cache keeps.
type partialEntry struct {
	key  partialKey
	tmpl *Template
}

// partialCache keeps the partials an environment parsed, by their text, so
// that renders that read the same text share its parse, whatever file system
// they read it from, and a partial whose text changed is parsed again. A
// parsed template is never written by a render, so renders share it. The
// least recently used partial goes first past partialCacheSize partials or
// partialCacheBytes bytes of text.
type partialCache struct {
	mu      sync.Mutex
	entries map[partialKey]*list.Element
	order   list.List // front: most recently used
	bytes   int
}

func (c *partialCache) get(k partialKey) *Template {
	c.mu.Lock()
	defer c.mu.Unlock()
	e, ok := c.entries[k]
	if !ok {
		return nil
	}
	c.order.MoveToFront(e)
	return e.Value.(*partialEntry).tmpl
}

func (c *partialCache) put(k partialKey, tmpl *Template) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.entries == nil {
		c.entries = make(map[partialKey]*list.Element)
	}
	if _, ok := c.entries[k]; ok {
		return
	}
	c.entries[k] = c.order.PushFront(&partialEntry{key: k, tmpl: tmpl})
	c.bytes += len(k.source)
	for c.order.Len() > partialCacheSize || (c.bytes > partialCacheBytes && c.order.Len() > 1) {
		oldest := c.order.Back()
		entry := oldest.Value.(*partialEntry)
		c.order.Remove(oldest)
		delete(c.entries, entry.key)
		c.bytes -= len(entry.key.source)
	}
}

func (c *partialCache) len() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.order.Len()
}

func (c *partialCache) size() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.bytes
}

// LoadPartial returns the partial templateName of the render context belongs
// to, read from the file system of its registers. A render keeps the
// partials it loaded (the cached_partials register), and the environment of
// parseContext keeps their parse for every render (partialCache): the file
// system is read at each render, so a changed partial is never stale.
func LoadPartial(templateName string, context interface {
	Registers() *Registers
}, parseContext ParseContextInterface) (*Template, error) {
	registers := context.Registers()
	cache, ok := registers.Get("cached_partials").(map[string]*Template)
	if !ok {
		cache = make(map[string]*Template)
		registers.Set("cached_partials", cache)
	}

	env := parseContext.Environment()
	key := partialKey{name: templateName, errorMode: "lax"}
	if env != nil {
		key.errorMode = env.ErrorMode()
	}
	if pc, ok := parseContext.(*ParseContext); ok {
		key.lineNumbers, _ = pc.GetOption("line_numbers").(bool)
	}
	renderKey := templateName + ":" + key.errorMode
	if cached, ok := cache[renderKey]; ok {
		return cached, nil
	}

	fs, ok := registers.Get("file_system").(FileSystem)
	if !ok {
		fs = &BlankFileSystem{}
	}
	source, err := fs.ReadTemplateFile(templateName)
	if err != nil {
		return nil, err
	}
	key.source = source

	if env != nil {
		if tmpl := env.partials.get(key); tmpl != nil {
			cache[renderKey] = tmpl
			return tmpl, nil
		}
	}

	// The partial is parsed with a parse context of its own: the tag's
	// parse context, which every render of the template shares, is only
	// read.
	tmpl := NewTemplate(&TemplateOptions{Environment: env})
	if err := tmpl.Parse(source, &TemplateOptions{Environment: env, LineNumbers: key.lineNumbers}); err != nil {
		nameError(err, templateName)
		return nil, err
	}
	tmpl.SetName(templateName)
	if env != nil {
		env.partials.put(key, tmpl)
	}
	cache[renderKey] = tmpl
	return tmpl, nil
}

// nameError sets the template name of a parse error of a partial.
func nameError(err error, name string) {
	switch e := err.(type) {
	case *Error:
		e.TemplateName = name
	case LiquidError:
		if inner := e.GetError(); inner != nil {
			inner.TemplateName = name
		}
	}
}
