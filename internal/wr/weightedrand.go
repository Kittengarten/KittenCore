// Package wr contains a performant data structure and algorithm used
// to randomly select an element from some kind of list, where the chances of
// each element to be selected not being equal, but defined by relative
// "weights" (or probabilities). This is called weighted random selection.
//
// This package creates a presorted cache optimized for binary search, allowing
// for repeated selections from the same set to be significantly faster,
// especially for large data sets.
package wr

import (
	"cmp"
	"errors"
	"math"
	"math/rand/v2"
	"slices"

	"golang.org/x/exp/constraints"
)

// Choice is a generic wrapper that can be used to add weights for any item.
type Choice[T any, W constraints.Integer] struct {
	Item   T
	Weight W
}

// NewChoice creates a new Choice with specified item and weight.
func NewChoice[T any, W constraints.Integer](item T, weight W) Choice[T, W] {
	return Choice[T, W]{Item: item, Weight: weight}
}

// A Chooser caches many possible Choices in a structure designed to improve
// performance on repeated calls for weighted random selection.
type Chooser[T any, W constraints.Integer] struct {
	data   []Choice[T, W]
	totals []uint64
	max    uint64
}

// NewChooser initializes a new Chooser for picking from the provided choices.
func NewChooser[T any, W constraints.Integer](choices ...Choice[T, W]) (*Chooser[T, W], error) {
	slices.SortFunc(choices, func(a, b Choice[T, W]) int {
		return cmp.Compare(a.Weight, b.Weight)
	})

	var (
		totals       = make([]uint64, len(choices))
		runningTotal uint64
	)
	for i, c := range choices {
		if c.Weight < 0 {
			continue // ignore negative weights, can never be picked
		}

		weight := uint64(c.Weight) // convert weight to uint64 for internal counter usage
		if math.MaxUint64-runningTotal <= weight {
			return nil, errWeightOverflow
		}
		runningTotal += weight
		totals[i] = runningTotal
	}

	if runningTotal < 1 {
		return nil, errNoValidChoices
	}

	return &Chooser[T, W]{data: choices, totals: totals, max: runningTotal}, nil
}

// Possible errors returned by NewChooser, preventing the creation of a Chooser
// with unsafe runtime states.
var (
	// If the sum of provided Choice weights exceed the maximum integer value
	// for the current platform (e.g. math.MaxInt32 or math.MaxInt64), then
	// the internal running total will overflow, resulting in an imbalanced
	// distribution generating improper results.
	errWeightOverflow = errors.New("sum of Choice Weights exceeds max int")
	// If there are no Choices available to the Chooser with a weight >= 1,
	// there are no valid choices and Pick would produce a runtime panic.
	errNoValidChoices = errors.New("zero Choices with Weight >= 1")
)

// Pick returns a single weighted random Choice.Item from the Chooser.
//
// Utilizes global rand as the source of randomness. Safe for concurrent usage.
func (c Chooser[T, W]) Pick() T {
	//nolint:gosec
	r := rand.N(c.max) + 1
	i, _ := slices.BinarySearch(c.totals, r)
	return c.data[i].Item
}
