package shopify

import (
	"fmt"
	"regexp"

	"github.com/Notifuse/liquidgo/liquid"
)

// CommentForm implements the comment_form block tag
type CommentForm struct {
	*liquid.Block
	variableName string
}

// NewCommentForm creates a new CommentForm tag
func NewCommentForm(tagName, markup string, parseContext liquid.ParseContextInterface) (*CommentForm, error) {
	// Syntax: comment_form [article]
	variableSignature := `[a-zA-Z_][\w\-]*(?:\.[a-zA-Z_][\w\-]*)*`
	re := regexp.MustCompile(fmt.Sprintf(`^(%s)`, variableSignature))
	matches := re.FindStringSubmatch(markup)

	if matches == nil {
		return nil, fmt.Errorf("Syntax Error in 'comment_form' - Valid syntax: comment_form [article]")
	}

	block := liquid.NewBlock(tagName, markup, parseContext)

	return &CommentForm{
		Block:        block,
		variableName: matches[1],
	}, nil
}

// RenderToOutputBuffer renders the comment_form tag
func (c *CommentForm) RenderToOutputBuffer(context liquid.TagContext, output *string) {
	ctx := context.Context().(*liquid.Context)
	article := ctx.FindVariable(c.variableName, false)

	// Create new scope
	ctx.Push(make(map[string]interface{}))
	defer ctx.Pop()

	// Get registers
	registers := context.Registers()
	
	// Set form context
	form := map[string]interface{}{
		"posted_successfully?": registers.Get("posted_successfully"),
		"errors":               ctx.FindVariable("comment.errors", false),
		"author":               ctx.FindVariable("comment.author", false),
		"email":                ctx.FindVariable("comment.email", false),
		"body":                 ctx.FindVariable("comment.body", false),
	}
	ctx.Set("form", form)

	// Render block content
	bodyOutput := c.Block.Render(context)

	// Wrap in form tag
	articleID := "unknown"
	if a, ok := article.(map[string]interface{}); ok {
		if id, ok := a["id"]; ok {
			articleID = fmt.Sprint(id)
		}
	}

	formHTML := fmt.Sprintf(`<form id="article-%s-comment-form" class="comment-form" method="post" action="">
%s
</form>`, articleID, bodyOutput)

	*output += formHTML
}

