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
	ResourceLimits     *ResourceLimits
	Environments       []map[string]interface{}
	StaticEnvironments []map[string]interface{}
	RethrowErrors      bool
}

// emptyEnvironment is the environment of a context built with none; nothing
// writes a context's environments, so contexts share it.
var emptyEnvironment = map[string]interface{}{}

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
	// self is the SelfDrop of the context, made when a template first reads
	// self (see FindVariable).
	self *SelfDrop
	// argScratch is the slice lent to the calls of standard filters, and
	// argLent whether a call holds it (see lendFilterArgs).
	argScratch []interface{}
	argLent    bool
	// fatalPanics makes a Go panic in a node end the render (Template.Render)
	// instead of being written as "Liquid error: internal".
	fatalPanics bool
	// rethrowErrors makes HandleError raise (panic with) the error it
	// records, where it returns its message: Ruby's rethrow_errors.
	rethrowErrors bool
	// diagnostics is the context whose errors and warnings this one adds
	// to: an isolated subcontext reports to the context that made it, as
	// Ruby's subcontexts share the errors array.
	diagnostics   *Context
	registers     *Registers
	stringScanner *StringScanner
	environment   *Environment
	templateName  string
	warnings      []error
	environments  []map[string]interface{}
	filters       []FilterDef
	// sf holds the standard filters of the render, which read its limits.
	sf                 StandardFilters
	interrupts         []interface{}
	errors             []error
	scopes             []map[string]interface{}
	staticEnvironments []map[string]interface{}
	baseScopeDepth     int
	// counters holds the increment and decrement counters of the render.
	// Ruby keeps them in the first environment; they are read after the
	// scopes and before the environments, as there.
	counters        map[string]int
	strictFilters   bool
	strictVariables bool
	partial         bool
	// safeHTML holds the texts of the captures of the render, which an
	// output does not escape (see EnvironmentConfig.OutputEscape). Only the
	// reporter context of the render keeps it.
	safeHTML map[string]struct{}
}

// BuildContext creates a new Context with the given configuration.
func BuildContext(config ContextConfig) *Context {
	env := config.Environment
	if env == nil {
		env = newEnvironment()
	}

	// As Ruby's Context.new, whose environments default to {}: a context
	// given none has one empty environment, which strict variables look a
	// name up in (and nothing writes)
	environments := config.Environments
	if environments == nil {
		environments = []map[string]interface{}{emptyEnvironment}
	}

	staticEnvironments := config.StaticEnvironments
	if staticEnvironments == nil {
		staticEnvironments = []map[string]interface{}{}
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
		scopes:             []map[string]interface{}{{}},
		registers:          registers,
		errors:             make([]error, 0, 4), // Pre-allocate for common error count
		warnings:           make([]error, 0, 4), // Pre-allocate for common warning count
		partial:            false,
		strictVariables:    false,
		strictFilters:      false,
		resourceLimits:     resourceLimits,
		baseScopeDepth:     0,
		interrupts:         make([]interface{}, 0, 4), // Pre-allocate for interrupts
		disabledTags:       make(map[string]int, 4),   // Pre-allocate map
		stringScanner:      NewStringScanner(""),
		templateName:       "",
	}

	// Initialize registers
	if registers.Get("file_system") == nil {
		registers.Set("file_system", env.FileSystem())
	}

	ctx.rethrowErrors = config.RethrowErrors

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

// AddFilters adds local filters to this context, which take precedence over
// the environment's and the standard ones, the last added first. A filter
// without a function or a name a template can write is an error, and none is
// added.
func (c *Context) AddFilters(filters []FilterDef) error {
	for _, f := range filters {
		if err := f.check(); err != nil {
			return err
		}
	}
	c.filters = append(c.filters, filters...)
	return nil
}

// standard returns the standard filters of the render.
func (c *Context) standard() *StandardFilters {
	c.sf.context = c
	return &c.sf
}

// filter returns the function of the filter name: a local filter of the
// render, the last added first, then the environment's, then the standard
// one; nil for an unknown filter.
// resolveFilter returns the filter name, and whether it is a standard
// filter that no local or environment filter takes the place of: such a
// filter never keeps its arguments (see lendFilterArgs). The local filters
// come first, the last added first, then the environment's.
func (c *Context) resolveFilter(name string) (FilterFunc, bool) {
	for i := len(c.filters) - 1; i >= 0; i-- {
		if c.filters[i].name == name {
			return c.filters[i].fn, false
		}
	}
	if c.environment != nil && len(c.environment.filters) > 0 {
		if fn, ok := c.environment.filters[name]; ok {
			return fn, false
		}
	}
	fn, ok := standardFilters[name]
	return fn, ok
}

// lendFilterArgs returns a slice of n arguments for a call of a standard
// filter, which never keeps them: the context lends the same slice at each
// call; give it back with returnFilterArgs. It returns nil while the slice is
// lent, and the caller makes a slice of its own.
func (c *Context) lendFilterArgs(n int) []interface{} {
	if c.argLent {
		return nil
	}
	if cap(c.argScratch) < n {
		c.argScratch = make([]interface{}, 0, max(n, 4))
	}
	c.argLent = true
	return c.argScratch[:n]
}

// returnFilterArgs takes back the slice lendFilterArgs lent, emptied so that
// it keeps no value alive.
func (c *Context) returnFilterArgs(args []interface{}) {
	clear(args)
	c.argLent = false
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
		liquidErr = asLiquidError(err)
		le := liquidErr.(LiquidError)
		e := le.GetError()
		e.TemplateName = c.templateName
		e.LineNumber = lineNumber
	}

	r := c.reporter()
	if c.rethrowErrors {
		// an error raised through enclosing blocks comes back here from
		// each of them: it is recorded once
		if !recorded(r.errors, liquidErr) {
			r.errors = append(r.errors, liquidErr)
		}
		panic(liquidErr)
	}
	r.errors = append(r.errors, liquidErr)
	return ToS(liquidErr, nil)
}

