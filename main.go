package main

import (
	"fmt"
	"log"
	"math/rand"
	"strconv"
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

	// Sample count flag - default to -1 to detect if it was passed
	var sampleCountOverride int
	flag.IntVarP(&sampleCountOverride, "samples", "n", -1, "Number of samples to simulate.")

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

	// Apply sample count override ONLY if it was passed
	if flag.Lookup("samples").Changed {
		config.Simulation.DefaultSamples = int32(sampleCountOverride)
		fmt.Printf("Sample count override: %d\n", config.Simulation.DefaultSamples)
	} else {
		fmt.Printf("Using config default samples: %d\n", config.Simulation.DefaultSamples)
	}

	// Apply Temperature Override by rewriting the equation
	if flag.Lookup("temp-override-value").Changed {
		config.Sensors.Temperature.Equation = strconv.Itoa(tempVal)
		fmt.Printf("Temperature equation overridden to constant: %s\n", config.Sensors.Temperature.Equation)
	}

	// Apply Pressure Override by rewriting the equation
	if flag.Lookup("pressure-override-value").Changed {
		config.Sensors.Pressure.Equation = strconv.Itoa(pressureVal)
		fmt.Printf("Pressure equation overridden to constant: %s\n", config.Sensors.Pressure.Equation)
	}

	// Initialize Processor
	processor := NewProcessor(config.Processing)
	// Initialize with default or overridden values so they don't start at zero
	processor.LatestTemperature = int32(tempVal)
	processor.LatestPressure = int32(pressureVal)
	fmt.Printf("Initial state: Temperature=%d, Pressure=%d\n", processor.LatestTemperature, processor.LatestPressure)

	// Initialize Output Handler
	outputHandler, err := GetOutputHandler(config.Output)
	if err != nil {
		log.Fatalf("Failed to initialize output handler: %v", err)
	}
	defer outputHandler.Close()

	// Start independent sensor simulations using config
	// All sensors are started, even if overridden, so that noise/filtering logic runs.
	flowCh := StartSensor(FlowSensor, config.Sensors.Flow)
	pressureCh := StartSensor(PressureSensor, config.Sensors.Pressure)
	tempCh := StartSensor(TemperatureSensor, config.Sensors.Temperature)

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
		case data := <-pressureCh:
			processor.UpdatePressure(data.Value)

		case data := <-tempCh:
			processor.UpdateTemperature(data.Value)

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
