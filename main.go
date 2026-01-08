package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	flag "github.com/spf13/pflag"
)

func main() {
	// Define command-line flags using pflag
	var configPath string
	flag.StringVarP(&configPath, "config", "c", "config.json", "Path to the configuration file")

	// Temp override flags
	var tempVal int
	flag.IntVarP(&tempVal, "temp-override-value", "T", 25, "Override temperature value (int32).")

	// Pressure override flags
	var pressureVal int
	flag.IntVarP(&pressureVal, "pressure-override-value", "P", 100, "Override pressure value (int32).")

	// Sample count flag
	var sampleCountOverride int
	flag.IntVarP(&sampleCountOverride, "samples", "n", 10000, "Number of samples to simulate.")

	flag.Parse()

	fmt.Println("Project initialized. Starting FlowMeter Simulation...")

	// Seed random number generator
	rand.Seed(time.Now().UnixNano())

	// Load Configuration
	config, err := LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load config from %s: %v", configPath, err)
	}
	fmt.Printf("Configuration loaded from %s.\n", configPath)

	// Apply sample count override (always use the flag value, which defaults to 10000)
	config.Simulation.DefaultSamples = int32(sampleCountOverride)
	fmt.Printf("Samples to simulate: %d\n", config.Simulation.DefaultSamples)

	// Initialize Processor
	processor := NewProcessor(config.Processing)

	// Always apply overrides (using default values if flags not passed)
	to := TempOverride{
		Value: int32(tempVal),
	}
	processor.ApplyTempOverride(to)
	fmt.Printf("Temperature override applied: Value=%d\n", to.Value)

	po := PressureOverride{
		Value: int32(pressureVal),
	}
	processor.ApplyPressureOverride(po)
	fmt.Printf("Pressure override applied: Value=%d\n", po.Value)

	// Initialize Output Handler
	outputHandler, err := GetOutputHandler(config.Output)
	if err != nil {
		log.Fatalf("Failed to initialize output handler: %v", err)
	}
	defer outputHandler.Close()

	// Start independent sensor simulations using config
	flowCh := StartSensor(FlowSensor, config.Sensors.Flow)
	
	// Since temperature and pressure are ALWAYS overridden now, we don't start their sensor simulations.
	// This simplifies the logic and follows the requirement to use the flag values (defaults).

	// Consume data
	// Calculate run duration based on samples and flow frequency
	runSecs := time.Duration(config.Simulation.DefaultSamples / config.Sensors.Flow.FrequencyHz)
	// Add a small buffer to ensure we get the last sample if timing is tight
	timeout := time.After(runSecs*time.Second + 500*time.Millisecond)
	startTime := time.Now()

	fmt.Println("Listening for sensor data...")
	var sampleCount int64
	maxSamples := int64(config.Simulation.DefaultSamples)

	for {
		select {
		case data := <-flowCh:
			sampleCount++
			// Calculate elapsed time for the equation
			elapsed := data.Timestamp.Sub(startTime).Seconds()

			// Calculate Final Flow
			calculated, err := processor.CalculateFlow(config.Processing.FlowEquation, data.Value, elapsed)
			if err != nil {
				log.Printf("Error calculating flow: %v", err)
				continue
			}

			// Prepare Output
			outData := OutputData{
				SampleNumber:   sampleCount,
				RawFlow:        data.Value,
				Pressure:       processor.LatestPressure,
				Temperature:    processor.LatestTemperature,
				CalculatedFlow: calculated,
			}

			if err := outputHandler.Write(outData); err != nil {
				log.Printf("Error writing output: %v", err)
			}

			if sampleCount >= maxSamples {
				fmt.Println("Simulation finished (sample limit reached).")
				return
			}

		case <-timeout:
			fmt.Println("Simulation finished (timeout).")
			return
		}
	}
}
