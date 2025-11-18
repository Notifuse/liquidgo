package liquid

// ResourceLimits tracks and enforces resource limits during template rendering.
type ResourceLimits struct {
	renderLengthLimit *int
	renderScoreLimit  *int
	assignScoreLimit  *int
	renderScore       int
	assignScore       int
	lastCaptureLength *int
	reachedLimit      bool
}

// ResourceLimitsConfig configures resource limits.
type ResourceLimitsConfig struct {
	RenderLengthLimit *int
	RenderScoreLimit  *int
	AssignScoreLimit  *int
}

// NewResourceLimits creates a new ResourceLimits instance.
func NewResourceLimits(config ResourceLimitsConfig) *ResourceLimits {
	rl := &ResourceLimits{
		renderLengthLimit: config.RenderLengthLimit,
		renderScoreLimit:  config.RenderScoreLimit,
		assignScoreLimit:  config.AssignScoreLimit,
	}
	rl.Reset()
	return rl
}

// IncrementRenderScore increments the render score.
func (rl *ResourceLimits) IncrementRenderScore(amount int) {
	rl.renderScore += amount
	if rl.renderScoreLimit != nil && rl.renderScore > *rl.renderScoreLimit {
		rl.raiseLimitsReached()
	}
}

// IncrementAssignScore increments the assign score.
func (rl *ResourceLimits) IncrementAssignScore(amount int) {
	rl.assignScore += amount
	if rl.assignScoreLimit != nil && rl.assignScore > *rl.assignScoreLimit {
		rl.raiseLimitsReached()
	}
}

// IncrementWriteScore updates either render_length or assign_score based on whether writes are captured.
func (rl *ResourceLimits) IncrementWriteScore(output string) {
	if rl.lastCaptureLength != nil {
		captured := len([]byte(output))
		increment := captured - *rl.lastCaptureLength
		rl.lastCaptureLength = &captured
		rl.IncrementAssignScore(increment)
	} else if rl.renderLengthLimit != nil && len([]byte(output)) > *rl.renderLengthLimit {
		rl.raiseLimitsReached()
	}
}

func (rl *ResourceLimits) raiseLimitsReached() {
	rl.reachedLimit = true
	panic(NewMemoryError("Memory limits exceeded"))
}

// Reached returns true if limits have been reached.
func (rl *ResourceLimits) Reached() bool {
	return rl.reachedLimit
}

// Reset resets all scores and flags.
func (rl *ResourceLimits) Reset() {
	rl.reachedLimit = false
	rl.lastCaptureLength = nil
	rl.renderScore = 0
	rl.assignScore = 0
}

// WithCapture executes a function with capture tracking.
func (rl *ResourceLimits) WithCapture(fn func()) {
	oldCaptureLength := rl.lastCaptureLength
	defer func() {
		rl.lastCaptureLength = oldCaptureLength
	}()

	zero := 0
	rl.lastCaptureLength = &zero
	fn()
}

// RenderScore returns the current render score.
func (rl *ResourceLimits) RenderScore() int {
	return rl.renderScore
}

// AssignScore returns the current assign score.
func (rl *ResourceLimits) AssignScore() int {
	return rl.assignScore
}

// RenderLengthLimit returns the render length limit.
func (rl *ResourceLimits) RenderLengthLimit() *int {
	return rl.renderLengthLimit
}

// AssignScoreLimit returns the assign score limit.
func (rl *ResourceLimits) AssignScoreLimit() *int {
	return rl.assignScoreLimit
}
