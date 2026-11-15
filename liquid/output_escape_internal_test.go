package liquid

import "testing"

// Reset clears the capture record, so a reused context does not carry one
// render's captures into the next.
func TestResetClearsTheCaptureRecord(t *testing.T) {
	env := NewEnvironment()
	env.SetOutputEscape(true)
	ctx := BuildContext(ContextConfig{Environment: env})
	ctx.MarkSafeHTML("<b>x</b>")
	if !ctx.isSafeHTML("<b>x</b>") {
		t.Fatal("MarkSafeHTML did not record the text")
	}
	ctx.Reset()
	if ctx.isSafeHTML("<b>x</b>") {
		t.Error("Reset kept the capture record")
	}
}
