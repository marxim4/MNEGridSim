package sim

import (
	"errors"
	"fmt"
	"math"
)

// Producer describes a power source and its available power in kW.
type Producer struct {
	ID          string
	AvailableKW int64
}

// TotalAvailableKW sums the power available from all producers.
// It returns zero and an error if any ID is empty or duplicated, any power is
// negative, or the sum exceeds math.MaxInt64.
func TotalAvailableKW(producers []Producer) (int64, error) {
	seen := make(map[string]struct{}, len(producers))
	var totalKW int64
	for _, producer := range producers {
		if producer.ID == "" {
			return 0, errors.New("producer ID must not be empty")
		}
		if producer.AvailableKW < 0 {
			return 0, fmt.Errorf("available power for producer %q must be non-negative: %d kW", producer.ID, producer.AvailableKW)
		}
		if _, exists := seen[producer.ID]; exists {
			return 0, fmt.Errorf("duplicate producer ID %q", producer.ID)
		}
		if producer.AvailableKW > math.MaxInt64-totalKW {
			return 0, fmt.Errorf("total available power overflows int64 when adding producer %q", producer.ID)
		}

		seen[producer.ID] = struct{}{}
		totalKW += producer.AvailableKW
	}
	return totalKW, nil
}
