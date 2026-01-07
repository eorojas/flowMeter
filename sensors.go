package main

import (
	"math/rand"
	"time"
)

// SensorType defines the category of the sensor.
type SensorType string

const (
	FlowSensor        SensorType = "Flow"
	PressureSensor    SensorType = "Pressure"
	TemperatureSensor SensorType = "Temperature"
)

// SensorData represents a standardized data structure for a sensor reading.
type SensorData struct {
	Type      SensorType
	Value     float64
	Timestamp time.Time
}

// readFlow simulates reading the flow sensor.
func readFlow() SensorData {
	return SensorData{
		Type:      FlowSensor,
		Value:     rand.Float64() * 100.0,
		Timestamp: time.Now(),
	}
}

// readPressure simulates reading the pressure sensor (8-bit ADC).
func readPressure() SensorData {
	return SensorData{
		Type:      PressureSensor,
		Value:     float64(rand.Intn(256)),
		Timestamp: time.Now(),
	}
}

// readTemperature simulates reading the temperature sensor.
func readTemperature() SensorData {
	return SensorData{
		Type:      TemperatureSensor,
		Value:     20.0 + rand.Float64()*10.0,
		Timestamp: time.Now(),
	}
}

// StartFlowSensor starts the flow sensor simulation on its own goroutine.
// Returns a receive-only channel dedicated to flow data.
func StartFlowSensor() <-chan SensorData {
	ch := make(chan SensorData)
	go func() {
		// Flow @ 100Hz
		ticker := time.NewTicker(time.Second / 100)
		defer ticker.Stop()
		for range ticker.C {
			ch <- readFlow()
		}
	}()
	return ch
}

// StartPressureSensor starts the pressure sensor simulation on its own goroutine.
// Returns a receive-only channel dedicated to pressure data.
func StartPressureSensor() <-chan SensorData {
	ch := make(chan SensorData)
	go func() {
		// Pressure @ 10Hz
		ticker := time.NewTicker(time.Second / 10)
		defer ticker.Stop()
		for range ticker.C {
			ch <- readPressure()
		}
	}()
	return ch
}

// StartTemperatureSensor starts the temperature sensor simulation on its own goroutine.
// Returns a receive-only channel dedicated to temperature data.
func StartTemperatureSensor() <-chan SensorData {
	ch := make(chan SensorData)
	go func() {
		// Temperature @ 10Hz
		ticker := time.NewTicker(time.Second / 10)
		defer ticker.Stop()
		for range ticker.C {
			ch <- readTemperature()
		}
	}()
	return ch
}