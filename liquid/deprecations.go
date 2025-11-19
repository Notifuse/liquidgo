package liquid

import (
	"fmt"
	"runtime"
	"sync"
)

// Deprecations handles deprecation warnings.
type Deprecations struct {
	warned map[string]bool
	mu     sync.Mutex
}

var globalDeprecations = &Deprecations{
	warned: make(map[string]bool),
}

// Warn issues a deprecation warning if it hasn't been warned about before.
func (d *Deprecations) Warn(name, alternative string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.warned[name] {
		return
	}

	d.warned[name] = true

	// Get caller location (skip Warn and the function that called Warn)
	pc, file, line, ok := runtime.Caller(2)
	if ok {
		fn := runtime.FuncForPC(pc)
		callerLocation := fmt.Sprintf("%s:%d", file, line)
		if fn != nil {
			callerLocation = fmt.Sprintf("%s in %s:%d", fn.Name(), file, line)
		}
		fmt.Printf("[DEPRECATION] %s is deprecated. Use %s instead. Called from %s\n", name, alternative, callerLocation)
	} else {
		fmt.Printf("[DEPRECATION] %s is deprecated. Use %s instead.\n", name, alternative)
	}
}

// Warn issues a deprecation warning using the global deprecations instance.
func Warn(name, alternative string) {
	globalDeprecations.Warn(name, alternative)
}

// Reset clears all warned deprecations (useful for testing).
func (d *Deprecations) Reset() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.warned = make(map[string]bool)
}

// ResetDeprecations clears all warned deprecations in the global instance (useful for testing).
func ResetDeprecations() {
	globalDeprecations.Reset()
}
