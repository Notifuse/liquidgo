package liquid

import (
	gocontext "context"
	"fmt"
	"reflect"
)

// ContextConfig configures a Context.
type ContextConfig struct {
	Registers          interface{}
	Environment        *Environment
	OuterScope         map[string]interface{}
	ResourceLimits     *ResourceLimits
	Environments       []map[string]interface{}
	StaticEnvironments []map[string]interface{}
	RethrowErrors      bool
}

// Context keeps the variable stack and resolves variables, as well as keywords.
type Context struct {
	disabledTags   map[string]int
	resourceLimits *ResourceLimits
	cancel         gocontext.Context
	done           <-chan struct{}
	// funcValues holds the value each func of the data gave in this render,
	// on the root context of the render (see reporter), keyed by the map the
	// func was read from and its key: the data itself is never written.
	funcValues map[funcValueKey]funcValue
	// fatalPanics makes a Go panic in a node end the render (Template.Render)
	// instead of being written as "Liquid error: internal".
	fatalPanics bool
	// rethrowErrors makes HandleError raise (panic with) the error it
	// records, where it returns its message: Ruby's rethrow_errors.
	rethrowErrors bool
	// diagnostics is the context whose errors and warnings this one adds
	// to: an isolated subcontext reports to the context that made it, as
	// Ruby's subcontexts share the errors array.
	diagnostics        *Context
	registers          *Registers
	stringScanner      *StringScanner
	strainer           *StrainerTemplate
	environment        *Environment
	templateName       string
	warnings           []error
	environments       []map[string]interface{}
	filters            []interface{}
	interrupts         []interface{}
	errors             []error
	scopes             []map[string]interface{}
	staticEnvironments []map[string]interface{}
	baseScopeDepth     int
	renderDepth        int // Track nesting depth during rendering
	// counters holds the increment and decrement counters of the render.
	// Ruby keeps them in the first environment; they are read after the
	// scopes and before the environments, as there.
	counters        map[string]int
	strictFilters   bool
	strictVariables bool
	partial         bool
	// safeHTML holds the texts of the captures of the render, which an
	// output does not escape (see Environment.SetOutputEscape). Only the
	// reporter context of the render keeps it.
	safeHTML map[string]struct{}
}

// BuildContext creates a new Context with the given configuration.
func BuildContext(config ContextConfig) *Context {
	env := config.Environment
	if env == nil {
		env = NewEnvironment()
	}

	// As Ruby's Context.new, whose environments default to {}: a context
	// given none has one empty environment, which strict variables look a
	// name up in (and nothing writes)
	environments := config.Environments
	if environments == nil {
		environments = []map[string]interface{}{emptyData}
	}

	staticEnvironments := config.StaticEnvironments
	if staticEnvironments == nil {
		staticEnvironments = []map[string]interface{}{}
	}

	outerScope := config.OuterScope
	if outerScope == nil {
		outerScope = make(map[string]interface{})
	}

	var registers *Registers
	if config.Registers != nil {
		if r, ok := config.Registers.(*Registers); ok {
			registers = r
		} else if m, ok := config.Registers.(map[string]interface{}); ok {
			registers = NewRegisters(m)
		} else {
			registers = NewRegisters(nil)
		}
	} else {
		registers = NewRegisters(nil)
	}

	resourceLimits := config.ResourceLimits
	if resourceLimits == nil {
		resourceLimits = NewResourceLimits(ResourceLimitsConfig{})
	}

	ctx := &Context{
		environment:        env,
		environments:       environments,
		staticEnvironments: staticEnvironments,
		scopes:             []map[string]interface{}{outerScope},
		registers:          registers,
		errors:             make([]error, 0, 4), // Pre-allocate for common error count
		warnings:           make([]error, 0, 4), // Pre-allocate for common warning count
		partial:            false,
		strictVariables:    false,
		strictFilters:      false,
		resourceLimits:     resourceLimits,
		baseScopeDepth:     0,
		interrupts:         make([]interface{}, 0, 4), // Pre-allocate for interrupts
		filters:            make([]interface{}, 0, 4), // Pre-allocate for filters
		disabledTags:       make(map[string]int, 4),   // Pre-allocate map
		strainer:           nil,
		stringScanner:      NewStringScanner(""),
		templateName:       "",
	}

	// Initialize registers
	if registers.Get("cached_partials") == nil {
		registers.Set("cached_partials", make(map[string]interface{}))
	}
	if registers.Get("file_system") == nil {
		registers.Set("file_system", env.FileSystem())
	}
	if registers.Get("template_factory") == nil {
		registers.Set("template_factory", NewTemplateFactory())
	}

	ctx.rethrowErrors = config.RethrowErrors

	ctx.squashInstanceAssignsWithEnvironments()
	return ctx
}

