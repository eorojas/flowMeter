package main

import (
	"fmt"
	"time"
)

func main() {
	fmt.Println("Project initialized. Starting FlowMeter Simulation with isolated sensors...")

	// Start independent sensor simulations
	flowCh := StartFlowSensor()
	pressureCh := StartPressureSensor()
	tempCh := StartTemperatureSensor()

	// Consume data from all channels independently
	// We run this for a fixed duration for demonstration
	timeout := time.After(2 * time.Second)

	fmt.Println("Listening for sensor data...")
	for {
		select {
		case data := <-flowCh:
			// Process Flow Data
			fmt.Printf("[%s] %-12s: %.2f L/min\n", 
				data.Timestamp.Format("15:04:05.000"), data.Type, data.Value)

		case data := <-pressureCh:
			// Process Pressure Data
			fmt.Printf("[%s] %-12s: %.0f (ADC)\n", 
				data.Timestamp.Format("15:04:05.000"), data.Type, data.Value)

		case data := <-tempCh:
			// Process Temperature Data
			fmt.Printf("[%s] %-12s: %.2f C\n", 
				data.Timestamp.Format("15:04:05.000"), data.Type, data.Value)

		case <-timeout:
			fmt.Println("Simulation finished.")
			return
		}
	}
}