package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"github.com/go-json-experiment/json"
	"net/http"
	"os"
	"strconv"
)

// OutputData represents the final calculated packet to be sent to receivers.
type OutputData struct {
	SampleNumber   int64 `json:"sample_number"`
	RawFlow        int32 `json:"raw_flow"`
	Pressure       int32 `json:"pressure"`
	Temperature    int32 `json:"temperature"`
	CalculatedFlow int32 `json:"calculated_flow"`
}

// OutputHandler defines the interface for different output destinations.
type OutputHandler interface {
	Write(data OutputData) error
	Close() error
}

// FileOutput implements OutputHandler for CSV file storage.
type FileOutput struct {
	file   *os.File
	writer *csv.Writer
}

func NewFileOutput(filename string) (*FileOutput, error) {
	file, err := os.Create(filename)
	if err != nil {
		return nil, err
	}

	writer := csv.NewWriter(file)
	// Write CSV Header
	header := []string{"sample_number", "raw_flow", "pressure", "temperature", "calculated_flow"}
	if err := writer.Write(header); err != nil {
		file.Close()
		return nil, err
	}
	writer.Flush()

	return &FileOutput{file: file, writer: writer}, nil
}

func (f *FileOutput) Write(data OutputData) error {
	record := []string{
		strconv.FormatInt(data.SampleNumber, 10),
		strconv.FormatInt(int64(data.RawFlow), 10),
		strconv.FormatInt(int64(data.Pressure), 10),
		strconv.FormatInt(int64(data.Temperature), 10),
		strconv.FormatInt(int64(data.CalculatedFlow), 10),
	}
	if err := f.writer.Write(record); err != nil {
		return err
	}
	f.writer.Flush()
	return nil
}

func (f *FileOutput) Close() error {
	f.writer.Flush()
	return f.file.Close()
}

// ConsoleOutput implements OutputHandler for stdout printing.
type ConsoleOutput struct{}

func NewConsoleOutput() *ConsoleOutput {
	return &ConsoleOutput{}
}

func (c *ConsoleOutput) Write(data OutputData) error {
	fmt.Printf("[%8d] Flow: %8d | P: %3d | T: %3d | Calc: %d\n",
		data.SampleNumber,
		data.RawFlow,
		data.Pressure,
		data.Temperature,
		data.CalculatedFlow)
	return nil
}

func (c *ConsoleOutput) Close() error {
	return nil
}

// NetworkOutput implements OutputHandler for HTTP POST requests.
type NetworkOutput struct {
	TargetURL string
	Client    *http.Client
}

func NewNetworkOutput(url string) *NetworkOutput {
	return &NetworkOutput{
		TargetURL: url,
		Client:    &http.Client{},
	}
}

func (n *NetworkOutput) Write(data OutputData) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	resp, err := n.Client.Post(n.TargetURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("server returned error status: %d", resp.StatusCode)
	}

	return nil
}

func (n *NetworkOutput) Close() error {
	// HTTP client doesn't need explicit closing
	return nil
}

// GetOutputHandler is a factory function to create the configured handler.
func GetOutputHandler(config OutputConfig) (OutputHandler, error) {
	switch config.Type {
	case "file":
		return NewFileOutput(config.Target)
	case "console":
		return NewConsoleOutput(), nil
	case "network":
		return NewNetworkOutput(config.Target), nil
	default:
		return NewConsoleOutput(), nil
	}
}