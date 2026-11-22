package sim

import (
	"cmp"
	"errors"
	"fmt"
	"slices"
)

// Request describes a consumer's power demand in kW.
type Request struct {
	ID          string
	Priority    int // Lower values have precedence.
	RequestedKW int64
}

// Allocation reports how much of a consumer's demand was served.
type Allocation struct {
	ID          string
	AllocatedKW int64
	UnmetKW     int64
}

// Result contains the allocations for one step and its unused power.
type Result struct {
	Allocations []Allocation
	SurplusKW   int64
}

// Allocate serves requests in ascending priority and lexicographic ID order.
// It leaves requests unchanged and returns a zero Result for invalid input.
func Allocate(availableKW int64, requests []Request) (Result, error) {
	if availableKW < 0 {
		return Result{}, fmt.Errorf("available power must be non-negative: %d kW", availableKW)
	}

	ordered := slices.Clone(requests)
	slices.SortFunc(ordered, func(a, b Request) int {
		if byPriority := cmp.Compare(a.Priority, b.Priority); byPriority != 0 {
			return byPriority
		}
		if byID := cmp.Compare(a.ID, b.ID); byID != 0 {
			return byID
		}
		// Equal IDs are invalid. Order their demands for consistent validation errors.
		return cmp.Compare(a.RequestedKW, b.RequestedKW)
	})

	seen := make(map[string]struct{}, len(ordered))
	for _, request := range ordered {
		if request.ID == "" {
			return Result{}, errors.New("consumer ID must not be empty")
		}
		if request.RequestedKW < 0 {
			return Result{}, fmt.Errorf("requested power for %q must be non-negative: %d kW", request.ID, request.RequestedKW)
		}
		if _, exists := seen[request.ID]; exists {
			return Result{}, fmt.Errorf("duplicate consumer ID %q", request.ID)
		}
		seen[request.ID] = struct{}{}
	}

	allocations := make([]Allocation, len(ordered))
	remainingKW := availableKW
	for i, request := range ordered {
		allocatedKW := min(remainingKW, request.RequestedKW)
		remainingKW -= allocatedKW
		allocations[i] = Allocation{
			ID:          request.ID,
			AllocatedKW: allocatedKW,
			UnmetKW:     request.RequestedKW - allocatedKW,
		}
	}

	return Result{Allocations: allocations, SurplusKW: remainingKW}, nil
}
