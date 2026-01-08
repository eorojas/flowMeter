package main

import (
	"math"
	"testing"
)

func TestSafeAdd32(t *testing.T) {
	if _, err := SafeAdd32(math.MaxInt32, 1); err != ErrOverflow {
		t.Error("Expected overflow for MaxInt32 + 1")
	}
	if v, err := SafeAdd32(100, 200); err != nil || v != 300 {
		t.Error("100 + 200 failed")
	}
}

func TestSafeMul32(t *testing.T) {
	if _, err := SafeMul32(math.MaxInt32/2+1, 2); err != ErrOverflow {
		t.Error("Expected overflow")
	}
	if v, err := SafeMul32(100, 20); err != nil || v != 2000 {
		t.Error("100 * 20 failed")
	}
}
