package sim

import "fmt"

// RunStep sums producer power and allocates it to the consumer requests.
// It leaves both inputs unchanged and returns a zero Result if either stage fails.
func RunStep(producers []Producer, requests []Request) (Result, error) {
	availableKW, err := TotalAvailableKW(producers)
	if err != nil {
		return Result{}, fmt.Errorf("aggregate production: %w", err)
	}

	result, err := Allocate(availableKW, requests)
	if err != nil {
		return Result{}, fmt.Errorf("allocate power: %w", err)
	}

	return result, nil
}
