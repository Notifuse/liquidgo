package liquid

import (
	"reflect"
	"sync"
)

// Environment is the container for all configuration options of Liquid, such as
// the registered tags, filters, and the default error mode.
type Environment struct {
	fileSystem                 FileSystem
	tags                       map[string]TagConstructor
	strainerTemplate           *StrainerTemplateClass
	defaultResourceLimits      map[string]interface{}
	strainerTemplateClassCache map[string]*StrainerTemplateClass
	// cacheMu guards strainerTemplateClassCache, which renders read and write
	// concurrently.
	cacheMu             sync.RWMutex
	errorMode           string
	outputEscape        bool
	rawFilterRegistered bool
	registeredFilters   []interface{} // Store filter instances for use when creating strainers
	// partials keeps the partials renders parsed (LoadPartial).
	partials partialCache
}

// NewEnvironment creates a new environment instance.
func NewEnvironment() *Environment {
	env := &Environment{
		errorMode:                  "lax",
		tags:                       make(map[string]TagConstructor),
		strainerTemplate:           NewStrainerTemplateClass(),
		fileSystem:                 &BlankFileSystem{},
		defaultResourceLimits:      EmptyHash,
		strainerTemplateClassCache: make(map[string]*StrainerTemplateClass),
		registeredFilters:          make([]interface{}, 0),
	}

	// Add standard filters
	standardFilters := &StandardFilters{}
	_ = env.strainerTemplate.AddFilter(standardFilters)

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

// RegisterFilter registers a new filter with the environment.
func (e *Environment) RegisterFilter(filter interface{}) error {
	// Clear cache
	e.cacheMu.Lock()
	e.strainerTemplateClassCache = make(map[string]*StrainerTemplateClass)
	e.cacheMu.Unlock()
	// Store the filter instance
	e.registeredFilters = append(e.registeredFilters, filter)
	return e.strainerTemplate.AddFilter(filter)
}

// RegisterFilters registers multiple filters with this environment.
func (e *Environment) RegisterFilters(filters []interface{}) error {
	e.cacheMu.Lock()
	e.strainerTemplateClassCache = make(map[string]*StrainerTemplateClass)
	e.cacheMu.Unlock()
	for _, filter := range filters {
		// Store the filter instance
		e.registeredFilters = append(e.registeredFilters, filter)
		if err := e.strainerTemplate.AddFilter(filter); err != nil {
			return err
		}
	}
	return nil
}

// CreateStrainer creates a new strainer instance with the given filters.
func (e *Environment) CreateStrainer(context interface{ Context() interface{} }, filters []interface{}, strictFilters bool) *StrainerTemplate {
	if len(filters) == 0 {
		// Use registered filters from the environment
		if len(e.registeredFilters) > 0 {
			return NewStrainerTemplateWithFilters(e.strainerTemplate, context, strictFilters, e.registeredFilters)
		}
		return NewStrainerTemplate(e.strainerTemplate, context, strictFilters)
	}

	// Create a key for caching based on filters
	cacheKey := e.createStrainerCacheKey(filters)
	// The strainer holds the environment's filter instances, then the
	// render's, which take precedence, as Ruby's class for local filters
	// subclasses the environment's strainer
	instances := make([]interface{}, 0, len(e.registeredFilters)+len(filters))
	instances = append(append(instances, e.registeredFilters...), filters...)

	// Check cache first. A cached class holds the filter methods only: the
	// filter instances of this render are given to the strainer as on a miss.
	e.cacheMu.RLock()
	cached, ok := e.strainerTemplateClassCache[cacheKey]
	e.cacheMu.RUnlock()
	if ok {
		return NewStrainerTemplateWithFilters(cached, context, strictFilters, instances)
	}

	// Create new class and cache it
	class := NewStrainerTemplateClass()
	// Copy base methods
	for method := range e.strainerTemplate.filterMethods {
		class.filterMethods[method] = true
	}
	// Add additional filters
	for _, filter := range filters {
		_ = class.AddFilter(filter)
	}

	// Cache the class
	e.cacheMu.Lock()
	e.strainerTemplateClassCache[cacheKey] = class
	e.cacheMu.Unlock()

	return NewStrainerTemplateWithFilters(class, context, strictFilters, instances)
}

// FilterMethodNames returns the names of all filter methods.
func (e *Environment) FilterMethodNames() []string {
	return e.strainerTemplate.FilterMethodNames()
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

// createStrainerCacheKey creates a cache key from a filters array.
// In Ruby, arrays are used directly as hash keys, but in Go we need to create a string key.
func (e *Environment) createStrainerCacheKey(filters []interface{}) string {
	if len(filters) == 0 {
		return ""
	}

	// Create a key based on filter types. A type's String() is not unique:
	// two packages may both have a filters.Filters, so the key has the
	// import path of the type too.
	key := ""
	for _, filter := range filters {
		filterType := reflect.TypeOf(filter)
		named := filterType
		for named.Kind() == reflect.Ptr {
			named = named.Elem()
		}
		key += named.PkgPath() + " " + filterType.String() + ":"
	}
	return key
}
