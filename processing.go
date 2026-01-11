package main

import (
	"sort"
	"strings"
)

// Filter defines the interface for data filters using int32.
type Filter interface {
	Process(value int32) int32
	Initialize(value int32) // Pre-populates the filter with a default value
}

// LowPassFilter implements an Exponential Moving Average (EMA) filter using integer math.
// It uses fixed-point arithmetic (scaled by 1024) to handle the alpha factor.
type LowPassFilter struct {
	AlphaScaled int32 // Alpha * 1024
	PrevValue   int32
	Initialized bool
}

func NewLowPassFilter(alpha float64) *LowPassFilter {
	// Clamp alpha between 0.0 and 1.0
	if alpha < 0 {
		alpha = 0
	}
	if alpha > 1 {
		alpha = 1
	}
	return &LowPassFilter{
		AlphaScaled: int32(alpha * 1024),
	}
}

func (f *LowPassFilter) Initialize(value int32) {
	f.PrevValue = value
	f.Initialized = true
}

func (f *LowPassFilter) Process(value int32) int32 {
	if !f.Initialized {
		f.PrevValue = value
		f.Initialized = true
		return value
	}
	// y[n] = y[n-1] + alpha * (x[n] - y[n-1])
	// Integer math: y[n] = y[n-1] + ((x[n] - y[n-1]) * alphaScaled) / 1024
	
	diff := value - f.PrevValue
	adjustment := (diff * f.AlphaScaled) / 1024
	
	f.PrevValue = f.PrevValue + adjustment
	return f.PrevValue
}

// MedianFilter implements a sliding window median filter using a maintained sorted slice.
// It uses a ring buffer to track insertion order and a sorted buffer for O(1) median retrieval.
// Complexity: O(N) per insert due to slice shifting (efficient for typical window sizes).
type MedianFilter struct {
	windowSize int
	ringBuffer []int32 // Stores data in chronological order
	sorted     []int32 // Stores data in numerical order
	head       int     // Current index in ringBuffer
	count      int     // Current number of elements
}

func NewMedianFilter(windowSize int) *MedianFilter {
	if windowSize <= 0 {
		windowSize = 5 // Default value
	}
	return &MedianFilter{
		windowSize: windowSize,
		ringBuffer: make([]int32, windowSize),
		sorted:     make([]int32, 0, windowSize),
		head:       0,
		count:      0,
	}
}

func (f *MedianFilter) Initialize(value int32) {
	for i := 0; i < f.windowSize; i++ {
		f.ringBuffer[i] = value
	}
	// Reset sorted slice and fill with copies of value
	f.sorted = make([]int32, f.windowSize)
	for i := 0; i < f.windowSize; i++ {
		f.sorted[i] = value
	}
	f.head = 0
	f.count = f.windowSize
}

func (f *MedianFilter) Process(value int32) int32 {
	// 1. If full, remove the oldest value from the sorted slice
	if f.count == f.windowSize {
		oldestValue := f.ringBuffer[f.head]
		
		// Binary search to find the index of the oldest value in 'sorted'
		// sort.Search returns the smallest index i such that f(i) is true.
		idx := sort.Search(len(f.sorted), func(i int) bool { return f.sorted[i] >= oldestValue })
		
		// Standard Binary Search finds the *first* occurrence. 
		// We verify we found it (we theoretically always should).
		if idx < len(f.sorted) && f.sorted[idx] == oldestValue {
			// Delete element at idx: copy(sorted[i:], sorted[i+1:])
			copy(f.sorted[idx:], f.sorted[idx+1:])
			f.sorted = f.sorted[:len(f.sorted)-1]
		}
	} else {
		f.count++
	}

	// 2. Add the new value to the ring buffer (overwrite old)
	f.ringBuffer[f.head] = value
	f.head = (f.head + 1) % f.windowSize

	// 3. Insert new value into sorted slice while maintaining order
	// Find insertion point
	insertIdx := sort.Search(len(f.sorted), func(i int) bool { return f.sorted[i] >= value })
	
	// Extend slice
	f.sorted = append(f.sorted, 0)
	// Shift elements right to make room
	copy(f.sorted[insertIdx+1:], f.sorted[insertIdx:])
	// Insert
	f.sorted[insertIdx] = value

	// 4. Return Median
	n := len(f.sorted)
	if n == 0 {
		return 0
	}
	mid := n / 2
	if n%2 == 0 {
		// Even: Average of two middle elements
		return (f.sorted[mid-1] + f.sorted[mid]) / 2
	}
	// Odd: Middle element
	return f.sorted[mid]
}

