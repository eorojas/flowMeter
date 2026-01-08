package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"time"
)

func main() {
	// Define command-line flags
	configPath := flag.String("config", "config.json", "Path to the configuration file")

	// Temp override flags
	tempVal := flag.Int("temp-override-value", 25, "Override temperature value (int32). Default is mid-range.")
	tempSrc := flag.String("temp-override-source", "cli", "Source of the temperature override.")

	// Pressure override flags
	pressureVal := flag.Int("pressure-override-value", 100, "Override pressure value (int32). Default is mid-range.")
	pressureSrc := flag.String("pressure-override-source", "cli", "Source of the pressure override.")

	flag.Parse()

	// Determine active overrides based on passed flags
	setFlags := make(map[string]bool)
	flag.Visit(func(f *flag.Flag) {
		setFlags[f.Name] = true
	})

	overrideTempActive := setFlags["temp-override-value"] || setFlags["temp-override-source"]
	overridePressureActive := setFlags["pressure-override-value"] || setFlags["pressure-override-source"]

	fmt.Println("Project initialized. Starting FlowMeter Simulation...")

	// Seed random number generator
	rand.Seed(time.Now().UnixNano())

	// Load Configuration
	config, err := LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config from %s: %v", *configPath, err)
	}
	fmt.Printf("Configuration loaded from %s.\n", *configPath)

	// Initialize Processor
	processor := NewProcessor(config.Processing)

	// Apply overrides if provided
	if overrideTempActive {
		to := TempOverride{
			Value:  int32(*tempVal),
			Source: *tempSrc,
		}
		processor.ApplyTempOverride(to)
		fmt.Printf("Temperature override applied: Value=%d, Source=%s\n", to.Value, to.Source)
	}

	if overridePressureActive {
		po := PressureOverride{
			Value:  int32(*pressureVal),
			Source: *pressureSrc,
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
	timeout := time.After(runSecs * time.Second)
	startTime := time.Now()

	fmt.Println("Listening for sensor data...")
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
				Timestamp:      data.Timestamp,
				RawFlow:        data.Value,
				Pressure:       processor.LatestPressure,
				Temperature:    processor.LatestTemperature,
				CalculatedFlow: calculated,
			}

			if err := outputHandler.Write(outData); err != nil {
				log.Printf("Error writing output: %v", err)
			}

		case <-timeout:
			fmt.Println("Simulation finished.")
			return
		}
	}
}