// NewContext creates a new Context with default settings.
func NewContext() *Context {
	return BuildContext(ContextConfig{})
}

// Environment returns the environment.
func (c *Context) Environment() *Environment {
	return c.environment
}

// Scopes returns the scopes.
func (c *Context) Scopes() []map[string]interface{} {
	return c.scopes
}

// Registers returns the registers.
func (c *Context) Registers() *Registers {
	return c.registers
}

// Errors returns the errors.
func (c *Context) Errors() []error {
	return c.reporter().errors
}

// Warnings returns the warnings.
func (c *Context) Warnings() []error {
	return c.reporter().warnings
}

// AddWarning adds a warning.
func (c *Context) AddWarning(warning error) {
	r := c.reporter()
	r.warnings = append(r.warnings, warning)
}

// reporter returns the context that holds the errors and warnings of c.
func (c *Context) reporter() *Context {
	if c.diagnostics != nil {
		return c.diagnostics
	}
	return c
}

// IncrementRenderDepth increments the render depth and checks for stack overflow.
func (c *Context) IncrementRenderDepth() {
	c.renderDepth++
	if c.renderDepth > blockMaxDepth {
		panic(NewStackLevelError("Nesting too deep"))
	}
}

// DecrementRenderDepth decrements the render depth.
func (c *Context) DecrementRenderDepth() {
	c.renderDepth--
}

// ResourceLimits returns the resource limits.
func (c *Context) ResourceLimits() *ResourceLimits {
	return c.resourceLimits
}

// SetResourceLimits sets the resource limits.
func (c *Context) SetResourceLimits(rl *ResourceLimits) {
	c.resourceLimits = rl
}

// TemplateName returns the template name.
func (c *Context) TemplateName() string {
	return c.templateName
}

// SetTemplateName sets the template name.
func (c *Context) SetTemplateName(name string) {
	c.templateName = name
}

// Partial returns whether this is a partial context.
func (c *Context) Partial() bool {
	return c.partial
}

// SetPartial sets whether this is a partial context.
func (c *Context) SetPartial(partial bool) {
	c.partial = partial
}

// StrictVariables returns whether strict variables mode is enabled.
func (c *Context) StrictVariables() bool {
	return c.strictVariables
}

// SetStrictVariables sets strict variables mode.
func (c *Context) SetStrictVariables(strict bool) {
	c.strictVariables = strict
}

// StrictFilters returns whether strict filters mode is enabled.
func (c *Context) StrictFilters() bool {
	return c.strictFilters
}

// SetStrictFilters sets strict filters mode.
func (c *Context) SetStrictFilters(strict bool) {
	c.strictFilters = strict
}

// Strainer returns the strainer (creates if needed).
func (c *Context) Strainer() *StrainerTemplate {
	if c.strainer == nil {
		c.strainer = c.environment.CreateStrainer(c, c.filters, c.strictFilters)
	}
	return c.strainer
}

// AddFilters adds filters to this context.
func (c *Context) AddFilters(filters []interface{}) {
	if filters == nil {
		return
	}
	c.filters = append(c.filters, filters...)
	c.strainer = nil // Reset strainer so it's recreated with new filters
}

