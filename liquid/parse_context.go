package liquid

// ParseContextOptions configures a ParseContext.
type ParseContextOptions struct {
	Environment     *Environment
	Locale          *I18n
	ExpressionCache map[string]interface{}
	TemplateOptions map[string]interface{}
	ErrorMode       ErrorMode
}

// ParseContext represents the context during template parsing.
type ParseContext struct {
	environment     *Environment
	locale          *I18n
	lineNumber      *int
	stringScanner   *StringScanner
	expressionCache map[string]interface{}
	templateOptions map[string]interface{}
	partialOptions  map[string]interface{}
	errorMode       ErrorMode
	warnings        []error
	depth           int
	trimWhitespace  bool
	partial         bool
}

// NewParseContext creates a new ParseContext.
func NewParseContext(options ParseContextOptions) *ParseContext {
	env := options.Environment
	if env == nil {
		env = newEnvironment()
	}

	locale := options.Locale
	if locale == nil {
		locale = NewI18n("en")
	}

	errorMode := options.ErrorMode
	if errorMode == "" {
		errorMode = env.ErrorMode()
	}

	templateOptions := options.TemplateOptions
	if templateOptions == nil {
		templateOptions = make(map[string]interface{})
	}

	pc := &ParseContext{
		environment:     env,
		locale:          locale,
		warnings:        []error{},
		errorMode:       errorMode,
		lineNumber:      nil,
		trimWhitespace:  false,
		depth:           0,
		partial:         false,
		stringScanner:   NewStringScanner(""),
		expressionCache: options.ExpressionCache,
		templateOptions: templateOptions,
	}

	if pc.expressionCache == nil {
		pc.expressionCache = make(map[string]interface{})
	}

	return pc
}

// Environment returns the environment.
func (pc *ParseContext) Environment() *Environment {
	return pc.environment
}

// Locale returns the locale.
func (pc *ParseContext) Locale() *I18n {
	return pc.locale
}

// Warnings returns the warnings.
func (pc *ParseContext) Warnings() []error {
	return pc.warnings
}

// AddWarning adds a warning.
func (pc *ParseContext) AddWarning(warning error) {
	pc.warnings = append(pc.warnings, warning)
}

// ErrorMode returns the error mode.
func (pc *ParseContext) ErrorMode() ErrorMode {
	return pc.errorMode
}

// LineNumber returns the line number.
func (pc *ParseContext) LineNumber() *int {
	return pc.lineNumber
}

// SetLineNumber sets the line number.
func (pc *ParseContext) SetLineNumber(ln *int) {
	pc.lineNumber = ln
}

// TrimWhitespace returns whether to trim whitespace.
func (pc *ParseContext) TrimWhitespace() bool {
	return pc.trimWhitespace
}

// SetTrimWhitespace sets whether to trim whitespace.
func (pc *ParseContext) SetTrimWhitespace(tw bool) {
	pc.trimWhitespace = tw
}

// Depth returns the depth.
func (pc *ParseContext) Depth() int {
	return pc.depth
}

// IncrementDepth increments the depth.
func (pc *ParseContext) IncrementDepth() {
	pc.depth++
}

// DecrementDepth decrements the depth.
func (pc *ParseContext) DecrementDepth() {
	pc.depth--
}

// Partial returns whether this is a partial parse.
func (pc *ParseContext) Partial() bool {
	return pc.partial
}

// SetPartial sets whether this is a partial parse.
func (pc *ParseContext) SetPartial(partial bool) {
	pc.partial = partial
	if partial {
		pc.partialOptions = pc.computePartialOptions()
		// Update error mode from partial options
		if mode, ok := pc.partialOptions["error_mode"].(string); ok {
			pc.errorMode = ErrorMode(mode)
		} else {
			pc.errorMode = pc.environment.ErrorMode()
		}
	} else {
		pc.partialOptions = nil
		pc.errorMode = pc.environment.ErrorMode()
	}
}

func (pc *ParseContext) computePartialOptions() map[string]interface{} {
	dontPass := pc.templateOptions["include_options_blacklist"]
	if dontPass == true {
		return map[string]interface{}{
			"locale": pc.locale,
		}
	}
	if blacklist, ok := dontPass.([]string); ok {
		result := make(map[string]interface{})
		for k, v := range pc.templateOptions {
			shouldInclude := true
			for _, blacklisted := range blacklist {
				if k == blacklisted {
					shouldInclude = false
					break
				}
			}
			if shouldInclude {
				result[k] = v
			}
		}
		return result
	}
	return pc.templateOptions
}

// GetOption gets an option value.
func (pc *ParseContext) GetOption(key string) interface{} {
	if pc.partial && pc.partialOptions != nil {
		return pc.partialOptions[key]
	}
	return pc.templateOptions[key]
}

// NewBlockBody creates a new BlockBody.
func (pc *ParseContext) NewBlockBody() *BlockBody {
	return NewBlockBody()
}

// NewParser creates a new Parser with the shared StringScanner.
func (pc *ParseContext) NewParser(input string) *Parser {
	pc.stringScanner.SetString(input)
	// Create parser from scanner (it will tokenize the scanner's string)
	p := NewParser(pc.stringScanner)
	p.rejectBareBrackets = rigidMode(pc.errorMode)
	return p
}

// NewTokenizer creates a new Tokenizer with the shared StringScanner.
func (pc *ParseContext) NewTokenizer(source string, lineNumbers bool, startLineNumber *int, forLiquidTag bool) *Tokenizer {
	return NewTokenizer(source, pc.stringScanner, lineNumbers, startLineNumber, forLiquidTag)
}

// SafeParseExpression parses the next expression of parser, as Ruby's
// safe_parse_expression: a lexer or a parser error is raised in every error
// mode. In lax mode the variable parse that tries the strict parse first
// catches it and parses the markup as lax, as Ruby's.
func (pc *ParseContext) SafeParseExpression(parser *Parser) interface{} {
	if err := parser.Error(); err != nil {
		panic(err)
	}
	expr, err := parser.Expression()
	if err != nil {
		panic(err)
	}
	return Parse(expr, pc.stringScanner, pc.expressionCache)
}

// ParseExpression parses an expression.
func (pc *ParseContext) ParseExpression(markup string) interface{} {
	return Parse(markup, pc.stringScanner, pc.expressionCache)
}

// ParseExpressionSafe parses an expression with safe flag.
func (pc *ParseContext) ParseExpressionSafe(markup string, safe bool) interface{} {
	if !safe && rigidMode(pc.errorMode) {
		panic(NewInternalError("unsafe parse_expression cannot be used in rigid mode"))
	}
	return Parse(markup, pc.stringScanner, pc.expressionCache)
}
