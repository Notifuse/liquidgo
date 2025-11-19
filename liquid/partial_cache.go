package liquid

// PartialCache provides caching for partial templates.
type PartialCache struct{}

// Load loads a partial template from cache or file system.
func (pc *PartialCache) Load(templateName string, context interface {
	Registers() *Registers
}, parseContext ParseContextInterface) (interface{}, error) {
	registers := context.Registers()
	cachedPartials := registers.Get("cached_partials")

	var cache map[string]interface{}
	if cp, ok := cachedPartials.(map[string]interface{}); ok {
		cache = cp
	} else {
		cache = make(map[string]interface{})
		registers.Set("cached_partials", cache)
	}

	// Create cache key
	errorMode := "lax" // Default
	if env := parseContext.Environment(); env != nil {
		errorMode = env.ErrorMode()
	}
	cacheKey := templateName + ":" + errorMode

	// Check cache
	if cached, ok := cache[cacheKey]; ok {
		return cached, nil
	}

	// Load from file system
	fileSystem := registers.Get("file_system")
	var fs FileSystem
	if f, ok := fileSystem.(FileSystem); ok {
		fs = f
	} else {
		fs = &BlankFileSystem{}
	}

	source, err := fs.ReadTemplateFile(templateName)
	if err != nil {
		return nil, err
	}

	// Set partial flag
	if pc, ok := parseContext.(*ParseContext); ok {
		pc.SetPartial(true)
		defer pc.SetPartial(false)
	}

	// Get template factory
	templateFactory := registers.Get("template_factory")
	var tf *TemplateFactory
	if t, ok := templateFactory.(*TemplateFactory); ok {
		tf = t
	} else {
		tf = NewTemplateFactory()
	}

	// Get template instance
	template := tf.For(templateName)
	var tmpl *Template
	if t, ok := template.(*Template); ok {
		tmpl = t
	} else {
		return nil, NewFileSystemError("template factory returned invalid template")
	}

	// Parse the template
	parseOptions := &TemplateOptions{}
	if pc, ok := parseContext.(*ParseContext); ok {
		parseOptions.Environment = pc.Environment()
	}
	err = tmpl.Parse(source, parseOptions)
	if err != nil {
		// Set template name on error if available
		if liquidErr, ok := err.(*Error); ok {
			if tmpl.Name() == "" {
				liquidErr.TemplateName = templateName
			} else {
				liquidErr.TemplateName = tmpl.Name()
			}
		}
		return nil, err
	}

	// Set name if not already set
	if tmpl.Name() == "" {
		tmpl.SetName(templateName)
	}

	// Cache the partial
	cache[cacheKey] = tmpl

	return tmpl, nil
}

// LoadPartial is a convenience function to load a partial.
func LoadPartial(templateName string, context interface {
	Registers() *Registers
}, parseContext ParseContextInterface) (interface{}, error) {
	pc := &PartialCache{}
	return pc.Load(templateName, context, parseContext)
}
