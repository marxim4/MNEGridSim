package sim

import (
	"errors"
	"reflect"
	"slices"
	"testing"
)

func TestRunScenario(t *testing.T) {
	demands := []ConsumerDemand{
		{ConsumerID: "industry", RequestedKW: 700},
		{ConsumerID: "hospital", RequestedKW: 600},
	}
	reference := Scenario{
		Consumers: []Consumer{{ID: "hospital", Priority: 1}, {ID: "industry", Priority: 2}},
		Ticks: []ScenarioTick{
			{Producers: []Producer{{ID: "hydro", AvailableKW: 1000}}, Demands: demands},
			{Producers: []Producer{{ID: "hydro", AvailableKW: 300}}, Demands: demands},
			{Producers: []Producer{{ID: "hydro", AvailableKW: 1200}}, Demands: demands},
		},
	}
	wantReference := []TickResult{
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
	permuted := Scenario{
		Consumers: slices.Clone(reference.Consumers),
		Ticks:     slices.Clone(reference.Ticks),
	}
	slices.Reverse(permuted.Consumers)
	reversedDemands := slices.Clone(demands)
	slices.Reverse(reversedDemands)
	for i := range permuted.Ticks {
		permuted.Ticks[i].Demands = reversedDemands
	}

	tests := []struct {
		name     string
		scenario Scenario
		want     []TickResult
	}{
		{name: "reference scenario", scenario: reference, want: wantReference},
		{name: "replay reference scenario", scenario: reference, want: wantReference},
		{name: "reverse consumers and demands", scenario: permuted, want: wantReference},
		{
			name: "demand changes while consumer priority stays fixed",
			scenario: Scenario{
				Consumers: []Consumer{{ID: "backup", Priority: 1}, {ID: "critical", Priority: 0}},
				Ticks: []ScenarioTick{
					{
						Producers: []Producer{{ID: "hydro", AvailableKW: 100}},
						Demands: []ConsumerDemand{
							{ConsumerID: "backup", RequestedKW: 100},
							{ConsumerID: "critical", RequestedKW: 100},
						},
					},
					{
						Producers: []Producer{{ID: "hydro", AvailableKW: 100}},
						Demands: []ConsumerDemand{
							{ConsumerID: "backup", RequestedKW: 100},
							{ConsumerID: "critical", RequestedKW: 0},
						},
					},
				},
			},
			want: []TickResult{
				{Tick: 0, Result: Result{Allocations: []Allocation{
					{ID: "critical", AllocatedKW: 100},
					{ID: "backup", UnmetKW: 100},
				}}},
				{Tick: 1, Result: Result{Allocations: []Allocation{
					{ID: "critical"},
					{ID: "backup", AllocatedKW: 100},
				}}},
			},
		},
		{name: "empty scenario", want: []TickResult{}},
		{
			name: "production without consumers",
			scenario: Scenario{Ticks: []ScenarioTick{
				{Producers: []Producer{{ID: "hydro", AvailableKW: 100}}, Demands: []ConsumerDemand{}},
			}},
			want: []TickResult{{Tick: 0, Result: Result{Allocations: []Allocation{}, SurplusKW: 100}}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := Scenario{
				Consumers: slices.Clone(tt.scenario.Consumers),
				Ticks:     slices.Clone(tt.scenario.Ticks),
			}
			for i, tick := range tt.scenario.Ticks {
				original.Ticks[i].Producers = slices.Clone(tick.Producers)
				original.Ticks[i].Demands = slices.Clone(tick.Demands)
			}

			got, err := RunScenario(tt.scenario)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RunScenario() = %+v, want %+v", got, tt.want)
			}
			if !reflect.DeepEqual(tt.scenario, original) {
				t.Error("RunScenario() changed its input scenario")
			}
		})
	}
}

func TestRunScenarioValidation(t *testing.T) {
	hospital := Consumer{ID: "hospital", Priority: 1}
	demand := ConsumerDemand{ConsumerID: "hospital", RequestedKW: 600}
	tests := []struct {
		name      string
		consumers []Consumer
		ticks     []ScenarioTick
		wantError string
	}{
		{
			name:      "empty consumer ID",
			consumers: []Consumer{hospital, {}},
			wantError: "consumer 1: ID must not be empty",
		},
		{
			name:      "duplicate consumer definition",
			consumers: []Consumer{hospital, hospital},
			wantError: `duplicate consumer ID "hospital"`,
		},
		{
			name:      "unknown consumer reference",
			consumers: []Consumer{hospital},
			ticks:     []ScenarioTick{{Demands: []ConsumerDemand{{ConsumerID: "unknown"}}}},
			wantError: `tick 0: unknown consumer ID "unknown"`,
		},
		{
			name:      "duplicate demand",
			consumers: []Consumer{hospital},
			ticks:     []ScenarioTick{{Demands: []ConsumerDemand{demand, demand}}},
			wantError: `tick 0: duplicate demand for consumer "hospital"`,
		},
		{
			name:      "missing demand on a later tick",
			consumers: []Consumer{hospital},
			ticks:     []ScenarioTick{{Demands: []ConsumerDemand{demand}}, {}},
			wantError: `tick 1: missing demand for consumer "hospital"`,
		},
		{
			name:      "negative demand",
			consumers: []Consumer{hospital},
			ticks:     []ScenarioTick{{Demands: []ConsumerDemand{{ConsumerID: "hospital", RequestedKW: -1}}}},
			wantError: `tick 0: requested power for consumer "hospital" must be non-negative: -1 kW`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RunScenario(Scenario{Consumers: tt.consumers, Ticks: tt.ticks})
			if got != nil || err == nil || err.Error() != tt.wantError {
				t.Errorf("RunScenario() = %+v, %v, want nil results and %q", got, err, tt.wantError)
			}
		})
	}
}

func TestRunScenarioWrapsTickError(t *testing.T) {
	scenario := Scenario{Ticks: []ScenarioTick{
		{Producers: []Producer{{ID: "hydro", AvailableKW: 100}}},
		{Producers: []Producer{{ID: "hydro", AvailableKW: -1}}},
	}}

	got, err := RunScenario(scenario)
	if got != nil {
		t.Errorf("RunScenario() = %+v, want nil results on error", got)
	}
	const cause = `available power for producer "hydro" must be non-negative: -1 kW`
	const stage = "aggregate production: " + cause
	const tick = "tick 1: " + stage
	for _, want := range []string{"run scenario: " + tick, tick, stage, cause} {
		if err == nil || err.Error() != want {
			t.Fatalf("error in RunScenario() chain = %v, want %q", err, want)
		}
		err = errors.Unwrap(err)
	}
}
