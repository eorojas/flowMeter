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
	var tempSrc string
	flag.IntVarP(&tempVal, "temp-override-value", "T", 25, "Override temperature value (int32). Default is mid-range.")
	flag.StringVar(&tempSrc, "temp-override-source", "cli", "Source of the temperature override.")

	// Pressure override flags
	var pressureVal int
	var pressureSrc string
	flag.IntVarP(&pressureVal, "pressure-override-value", "P", 100, "Override pressure value (int32). Default is mid-range.")
	flag.StringVar(&pressureSrc, "pressure-override-source", "cli", "Source of the pressure override.")

	// Sample count flag
	var sampleCountOverride int
	flag.IntVarP(&sampleCountOverride, "samples", "n", -1, "Override number of samples to simulate.")

	flag.Parse()

	// Determine active overrides based on passed flags
	// pflag provides a simple way to check if a flag was changed
	overrideTempActive := flag.Lookup("temp-override-value").Changed || flag.Lookup("temp-override-source").Changed
	overridePressureActive := flag.Lookup("pressure-override-value").Changed || flag.Lookup("pressure-override-source").Changed

	fmt.Println("Project initialized. Starting FlowMeter Simulation...")

	// Seed random number generator
	rand.Seed(time.Now().UnixNano())

	// Load Configuration
	config, err := LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load config from %s: %v", configPath, err)
	}
	fmt.Printf("Configuration loaded from %s.\n", configPath)

	// Apply sample count override if provided
	if sampleCountOverride != -1 {
		config.Simulation.DefaultSamples = int32(sampleCountOverride)
		fmt.Printf("Sample count override: %d\n", config.Simulation.DefaultSamples)
	}

	// Initialize Processor
	processor := NewProcessor(config.Processing)

	// Apply overrides if provided
	if overrideTempActive {
		to := TempOverride{
			Value:  int32(tempVal),
			Source: tempSrc,
		}
		processor.ApplyTempOverride(to)
		fmt.Printf("Temperature override applied: Value=%d, Source=%s\n", to.Value, to.Source)
	}

	if overridePressureActive {
		po := PressureOverride{
			Value:  int32(pressureVal),
			Source: pressureSrc,
		}
		processor.ApplyPressureOverride(po)
		fmt.Printf("Pressure override applied: Value=%d, Source=%s\n", po.Value, po.Source)
	}

	// Initialize Output Handler
	outputHandler, err := GetOutputHandler(config.Output)
	if err != nil {
		log.Fatalf("Failed to initialize output handler: %v", err)
	}
	defer outputHandler.Close()

	// Start independent sensor simulations using config
	flowCh := StartSensor(FlowSensor, config.Sensors.Flow)

	// Only start sensors if they are not overridden
	var pressureCh <-chan SensorData
	if !overridePressureActive {
		pressureCh = StartSensor(PressureSensor, config.Sensors.Pressure)
	}

	var tempCh <-chan SensorData
	if !overrideTempActive {
		tempCh = StartSensor(TemperatureSensor, config.Sensors.Temperature)
	}

	// Consume data
	// Calculate run duration based on samples and flow frequency
	runSecs := time.Duration(config.Simulation.DefaultSamples / config.Sensors.Flow.FrequencyHz)
	// Add a small buffer to ensure we get the last sample if timing is tight
	timeout := time.After(runSecs*time.Second + 100*time.Millisecond)
	startTime := time.Now()

	fmt.Println("Listening for sensor data...")
	var sampleCount int64
	// Limit loop by sample count as well
	maxSamples := int64(config.Simulation.DefaultSamples)

	for {
		select {
		case data, ok := <-pressureCh:
			if ok {
				processor.UpdatePressure(data.Value)
			}

		case data, ok := <-tempCh:
			if ok {
				processor.UpdateTemperature(data.Value)
			}

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