// Package filters is a test filter type whose name another package shares.
package filters

// Filters has a filter only this package defines.
type Filters struct{}

// AOnly is a filter.
func (f *Filters) AOnly(input interface{}) string { return "a" }
