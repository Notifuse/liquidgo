package liquid

import "fmt"

// Environment holds what the templates parsed with it share: their tags,
// their filters, their error mode, the file system of their partials, the
// resource limits of their renders and the output escape. It is built once
// by NewEnvironment and never changes after, so renders can share it.
type Environment struct {
	fileSystem   FileSystem
	tags         map[string]TagConstructor
	limits       ResourceLimitsConfig
	errorMode    ErrorMode
	outputEscape bool
	// filters are the filters the environment declares, which take
	// precedence over the standard ones.
	filters map[string]FilterFunc
	// partials keeps the partials renders parsed (LoadPartial).
	partials partialCache
}

// EnvironmentConfig configures an environment. NewEnvironment copies it: a
// change to the config, or to its maps, after the call changes nothing.
type EnvironmentConfig struct {
	// Tags are the tags templates can use, by name: tags.Standard() gives
	// Liquid's, to which a map entry adds or replaces one.
	Tags map[string]TagConstructor
	// Filters are declared next to the standard ones, over which they take
	// precedence for a name both define.
	Filters []FilterDef
	// ErrorMode is how templates parse: ErrorModeLax (the default),
	// ErrorModeStrict, ErrorModeStrict2 (or ErrorModeRigid, its other name)
	// or ErrorModeWarn.
	ErrorMode ErrorMode
	// FileSystem reads the partials of render and include; nil reads none.
	FileSystem FileSystem
	// Limits are the resource limits of each render; a render's WithLimits
	// replaces them field by field.
	Limits ResourceLimitsConfig
	// OutputEscape HTML-escapes the text of every {{ }} output, echo and
	// cycle, as LiquidJS's outputEscape option does, so that a value of the
	// render data cannot add markup to an HTML template. It is off by
	// default, as in Shopify Liquid.
	//
	// An output is written as it is when its value is HTML on purpose:
	//   - the result of raw ({{ feed.html | raw }}), or of escape,
	//     escape_once or h, which have escaped it already;
	//   - the text of a capture of the render, however it is reached
	//     ({{ block }}, or {{ copy }} after an assign): its outputs were
	//     escaped when it was captured, and the rest is the template's own
	//     markup;
	//   - either of these through strip, lstrip, rstrip, strip_newlines,
	//     newline_to_br, truncate or truncatewords ({{ bio | escape |
	//     newline_to_br }}).
	//
	// Any other filter after them makes the result ordinary text again.
	// Values are not escaped where they are read, so assign, capture,
	// conditions and filters see the data as it is. The option adds the raw
	// filter, which Shopify Liquid does not have.
	OutputEscape bool
	// PartialCache bounds the partials the environment keeps parsed for its
	// renders; nil keeps the defaults (see PartialCacheLimits).
	PartialCache *PartialCacheLimits
}

// PartialCacheLimits bounds the partials an environment keeps parsed: at most
// Partials of them, and Bytes of their text (1,024 and 4 MiB by default); a
// parsed partial holds 10 to 30 times its text. A partial larger than an
// eighth of Bytes is parsed for each render and not kept. 0 turns the cache
// off.
type PartialCacheLimits struct {
	Partials int
	Bytes    int
}

// NewEnvironment builds an environment from config. It refuses an unknown
// error mode, a tag without a name or a constructor, an invalid filter
// declaration and negative partial cache limits.
func NewEnvironment(config EnvironmentConfig) (*Environment, error) {
	env := newEnvironment()
	switch config.ErrorMode {
	case "":
	case ErrorModeLax, ErrorModeStrict, ErrorModeStrict2, ErrorModeRigid, ErrorModeWarn:
		env.errorMode = config.ErrorMode
	default:
		return nil, fmt.Errorf("liquid: unknown error mode %q", config.ErrorMode)
	}
	for name, ctor := range config.Tags {
		if name == "" || ctor == nil {
			return nil, fmt.Errorf("liquid: the tag %q has no name or no constructor", name)
		}
		env.tags[name] = ctor
	}
	filters := config.Filters
	if config.OutputEscape {
		env.outputEscape = true
		// raw first, so that a filter of the config can replace it
		filters = append([]FilterDef{Filter("raw", rawFilter)}, filters...)
	}
	for _, f := range filters {
		if err := f.check(); err != nil {
			return nil, err
		}
		if env.filters == nil {
			env.filters = make(map[string]FilterFunc, len(filters))
		}
		env.filters[f.name] = f.fn
	}
	if config.FileSystem != nil {
		env.fileSystem = config.FileSystem
	}
	env.limits = copyLimits(config.Limits)
	if pc := config.PartialCache; pc != nil {
		if pc.Partials < 0 || pc.Bytes < 0 {
			return nil, fmt.Errorf("liquid: negative partial cache limits (%d partials, %d bytes)", pc.Partials, pc.Bytes)
		}
		env.partials.setLimits(pc.Partials, pc.Bytes)
	}
	return env, nil
}

// newEnvironment returns the environment of a template parsed without one:
// lax, without tags, filters other than the standard ones, file system or
// limits.
func newEnvironment() *Environment {
	return &Environment{
		errorMode:  ErrorModeLax,
		tags:       make(map[string]TagConstructor),
		fileSystem: &BlankFileSystem{},
	}
}

// copyLimits returns limits with integers of its own, so that the caller's
// later writes through its pointers change nothing.
func copyLimits(limits ResourceLimitsConfig) ResourceLimitsConfig {
	own := func(p *int) *int {
		if p == nil {
			return nil
		}
		v := *p
		return &v
	}
	return ResourceLimitsConfig{
		RenderLengthLimit: own(limits.RenderLengthLimit),
		RenderScoreLimit:  own(limits.RenderScoreLimit),
		AssignScoreLimit:  own(limits.AssignScoreLimit),
	}
}

// ErrorMode returns the error mode.
func (e *Environment) ErrorMode() ErrorMode {
	return e.errorMode
}

// FilterNames returns the names of the filters of the environment, the
// standard ones included, sorted.
func (e *Environment) FilterNames() []string {
	return filterNames(standardFilters, e.filters)
}

// TagForName returns the constructor of the tag name, or nil.
func (e *Environment) TagForName(name string) TagConstructor {
	return e.tags[name]
}

// FileSystem returns the file system.
func (e *Environment) FileSystem() FileSystem {
	return e.fileSystem
}

// OutputEscape reports whether the output escape of the environment is on.
func (e *Environment) OutputEscape() bool {
	return e != nil && e.outputEscape
}
