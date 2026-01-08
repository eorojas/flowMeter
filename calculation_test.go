package main

import (
	"math"
	"testing"
)

func TestCalculateFlowAccuracy(t *testing.T) {
	// Formula: F * (P / 255) * (T / 255)
	// Note: P and T are typically 8-bit (0-255). F is 24-bit.

	tests := []struct {
		name          string
		flow          int32
		pressure      int32
		temperature   int32
		expected      int32
		tolerance     int32 // Allow +/- 1 due to int casting/float precision
	}{
		{
			name:        "Max Values",
			flow:        1000,
			pressure:    255,
			temperature: 255,
			expected:    1000, // 1000 * 1 * 1
			tolerance:   0,
		},
		{
			name:        "Zero Flow",
			flow:        0,
			pressure:    100,
			temperature: 100,
			expected:    0,
			tolerance:   0,
		},
		{
			name:        "Zero Pressure",
			flow:        1000,
			pressure:    0,
			temperature: 255,
			expected:    0,
			tolerance:   0,
		},
		{
			name:        "Half Scale",
			flow:        5000,
			pressure:    127,
			temperature: 127,
			// 5000 * (127/255) * (127/255)
			// 5000 * 0.498039 * 0.498039
			// 5000 * 0.248043 ~= 1240.21
			expected:    1240,
			tolerance:   1,
		},
		{
			name:        "Typical Operating Point",
			flow:        5000,
			pressure:    100,
			temperature: 25,
			// 5000 * (100/255) * (25/255)
			// 5000 * 0.39215 * 0.09803
			// 5000 * 0.038446 ~= 192.23
			expected:    192,
			tolerance:   1,
		},
	}

	// Setup a processor with no filters for basic equation testing
	config := ProcessingConfig{
		FlowEquation: "F * (P / 255) * (T / 255)",
		Filters:      []FilterConfig{},
	}
	processor := NewProcessor(config)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Manually set state
			processor.LatestPressure = tc.pressure
			processor.LatestTemperature = tc.temperature

			// Calculate
			result, err := processor.CalculateFlow(config.FlowEquation, tc.flow, 0)
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
	// Verify that filters are applied to the inputs *before* calculation if setup
	
	// Setup processor with filters on Pressure
	// Low Pass Alpha 0.5
	config := ProcessingConfig{
		FlowEquation: "F * (P / 255) * (T / 255)",
		Filters: []FilterConfig{
			{Type: "low_pass", Target: "pressure", Alpha: 0.5},
		},
	}
	processor := NewProcessor(config)

	// Set Temperature constant Max
	processor.LatestTemperature = 255

	// 1. Update Pressure with 0 -> Stored 0
	processor.UpdatePressure(0)
	
	// 2. Update Pressure with 255. 
	// Filter: Prev=0, New=255. Alpha=0.5. Result = 0 + 0.5*(255-0) = 127 (integer math)
	processor.UpdatePressure(255)

	// Processor.LatestPressure should now be 127, NOT 255.

	if processor.LatestPressure != 127 {
		t.Errorf("Filter expected to dampen pressure to 127, got %d", processor.LatestPressure)
	}

	// Calculate Flow: F=1000, P=127, T=255
	// Expected: 1000 * (127/255) * 1 = 498
	
	result, err := processor.CalculateFlow(config.FlowEquation, 1000, 0)
	if err != nil {
		t.Fatalf("Calculation error: %v", err)
	}

	expected := int32(498)
	if result != expected {
		t.Errorf("Filtered Flow Calculation: expected %d, got %d", expected, result)
	}
}
