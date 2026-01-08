package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"

	flag "github.com/spf13/pflag"
)

func main() {
	// Pre-scan for config file to load defaults
	configPath := "config.json"
	// We do a simple manual scan because pflag parsing requires all flags to be defined
	for i, arg := range os.Args {
		if (arg == "-c" || arg == "--config") && i+1 < len(os.Args) {
			configPath = os.Args[i+1]
			break
		}
	}

	// Load Configuration
	config, err := LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load config from %s: %v", configPath, err)
	}
	fmt.Printf("Configuration loaded from %s.\n", configPath)

	// Define command-line flags using pflag
	// We redefine the config flag here so it appears in help, even though we manually parsed it above
	var configPathFlag string
	flag.StringVarP(&configPathFlag, "config", "c", configPath, "Path to the configuration file")

	// Temp override flags - Default from Config
	var tempVal int
	flag.IntVarP(&tempVal, "temp-override-value", "T", int(config.Simulation.DefaultTemperature), "Override temperature value (int32).")

	// Pressure override flags - Default from Config
	var pressureVal int
	flag.IntVarP(&pressureVal, "pressure-override-value", "P", int(config.Simulation.DefaultPressure), "Override pressure value (int32).")

	// Sample count flag - default to -1 to detect if it was passed
	var sampleCountOverride int
	flag.IntVarP(&sampleCountOverride, "samples", "n", -1, "Number of samples to simulate.")

	// Filter type flag
	var useMedian bool
	flag.BoolVarP(&useMedian, "median", "m", false, "Use median filter instead of default (low_pass).")

	// Random seed flag
	var randomSeed bool
	flag.BoolVarP(&randomSeed, "random-seed", "r", false, "Use time-based random seed (default is deterministic seed 0).")

	flag.Parse()

	fmt.Println("Project initialized. Starting FlowMeter Simulation...")

	// Seed random number generator
	var baseSeed int64
	if randomSeed {
		baseSeed = time.Now().UnixNano()
		fmt.Printf("Using random base seed: %d\n", baseSeed)
	} else {
		baseSeed = 0
		fmt.Println("Using deterministic base seed 0.")
	}
	// Note: We still seed the global rand just in case, but sensors use local Rand
	rand.Seed(baseSeed)

	// Apply sample count override ONLY if it was passed
	if flag.Lookup("samples").Changed {
		config.Simulation.DefaultSamples = int32(sampleCountOverride)
		fmt.Printf("Sample count override: %d\n", config.Simulation.DefaultSamples)
	} else {
		fmt.Printf("Using config default samples: %d\n", config.Simulation.DefaultSamples)
	}

	// Apply Temperature Override by rewriting the equation IF CHANGED from default
	if flag.Lookup("temp-override-value").Changed {
		config.Sensors.Temperature.Equation = strconv.Itoa(tempVal)
		fmt.Printf("Temperature equation overridden to constant: %s\n", config.Sensors.Temperature.Equation)
	}

	// Apply Pressure Override by rewriting the equation
	if flag.Lookup("pressure-override-value").Changed {
		config.Sensors.Pressure.Equation = strconv.Itoa(pressureVal)
		fmt.Printf("Pressure equation overridden to constant: %s\n", config.Sensors.Pressure.Equation)
	}

	// Determine Filter Type
	filterType := config.Processing.DefaultFilterType
	if filterType == "" {
		filterType = "low_pass" // Fallback default
	}
	if useMedian {
		filterType = "median"
		fmt.Println("Filter type overridden to: median")
	} else {
		fmt.Printf("Using configured filter type: %s\n", filterType)
	}

	// Apply filter type to all filters
	for i := range config.Processing.Filters {
		config.Processing.Filters[i].Type = filterType
	}

	// Initialize Processor
	processor := NewProcessor(config.Processing)
	// Initialize with defaults (which now come from flags/config)
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
	refParams := map[string]interface{}{
		"RefP": float64(config.Simulation.DefaultPressure),
		"RefT": float64(config.Simulation.DefaultTemperature),
	}
	// Use unique seeds derived from baseSeed for each sensor
	flowCh := StartSensor(FlowSensor, config.Sensors.Flow, refParams, baseSeed)
	pressureCh := StartSensor(PressureSensor, config.Sensors.Pressure, refParams, baseSeed+1)
	tempCh := StartSensor(TemperatureSensor, config.Sensors.Temperature, refParams, baseSeed+2)

	// Consume data
	// Calculate run duration based on samples and flow frequency
	runSecs := time.Duration(config.Simulation.DefaultSamples / config.Sensors.Flow.FrequencyHz)
	// Add a small buffer to ensure we get the last sample if timing is tight
	timeout := time.After(runSecs*time.Second + 500*time.Millisecond)
	startTime := time.Now()

	fmt.Println("Listening for sensor data...")
	var sampleCount int64
	maxSamples := int64(config.Simulation.DefaultSamples)

	// Capture reference values for the equation
	refPressure := config.Simulation.DefaultPressure
	refTemperature := config.Simulation.DefaultTemperature

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

			// Calculate Final Flow using updated signature
			calculated, err := processor.CalculateFlow(config.Processing.FlowEquation, data.Value, elapsed, refPressure, refTemperature)
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