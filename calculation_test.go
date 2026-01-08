package main

import (
	"math"
	"testing"
)

func TestCalculateFlowAccuracy(t *testing.T) {
	// Formula: F + F * ((P - RefP) / 255) * ((T - RefT) / 255)
	
	equation := "F + F * ((P - RefP) / 255) * ((T - RefT) / 255)"
	
	refP := int32(100)
	refT := int32(125)

	tests := []struct {
		name          string
		flow          int32
		pressure      int32
		temperature   int32
		expected      int32
		tolerance     int32 
	}{
		{
			name:        "Reference Point (No Deviation)",
			flow:        1000,
			pressure:    100,
			temperature: 125,
			expected:    1000, // 1000 + 1000 * 0 * 0
			tolerance:   0,
		},
		{
			name:        "Max Values",
			flow:        1000,
			pressure:    255,
			temperature: 255,
			// 1000 + 1000 * (155/255) * (130/255)
			// 1000 + 1000 * 0.6078 * 0.5098
			// 1000 + 1000 * 0.3098 ~= 1309.8
			expected:    1310,
			tolerance:   1,
		},
		{
			name:        "Zero Values",
			flow:        1000,
			pressure:    0,
			temperature: 0,
			// 1000 + 1000 * (-100/255) * (-125/255)
			// 1000 + 1000 * (-0.3921) * (-0.4901)
			// 1000 + 1000 * 0.1922 ~= 1192.2
			expected:    1192,
			tolerance:   1,
		},
		{
			name:        "Negative Deviation (Pressure)",
			flow:        1000,
			pressure:    50,
			temperature: 255,
			// 1000 + 1000 * (-50/255) * (130/255)
			// 1000 + 1000 * -0.196 * 0.5098
			// 1000 + 1000 * -0.1 ~= 900
			expected:    900,
			tolerance:   1,
		},
	}

	config := ProcessingConfig{
		FlowEquation: equation,
		Filters:      []FilterConfig{},
	}
	processor := NewProcessor(config)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			processor.LatestPressure = tc.pressure
			processor.LatestTemperature = tc.temperature

			result, err := processor.CalculateFlow(config.FlowEquation, tc.flow, 0, refP, refT)
			if err != nil {
				t.Fatalf("Calculation error: %v", err)
			}

			diff := int32(math.Abs(float64(result - tc.expected)))
			if diff > tc.tolerance {
				t.Errorf("Inputs(F=%d, P=%d, T=%d): expected ~%d, got %d (diff %d > %d)",
					tc.flow, tc.pressure, tc.temperature, tc.expected, result, diff, tc.tolerance)
			}
		})
	}
}

func TestCalculateFlowWithFilters(t *testing.T) {
	// Verify that filters are applied to the inputs *before* calculation
	
	equation := "F + F * ((P - RefP) / 255) * ((T - RefT) / 255)"
	config := ProcessingConfig{
		FlowEquation: equation,
		Filters: []FilterConfig{
			{Type: "low_pass", Target: "pressure", Alpha: 0.5},
		},
	}
	processor := NewProcessor(config)
	
	refP := int32(100)
	refT := int32(125)

	// Set Temperature to reference point (125) so it doesn't affect calculation
	processor.LatestTemperature = 125

	// 1. Update Pressure with 100 (Reference) -> Stored 100
	processor.UpdatePressure(100)
	
	// 2. Update Pressure with 200. 
	// Filter: Prev=100, New=200. Alpha=0.5. Result = 100 + 0.5*(200-100) = 150
	processor.UpdatePressure(200)

	if processor.LatestPressure != 150 {
		t.Errorf("Filter expected to dampen pressure to 150, got %d", processor.LatestPressure)
	}

	// Calculate Flow: F=1000, P=150, T=125
	// Since T=125, the second term (T-125)/255 is 0.
	// Expected: 1000 + 1000 * ((150-100)/255) * 0 = 1000
	
	result, err := processor.CalculateFlow(config.FlowEquation, 1000, 0, refP, refT)
	if err != nil {
		t.Fatalf("Calculation error: %v", err)
	}

	if result != 1000 {
		t.Errorf("Filtered Flow Calculation: expected %d, got %d", 1000, result)
	}
}
