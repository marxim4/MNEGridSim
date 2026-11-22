package sim

import (
	"math"
	"math/big"
	"reflect"
	"slices"
	"testing"
)

func TestAllocate(t *testing.T) {
	tests := []struct {
		name        string
		availableKW int64
		requests    []Request
		want        Result
	}{
		{
			name:        "hospital and industry",
			availableKW: 1000,
			requests: []Request{
				{ID: "industry", Priority: 2, RequestedKW: 700},
				{ID: "hospital", Priority: 1, RequestedKW: 600},
			},
			want: Result{Allocations: []Allocation{
				{ID: "hospital", AllocatedKW: 600},
				{ID: "industry", AllocatedKW: 400, UnmetKW: 300},
			}},
		},
		{
			name:        "ID breaks ties and partial service exhausts power",
			availableKW: 7,
			requests: []Request{
				{ID: "factory", Priority: 10, RequestedKW: 6},
				{ID: "beta", Priority: 5, RequestedKW: 6},
				{ID: "zeta", Priority: 5, RequestedKW: 1},
				{ID: "alpha", Priority: 5, RequestedKW: 4},
			},
			want: Result{Allocations: []Allocation{
				{ID: "alpha", AllocatedKW: 4},
				{ID: "beta", AllocatedKW: 3, UnmetKW: 3},
				{ID: "zeta", UnmetKW: 1},
				{ID: "factory", UnmetKW: 6},
			}},
		},
		{
			name: "no available power",
			requests: []Request{
				{ID: "load", Priority: 2, RequestedKW: 5},
				{ID: "idle", Priority: 1},
			},
			want: Result{Allocations: []Allocation{
				{ID: "idle"},
				{ID: "load", UnmetKW: 5},
			}},
		},
		{
			name:        "zero demand and surplus",
			availableKW: 10,
			requests: []Request{
				{ID: "idle", Priority: 1},
				{ID: "load", Priority: 2, RequestedKW: 4},
			},
			want: Result{
				Allocations: []Allocation{{ID: "idle"}, {ID: "load", AllocatedKW: 4}},
				SurplusKW:   6,
			},
		},
		{
			name:        "nil requests preserve maximum available power",
			availableKW: math.MaxInt64,
			want:        Result{Allocations: []Allocation{}, SurplusKW: math.MaxInt64},
		},
		{
			name:        "empty requests",
			availableKW: 20,
			requests:    []Request{},
			want:        Result{Allocations: []Allocation{}, SurplusKW: 20},
		},
		{
			name:        "exact balance at maximum available power",
			availableKW: math.MaxInt64,
			requests: []Request{
				{ID: "z", Priority: math.MinInt, RequestedKW: math.MaxInt64 - 1},
				{ID: "a", Priority: math.MaxInt, RequestedKW: 1},
			},
			want: Result{Allocations: []Allocation{
				{ID: "z", AllocatedKW: math.MaxInt64 - 1},
				{ID: "a", AllocatedKW: 1},
			}},
		},
		{
			name:        "total demand exceeds int64",
			availableKW: math.MaxInt64,
			requests: []Request{
				{ID: "first", Priority: 1, RequestedKW: math.MaxInt64 - 1},
				{ID: "next", Priority: 2, RequestedKW: math.MaxInt64},
			},
			want: Result{Allocations: []Allocation{
				{ID: "first", AllocatedKW: math.MaxInt64 - 1},
				{ID: "next", AllocatedKW: 1, UnmetKW: math.MaxInt64 - 1},
			}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Allocate(tt.availableKW, tt.requests)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("Allocate() = %+v, want %+v", got, tt.want)
			}
			assertAllocationInvariants(t, tt.availableKW, tt.requests, got)
		})
	}
}

