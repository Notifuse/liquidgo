package liquid

// Usage provides usage tracking functionality.
// Currently a placeholder for future implementation.
type Usage struct{}

// Increment increments usage count for a given name.
// Currently a no-op, to be implemented when usage tracking is needed.
func (u *Usage) Increment(name string) {
	// TODO: Implement usage tracking
}

// GlobalUsage is the global usage tracker.
var GlobalUsage = &Usage{}

// IncrementUsage increments usage for the global tracker.
func IncrementUsage(name string) {
	GlobalUsage.Increment(name)
}
