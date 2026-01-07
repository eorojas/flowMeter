# Gemini Project Context: FlowMeter

## Project Overview
### FlowMeter Simulation
A simulation of a multi-sensor flow meter.

#### Components
- **ADC Converters:** 
  Simulates hardware interrupts via timer-driven methods.
- **Primary Flow Input:** 
  A timer-driven method that delivers samples to a consumer at 100Hz.
- **Pressure Input:** 
  A timer-driven method that delivers 8-bit ADC samples to a consumer at 
  10Hz.
- **Temperature Input:** 
  A timer-driven method that delivers samples to a consumer at 10Hz.
- **Sample Method:** 
  These simulate interrupts from the ADCs by delivering samples to a 
  consumer.

#### Configuration
- **JSON Config:** 
  A file that describes the equation used for flow calculation.

## Goals
1. **Primary:** Learn the Go programming language (Golang).
2. **Secondary:** Build a functional application to simulate a flow meter.

## Technology Stack
- **Language:** Go (Latest stable)
- **Key Libraries:** (To be added, e.g., Cobra, Gin)

## Coding Preferences & Conventions
- **Style:** Idiomatic Go (Effective Go standards).
- **Comments:** Provide detailed comments explaining *why* certain Go 
  features (like channels, goroutines, interfaces) are used.
- **Error Handling:** Explicit error checking (standard `if err != nil`).

## Roadmap
- [x] Initialize Go module (`go mod init`)
- [x] Create basic "Hello World" entry point
- [ ] Implement core logic
- [ ] Add unit tests
- [ ] Refactor for better structure

## Useful Commands
- Run project: `go run .`
- Test project: `go test ./...`
- Build: `go build -o flowMeter`