// Processor maintains the state of the sensors and applies filters.
type Processor struct {
	// Latest filtered values
	LatestPressure    int32
	LatestTemperature int32

	// Filters for each sensor type
	PressureFilters    []Filter
	TemperatureFilters []Filter
	FlowFilters        []Filter
}

// NewProcessor creates a Processor and initializes filters based on config.
func NewProcessor(config ProcessingConfig) *Processor {
	p := &Processor{
		PressureFilters:    []Filter{},
		TemperatureFilters: []Filter{},
		FlowFilters:        []Filter{},
	}

	for _, fc := range config.Filters {
		var f Filter
		switch fc.Type {
		case "low_pass":
			f = NewLowPassFilter(fc.Alpha)
		case "median":
			winSize := fc.WindowSize
			if winSize <= 0 {
				winSize = 5 // Default if not specified
			}
			f = NewMedianFilter(winSize)
		default:
			continue
		}

		switch strings.ToLower(fc.Target) {
		case "pressure":
			p.PressureFilters = append(p.PressureFilters, f)
		case "temperature":
			p.TemperatureFilters = append(p.TemperatureFilters, f)
		case "flow":
			p.FlowFilters = append(p.FlowFilters, f)
		}
	}
	return p
}

// InitializeFilters pre-populates all filters with reference values.
func (p *Processor) InitializeFilters(flowRef, pressRef, tempRef int32) {
	for _, f := range p.FlowFilters {
		f.Initialize(flowRef)
	}
	for _, f := range p.PressureFilters {
		f.Initialize(pressRef)
	}
	for _, f := range p.TemperatureFilters {
		f.Initialize(tempRef)
	}
}

// UpdatePressure processes a raw pressure value through filters and updates state.
func (p *Processor) UpdatePressure(raw int32) {
	val := raw
	for _, f := range p.PressureFilters {
		val = f.Process(val)
	}
	p.LatestPressure = val
}

// UpdateTemperature processes a raw temperature value through filters and updates state.
func (p *Processor) UpdateTemperature(raw int32) {
	val := raw
	for _, f := range p.TemperatureFilters {
		val = f.Process(val)
	}
	p.LatestTemperature = val
}

// ProcessFlow processes a raw flow value through filters and returns the filtered flow.
func (p *Processor) ProcessFlow(raw int32) int32 {
	val := raw
	for _, f := range p.FlowFilters {
		val = f.Process(val)
	}
	return val
}

// CalculateFlow computes the final flow rate using the configured equation.
// It uses the latest filtered pressure and temperature values from the processor state.
//
// Assumptions:
// 1. Input sensors (Flow, Pressure, Temperature) are within their configured bit-depths.
// 2. The provided equation results in a value that fits within int32 range.
// 3. Intermediate floating-point calculations in the expression engine are used to handle
//    scaling (e.g. / 255.0) but the final result is cast to int32 with overflow checking.
func (p *Processor) CalculateFlow(equation string, rawFlow int32, timeSecs float64, refFlow int32, refPressure int32, refTemperature int32) (int32, error) {
	// Filter the raw flow first
	filteredFlow := p.ProcessFlow(rawFlow)

	// Prepare parameters
	// We pass values as float64 to the engine to support division scaling (e.g. / 255.0)
	params := map[string]interface{}{
		"flow":        float64(filteredFlow),
		"pressure":    float64(p.LatestPressure),
		"temperature": float64(p.LatestTemperature),
		"t":           timeSecs,
		// Short aliases
		"F": float64(filteredFlow),
		"P": float64(p.LatestPressure),
		"T": float64(p.LatestTemperature),
		// Reference values
		"RefF": float64(refFlow),
		"RefP": float64(refPressure),
		"RefT": float64(refTemperature),
	}

	resultFloat, err := EvaluateEquation(equation, params)
	if err != nil {
		return 0, err
	}

	// Explicit Overflow Check for int32
	// MaxInt32 = 2147483647
	// MinInt32 = -2147483648
	if resultFloat > 2147483647 || resultFloat < -2147483648 {
		return 0, ErrOverflow
	}

	return int32(resultFloat), nil
}
