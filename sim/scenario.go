package sim

import "fmt"

// Scenario defines the consumers and ordered ticks of a simulation.
type Scenario struct {
	Consumers []Consumer
	Ticks     []ScenarioTick
}

// ScenarioTick supplies producers and consumer demand readings for one tick.
type ScenarioTick struct {
	Producers []Producer
	Demands   []ConsumerDemand
}

// RunScenario resolves consumer demands and runs the resulting simulation ticks.
// It requires one demand per consumer per tick and returns nil results on error.
func RunScenario(scenario Scenario) ([]TickResult, error) {
	priorities := make(map[string]int, len(scenario.Consumers))
	for i, consumer := range scenario.Consumers {
		if consumer.ID == "" {
			return nil, fmt.Errorf("consumer %d: ID must not be empty", i)
		}
		if _, exists := priorities[consumer.ID]; exists {
			return nil, fmt.Errorf("duplicate consumer ID %q", consumer.ID)
		}
		priorities[consumer.ID] = consumer.Priority
	}

	inputs := make([]TickInput, len(scenario.Ticks))
	for i, tick := range scenario.Ticks {
		requests := make([]Request, len(tick.Demands))
		seen := make(map[string]struct{}, len(tick.Demands))
		for j, demand := range tick.Demands {
			priority, exists := priorities[demand.ConsumerID]
			if !exists {
				return nil, fmt.Errorf("tick %d: unknown consumer ID %q", i, demand.ConsumerID)
			}
			if _, duplicate := seen[demand.ConsumerID]; duplicate {
				return nil, fmt.Errorf("tick %d: duplicate demand for consumer %q", i, demand.ConsumerID)
			}
			if demand.RequestedKW < 0 {
				return nil, fmt.Errorf("tick %d: requested power for consumer %q must be non-negative: %d kW", i, demand.ConsumerID, demand.RequestedKW)
			}
			seen[demand.ConsumerID] = struct{}{}
			requests[j] = Request{
				ID:          demand.ConsumerID,
				Priority:    priority,
				RequestedKW: demand.RequestedKW,
			}
		}
		for _, consumer := range scenario.Consumers {
			if _, exists := seen[consumer.ID]; !exists {
				return nil, fmt.Errorf("tick %d: missing demand for consumer %q", i, consumer.ID)
			}
		}
		inputs[i] = TickInput{Producers: tick.Producers, Requests: requests}
	}

	results, err := RunTicks(inputs)
	if err != nil {
		return nil, fmt.Errorf("run scenario: %w", err)
	}
	return results, nil
}
