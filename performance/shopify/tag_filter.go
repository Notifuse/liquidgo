package shopify

import (
	"fmt"
	"strings"
)

// TagFilter provides tag-related template filters
type TagFilter struct {
	context map[string]interface{}
}

// SetContext sets the rendering context
func (f *TagFilter) SetContext(ctx map[string]interface{}) {
	f.context = ctx
}

// LinkToTag generates a link to a tag
func (f *TagFilter) LinkToTag(label, tag string) string {
	handle := ""
	if f.context != nil {
		if h, ok := f.context["handle"].(string); ok {
			handle = h
		}
	}
	return fmt.Sprintf(`<a title="Show tag %s" href="/collections/%s/%s">%s</a>`, tag, handle, tag, label)
}

// HighlightActiveTag highlights the active tag with CSS class
func (f *TagFilter) HighlightActiveTag(tag string, cssClass ...string) string {
	class := "active"
	if len(cssClass) > 0 {
		class = cssClass[0]
	}

	if f.context != nil {
		if currentTags, ok := f.context["current_tags"].([]string); ok {
			for _, t := range currentTags {
				if t == tag {
					return fmt.Sprintf(`<span class="%s">%s</span>`, class, tag)
				}
			}
		}
	}

	return tag
}

// LinkToAddTag generates a link that adds a tag to current tags
func (f *TagFilter) LinkToAddTag(label, tag string) string {
	handle := ""
	var currentTags []string

	if f.context != nil {
		if h, ok := f.context["handle"].(string); ok {
			handle = h
		}
		if ct, ok := f.context["current_tags"].([]string); ok {
			currentTags = ct
		}
	}

	// Add tag if not already present
	tags := append([]string{}, currentTags...)
	found := false
	for _, t := range tags {
		if t == tag {
			found = true
			break
		}
	}
	if !found {
		tags = append(tags, tag)
	}

	tagsStr := strings.Join(tags, "+")
	return fmt.Sprintf(`<a title="Show tag %s" href="/collections/%s/%s">%s</a>`, tag, handle, tagsStr, label)
}

// LinkToRemoveTag generates a link that removes a tag from current tags
func (f *TagFilter) LinkToRemoveTag(label, tag string) string {
	handle := ""
	var currentTags []string

	if f.context != nil {
		if h, ok := f.context["handle"].(string); ok {
			handle = h
		}
		if ct, ok := f.context["current_tags"].([]string); ok {
			currentTags = ct
		}
	}

	// Remove tag
	var tags []string
	for _, t := range currentTags {
		if t != tag {
			tags = append(tags, t)
		}
	}

	tagsStr := strings.Join(tags, "+")
	return fmt.Sprintf(`<a title="Show tag %s" href="/collections/%s/%s">%s</a>`, tag, handle, tagsStr, label)
}