// recordError records err, a Liquid error, without writing or raising it.
func (c *Context) recordError(err error) {
	if le, ok := err.(LiquidError); ok {
		if e := le.GetError(); e != nil && e.TemplateName == "" {
			e.TemplateName = c.templateName
		}
	}
	r := c.reporter()
	r.errors = append(r.errors, err)
}

// recorded reports whether err is already in errs, the same value.
func recorded(errs []error, err error) bool {
	for _, e := range errs {
		if sameError(e, err) {
			return true
		}
	}
	return false
}

// sameError reports whether a and b are the same error value. Two values of
// a type that holds an uncomparable value (a slice in an interface field)
// are not the same: comparing them would panic.
func sameError(a, b error) (same bool) {
	if reflect.TypeOf(a) != reflect.TypeOf(b) {
		return false
	}
	defer func() {
		if recover() != nil {
			same = false
		}
	}()
	return a == b
}

// Invoke invokes a filter method. A call the filter cannot take (too many arguments)
// gives nil and is recorded, as a filter that fails is, rather than passing its input
// through: a filter such as escape must not be skipped silently.
func (c *Context) Invoke(method string, obj interface{}, args ...interface{}) interface{} {
	fn, _ := c.resolveFilter(method)
	return c.invokeFilter(method, fn, obj, args)
}

