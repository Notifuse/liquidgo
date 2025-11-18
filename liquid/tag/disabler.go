package tag

import (
	"github.com/Notifuse/liquidgo/liquid"
)

// Disabler provides functionality for tags that disable other tags.
// Tags that embed this struct will disable specified tags during rendering.
type Disabler struct{}

// RenderToOutputBuffer wraps tag rendering to disable specified tags during rendering.
// It calls context.WithDisabledTags with the provided disabled tags list.
func (d *Disabler) RenderToOutputBuffer(
	disabledTags []string,
	context liquid.TagContext,
	output *string,
	renderFn func(),
) {
	context.WithDisabledTags(disabledTags, func() {
		renderFn()
	})
}
