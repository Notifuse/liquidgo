package liquid

// TemplateFactory creates template instances.
type TemplateFactory struct{}

// NewTemplateFactory creates a new TemplateFactory.
func NewTemplateFactory() *TemplateFactory {
	return &TemplateFactory{}
}

// For returns a template instance for the given template name.
func (tf *TemplateFactory) For(templateName string) interface{} {
	// Return a new Template instance (name is not used in Ruby implementation)
	return NewTemplate(nil)
}