// Interrupt returns true if there are any unhandled interrupts.
func (c *Context) Interrupt() bool {
	return len(c.interrupts) > 0
}

// PushInterrupt pushes an interrupt to the stack.
func (c *Context) PushInterrupt(interrupt interface{}) {
	c.interrupts = append(c.interrupts, interrupt)
}

// PopInterrupt pops an interrupt from the stack.
func (c *Context) PopInterrupt() interface{} {
	if len(c.interrupts) == 0 {
		return nil
	}
	interrupt := c.interrupts[len(c.interrupts)-1]
	c.interrupts = c.interrupts[:len(c.interrupts)-1]
	return interrupt
}

// HandleError handles an error and returns the rendered error message.
func (c *Context) HandleError(err error, lineNumber *int) string {
	liquidErr := err

	// Check if it's a LiquidError (has Err *Error field)
	if le, ok := err.(LiquidError); ok {
		e := le.GetError()
		if e.TemplateName == "" {
			e.TemplateName = c.templateName
		}
		if e.LineNumber == nil {
			e.LineNumber = lineNumber
		}
	} else if e, ok := err.(*Error); ok {
		// Handle base Error type
		if e.TemplateName == "" {
			e.TemplateName = c.templateName
		}
		if e.LineNumber == nil {
			e.LineNumber = lineNumber
		}
	} else {
		// Unknown error type, wrap as InternalError
		liquidErr = NewInternalError("internal")
		le := liquidErr.(LiquidError)
		e := le.GetError()
		e.TemplateName = c.templateName
		e.LineNumber = lineNumber
	}

	r := c.reporter()
	r.errors = append(r.errors, liquidErr)
	if c.rethrowErrors {
		panic(liquidErr)
	}
	return ToS(liquidErr, nil)
}

// Invoke invokes a filter method. A call the filter cannot take (too many arguments)
// gives nil, as a filter that fails does, rather than passing its input through: a
// filter such as escape must not be skipped silently.
func (c *Context) Invoke(method string, obj interface{}, args ...interface{}) interface{} {
	result, err := c.Strainer().Invoke(method, append([]interface{}{obj}, args...)...)
	if err != nil {
		if c.strictFilters {
			panic(err)
		}
		return nil
	}
	// A filter result is checked as a write to the output would be, so that
	// a chain of filters that each grow their input stays within the render
	// length limit.
	if str, ok := result.(string); ok && c.resourceLimits != nil {
		if limit := c.resourceLimits.renderLengthLimit; limit != nil && len(str) > *limit {
			c.resourceLimits.raiseLimitsReached()
		}
	}
	return ToLiquid(result)
}

// Push pushes a new local scope on the stack.
func (c *Context) Push(newScope map[string]interface{}) {
	if newScope == nil {
		newScope = make(map[string]interface{})
	}
	c.scopes = append([]map[string]interface{}{newScope}, c.scopes...)
	c.checkOverflow()
}

// Merge merges variables into the current local scope.
func (c *Context) Merge(newScopes map[string]interface{}) {
	if len(c.scopes) == 0 {
		c.scopes = []map[string]interface{}{make(map[string]interface{})}
	}
	for k, v := range newScopes {
		c.scopes[0][k] = v
	}
}

// Pop pops from the stack.
func (c *Context) Pop() {
	if len(c.scopes) <= 1 {
		panic(NewContextError("cannot pop from context stack"))
	}
	c.scopes = c.scopes[1:]
}

// Stack pushes a new scope, executes the function, then pops.
func (c *Context) Stack(newScope map[string]interface{}, fn func()) {
	c.Push(newScope)
	defer c.Pop()
	fn()
}

// Set sets a variable in the current scope (innermost/newest scope).
func (c *Context) Set(key string, value interface{}) {
	if len(c.scopes) == 0 {
		c.scopes = []map[string]interface{}{make(map[string]interface{})}
	}
	c.scopes[0][key] = value
}

