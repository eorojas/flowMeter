package main

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/Knetic/govaluate"
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
	Value     uint32
	Timestamp time.Time
}

// getFunctions returns a map of supported math functions for govaluate.
func getFunctions() map[string]govaluate.ExpressionFunction {
	return map[string]govaluate.ExpressionFunction{
		"sin": func(args ...interface{}) (interface{}, error) {
			val := args[0].(float64)
			return math.Sin(val), nil
		},
		"cos": func(args ...interface{}) (interface{}, error) {
			val := args[0].(float64)
			return math.Cos(val), nil
		},
		"tan": func(args ...interface{}) (interface{}, error) {
			val := args[0].(float64)
			return math.Tan(val), nil
		},
		"sqrt": func(args ...interface{}) (interface{}, error) {
			val := args[0].(float64)
			return math.Sqrt(val), nil
		},
	}
}

// evaluateEquation parses and evaluates the sensor equation at time t (seconds).
func evaluateEquation(equation string, t float64) (float64, error) {
	functions := getFunctions()
	expression, err := govaluate.NewEvaluableExpressionWithFunctions(equation, functions)
	if err != nil {
		return 0, err
	}

	parameters := make(map[string]interface{})
	parameters["t"] = t

	result, err := expression.Evaluate(parameters)
	if err != nil {
		return 0, err
	}

	// Helper to convert result to float64 safely
	if val, ok := result.(float64); ok {
		return val, nil
	}
	return 0, fmt.Errorf("equation result is not a float64")
}

// readSensorValue calculates the sensor value based on the equation and noise.
func readSensorValue(config SensorConfig, startTime time.Time) (uint32, error) {
	elapsed := time.Since(startTime).Seconds()
	
	// Evaluate the base value from the equation
	baseValue, err := evaluateEquation(config.Equation, elapsed)
	if err != nil {
		return 0, err
	}

	// Add random noise
	// noise is random value between [-noiseAmplitude, +noiseAmplitude]
	noise := (rand.Float64()*2 - 1) * config.NoiseAmplitude
	finalValue := baseValue + noise

	// Clamp and scale based on resolution bits
	maxVal := float64(uint64(1)<<config.ResolutionBits - 1)
	if finalValue < 0 {
		finalValue = 0
	}
	if finalValue > maxVal {
		finalValue = maxVal
	}

	return uint32(finalValue), nil
}

// StartSensor starts a generic sensor simulation.
// It returns a channel for that specific sensor type.
func StartSensor(sType SensorType, config SensorConfig) <-chan SensorData {
	ch := make(chan SensorData)
	startTime := time.Now()

	go func() {
		ticker := time.NewTicker(time.Second / time.Duration(config.FrequencyHz))
		defer ticker.Stop()

		for range ticker.C {
			val, err := readSensorValue(config, startTime)
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
