package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"
)

func main() {
	fmt.Println("Project initialized. Starting FlowMeter Simulation...")
	
	// Seed random number generator
	rand.Seed(time.Now().UnixNano())

	// Load Configuration
	config, err := LoadConfig("config.json")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	fmt.Println("Configuration loaded.")

	// Start independent sensor simulations using config
	flowCh := StartSensor(FlowSensor, config.Sensors.Flow)
	pressureCh := StartSensor(PressureSensor, config.Sensors.Pressure)
	tempCh := StartSensor(TemperatureSensor, config.Sensors.Temperature)

	// Consume data
	timeout := time.After(2 * time.Second)

	fmt.Println("Listening for sensor data...")
	for {
		select {
		case data := <-flowCh:
			fmt.Printf("[%s] %-12s: %d\n", data.Timestamp.Format("15:04:05.000"), data.Type, data.Value)
		case data := <-pressureCh:
			fmt.Printf("[%s] %-12s: %d\n", data.Timestamp.Format("15:04:05.000"), data.Type, data.Value)
		case data := <-tempCh:
			fmt.Printf("[%s] %-12s: %d\n", data.Timestamp.Format("15:04:05.000"), data.Type, data.Value)
		case <-timeout:
			fmt.Println("Simulation finished.")
			return
		}
	}
}