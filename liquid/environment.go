package liquid

import ()

// Environment is the container for all configuration options of Liquid, such as
// the registered tags, filters, and the default error mode.
type Environment struct {
	fileSystem            FileSystem
	tags                  map[string]TagConstructor
	defaultResourceLimits map[string]interface{}
	errorMode             string
	outputEscape          bool
	rawFilterRegistered   bool
	// filters are the filters the environment declares, which take
	// precedence over the standard ones (RegisterFilter).
	filters map[string]FilterFunc
	// partials keeps the partials renders parsed (LoadPartial).
	partials partialCache
}

// NewEnvironment creates a new environment instance.
func NewEnvironment() *Environment {
	env := &Environment{
		errorMode:             "lax",
		tags:                  make(map[string]TagConstructor),
		fileSystem:            &BlankFileSystem{},
		defaultResourceLimits: EmptyHash,
	}

	return env
}

// NewEnvironmentWithStandardTags creates a new environment and registers all standard tags.
// This function avoids import cycles by being called from outside the liquid package.
func NewEnvironmentWithStandardTags() *Environment {
	env := NewEnvironment()
	// Tags will be registered via tags.RegisterStandardTags from outside
	return env
}

// ErrorMode returns the error mode.
func (e *Environment) ErrorMode() string {
	return e.errorMode
}

// SetErrorMode sets the error mode.
func (e *Environment) SetErrorMode(mode string) {
	e.errorMode = mode
}

// Tags returns the tags map.
func (e *Environment) Tags() map[string]TagConstructor {
	return e.tags
}

// RegisterTag registers a new tag with the environment.
func (e *Environment) RegisterTag(name string, constructor TagConstructor) {
	e.tags[name] = constructor
}

// RegisterFilter declares filters of the environment, which take precedence
// over the standard ones of the same name.
func (e *Environment) RegisterFilter(filters ...FilterDef) error {
	for _, f := range filters {
		if err := f.check(); err != nil {
			return err
		}
	}
	if e.filters == nil {
		e.filters = make(map[string]FilterFunc, len(filters))
	}
	for _, f := range filters {
		e.filters[f.name] = f.fn
	}
	return nil
}

// FilterNames returns the names of the filters of the environment, the
// standard ones included, sorted.
func (e *Environment) FilterNames() []string {
	return filterNames(standardFilters, e.filters)
}

// TagForName returns the tag class for the given tag name.
func (e *Environment) TagForName(name string) TagConstructor {
	return e.tags[name]
}

// FileSystem returns the file system.
func (e *Environment) FileSystem() FileSystem {
	return e.fileSystem
}

// SetFileSystem sets the file system.
func (e *Environment) SetFileSystem(fs FileSystem) {
	e.fileSystem = fs
}

// SetDefaultResourceLimits sets the default resource limits.
func (e *Environment) SetDefaultResourceLimits(limits map[string]interface{}) {
	e.defaultResourceLimits = limits
}
