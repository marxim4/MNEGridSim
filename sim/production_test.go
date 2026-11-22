package sim

import (
	"math"
	"testing"
)

func TestTotalAvailableKW(t *testing.T) {
	tests := []struct {
		name      string
		producers []Producer
		wantKW    int64
		wantError string
	}{
		{
			name: "nil producers",
		},
		{
			name:      "empty producers",
			producers: []Producer{},
		},
		{
			name:      "zero available power",
			producers: []Producer{{ID: "solar"}},
		},
		{
			name: "sum multiple producers",
			producers: []Producer{
				{ID: "hydro", AvailableKW: 600},
				{ID: "solar", AvailableKW: 400},
				{ID: "wind"},
			},
			wantKW: 1000,
		},
		{
			name: "sum reaches max int64 and accepts zero afterward",
			producers: []Producer{
				{ID: "first", AvailableKW: math.MaxInt64 - 1},
				{ID: "second", AvailableKW: 1},
				{ID: "idle"},
			},
			wantKW: math.MaxInt64,
		},
		{
			name: "sum exceeds max int64 by one",
			producers: []Producer{
				{ID: "first", AvailableKW: math.MaxInt64 - 1},
				{ID: "second", AvailableKW: 2},
			},
			wantError: `total available power overflows int64 when adding producer "second"`,
		},
		{
			name: "negative power after valid producer",
			producers: []Producer{
				{ID: "hydro", AvailableKW: 600},
				{ID: "solar", AvailableKW: math.MinInt64},
			},
			wantError: `available power for producer "solar" must be non-negative: -9223372036854775808 kW`,
		},
		{
			name: "empty producer ID",
			producers: []Producer{
				{ID: "hydro", AvailableKW: 600},
				{AvailableKW: 400},
			},
			wantError: "producer ID must not be empty",
		},
		{
			name: "nonconsecutive duplicate producer ID",
			producers: []Producer{
				{ID: "hydro", AvailableKW: 600},
				{ID: "solar", AvailableKW: 400},
				{ID: "hydro"},
			},
			wantError: `duplicate producer ID "hydro"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotKW, err := TotalAvailableKW(tt.producers)
			if tt.wantError == "" {
				if err != nil {
					t.Errorf("TotalAvailableKW() unexpected error: %v", err)
				}
			} else if err == nil || err.Error() != tt.wantError {
				t.Errorf("TotalAvailableKW() error = %v, want %q", err, tt.wantError)
			}
			if gotKW != tt.wantKW {
				t.Errorf("TotalAvailableKW() = %d kW, want %d kW", gotKW, tt.wantKW)
			}
		})
	}
}
