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
	overrideTemp := flag.Int("temp", -1, "Override temperature with a constant value (int32)")
	overridePressure := flag.Int("pressure", -1, "Override pressure with a constant value (int32)")
	flag.Parse()

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
	if *overrideTemp != -1 {
		processor.LatestTemperature = int32(*overrideTemp)
		fmt.Printf("Temperature override: %d\n", processor.LatestTemperature)
	}
	if *overridePressure != -1 {
		processor.LatestPressure = int32(*overridePressure)
		fmt.Printf("Pressure override: %d\n", processor.LatestPressure)
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
	if *overridePressure == -1 {
		pressureCh = StartSensor(PressureSensor, config.Sensors.Pressure)
	}

	var tempCh <-chan SensorData
	if *overrideTemp == -1 {
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
				CalculatedFlow: int32(calculated), // calculated is int64, but OutputData struct expects int32.
			}
            
            // Wait, OutputData.CalculatedFlow is int32. The logic in processing.go returns int32 now? 
            // I checked processing.go, I changed it to return int32 in the last step.
            // So casting int32(calculated) is redundant if it returns int32, but safe if I misremembered.
            // Actually, wait, let me check processing.go again to be sure.

			if err := outputHandler.Write(outData); err != nil {
				log.Printf("Error writing output: %v", err)
			}

		case <-timeout:
			fmt.Println("Simulation finished.")
			return
		}
	}
}