func TestAllocateValidation(t *testing.T) {
	tests := []struct {
		name        string
		availableKW int64
		requests    []Request
		wantError   string
	}{
		{
			name:        "negative available power",
			availableKW: math.MinInt64,
			wantError:   "available power must be non-negative: -9223372036854775808 kW",
		},
		{
			name: "negative demand after a valid request",
			requests: []Request{
				{ID: "bad", Priority: 2, RequestedKW: math.MinInt64},
				{ID: "valid", Priority: 1, RequestedKW: 1},
			},
			wantError: `requested power for "bad" must be non-negative: -9223372036854775808 kW`,
		},
		{
			name:      "empty ID",
			requests:  []Request{{RequestedKW: 1}},
			wantError: "consumer ID must not be empty",
		},
		{
			name: "duplicate ID across priorities",
			requests: []Request{
				{ID: "same", Priority: 3},
				{ID: "other", Priority: 2},
				{ID: "same", Priority: 1},
			},
			wantError: `duplicate consumer ID "same"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := slices.Clone(tt.requests)
			got, err := Allocate(tt.availableKW, tt.requests)
			if err == nil || err.Error() != tt.wantError {
				t.Errorf("Allocate() error = %v, want %q", err, tt.wantError)
			}
			if !reflect.DeepEqual(got, Result{}) {
				t.Errorf("Allocate() returned %+v for invalid input", got)
			}
			if !slices.Equal(tt.requests, original) {
				t.Errorf("Allocate() changed the input to %+v, want %+v", tt.requests, original)
			}
		})
	}
}

func TestAllocatePermutations(t *testing.T) {
	alpha := Request{ID: "alpha", Priority: 1, RequestedKW: 3}
	beta := Request{ID: "beta", Priority: 1, RequestedKW: 4}
	factory := Request{ID: "factory", Priority: 2, RequestedKW: 5}
	orders := [][]Request{
		{alpha, beta, factory},
		{alpha, factory, beta},
		{beta, alpha, factory},
		{beta, factory, alpha},
		{factory, alpha, beta},
		{factory, beta, alpha},
	}
	want := Result{Allocations: []Allocation{
		{ID: "alpha", AllocatedKW: 3},
		{ID: "beta", AllocatedKW: 2, UnmetKW: 2},
		{ID: "factory", UnmetKW: 5},
	}}

	for _, order := range orders {
		t.Run(order[0].ID+"-"+order[1].ID+"-"+order[2].ID, func(t *testing.T) {
			backing := append(slices.Clone(order), Request{ID: "outside the input slice"})
			original := slices.Clone(backing)
			got, err := Allocate(5, backing[:len(order)])
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("Allocate() = %+v, want %+v", got, want)
			}
			if !slices.Equal(backing, original) {
				t.Errorf("Allocate() changed the backing array to %+v, want %+v", backing, original)
			}
		})
	}
}

func TestAllocateValidationOrder(t *testing.T) {
	first := Request{ID: "duplicate", RequestedKW: -1}
	second := Request{ID: "duplicate", RequestedKW: -2}
	for _, requests := range [][]Request{{first, second}, {second, first}} {
		_, err := Allocate(0, requests)
		const want = `requested power for "duplicate" must be non-negative: -2 kW`
		if err == nil || err.Error() != want {
			t.Errorf("Allocate(0, %+v) error = %v, want %q", requests, err, want)
		}
	}
}

func assertAllocationInvariants(t *testing.T, availableKW int64, requests []Request, result Result) {
	t.Helper()

	// Exact totals keep the test valid even when the sum exceeds int64.
	requestedByID := make(map[string]int64, len(requests))
	var totalRequested, totalAllocated big.Int
	for _, request := range requests {
		requestedByID[request.ID] = request.RequestedKW
		totalRequested.Add(&totalRequested, big.NewInt(request.RequestedKW))
	}

	exhausted := false
	for _, allocation := range result.Allocations {
		requested := requestedByID[allocation.ID]
		allocated := allocation.AllocatedKW
		if allocated < 0 || allocated > requested {
			t.Fatalf("%q allocated %d kW outside [0, %d]", allocation.ID, allocated, requested)
		}
		if allocation.UnmetKW != requested-allocated {
			t.Errorf("%q: allocated %d and unmet %d do not balance requested %d", allocation.ID, allocated, allocation.UnmetKW, requested)
		}
		if exhausted && allocated != 0 {
			t.Errorf("%q received %d kW after a previous request went unmet", allocation.ID, allocated)
		}
		exhausted = exhausted || allocation.UnmetKW > 0
		totalAllocated.Add(&totalAllocated, big.NewInt(allocated))
	}

	available := big.NewInt(availableKW)
	wantAllocated := available
	if totalRequested.Cmp(available) < 0 {
		wantAllocated = &totalRequested
	}
	if totalAllocated.Cmp(wantAllocated) != 0 {
		t.Errorf("total allocated = %s kW, want %s kW", &totalAllocated, wantAllocated)
	}

	var accounted big.Int
	accounted.Add(&totalAllocated, big.NewInt(result.SurplusKW))
	if accounted.Cmp(available) != 0 {
		t.Errorf("allocated plus surplus = %s kW, want %s kW", &accounted, available)
	}
}
