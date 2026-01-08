package main

import (
	"fmt"
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
	Value     int32
	Timestamp time.Time
}

// readSensorValue calculates the sensor value based on the equation and noise.
func readSensorValue(config SensorConfig, startTime time.Time, params map[string]interface{}, r *rand.Rand) (int32, error) {
	elapsed := time.Since(startTime).Seconds()
	
	// Prepare parameters for the equation
	parameters := make(map[string]interface{})
	for k, v := range params {
		parameters[k] = v
	}
	parameters["t"] = elapsed
	
	baseValue, err := EvaluateEquation(config.Equation, parameters)
	if err != nil {
		return 0, err
	}

	// Add random noise
	var noise float64
	if config.NoiseDistribution == "normal" {
		// Normal distribution (Gaussian): Mean 0, StdDev 1 * Amplitude
		// This treats Amplitude as the standard deviation
		noise = r.NormFloat64() * config.NoiseAmplitude
	} else {
		// Default: Uniform distribution [-Amplitude, +Amplitude]
		noise = (r.Float64()*2 - 1) * config.NoiseAmplitude
	}
	finalValue := baseValue + noise

	// Clamp and scale based on resolution bits
	maxVal := float64(uint64(1)<<config.ResolutionBits - 1)
	if finalValue < 0 {
		finalValue = 0
	}
	if finalValue > maxVal {
		finalValue = maxVal
	}

	return int32(finalValue), nil
}

// StartSensor starts a generic sensor simulation.
// It returns a channel for that specific sensor type.
func StartSensor(sType SensorType, config SensorConfig, params map[string]interface{}, seed int64) <-chan SensorData {
	ch := make(chan SensorData)
	startTime := time.Now()
	// Create local random source
	r := rand.New(rand.NewSource(seed))

	go func() {
		ticker := time.NewTicker(time.Second / time.Duration(config.FrequencyHz))
		defer ticker.Stop()

		for range ticker.C {
			val, err := readSensorValue(config, startTime, params, r)
			if err != nil {
				fmt.Printf("Error reading %s: %v\n", sType, err)
				continue
			}

			ch <- SensorData{
				Type:      sType,
				Value:     val,
				Timestamp: time.Now(),
			}
		}
	}()
	return ch
}
