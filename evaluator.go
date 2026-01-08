package main

import (
	"fmt"
	"math"

	"github.com/Knetic/govaluate"
)

// getFunctions returns a map of supported math functions for govaluate.
func getFunctions() map[string]govaluate.ExpressionFunction {
	return map[string]govaluate.ExpressionFunction{
		"sin": func(args ...interface{}) (interface{}, error) {
			val := args[0].(float64)
			return math.Sin(val), nil
		},
		"cos": func(args ...interface{}) (interface{}, error) {
			val := args[0].(float64)
			return math.Cos(val), nil
		},
		"tan": func(args ...interface{}) (interface{}, error) {
			val := args[0].(float64)
			return math.Tan(val), nil
		},
		"sqrt": func(args ...interface{}) (interface{}, error) {
			val := args[0].(float64)
			return math.Sqrt(val), nil
		},
	}
}

// EvaluateEquation parses and evaluates an equation with the given parameters.
func EvaluateEquation(equation string, parameters map[string]interface{}) (float64, error) {
	functions := getFunctions()
	expression, err := govaluate.NewEvaluableExpressionWithFunctions(equation, functions)
	if err != nil {
		return 0, err
	}

	result, err := expression.Evaluate(parameters)
	if err != nil {
		return 0, err
	}

	// Helper to convert result to float64 safely
	if val, ok := result.(float64); ok {
		return val, nil
	}
	return 0, fmt.Errorf("equation result is not a float64")
}
