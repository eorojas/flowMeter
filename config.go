package main

import (
    "fmt"
    "github.com/go-json-experiment/json"
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
    DefaultSamples     int32 `json:"default_samples"`
    DefaultPressure    int32 `json:"default_pressure"`
    DefaultTemperature int32 `json:"default_temperature"`
    DefaultFlow        int32 `json:"default_flow"`
}

type SensorsConfig struct {
    Flow        SensorConfig `json:"flow"`
    Pressure    SensorConfig `json:"pressure"`
    Temperature SensorConfig `json:"temperature"`
}

type SensorConfig struct {
    FrequencyHz       int32   `json:"frequency_hz"`
    ResolutionBits    int32   `json:"resolution_bits"`
    Equation          string  `json:"equation"`
    NoiseAmplitude    float64 `json:"noise_amplitude"`
    // "uniform" (default) or "normal"
    NoiseDistribution string  `json:"noise_distribution,omitempty"`
}

type ProcessingConfig struct {
    FlowEquation      string         `json:"flow_equation"`
    // "low_pass" or "median"
    DefaultFilterType string         `json:"default_filter_type"`
    Filters           []FilterConfig `json:"filters"`
}

type FilterConfig struct {
    // e.g., "low_pass", "median"
    Type       string  `json:"type"`
    // e.g., "pressure"
    Target     string  `json:"target"`
    // For Low Pass
    Alpha      float64 `json:"alpha,omitempty"`
    // For Median
    WindowSize int     `json:"window_size,omitempty"`
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

    if err := config.Validate(); err != nil {
        return nil, err
    }

    return &config, nil
}

// Validate checks configuration constraints.
func (c *Config) Validate() error {
    if c.Simulation.DefaultPressure < 10 ||
        c.Simulation.DefaultPressure > 250 {
        return fmt.Errorf("default_pressure must be between 10 and 250, got %d",
            c.Simulation.DefaultPressure)
    }
    if c.Simulation.DefaultTemperature < 10 ||
        c.Simulation.DefaultTemperature > 250 {
        return fmt.Errorf("default_temperature must be 10-250, got %d",
            c.Simulation.DefaultTemperature)
    }
    if c.Simulation.DefaultFlow < 0 ||
        c.Simulation.DefaultFlow > 16777215 {
        return fmt.Errorf("default_flow must be within 24-bit range, got %d",
            c.Simulation.DefaultFlow)
    }
    if c.Processing.DefaultFilterType != "" {
        if c.Processing.DefaultFilterType != "low_pass" &&
            c.Processing.DefaultFilterType != "median" {
            return fmt.Errorf("default_filter_type must be "+
                "'low_pass' or 'median', got %s",
                c.Processing.DefaultFilterType)
        }
    }
    return nil
}