// SetLast sets a variable in the last scope (outermost/oldest scope).
// This is used by assign and capture tags, matching Ruby's context.scopes.last[@key] = value.
func (c *Context) SetLast(key string, value interface{}) {
	if len(c.scopes) == 0 {
		c.scopes = []map[string]interface{}{make(map[string]interface{})}
	}
	c.scopes[len(c.scopes)-1][key] = value
}

// Get gets a variable by evaluating an expression.
func (c *Context) Get(expression string) interface{} {
	expr := Parse(expression, c.stringScanner, nil)
	if expr == nil {
		return nil
	}
	return c.Evaluate(expr)
}

// Key returns true if the key exists.
func (c *Context) Key(key string) bool {
	return c.FindVariable(key, false) != nil
}

// Evaluate evaluates an object (calls evaluate if it has that method).
func (c *Context) Evaluate(object interface{}) interface{} {
	if object == nil {
		return nil
	}

	// Check if it's a VariableLookup
	if vl, ok := object.(*VariableLookup); ok {
		return vl.Evaluate(c)
	}

	// Check if it's a RangeLookup
	if rl, ok := object.(*RangeLookup); ok {
		startVal := c.Evaluate(rl.StartObj())
		endVal := c.Evaluate(rl.EndObj())
		startInt, _ := ToInteger(startVal)
		endInt, _ := ToInteger(endVal)
		return &Range{Start: startInt, End: endInt}
	}

	// Check if it has Evaluate method
	if evaluable, ok := object.(interface {
		Evaluate(context *Context) interface{}
	}); ok {
		return evaluable.Evaluate(c)
	}

	return object
}

// FindVariable finds a variable starting at local scope and moving up.
func (c *Context) FindVariable(key string, raiseOnNotFound bool) interface{} {
	// Key is already a string
	keyStr := key

	// As Ruby's find_variable: every scope, the outer one included, before
	// the environments. The outer scope holds the template's assign and
	// capture; the instance assigns kept from an earlier render were
	// replaced by the render data when the context was built
	// (squashInstanceAssignsWithEnvironments).
	for _, scope := range c.scopes {
		if _, ok := scope[keyStr]; ok {
			return c.lookupInScope(scope, keyStr)
		}
	}
	if n, ok := c.counters[keyStr]; ok {
		return n
	}

	// As Ruby: a name that an environment holds, nil included, is found;
	// with strict variables, a name the first environment lacks is an
	// UndefinedVariable (tryVariableFindInEnvironments)
	return c.tryVariableFindInEnvironments(keyStr, raiseOnNotFound)
}

// LookupAndEvaluate looks up and evaluates a value from an object.
func (c *Context) LookupAndEvaluate(obj map[string]interface{}, key string, raiseOnNotFound bool) interface{} {
	return c.lookupAndEvaluate(obj, key, raiseOnNotFound)
}

func (c *Context) lookupAndEvaluate(obj map[string]interface{}, key string, raiseOnNotFound bool) interface{} {
	if c.strictVariables && raiseOnNotFound {
		if _, ok := obj[key]; !ok {
			panic(NewUndefinedVariable("undefined variable " + key))
		}
	}

	value, exists := obj[key]
	if !exists {
		return nil
	}

	// A func is called once per render, as Ruby calls a proc once and keeps
	// its value; the value is kept with the render, not written in obj, which
	// may be the caller's data
	switch value.(type) {
	case func() interface{}, func(*Context) interface{}:
		value = c.funcValueOf(obj, key, value)
	}
	return c.liquidValue(value)
}

