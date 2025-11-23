package integration

import (
	"fmt"
	"testing"

	"github.com/Notifuse/liquidgo/liquid"
)

func TestDebugInvokable(t *testing.T) {
	drop := &ThingWithToLiquid{}
	methods := liquid.GetInvokableMethods(drop)
	fmt.Printf("Methods for ThingWithToLiquid: %v\n", methods)

	isInvokable := liquid.IsInvokable(drop, "to_liquid")
	fmt.Printf("IsInvokable(drop, 'to_liquid'): %v\n", isInvokable)
}
