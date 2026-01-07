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

	fmt.Println("Listening for sensor data (Raw ADC Values)...")
	for {
		select {
		case data := <-flowCh:
			// Process Flow Data (24-bit)
			fmt.Printf("[%s] %-12s: %d (24-bit)\n", 
				data.Timestamp.Format("15:04:05.000"), data.Type, data.Value)

		case data := <-pressureCh:
			// Process Pressure Data (8-bit)
			fmt.Printf("[%s] %-12s: %d (8-bit)\n", 
				data.Timestamp.Format("15:04:05.000"), data.Type, data.Value)

		case data := <-tempCh:
			// Process Temperature Data (8-bit)
			fmt.Printf("[%s] %-12s: %d (8-bit)\n", 
				data.Timestamp.Format("15:04:05.000"), data.Type, data.Value)

		case <-timeout:
			fmt.Println("Simulation finished.")
			return
		}
	}
}
