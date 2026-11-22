package sim

import "fmt"

// TickInput contains the producers and requests for a single simulation tick.
type TickInput struct {
	Producers []Producer
	Requests  []Request
}

// TickResult pairs a simulation result with its zero-based tick index.
type TickResult struct {
	Tick   int
	Result Result
}

// RunTicks runs independent simulation steps in input order.
// Empty input returns a non-nil empty slice. A failed step returns nil results
// and an error that includes its tick index and wraps the step error.
func RunTicks(inputs []TickInput) ([]TickResult, error) {
	results := make([]TickResult, len(inputs))
	for i, input := range inputs {
		result, err := RunStep(input.Producers, input.Requests)
		if err != nil {
			return nil, fmt.Errorf("tick %d: %w", i, err)
		}
		results[i] = TickResult{Tick: i, Result: result}
	}
	return results, nil
}