// invokeFilter is Invoke for fn, the filter method resolves to (nil for an
// unknown one).
func (c *Context) invokeFilter(method string, fn FilterFunc, obj interface{}, args []interface{}) interface{} {
	var result interface{}
	var err error
	if fn != nil {
		result, err = fn(c, obj, args...)
	} else if c.strictFilters {
		// An unknown filter stays an unknown filter: it never reads a
		// property or calls a method of its input, as in Ruby.
		err = NewUndefinedFilter("undefined filter " + method)
	} else {
		result = obj
	}
	if err != nil {
		// An error of a filter is raised with strict filters; otherwise it
		// is recorded and the filter gives nil, so that no error text is
		// written (a deliberate divergence from Ruby).
		err = asLiquidError(err)
		if c.strictFilters {
			panic(err)
		}
		c.recordError(err)
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

// Push pushes a new local scope on the stack. newScope belongs to the
// render from then on, as in Ruby: the render writes its assigns and the
// value of a func it holds in it. Push a map of the render, never one that
// renders share.
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

// Stack pushes a new scope, executes the function, then pops. newScope
// belongs to the render, as with Push.
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
	switch o := object.(type) {
	case nil:
		return nil
	case string, int, float64, bool:
		return object // a literal: no method lookup
	case *VariableLookup:
		return o.Evaluate(c)
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
	// capture.
	for _, scope := range c.scopes {
		if len(scope) == 0 {
			continue // no hash of the key for an empty scope
		}
		if value, ok := scope[keyStr]; ok {
			return c.lookupInScope(scope, keyStr, value)
		}
	}
	if len(c.counters) > 0 {
		if n, ok := c.counters[keyStr]; ok {
			return n
		}
	}

	// As Ruby: the first environment that gives a value; with strict
	// variables, the first environment decides (tryVariableFindInEnvironments).
	// self, found nowhere, is the SelfDrop of the context (Shopify Liquid 5.13).
	isSelf := keyStr == selfKeyword
	value := c.tryVariableFindInEnvironments(keyStr, raiseOnNotFound && !isSelf)
	if isSelf && value == nil {
		if c.self == nil {
			c.self = &SelfDrop{context: c}
		}
		return c.self
	}
	return value
}

// variableDefined reports whether key is in a scope or an environment, nil
// values included, as Ruby's variable_defined?.
func (c *Context) variableDefined(key string) bool {
	for _, scopes := range [3][]map[string]interface{}{c.scopes, c.environments, c.staticEnvironments} {
		for _, s := range scopes {
			if _, ok := s[key]; ok {
				return true
			}
		}
	}
	return false
}

// valueIn returns the value of key in obj, a map the render does not own
// (the data, the template's assigns, a map of them), before its liquid
// form: a func is called once per render, as Ruby calls a proc once and
// keeps its value; the value is kept with the render, not written in obj.
func (c *Context) valueIn(obj map[string]interface{}, key string) interface{} {
	return c.mapValue(obj, key, obj[key])
}

// mapValue is valueIn for value, read from obj under key.
func (c *Context) mapValue(obj map[string]interface{}, key string, value interface{}) interface{} {
	switch value.(type) {
	case func() interface{}, func(*Context) interface{}:
		value = c.funcValueOf(obj, key, value)
	}
	return value
}

// lookupInScope returns value, the value of key in scope, a scope of the render
// that holds it. A func there is called and its value written back in the
// scope, as Ruby's lookup_and_evaluate: a scope belongs to the render, and
// the next value the render sets under the key (a loop variable, an assign)
// is read anew.
func (c *Context) lookupInScope(scope map[string]interface{}, key string, value interface{}) interface{} {
	switch value.(type) {
	case func() interface{}, func(*Context) interface{}:
		value = c.callFunc(value)
		scope[key] = value
	}
	return ToLiquid(value)
}

// itemOf returns item i of list, a list the render does not own: a func is
// called once per render, as Ruby's lookup_and_evaluate calls a proc read by
// an index, and the list is never written.
func (c *Context) itemOf(list []interface{}, i int) interface{} {
	value := list[i]
	switch value.(type) {
	case func() interface{}, func(*Context) interface{}:
		root := c.reporter()
		k := funcValueKey{m: reflect.ValueOf(list).Pointer(), index: i + 1}
		if kept, ok := root.funcValues[k]; ok {
			return kept.v
		}
		v := c.callFunc(value)
		root.keepFuncValue(k, list, v)
		return v
	}
	return value
}

// funcValueKey identifies a func of the data: the map it is in and its key,
// or the list it is in and its index plus one.
type funcValueKey struct {
	m     uintptr
	key   string
	index int
}

// funcValue is the value a func of the data gave, with the map or list it
// was read from: the memo keeps it alive, so that no other map or list, one a
// filter or a host tag makes during the render, takes its address.
type funcValue struct {
	in interface{}
	v  interface{}
}

// funcValueOf returns the value of fn, read from obj[key], calling it the
// first time the render reads it. obj is a map the render does not own (the
// data, the template's assigns, a map of them): the caller does not change
// it during the render, so its func at key stays the same.
func (c *Context) funcValueOf(obj map[string]interface{}, key string, fn interface{}) interface{} {
	root := c.reporter()
	k := funcValueKey{m: reflect.ValueOf(obj).Pointer(), key: key}
	if kept, ok := root.funcValues[k]; ok {
		return kept.v
	}
	v := c.callFunc(fn)
	root.keepFuncValue(k, obj, v)
	return v
}

// callFunc returns the value of a func of the data.
func (c *Context) callFunc(fn interface{}) interface{} {
	switch f := fn.(type) {
	case func() interface{}:
		return f()
	case func(*Context) interface{}:
		return f(c)
	}
	return fn
}

// keepFuncValue keeps v, the value of the func at k in in, for the render.
func (c *Context) keepFuncValue(k funcValueKey, in interface{}, v interface{}) {
	if c.funcValues == nil {
		c.funcValues = make(map[funcValueKey]funcValue)
	}
	c.funcValues[k] = funcValue{in: in, v: v}
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
	subCtx.diagnostics = c.reporter()
	subCtx.disabledTags = c.disabledTags
	subCtx.cancel, subCtx.done = c.cancel, c.done
	subCtx.fatalPanics = c.fatalPanics

	return subCtx
}

func (c *Context) tryVariableFindInEnvironments(key string, raiseOnNotFound bool) interface{} {
	// As Ruby's try_variable_find_in_environments: the first environment
	// that gives a value other than nil; with strict variables, the first
	// environment decides, nil included, and one that lacks the name raises
	strict := c.strictVariables && raiseOnNotFound
	for _, envs := range [2][]map[string]interface{}{c.environments, c.staticEnvironments} {
		for _, env := range envs {
			value, ok := env[key]
			if !ok {
				if strict {
					panic(NewUndefinedVariable("undefined variable " + key))
				}
				continue
			}
			// as Ruby, the value decides before its liquid form
			if v := c.mapValue(env, key, value); v != nil || strict {
				return ToLiquid(v)
			}
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
