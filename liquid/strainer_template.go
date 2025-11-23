package liquid

import (
	"fmt"
	"reflect"
	"strings"
)

// StrainerTemplate is the computed class for the filters system.
// New filters are mixed into the strainer class which is then instantiated for each liquid template render run.
type StrainerTemplate struct {
	context         interface{ Context() interface{} }
	filterMethods   map[string]bool
	filterInstances map[string]interface{}
	strictFilters   bool
}

// StrainerTemplateClass represents a strainer template class that can have filters added.
type StrainerTemplateClass struct {
	filterMethods map[string]bool
}

// NewStrainerTemplateClass creates a new strainer template class.
func NewStrainerTemplateClass() *StrainerTemplateClass {
	return &StrainerTemplateClass{
		filterMethods: make(map[string]bool),
	}
}

// AddFilter adds a filter module to the strainer template class.
func (stc *StrainerTemplateClass) AddFilter(filter interface{}) error {
	// Get all methods from the filter
	filterType := reflect.TypeOf(filter)
	if filterType.Kind() != reflect.Ptr {
		return fmt.Errorf("filter must be a pointer to a struct")
	}

	// Get methods from the filter
	for i := 0; i < filterType.NumMethod(); i++ {
		method := filterType.Method(i)
		// Only include exported methods
		if method.PkgPath == "" {
			stc.filterMethods[method.Name] = true
		}
	}

	return nil
}

// Invokable checks if a method name is invokable.
func (stc *StrainerTemplateClass) Invokable(method string) bool {
	return stc.filterMethods[method]
}

// FilterMethodNames returns all filter method names.
func (stc *StrainerTemplateClass) FilterMethodNames() []string {
	names := make([]string, 0, len(stc.filterMethods))
	for name := range stc.filterMethods {
		names = append(names, name)
	}
	return names
}

// NewStrainerTemplate creates a new strainer template instance.
func NewStrainerTemplate(class *StrainerTemplateClass, context interface{ Context() interface{} }, strictFilters bool) *StrainerTemplate {
	st := &StrainerTemplate{
		context:         context,
		filterMethods:   class.filterMethods,
		strictFilters:   strictFilters,
		filterInstances: make(map[string]interface{}),
	}

	// Always add StandardFilters as a base filter with context
	var ctx *Context
	if context != nil {
		if c, ok := context.Context().(*Context); ok {
			ctx = c
		}
	}
	sf := &StandardFilters{context: ctx}
	st.filterInstances["*liquid.StandardFilters"] = sf

	return st
}

// NewStrainerTemplateWithFilters creates a new strainer template instance with additional filters.
func NewStrainerTemplateWithFilters(class *StrainerTemplateClass, context interface{ Context() interface{} }, strictFilters bool, filters []interface{}) *StrainerTemplate {
	st := NewStrainerTemplate(class, context, strictFilters)

	// Add additional filter instances
	for _, filter := range filters {
		filterType := reflect.TypeOf(filter)
		if filterType.Kind() == reflect.Ptr {
			st.filterInstances[filterType.String()] = filter
		}
	}

	return st
}

