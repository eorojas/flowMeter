package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNetworkOutput(t *testing.T) {
	// Create a test server
	var receivedData OutputData
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		
		// Decode the body
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&receivedData); err != nil {
			t.Errorf("Failed to decode request body: %v", err)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Initialize NetworkOutput
	output := NewNetworkOutput(server.URL)

	// Send data
	data := OutputData{
		SampleNumber:   123,
		RawFlow:        5000,
		Pressure:       100,
		Temperature:    125,
		CalculatedFlow: 5000,
	}

	if err := output.Write(data); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	// Verify received data
	if receivedData.SampleNumber != 123 {
		t.Errorf("Expected SampleNumber 123, got %d", receivedData.SampleNumber)
	}
	if receivedData.CalculatedFlow != 5000 {
		t.Errorf("Expected CalculatedFlow 5000, got %d", receivedData.CalculatedFlow)
	}
}
