package sim

import (
	"errors"
	"reflect"
	"slices"
	"testing"
)

func TestRunTicks(t *testing.T) {
	requests := []Request{
		{ID: "industry", Priority: 2, RequestedKW: 700},
		{ID: "hospital", Priority: 1, RequestedKW: 600},
	}
	inputs := []TickInput{
		{Producers: []Producer{{ID: "hydro", AvailableKW: 1000}}, Requests: requests},
		{Producers: []Producer{{ID: "hydro", AvailableKW: 300}}, Requests: requests},
		{Producers: []Producer{{ID: "hydro", AvailableKW: 1200}}, Requests: requests},
	}
	want := []TickResult{
		{Tick: 0, Result: Result{Allocations: []Allocation{
			{ID: "hospital", AllocatedKW: 600},
			{ID: "industry", AllocatedKW: 400, UnmetKW: 300},
		}}},
		{Tick: 1, Result: Result{Allocations: []Allocation{
			{ID: "hospital", AllocatedKW: 300, UnmetKW: 300},
			{ID: "industry", UnmetKW: 700},
		}}},
		{Tick: 2, Result: Result{Allocations: []Allocation{
			{ID: "hospital", AllocatedKW: 600},
			{ID: "industry", AllocatedKW: 600, UnmetKW: 100},
		}}},
	}

	original := make([]TickInput, len(inputs))
	for i, input := range inputs {
		original[i] = TickInput{
			Producers: slices.Clone(input.Producers),
			Requests:  slices.Clone(input.Requests),
		}
	}

	for run := 1; run <= 2; run++ {
		got, err := RunTicks(inputs)
		if err != nil {
			t.Fatalf("run %d: %v", run, err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("run %d: RunTicks() = %+v, want %+v", run, got, want)
		}
		if !reflect.DeepEqual(inputs, original) {
			t.Fatalf("run %d: RunTicks() changed inputs to %+v, want %+v", run, inputs, original)
		}
	}
}

func TestRunTicksEmpty(t *testing.T) {
	tests := []struct {
		name   string
		inputs []TickInput
	}{
		{name: "nil input"},
		{name: "empty input", inputs: []TickInput{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RunTicks(tt.inputs)
			if err != nil || got == nil || len(got) != 0 {
				t.Errorf("RunTicks() = %#v, %v, want a non-nil empty slice and nil error", got, err)
			}
		})
	}
}

func TestRunTicksError(t *testing.T) {
	inputs := []TickInput{
		{
			Producers: []Producer{{ID: "hydro", AvailableKW: 600}},
			Requests:  []Request{{ID: "hospital", RequestedKW: 600}},
		},
		{Requests: []Request{{RequestedKW: 600}}},
	}
	got, err := RunTicks(inputs)
	if got != nil {
		t.Errorf("RunTicks() = %+v, want nil results on failure", got)
	}
	const causeMessage = "consumer ID must not be empty"
	const stepMessage = "allocate power: " + causeMessage
	if err == nil || err.Error() != "tick 1: "+stepMessage {
		t.Fatalf("RunTicks() error = %v, want %q", err, "tick 1: "+stepMessage)
	}
	stepErr := errors.Unwrap(err)
	if stepErr == nil || stepErr.Error() != stepMessage {
		t.Fatalf("RunTicks() wrapped step error = %v, want %q", stepErr, stepMessage)
	}
	if cause := errors.Unwrap(stepErr); cause == nil || cause.Error() != causeMessage {
		t.Errorf("RunTicks() wrapped root cause = %v, want %q", cause, causeMessage)
	}
}