// lookupInScope returns the value of key in scope, a scope of the render
// that holds it. A func there is called and its value written back in the
// scope, as Ruby's lookup_and_evaluate: a scope belongs to the render, and
// the next value the render sets under the key (a loop variable, an assign)
// is read anew.
func (c *Context) lookupInScope(scope map[string]interface{}, key string) interface{} {
	value := scope[key]
	switch f := value.(type) {
	case func() interface{}:
		value = f()
		scope[key] = value
	case func(*Context) interface{}:
		value = f(c)
		scope[key] = value
	}
	return c.liquidValue(value)
}

// liquidValue returns the liquid form of a value read from a map, a drop
// bound to this render.
func (c *Context) liquidValue(value interface{}) interface{} {
	liquidValue := ToLiquid(value)

	// Set context on drops
	if drop, ok := liquidValue.(interface {
		SetContext(*Context)
	}); ok {
		drop.SetContext(c)
	}

	return liquidValue
}

// funcValueKey identifies a func of the data: the map it is in, and its key.
type funcValueKey struct {
	m   uintptr
	key string
}

// funcValue is the value a func of the data gave, with the map it was read
// from: the memo keeps the map alive, so that no other map takes its address
// during the render.
type funcValue struct {
	m map[string]interface{}
	v interface{}
}

// funcValueOf returns the value of fn, read from obj[key], calling it the
// first time the render reads it. obj is a map the render does not own (the
// data, the template's assigns): the caller does not change it during the
// render, so its func at key stays the same.
func (c *Context) funcValueOf(obj map[string]interface{}, key string, fn interface{}) interface{} {
	root := c.reporter()
	k := funcValueKey{m: reflect.ValueOf(obj).Pointer(), key: key}
	if kept, ok := root.funcValues[k]; ok {
		return kept.v
	}
	var v interface{}
	switch f := fn.(type) {
	case func() interface{}:
		v = f()
	case func(*Context) interface{}:
		v = f(c)
	}
	if root.funcValues == nil {
		root.funcValues = make(map[funcValueKey]funcValue)
	}
	root.funcValues[k] = funcValue{m: obj, v: v}
	return v
}

// WithDisabledTags executes a function with disabled tags.
func (c *Context) WithDisabledTags(tagNames []string, fn func()) {
	for _, name := range tagNames {
		c.disabledTags[name] = c.disabledTags[name] + 1
	}
	defer func() {
		for _, name := range tagNames {
			c.disabledTags[name] = c.disabledTags[name] - 1
			if c.disabledTags[name] <= 0 {
				delete(c.disabledTags, name)
			}
		}
	}()
	fn()
}

// TagDisabled returns true if a tag is disabled.
func (c *Context) TagDisabled(tagName string) bool {
	return c.disabledTags[tagName] > 0
}

// Counter returns the value of the increment and decrement counter name: 0
// until a tag sets it. A counter lasts the whole render of the context, and a
// lookup reads it unless a scope holds the same name.
func (c *Context) Counter(name string) int {
	return c.counters[name]
}

// SetCounter sets the increment and decrement counter name.
func (c *Context) SetCounter(name string, value int) {
	if c.counters == nil {
		c.counters = make(map[string]int)
	}
	c.counters[name] = value
}

// NewIsolatedSubcontext creates a new isolated subcontext.
func (c *Context) NewIsolatedSubcontext() *Context {
	c.checkOverflow()

	subCtx := BuildContext(ContextConfig{
		Environment:        c.environment,
		ResourceLimits:     c.resourceLimits,
		StaticEnvironments: c.staticEnvironments,
		Registers:          NewRegisters(c.registers),
	})

	subCtx.baseScopeDepth = c.baseScopeDepth + 1
	subCtx.rethrowErrors = c.rethrowErrors
	subCtx.filters = c.filters
	subCtx.strainer = nil
	subCtx.diagnostics = c.reporter()
	subCtx.disabledTags = c.disabledTags
	subCtx.cancel, subCtx.done = c.cancel, c.done
	subCtx.fatalPanics = c.fatalPanics

	return subCtx
}

