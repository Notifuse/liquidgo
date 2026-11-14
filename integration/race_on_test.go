//go:build race

package integration

// raceEnabled is true when the tests run with the race detector, which makes
// each step of a long computation about ten times slower.
const raceEnabled = true
