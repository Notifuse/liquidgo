package tag

import (
	"github.com/Notifuse/liquidgo/v5/liquid"
)

// Disabler provides functionality for tags that disable other tags.
// Tags that embed this struct will disable specified tags during rendering.
type Disabler struct{}

// Render calls renderFn with disabledTags disabled in context. A tag that
// embeds Disabler calls it from its own RenderToOutput: the name is not
// RenderToOutput, so that it never hides the tag's Node method.
func (d *Disabler) Render(
	disabledTags []string,
	context liquid.TagContext,
	renderFn func(),
) {
	context.WithDisabledTags(disabledTags, renderFn)
}