// Invoke invokes a filter method.
func (st *StrainerTemplate) Invoke(method string, args ...interface{}) (interface{}, error) {
	// Check if method is invokable (try both lowercase and capitalized)
	methodInvokable := st.filterMethods[method]
	if !methodInvokable && len(method) > 0 {
		// Try CamelCase version (converts snake_case to CamelCase for Go method names)
		// e.g., find_index -> FindIndex, sort_natural -> SortNatural
		camelMethod := snakeToCamelCase(method)
		methodInvokable = st.filterMethods[camelMethod]
		if methodInvokable {
			// Use CamelCase version for lookup
			method = camelMethod
		} else {
			// Try case-insensitive match for acronyms (e.g., StripHtml -> StripHTML)
			// This handles cases where the method uses uppercase acronyms like HTML, XML, etc.
			for registeredMethod := range st.filterMethods {
				if strings.EqualFold(registeredMethod, camelMethod) {
					methodInvokable = true
					method = registeredMethod
					break
				}
			}
			if !methodInvokable {
				// Fallback: try simple capitalization (for single-word filters)
				capitalizedMethod := strings.ToUpper(method[:1]) + method[1:]
				methodInvokable = st.filterMethods[capitalizedMethod]
				if methodInvokable {
					// Use capitalized version for lookup
					method = capitalizedMethod
				}
			}
		}
	}
	if !methodInvokable {
		// Before failing, try property access on the first argument
		// This enables patterns like: {{ posts | first | title }}
		if len(args) > 0 && args[0] != nil {
			// Try to access property using InvokeDropOn
			result := InvokeDropOn(args[0], method)
			if result != nil {
				return result, nil
			}
		}

		if st.strictFilters {
			return nil, NewUndefinedFilter("undefined filter " + method)
		}
		// In non-strict mode, return first arg
		if len(args) > 0 {
			return args[0], nil
		}
		return nil, nil
	}

	// Method is invokable - use reflection to find and call it on filter instances
	// Try each filter instance until we find one with the method
	for _, filterInstance := range st.filterInstances {
		filterValue := reflect.ValueOf(filterInstance)

		// Look for the method - try both original case and capitalized version
		// Go method names are capitalized, but Liquid filter names are lowercase
		methodValue := filterValue.MethodByName(method)
		if !methodValue.IsValid() && len(method) > 0 {
			// Try capitalized version (first letter uppercase)
			capitalizedMethod := strings.ToUpper(method[:1]) + method[1:]
			methodValue = filterValue.MethodByName(capitalizedMethod)
		}
		if !methodValue.IsValid() {
			continue
		}

		// Check if method signature matches (first arg is input, rest are filter args)
		methodType := methodValue.Type()
		if methodType.NumIn() < 1 {
			continue
		}

		// Prepare arguments
		expectedArgs := methodType.NumIn()
		callArgs := make([]reflect.Value, expectedArgs)

		// First arg is the input (from args[0])
		if len(args) == 0 {
			continue
		}
		
		// Build all arguments
		for i := 0; i < expectedArgs; i++ {
			if i < len(args) {
				// Argument provided - use it
				argValue := args[i]
				if argValue == nil {
					// nil argument - use zero value for the parameter type
					paramType := methodType.In(i)
					callArgs[i] = reflect.Zero(paramType)
				} else {
					callArgs[i] = reflect.ValueOf(argValue)
				}
			} else {
				// Argument missing - pad with zero value for the parameter type
				// This allows filters with optional parameters to work correctly
				// and matches Ruby Liquid behavior where optional parameters default to nil
				paramType := methodType.In(i)
				callArgs[i] = reflect.Zero(paramType)
			}
		}

		// Call the method
		results := methodValue.Call(callArgs)
		if len(results) > 0 {
			return results[0].Interface(), nil
		}
		return nil, nil
	}

	// Method not found in any filter - this shouldn't happen if filterMethods is correct
	// but handle gracefully
	if len(args) > 0 {
		return args[0], nil
	}
	return nil, nil
}

// snakeToCamelCase converts snake_case to CamelCase.
// e.g., find_index -> FindIndex, sort_natural -> SortNatural, strip_html -> StripHTML
func snakeToCamelCase(s string) string {
	if s == "" {
		return ""
	}

	// Split by underscore
	parts := strings.Split(s, "_")

	// Common acronyms that should be uppercase
	acronyms := map[string]string{
		"html": "HTML",
		"xml":  "XML",
		"json": "JSON",
		"url":  "URL",
		"id":   "ID",
		"api":  "API",
		"css":  "CSS",
		"js":   "JS",
	}

	// Capitalize each part
	for i, part := range parts {
		if len(part) > 0 {
			lowerPart := strings.ToLower(part)
			if acronym, ok := acronyms[lowerPart]; ok {
				// Use uppercase acronym
				parts[i] = acronym
			} else {
				// Capitalize first letter
				parts[i] = strings.ToUpper(part[:1]) + strings.ToLower(part[1:])
			}
		}
	}

	return strings.Join(parts, "")
}
