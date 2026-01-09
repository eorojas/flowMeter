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
- **Output Method:** 
  These simulate interrupts from the ADCs by delivering samples to a 
  - Filter Out Data
    - The output data will be filtered. There should be options to implement
    a low pass filter (details to be decided), and/or a median filter.
    - The configuration will have math function for each sensor type.
    - There will be the ability to add noise on top of the sensor values.
    - A each flow value we will calculate the current filtered temperature and 
        pressure values, which will be allied to the flow equation.
    - The calculated flow rate (CFRate) (and any ancillary data) will
        be packaged and push to a configured end-point, e.g.,
        a file or network end-point.
    - For testing the number of samples should be command-line configurable
        and default to 10000.
- **Receiver Method:** 
    Receives the data sent by the Output Method, i.e., the CRRate data.
    - There are two types of Receiver methods, configurable, file
        and network (HTTP, CSV-receiver, other?).

#### Configuration
- **JSON Config:** 
  A file, config.json
  - Equation that describes for Flow calculation, in (F, P, T, t).
  - Equation that describes for Flow data values, in t.
  - Equation that describes for Pressure data values, in t.
  - Equation that describes for Temperature data values, in t.
  - Data value equations should include random noise.

## Goals
1. **Primary:** Learn the Go programming language (Golang).
2. **Secondary:** Build a functional application to simulate a flow meter.

## Technology Stack
- **Language:** Go (Latest stable)
- **Key Libraries:** 
    - `github.com/Knetic/govaluate` (Expression evaluation)
    - `github.com/go-json-experiment/json` (JSON v2)

## Coding Preferences & Conventions
- **Style:** Idiomatic Go (Effective Go standards).
- **Comments:** Provide detailed comments explaining *why* certain Go 
  features (like channels, goroutines, interfaces) are used.
- **Error Handling:** Explicit error checking (standard `if err != nil`).

## Roadmap
- [x] Initialize Go module (`go mod init`)
- [x] Create basic "Hello World" entry point
- [x] Implement core logic
    - [x] Define JSON configuration structure
    - [x] Integrate Expression Engine for Sensor Values
    - [x] Implement Filtering Logic
    - [x] Implement Output Handlers
- [x] Add unit tests
- [ ] Refactor for better structure

## Useful Commands
- Run project: `go run .`
- Test project: `go test ./...`
- Build: `go build -o flowMeter`

## Misc Requirements
- Argument parsing
    - [x] select config file.
    - [x] override temperature
    - [x] override pressure

## Exploration & Research
- **Floating Point Precision**: Evaluate the implications of using `float32` vs `float64`. 
    - *Current status*: Using `float64` for maximum compatibility with `math` and `govaluate`.
    - *Research*: Impact on precision for high-value flow references (~8M) and accumulation error.
- **Integer-Space Filtering**: Investigate implementing filters in integer/fixed-point space.
    - *FIR Filters*: Research efficient integer implementations to avoid floating point overhead.
    - *Median Filters*: Current implementation uses `int32`, verify efficiency for large windows.
    - *Scaling*: Using fixed-point arithmetic (e.g., Q16.16) for fractional logic in integer space.

