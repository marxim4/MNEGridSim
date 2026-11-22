package sim

import (
	"errors"
	"math"
	"reflect"
	"slices"
	"testing"
)

func TestRunStep(t *testing.T) {
	tests := []struct {
		name      string
		producers []Producer
		requests  []Request
		want      Result
	}{
		{
			name: "multiple producers with insufficient power",
			producers: []Producer{
				{ID: "hydro", AvailableKW: 600},
				{ID: "solar", AvailableKW: 300},
				{ID: "wind", AvailableKW: 100},
			},
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
			name:      "surplus after serving all demand",
			producers: []Producer{{ID: "hydro", AvailableKW: 900}, {ID: "solar", AvailableKW: 100}},
			requests:  []Request{{ID: "hospital", RequestedKW: 600}},
			want: Result{
				Allocations: []Allocation{{ID: "hospital", AllocatedKW: 600}},
				SurplusKW:   400,
			},
		},
		{
			name:     "demand without production",
			requests: []Request{{ID: "hospital", RequestedKW: 600}},
			want:     Result{Allocations: []Allocation{{ID: "hospital", UnmetKW: 600}}},
		},
		{
			name:      "empty inputs",
			producers: []Producer{},
			requests:  []Request{},
			want:      Result{Allocations: []Allocation{}},
		},
		{
			name: "nil inputs",
			want: Result{Allocations: []Allocation{}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RunStep(tt.producers, tt.requests)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("RunStep() = %+v, want %+v", got, tt.want)
			}

			var productionKW int64
			for _, producer := range tt.producers {
				productionKW += producer.AvailableKW
			}
			accountedKW := got.SurplusKW
			for _, allocation := range got.Allocations {
				accountedKW += allocation.AllocatedKW
			}
			if accountedKW != productionKW {
				t.Errorf("allocated power plus surplus = %d kW, total production = %d kW", accountedKW, productionKW)
			}
		})
	}
}

func TestRunStepErrors(t *testing.T) {
	tests := []struct {
		name        string
		producers   []Producer
		requests    []Request
		wantContext string
		wantCause   string
	}{
		{
			name:        "invalid producer",
			producers:   []Producer{{ID: "hydro", AvailableKW: -1}},
			requests:    []Request{{ID: "hospital", RequestedKW: 600}},
			wantContext: "aggregate production",
			wantCause:   `available power for producer "hydro" must be non-negative: -1 kW`,
		},
		{
			name:        "invalid request after successful aggregation",
			producers:   []Producer{{ID: "hydro", AvailableKW: 600}},
			requests:    []Request{{ID: "hospital", RequestedKW: 400}, {RequestedKW: 100}},
			wantContext: "allocate power",
			wantCause:   "consumer ID must not be empty",
		},
		{
			name:        "production overflow",
			producers:   []Producer{{ID: "hydro", AvailableKW: math.MaxInt64}, {ID: "solar", AvailableKW: 1}},
			requests:    []Request{{ID: "hospital", RequestedKW: 600}},
			wantContext: "aggregate production",
			wantCause:   `total available power overflows int64 when adding producer "solar"`,
		},
		{
			name:        "production error takes precedence when both inputs are invalid",
			producers:   []Producer{{AvailableKW: 600}},
			requests:    []Request{{RequestedKW: 600}},
			wantContext: "aggregate production",
			wantCause:   "producer ID must not be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RunStep(tt.producers, tt.requests)
			if !reflect.DeepEqual(got, Result{}) {
				t.Errorf("RunStep() = %+v, want zero Result", got)
			}
			if err == nil {
				t.Fatal("RunStep() succeeded, want an error")
			}
			if want := tt.wantContext + ": " + tt.wantCause; err.Error() != want {
				t.Errorf("RunStep() error = %q, want %q", err, want)
			}
			if cause := errors.Unwrap(err); cause == nil || cause.Error() != tt.wantCause {
				t.Errorf("RunStep() wrapped cause = %v, want %q", cause, tt.wantCause)
			}
		})
	}
}

func TestRunStepDeterminism(t *testing.T) {
	hydro := Producer{ID: "hydro", AvailableKW: 5}
	solar := Producer{ID: "solar", AvailableKW: 2}
	alpha := Request{ID: "alpha", Priority: 1, RequestedKW: 3}
	beta := Request{ID: "beta", Priority: 1, RequestedKW: 5}

	producers := []Producer{hydro, solar}
	requests := []Request{beta, alpha}
	reversedProducers := []Producer{solar, hydro}
	reversedRequests := []Request{alpha, beta}
	tests := []struct {
		name      string
		producers []Producer
		requests  []Request
	}{
		{"original order", producers, requests},
		{"repeat original order", producers, requests},
		{"reverse producers", reversedProducers, requests},
		{"reverse requests", producers, reversedRequests},
		{"reverse both inputs", reversedProducers, reversedRequests},
	}
	want := Result{Allocations: []Allocation{
		{ID: "alpha", AllocatedKW: 3},
		{ID: "beta", AllocatedKW: 4, UnmetKW: 1},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalProducers := slices.Clone(tt.producers)
			originalRequests := slices.Clone(tt.requests)
			got, err := RunStep(tt.producers, tt.requests)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("RunStep() = %+v, want %+v", got, want)
			}
			if !slices.Equal(tt.producers, originalProducers) {
				t.Errorf("RunStep() changed producers to %+v, want %+v", tt.producers, originalProducers)
			}
			if !slices.Equal(tt.requests, originalRequests) {
				t.Errorf("RunStep() changed requests to %+v, want %+v", tt.requests, originalRequests)
			}
		})
	}
}
