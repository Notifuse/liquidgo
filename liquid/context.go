package liquid

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
	disabledTags       map[string]int
	resourceLimits     *ResourceLimits
	profiler           *Profiler
	exceptionRenderer  func(error) interface{}
	registers          *Registers
	stringScanner      *StringScanner
	strainer           *StrainerTemplate
	environment        *Environment
	globalFilter       func(interface{}) interface{}
	templateName       string
	warnings           []error
	environments       []map[string]interface{}
	filters            []interface{}
	interrupts         []interface{}
	errors             []error
	scopes             []map[string]interface{}
	staticEnvironments []map[string]interface{}
	baseScopeDepth     int
	renderDepth        int             // Track nesting depth during rendering
	dropInvokeStack    map[string]bool // Track drop method invocations to prevent infinite recursion
	strictFilters      bool
	strictVariables    bool
	partial            bool
}

// BuildContext creates a new Context with the given configuration.
func BuildContext(config ContextConfig) *Context {
	env := config.Environment
	if env == nil {
		env = NewEnvironment()
	}

	environments := config.Environments
	if environments == nil {
		environments = []map[string]interface{}{}
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
		globalFilter:       nil,
		disabledTags:       make(map[string]int, 4), // Pre-allocate map
		strainer:           nil,
		stringScanner:      NewStringScanner(""),
		templateName:       "",
		exceptionRenderer:  env.ExceptionRenderer(),
		dropInvokeStack:    make(map[string]bool),
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

	if config.RethrowErrors {
		ctx.exceptionRenderer = func(err error) interface{} {
			panic(err)
		}
	}

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
	return c.errors
}

// Warnings returns the warnings.
func (c *Context) Warnings() []error {
	return c.warnings
}

// AddWarning adds a warning.
func (c *Context) AddWarning(warning error) {
	c.warnings = append(c.warnings, warning)
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

// GlobalFilter returns the global filter function.
func (c *Context) GlobalFilter() func(interface{}) interface{} {
	return c.globalFilter
}

// SetGlobalFilter sets the global filter function.
func (c *Context) SetGlobalFilter(filter func(interface{}) interface{}) {
	c.globalFilter = filter
}

// ExceptionRenderer returns the exception renderer.
func (c *Context) ExceptionRenderer() func(error) interface{} {
	return c.exceptionRenderer
}

// SetExceptionRenderer sets the exception renderer.
func (c *Context) SetExceptionRenderer(renderer func(error) interface{}) {
	c.exceptionRenderer = renderer
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

// ApplyGlobalFilter applies the global filter to an object.
func (c *Context) ApplyGlobalFilter(obj interface{}) interface{} {
	if c.globalFilter == nil {
		return obj
	}
	return c.globalFilter(obj)
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

	c.errors = append(c.errors, liquidErr)
	result := c.exceptionRenderer(liquidErr)
	return ToS(result, nil)
}

// Invoke invokes a filter method.
func (c *Context) Invoke(method string, obj interface{}, args ...interface{}) interface{} {
	result, err := c.Strainer().Invoke(method, append([]interface{}{obj}, args...)...)
	if err != nil {
		if c.strictFilters {
			panic(err)
		}
		return obj
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

	// Check scopes except the last one (outerScope)
	// We want to check environments before outerScope so custom assigns override instance assigns
	scopesToCheck := c.scopes
	var outerScope map[string]interface{}
	if len(scopesToCheck) > 0 {
		outerScope = scopesToCheck[len(scopesToCheck)-1]
		scopesToCheck = scopesToCheck[:len(scopesToCheck)-1]
	}

	for _, scope := range scopesToCheck {
		if _, ok := scope[keyStr]; ok {
			return c.lookupAndEvaluate(scope, keyStr, raiseOnNotFound)
		}
	}

	// Check environments (includes custom assigns which should override instance assigns)
	variable := c.tryVariableFindInEnvironments(keyStr, raiseOnNotFound)
	if variable != nil {
		return variable
	}

	// Check outerScope (instance assigns) last, after environments
	if outerScope != nil {
		if _, ok := outerScope[keyStr]; ok {
			return c.lookupAndEvaluate(outerScope, keyStr, raiseOnNotFound)
		}
	}

	// Check if there's a drop in the outermost scope that can handle this key
	// This allows drops to be used as context (Ruby behavior)
	if outerScope != nil {
		if drop, ok := outerScope["__drop__"]; ok {
			// Check for infinite recursion in drop invocation
			// This prevents cycles like: FindVariable -> InvokeDropOn -> LiquidMethodMissing -> Context.Get -> FindVariable
			if c.dropInvokeStack[keyStr] {
				// Already invoking drop for this key, return nil to break the cycle
				return nil
			}

			// Mark this key as being invoked on the drop
			c.dropInvokeStack[keyStr] = true
			defer func() {
				// Clean up after invocation
				delete(c.dropInvokeStack, keyStr)
			}()

			// Always try to invoke on the drop - if the method doesn't exist,
			// InvokeDropOn will call LiquidMethodMissing as a fallback
			return InvokeDropOn(drop, keyStr)
		}
	}

	if raiseOnNotFound && c.strictVariables {
		panic(NewUndefinedVariable("undefined variable " + keyStr))
	}

	return nil
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

	// Handle procs/functions
	if fn, ok := value.(func() interface{}); ok {
		value = fn()
		obj[key] = value
	} else if fn, ok := value.(func(*Context) interface{}); ok {
		value = fn(c)
		obj[key] = value
	}

	// Convert to liquid
	liquidValue := ToLiquid(value)

	// Set context on drops
	if drop, ok := liquidValue.(interface {
		SetContext(*Context)
	}); ok {
		drop.SetContext(c)
	}

	return liquidValue
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
	subCtx.exceptionRenderer = c.exceptionRenderer
	subCtx.filters = c.filters
	subCtx.strainer = nil
	subCtx.errors = c.errors
	subCtx.warnings = c.warnings
	subCtx.disabledTags = c.disabledTags
	subCtx.profiler = c.profiler

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

// Profiler returns the profiler.
func (c *Context) Profiler() *Profiler {
	return c.profiler
}

// SetProfiler sets the profiler.
func (c *Context) SetProfiler(profiler *Profiler) {
	c.profiler = profiler
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

	// Clear maps
	if c.disabledTags != nil {
		for k := range c.disabledTags {
			delete(c.disabledTags, k)
		}
	}

	// Nil out pointer fields
	c.resourceLimits = nil
	c.profiler = nil
	c.exceptionRenderer = nil
	c.registers = nil
	c.stringScanner = nil
	c.strainer = nil
	c.environment = nil
	c.globalFilter = nil

	// Reset primitive fields
	c.templateName = ""
	c.baseScopeDepth = 0
	c.strictFilters = false
	c.strictVariables = false
	c.partial = false
}
