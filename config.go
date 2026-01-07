package main

import (
	"encoding/json"
	"os"
)

// Config represents the top-level configuration structure.
type Config struct {
	Simulation SimulationConfig `json:"simulation"`
	Sensors    SensorsConfig    `json:"sensors"`
	Processing ProcessingConfig `json:"processing"`
	Output     OutputConfig     `json:"output"`
}

type SimulationConfig struct {
	DefaultSamples int32 `json:"default_samples"`
}

type SensorsConfig struct {
	Flow        SensorConfig `json:"flow"`
	Pressure    SensorConfig `json:"pressure"`
	Temperature SensorConfig `json:"temperature"`
}

type SensorConfig struct {
	FrequencyHz    int32   `json:"frequency_hz"`
	ResolutionBits int32   `json:"resolution_bits"`
	Equation       string  `json:"equation"`
	NoiseAmplitude float64 `json:"noise_amplitude"`
}

type ProcessingConfig struct {
	FlowEquation string         `json:"flow_equation"`
	Filters      []FilterConfig `json:"filters"`
}

type FilterConfig struct {
	Type   string  `json:"type"`   // e.g., "low_pass", "median"
	Target string  `json:"target"` // e.g., "pressure"
	Alpha  float64 `json:"alpha,omitempty"`
}

type OutputConfig struct {
	Type   string `json:"type"`   // e.g., "file", "network"
	Target string `json:"target"` // e.g., filename or URL
}

// LoadConfig reads and parses the config.json file.
func LoadConfig(filename string) (*Config, error) {
	file, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config Config
	err = json.Unmarshal(file, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
