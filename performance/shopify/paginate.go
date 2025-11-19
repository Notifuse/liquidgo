package shopify

import (
	"fmt"
	"math"
	"regexp"
	"strconv"

	"github.com/Notifuse/liquidgo/liquid"
)

// Paginate implements the paginate block tag
type Paginate struct {
	*liquid.Block
	collectionName string
	pageSize       int
	windowSize     int
}

// NewPaginate creates a new Paginate tag
func NewPaginate(tagName, markup string, parseContext liquid.ParseContextInterface) (*Paginate, error) {
	// Syntax: paginate [collection] by number
	quotedFragment := `(?:[^\s,\|'"]|"[^"]*"|'[^']*')+`
	re := regexp.MustCompile(fmt.Sprintf(`^(%s)\s*(?:by\s*(\d+))?`, quotedFragment))
	matches := re.FindStringSubmatch(markup)

	if matches == nil {
		return nil, fmt.Errorf("syntax error in tag 'paginate' - valid syntax: paginate [collection] by number")
	}

	block := liquid.NewBlock(tagName, markup, parseContext)

	p := &Paginate{
		Block:          block,
		collectionName: matches[1],
		pageSize:       20, // default
		windowSize:     3,  // default
	}

	if matches[2] != "" {
		if size, err := strconv.Atoi(matches[2]); err == nil {
			p.pageSize = size
		}
	}

	// Parse attributes
	attrRe := regexp.MustCompile(`(\w+)\s*:\s*('[^']*'|"[^"]*"|\d+)`)
	attrMatches := attrRe.FindAllStringSubmatch(markup, -1)
	for _, match := range attrMatches {
		if match[1] == "window_size" {
			if size, err := strconv.Atoi(match[2]); err == nil {
				p.windowSize = size
			}
		}
	}

	return p, nil
}

// RenderToOutputBuffer renders the paginate tag
func (p *Paginate) RenderToOutputBuffer(context liquid.TagContext, output *string) {
	ctx := context.Context().(*liquid.Context)
	ctx.Push(make(map[string]interface{}))
	defer ctx.Pop()

	currentPage := 5 // default for benchmarking

	if cp := ctx.FindVariable("current_page", false); cp != nil {
		if i, ok := cp.(int); ok {
			currentPage = i
		}
	}

	pagination := map[string]interface{}{
		"page_size":      p.pageSize,
		"current_page":   currentPage,
		"current_offset": p.pageSize * currentPage,
	}

	ctx.Set("paginate", pagination)

	collection := ctx.FindVariable(p.collectionName, false)
	if collection == nil {
		// In non-error mode, just render the block
		bodyOutput := p.Render(context)
		*output += bodyOutput
		return
	}

	collectionSize := 0
	switch c := collection.(type) {
	case []interface{}:
		collectionSize = len(c)
	case map[string]interface{}:
		collectionSize = len(c)
	default:
		// In non-error mode, just render the block
		bodyOutput := p.Render(context)
		*output += bodyOutput
		return
	}

	pageCount := int(math.Ceil(float64(collectionSize)/float64(p.pageSize))) + 1

	pagination["items"] = collectionSize
	pagination["pages"] = pageCount - 1

	// Add previous link
	if currentPage > 1 {
		pagination["previous"] = map[string]interface{}{
			"title":   "&laquo; Previous",
			"url":     p.link(currentPage - 1),
			"is_link": true,
		}
	}

	// Add next link
	if currentPage+1 < pageCount {
		pagination["next"] = map[string]interface{}{
			"title":   "Next &raquo;",
			"url":     p.link(currentPage + 1),
			"is_link": true,
		}
	}

	// Build parts
	var parts []interface{}
	hellipBreak := false

	if pageCount > 2 {
		for page := 1; page < pageCount; page++ {
			if currentPage == page {
				parts = append(parts, p.noLink(fmt.Sprint(page)))
			} else if page == 1 {
				parts = append(parts, p.link(page))
			} else if page == pageCount-1 {
				parts = append(parts, p.link(page))
			} else if page <= currentPage-p.windowSize || page >= currentPage+p.windowSize {
				if !hellipBreak {
					parts = append(parts, p.noLink("&hellip;"))
					hellipBreak = true
				}
				continue
			} else {
				parts = append(parts, p.link(page))
			}
			hellipBreak = false
		}
	}

	pagination["parts"] = parts

	// Render block content
	bodyOutput := p.Render(context)
	*output += bodyOutput
}

func (p *Paginate) noLink(title string) map[string]interface{} {
	return map[string]interface{}{
		"title":   title,
		"is_link": false,
	}
}

func (p *Paginate) link(page int) map[string]interface{} {
	return map[string]interface{}{
		"title":   fmt.Sprint(page),
		"url":     fmt.Sprintf("%s?page=%d", p.currentURL(), page),
		"is_link": true,
	}
}

func (p *Paginate) currentURL() string {
	return "/collections/frontpage"
}
