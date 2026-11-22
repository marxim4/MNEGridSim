package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/marxim4/MNEGridSim/sim"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	demands := []sim.ConsumerDemand{
		{ConsumerID: "hospital", RequestedKW: 600},
		{ConsumerID: "industry", RequestedKW: 700},
	}
	scenario := sim.Scenario{
		Consumers: []sim.Consumer{
			{ID: "hospital", Priority: 1},
			{ID: "industry", Priority: 2},
		},
		Ticks: []sim.ScenarioTick{
			{
				Producers: []sim.Producer{
					{ID: "hydro", AvailableKW: 600},
					{ID: "solar", AvailableKW: 300},
					{ID: "wind", AvailableKW: 100},
				},
				Demands: demands,
			},
			{Producers: []sim.Producer{{ID: "hydro", AvailableKW: 300}}, Demands: demands},
			{Producers: []sim.Producer{{ID: "hydro", AvailableKW: 1200}}, Demands: demands},
		},
	}

	results, err := sim.RunScenario(scenario)
	if err != nil {
		return err
	}

	var output strings.Builder
	fmt.Fprintln(&output, "MNEGridSim synthetic scenario")
	for _, tick := range results {
		fmt.Fprintf(&output, "\nTick %d\n", tick.Tick)
		for _, allocation := range tick.Result.Allocations {
			fmt.Fprintf(&output, "  %s: %d kW allocated, %d kW unmet\n",
				allocation.ID, allocation.AllocatedKW, allocation.UnmetKW)
		}
		fmt.Fprintf(&output, "  Surplus: %d kW\n", tick.Result.SurplusKW)
	}
	if _, err := fmt.Print(output.String()); err != nil {
		return fmt.Errorf("write results: %w", err)
	}
	return nil
}
