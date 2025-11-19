package liquid

import (
	"reflect"
)

// Environment is the container for all configuration options of Liquid, such as
// the registered tags, filters, and the default error mode.
type Environment struct {
	fileSystem                 FileSystem
	tags                       map[string]interface{}
	strainerTemplate           *StrainerTemplateClass
	exceptionRenderer          func(error) interface{}
	defaultResourceLimits      map[string]interface{}
	strainerTemplateClassCache map[string]*StrainerTemplateClass
	errorMode                  string
}

// NewEnvironment creates a new environment instance.
func NewEnvironment() *Environment {
	env := &Environment{
		errorMode:                  "lax",
		tags:                       make(map[string]interface{}),
		strainerTemplate:           NewStrainerTemplateClass(),
		exceptionRenderer:          func(err error) interface{} { return err },
		fileSystem:                 &BlankFileSystem{},
		defaultResourceLimits:      EmptyHash,
		strainerTemplateClassCache: make(map[string]*StrainerTemplateClass),
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
func (e *Environment) Tags() map[string]interface{} {
	return e.tags
}

// RegisterTag registers a new tag with the environment.
func (e *Environment) RegisterTag(name string, tagClass interface{}) {
	e.tags[name] = tagClass
}

// RegisterFilter registers a new filter with the environment.
func (e *Environment) RegisterFilter(filter interface{}) error {
	// Clear cache
	e.strainerTemplateClassCache = make(map[string]*StrainerTemplateClass)
	return e.strainerTemplate.AddFilter(filter)
}

// RegisterFilters registers multiple filters with this environment.
func (e *Environment) RegisterFilters(filters []interface{}) error {
	e.strainerTemplateClassCache = make(map[string]*StrainerTemplateClass)
	for _, filter := range filters {
		if err := e.strainerTemplate.AddFilter(filter); err != nil {
			return err
		}
	}
	return nil
}

// CreateStrainer creates a new strainer instance with the given filters.
func (e *Environment) CreateStrainer(context interface{ Context() interface{} }, filters []interface{}, strictFilters bool) *StrainerTemplate {
	if len(filters) == 0 {
		return NewStrainerTemplate(e.strainerTemplate, context, strictFilters)
	}

	// Create a key for caching based on filters
	cacheKey := e.createStrainerCacheKey(filters)

	// Check cache first
	if cached, ok := e.strainerTemplateClassCache[cacheKey]; ok {
		return NewStrainerTemplate(cached, context, strictFilters)
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
	e.strainerTemplateClassCache[cacheKey] = class

	return NewStrainerTemplateWithFilters(class, context, strictFilters, filters)
}

// FilterMethodNames returns the names of all filter methods.
func (e *Environment) FilterMethodNames() []string {
	return e.strainerTemplate.FilterMethodNames()
}

// TagForName returns the tag class for the given tag name.
func (e *Environment) TagForName(name string) interface{} {
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

// ExceptionRenderer returns the exception renderer.
func (e *Environment) ExceptionRenderer() func(error) interface{} {
	return e.exceptionRenderer
}

// SetExceptionRenderer sets the exception renderer.
func (e *Environment) SetExceptionRenderer(renderer func(error) interface{}) {
	e.exceptionRenderer = renderer
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

	// Create a key based on filter types
	key := ""
	for _, filter := range filters {
		filterType := reflect.TypeOf(filter)
		key += filterType.String() + ":"
	}
	return key
}
