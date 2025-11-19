package liquid

// SnippetDrop represents a snippet drop.
type SnippetDrop struct {
	*Drop
	body     string
	name     string
	filename string
}

// NewSnippetDrop creates a new SnippetDrop.
func NewSnippetDrop(body string, name string, filename string) *SnippetDrop {
	return &SnippetDrop{
		Drop:     NewDrop(),
		body:     body,
		name:     name,
		filename: filename,
	}
}

// Body returns the body of the snippet.
func (s *SnippetDrop) Body() string {
	return s.body
}

// Name returns the name of the snippet.
func (s *SnippetDrop) Name() string {
	return s.name
}

// Filename returns the filename of the snippet.
func (s *SnippetDrop) Filename() string {
	return s.filename
}

// ToPartial returns the body as a partial.
func (s *SnippetDrop) ToPartial() string {
	return s.body
}

// String returns the string representation.
func (s *SnippetDrop) String() string {
	return "SnippetDrop"
}

// InvokeDrop invokes a method on the snippet drop.
func (s *SnippetDrop) InvokeDrop(methodOrKey string) interface{} {
	return InvokeDropOn(s, methodOrKey)
}
