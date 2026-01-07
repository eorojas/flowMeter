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

	// 24-bit max value is 16,777,215
	if data.Value > 16777215 {
		t.Errorf("Flow value %d out of expected 24-bit range [0, 16777215]", data.Value)
	}
}

func TestReadPressure(t *testing.T) {
	data := readPressure()

	if data.Type != PressureSensor {
		t.Errorf("Expected sensor type %s, got %s", PressureSensor, data.Type)
	}

	// 8-bit max value is 255
	if data.Value > 255 {
		t.Errorf("Pressure value %d out of expected 8-bit range [0, 255]", data.Value)
	}
}

func TestReadTemperature(t *testing.T) {
	data := readTemperature()

	if data.Type != TemperatureSensor {
		t.Errorf("Expected sensor type %s, got %s", TemperatureSensor, data.Type)
	}

	// 8-bit max value is 255
	if data.Value > 255 {
		t.Errorf("Temperature value %d out of expected 8-bit range [0, 255]", data.Value)
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
