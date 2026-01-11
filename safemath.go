package main

import (
    "errors"
    "math"
)

var ErrOverflow = errors.New("integer overflow")

// SafeAdd32 adds two int32 numbers and checks for overflow.
func SafeAdd32(a, b int32) (int32, error) {
    if b > 0 {
        if a > math.MaxInt32-b {
            return 0, ErrOverflow
        }
    } else {
        if a < math.MinInt32-b {
            return 0, ErrOverflow
        }
    }
    return a + b, nil
}

// SafeSub32 subtracts two int32 numbers and checks for overflow.
func SafeSub32(a, b int32) (int32, error) {
    if b > 0 {
        if a < math.MinInt32+b {
            return 0, ErrOverflow
        }
    } else {
        if a > math.MaxInt32+b {
            return 0, ErrOverflow
        }
    }
    return a - b, nil
}

// SafeMul32 multiplies two int32 numbers and checks for overflow.
func SafeMul32(a, b int32) (int32, error) {
    if a == 0 || b == 0 {
        return 0, nil
    }
    result := int64(a) * int64(b)
    if result > math.MaxInt32 || result < math.MinInt32 {
        return 0, ErrOverflow
    }
    return int32(result), nil
}

// SafeDiv32 divides two int32 numbers and checks for /
// overflow (only for MinInt32 / -1).
func SafeDiv32(a, b int32) (int32, error) {
    if b == 0 {
        return 0, errors.New("division by zero")
    }
    if a == math.MinInt32 && b == -1 {
        return 0, ErrOverflow
    }
    return a / b, nil
}

