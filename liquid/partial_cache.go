package liquid

import (
	"container/list"
	"hash/maphash"
	"sync"
)

// partialCacheSize is the most partials an environment keeps parsed, unless
// EnvironmentConfig.PartialCache sets another bound.
const partialCacheSize = 1024

// partialCacheBytes is the most bytes of partial text an environment keeps
// parsed, unless EnvironmentConfig.PartialCache sets another bound. A parsed partial
// holds 10 to 30 times its text: 4 MiB of text is about 100 MiB of memory.
const partialCacheBytes = 4 << 20

// partialSeed seeds the hash of the texts of partials.
var partialSeed = maphash.MakeSeed()

// partialKey identifies the parse of a partial: its name, which the template
// keeps for its errors, the options that change the parse, and the hash and
// length of its text, which the entry keeps to compare.
type partialKey struct {
	name        string
	lineNumbers bool
	hash        uint64
	length      int
}

// partialEntry is a partial the cache keeps; it never changes once made.
type partialEntry struct {
	key    partialKey
	source string
	tmpl   *Template
}

// partialCache keeps the partials an environment parsed, by their text, so
// that renders that read the same text share its parse, whatever file system
// they read it from, and a partial whose text changed is parsed again. A
// parsed template is never written by a render, so renders share it. The
// least recently used partial goes first past maxEntries partials or maxBytes
// bytes of text; a partial larger than an eighth of maxBytes is not kept.
type partialCache struct {
	mu         sync.Mutex
	entries    map[partialKey]*list.Element
	order      list.List // front: most recently used
	bytes      int
	limited    bool // the bounds were set (EnvironmentConfig.PartialCache)
	maxEntries int
	maxBytes   int
}

// limits returns the bounds of the cache; the caller holds the lock.
func (c *partialCache) limits() (int, int) {
	if !c.limited {
		return partialCacheSize, partialCacheBytes
	}
	return c.maxEntries, c.maxBytes
}

// get returns the partial of k whose text is source. The text is compared
// after the lock is released: two texts of one hash and length are a miss.
func (c *partialCache) get(k partialKey, source string) *Template {
	c.mu.Lock()
	e, ok := c.entries[k]
	if ok {
		c.order.MoveToFront(e)
	}
	c.mu.Unlock()
	if !ok {
		return nil
	}
	entry := e.Value.(*partialEntry)
	if entry.source != source {
		return nil
	}
	return entry.tmpl
}

func (c *partialCache) put(k partialKey, source string, tmpl *Template) {
	c.mu.Lock()
	defer c.mu.Unlock()
	maxEntries, maxBytes := c.limits()
	if maxEntries <= 0 || maxBytes <= 0 || len(source) > maxBytes/8 {
		return
	}
	if c.entries == nil {
		c.entries = make(map[partialKey]*list.Element)
	}
	if _, ok := c.entries[k]; ok {
		return
	}
	c.entries[k] = c.order.PushFront(&partialEntry{key: k, source: source, tmpl: tmpl})
	c.bytes += len(source)
	c.evict(maxEntries, maxBytes)
}

// evict removes the least recently used partials past the bounds; the caller
// holds the lock.
func (c *partialCache) evict(maxEntries, maxBytes int) {
	for c.order.Len() > 0 && (c.order.Len() > maxEntries || c.bytes > maxBytes) {
		oldest := c.order.Back()
		entry := oldest.Value.(*partialEntry)
		c.order.Remove(oldest)
		delete(c.entries, entry.key)
		c.bytes -= len(entry.source)
	}
}

// setLimits sets the bounds of the cache, and removes what is past them.
func (c *partialCache) setLimits(maxEntries, maxBytes int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.limited, c.maxEntries, c.maxBytes = true, maxEntries, maxBytes
	c.evict(maxEntries, maxBytes)
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
// system is read at each render, so a changed partial is never stale. The
// template returned is shared: a caller must not change it.
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
	// the cache of the environment needs no error mode in its key: an
	// environment has one, which never changes
	key := partialKey{name: templateName}
	errorMode := ErrorModeLax
	if env != nil {
		errorMode = env.ErrorMode()
	}
	if pc, ok := parseContext.(*ParseContext); ok {
		key.lineNumbers, _ = pc.GetOption("line_numbers").(bool)
	}
	renderKey := templateName + ":" + string(errorMode)
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
	key.hash, key.length = maphash.String(partialSeed, source), len(source)

	if env != nil {
		if tmpl := env.partials.get(key, source); tmpl != nil {
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
		env.partials.put(key, source, tmpl)
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
