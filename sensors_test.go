package main

import (
	"testing"
	"time"
)

func TestReadFlow(t *testing.T) {
	data := readFlow()

	if data.Type != FlowSensor {
		t.Errorf("Expected sensor type %s, got %s", FlowSensor, data.Type)
	}

	if data.Value < 0 || data.Value > 100 {
		t.Errorf("Flow value %f out of expected range [0, 100]", data.Value)
	}
}

func TestReadPressure(t *testing.T) {
	data := readPressure()

	if data.Type != PressureSensor {
		t.Errorf("Expected sensor type %s, got %s", PressureSensor, data.Type)
	}

	// Check for 8-bit resolution limits (0-255)
	if data.Value < 0 || data.Value > 255 {
		t.Errorf("Pressure value %f out of expected range [0, 255]", data.Value)
	}
	
	if data.Value != float64(int(data.Value)) {
		t.Errorf("Pressure value %f should be an integer (simulating 8-bit ADC)", data.Value)
	}
}

func TestReadTemperature(t *testing.T) {
	data := readTemperature()

	if data.Type != TemperatureSensor {
		t.Errorf("Expected sensor type %s, got %s", TemperatureSensor, data.Type)
	}

	if data.Value < 20 || data.Value > 30 {
		t.Errorf("Temperature value %f out of expected range [20, 30]", data.Value)
	}
}

func TestStartFlowSensor(t *testing.T) {
	ch := StartFlowSensor()
	
	select {
	case data := <-ch:
		if data.Type != FlowSensor {
			t.Errorf("Expected FlowSensor data from channel, got %s", data.Type)
		}
	case <-time.After(1 * time.Second):
		t.Error("Timed out waiting for Flow sensor data")
	}
}