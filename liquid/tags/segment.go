package tags

import (
	"math"

	"github.com/Notifuse/liquidgo/v5/liquid"
)

// segment is the part of a collection that a for or tablerow loop iterates:
// the items an offset and a limit keep, in order or reversed. The integers
// of a range are computed as the loop reaches them, never built.
type segment struct {
	items    []interface{}
	first    int // the first integer of a range segment
	length   int
	isRange  bool
	reversed bool
}

// Len returns the number of items.
func (s segment) Len() int { return s.length }

// At returns the item at index i of the loop.
func (s segment) At(i int) interface{} {
	if s.reversed {
		i = s.length - 1 - i
	}
	if s.isRange {
		return s.first + i
	}
	return s.items[i]
}

// addSaturated returns a + b, or the int bound it passes.
func addSaturated(a, b int) int {
	if b > 0 && a > math.MaxInt-b {
		return math.MaxInt
	}
	if b < 0 && a < math.MinInt-b {
		return math.MinInt
	}
	return a + b
}

// sliceSegment returns the items of collection whose index is at least from
// and below to, when to is set, as Ruby's Utils.slice_collection.
func sliceSegment(collection interface{}, from int, to *int) segment {
	r, ok := collection.(*liquid.Range)
	if !ok {
		items := liquid.SliceCollection(collection, from, to)
		return segment{items: items, length: len(items)}
	}
	lo, hi := max(from, 0), r.Len()
	if to != nil {
		hi = min(hi, *to)
	}
	if hi <= lo {
		return segment{isRange: true}
	}
	return segment{isRange: true, first: r.Start + lo, length: hi - lo}
}

// chargeItem counts an item of a range the loop visits in the render score,
// as Shopify Liquid 5.14's RangeSlice, so that a range costs what it is
// iterated over, not what it holds. A body without nodes is already counted 1
// at each iteration (BlockBody), which stands for the item: the item is
// counted apart only with a body, so that the score is Ruby's in both cases.
func (s segment) chargeItem(ctx *liquid.Context, body []interface{}) {
	if !s.isRange || len(body) == 0 {
		return
	}
	if rl := ctx.ResourceLimits(); rl != nil {
		rl.IncrementRenderScore(1)
	}
}
