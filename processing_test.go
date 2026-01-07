package main

import "testing"

func TestLowPassFilter(t *testing.T) {
	// Alpha = 0.5
	filter := NewLowPassFilter(0.5)

	// Initial value
	val := filter.Process(100)
	if val != 100 {
		t.Errorf("First value should be 100, got %d", val)
	}

	// Next value 200.
	// AlphaScaled = 0.5 * 1024 = 512
	// Diff = 200 - 100 = 100
	// Adjustment = (100 * 512) / 1024 = 50
	// Result = 100 + 50 = 150
	val = filter.Process(200)
	if val != 150 {
		t.Errorf("Second value should be 150, got %d", val)
	}

	// Next value 150.
	// Diff = 150 - 150 = 0
	// Result = 150
	val = filter.Process(150)
	if val != 150 {
		t.Errorf("Third value should be 150, got %d", val)
	}
}

func TestMedianFilter(t *testing.T) {
	filter := NewMedianFilter(3)

	// [10]
	if v := filter.Process(10); v != 10 {
		t.Errorf("Expected 10, got %d", v)
	}

	// [10, 50] -> sort [10, 50] -> avg(10, 50) = 30
	if v := filter.Process(50); v != 30 {
		t.Errorf("Expected 30, got %d", v)
	}

	// [10, 50, 20] -> sort [10, 20, 50] -> mid 20
	if v := filter.Process(20); v != 20 {
		t.Errorf("Expected 20, got %d", v)
	}

	// Window slides: [50, 20, 100] -> sort [20, 50, 100] -> mid 50
	if v := filter.Process(100); v != 50 {
		t.Errorf("Expected 50, got %d", v)
	}
}