// ClearInstanceAssigns clears the current scope.
func (c *Context) ClearInstanceAssigns() {
	if len(c.scopes) > 0 {
		c.scopes[0] = make(map[string]interface{})
	}
}

func (c *Context) tryVariableFindInEnvironments(key string, raiseOnNotFound bool) interface{} {
	// Check dynamic environments
	for _, env := range c.environments {
		if _, ok := env[key]; ok {
			return c.lookupAndEvaluate(env, key, raiseOnNotFound)
		}
		if c.strictVariables && raiseOnNotFound {
			panic(NewUndefinedVariable("undefined variable " + key))
		}
	}

	// Check static environments
	for _, env := range c.staticEnvironments {
		if _, ok := env[key]; ok {
			return c.lookupAndEvaluate(env, key, raiseOnNotFound)
		}
		if c.strictVariables && raiseOnNotFound {
			panic(NewUndefinedVariable("undefined variable " + key))
		}
	}

	return nil
}

func (c *Context) checkOverflow() {
	if c.overflow() {
		panic(NewStackLevelError("Nesting too deep"))
	}
}

func (c *Context) overflow() bool {
	return c.baseScopeDepth+len(c.scopes) > blockMaxDepth
}

func (c *Context) squashInstanceAssignsWithEnvironments() {
	if len(c.scopes) == 0 {
		return
	}
	lastScope := c.scopes[len(c.scopes)-1]
	for k := range lastScope {
		for _, env := range c.environments {
			if _, ok := env[k]; ok {
				lastScope[k] = c.lookupAndEvaluate(env, k, false)
				break
			}
		}
	}
}

// Context interface for TagContext
func (c *Context) Context() interface{} {
	return c
}

// ParseContext returns a ParseContextInterface (not implemented yet, returns nil).
// This is needed for TagContext interface but Context doesn't have a ParseContext.
func (c *Context) ParseContext() ParseContextInterface {
	// TODO: Create ParseContext from Context when needed
	return nil
}

// Reset clears the Context for reuse from the pool.
// This method must reset all fields to their zero values.
func (c *Context) Reset() {
	// Clear slices (keep capacity for reuse)
	c.scopes = c.scopes[:0]
	c.errors = c.errors[:0]
	c.warnings = c.warnings[:0]
	c.filters = c.filters[:0]
	c.interrupts = c.interrupts[:0]
	c.environments = c.environments[:0]
	c.staticEnvironments = c.staticEnvironments[:0]

	c.counters = nil
	c.safeHTML = nil

	// Clear maps
	if c.disabledTags != nil {
		for k := range c.disabledTags {
			delete(c.disabledTags, k)
		}
	}

	// Nil out pointer fields
	c.resourceLimits = nil
	c.registers = nil
	c.stringScanner = nil
	c.strainer = nil
	c.environment = nil
	c.diagnostics = nil
	c.cancel, c.done = nil, nil

	// Reset primitive fields
	c.templateName = ""
	c.baseScopeDepth = 0
	c.strictFilters = false
	c.strictVariables = false
	c.partial = false
}

// SetCancellation makes the render stop once goCtx is done: the next node or
// loop iteration raises a CanceledError, which ends the render.
func (c *Context) SetCancellation(goCtx gocontext.Context) {
	c.cancel = goCtx
	c.done = nil
	if goCtx != nil {
		c.done = goCtx.Done()
	}
}

// checkCanceled raises a CanceledError if the render was cancelled.
func (c *Context) checkCanceled() {
	if c == nil || c.done == nil {
		return
	}
	select {
	case <-c.done:
		err := c.cancel.Err()
		if cause := gocontext.Cause(c.cancel); cause != nil && cause != err {
			err = fmt.Errorf("%w: %w", err, cause)
		}
		panic(NewCanceledError(err))
	default:
	}
}

// CheckCanceled raises a CanceledError if the render was cancelled. A tag
// that loops calls it at each iteration.
func (c *Context) CheckCanceled() {
	c.checkCanceled()
}
