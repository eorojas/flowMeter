package main

import (
	"math"
	"testing"
	"time"
)

func TestEvaluateEquation(t *testing.T) {
	tests := []struct {
		equation string
		t        float64
		expected float64
	}{
		{"10 + t", 5.0, 15.0},
		{"sin(t)", 0.0, 0.0},
		{"t * 2", 10.0, 20.0},
	}

	for _, test := range tests {
		params := map[string]interface{}{"t": test.t}
		result, err := EvaluateEquation(test.equation, params)
		if err != nil {
			t.Errorf("Error evaluating '%s': %v", test.equation, err)
		}
		if math.Abs(result-test.expected) > 0.001 {
			t.Errorf("Equation '%s' at t=%.1f: expected %.3f, got %.3f", 
				test.equation, test.t, test.expected, result)
		}
	}
}

func TestReadSensorValue(t *testing.T) {
	config := SensorConfig{
		Equation:       "100",
		NoiseAmplitude: 0.0,
		ResolutionBits: 8,
	}

	val, err := readSensorValue(config, time.Now(), nil)
	if err != nil {
		t.Fatalf("readSensorValue failed: %v", err)
	}

	if val != 100 {
		t.Errorf("Expected 100, got %d", val)
	}
}

func TestStartSensor(t *testing.T) {
	config := SensorConfig{
		FrequencyHz:    10,
		ResolutionBits: 8,
		Equation:       "50",
		NoiseAmplitude: 0.0,
	}

	ch := StartSensor(FlowSensor, config, nil)

	select {
	case data := <-ch:
		if data.Type != FlowSensor {
			t.Errorf("Expected FlowSensor, got %s", data.Type)
		}
		if data.Value != 50 {
			t.Errorf("Expected value 50, got %d", data.Value)
		}
	case <-time.After(1 * time.Second):
		t.Error("Timed out waiting for sensor data")
	}
}