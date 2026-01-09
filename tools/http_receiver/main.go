package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-json-experiment/json"
)

// OutputData matches the structure sent by flowMeter
type OutputData struct {
	SampleNumber   int64 `json:"sample_number"`
	RawFlow        int32 `json:"raw_flow"`
	Pressure       int32 `json:"pressure"`
	Temperature    int32 `json:"temperature"`
	CalculatedFlow int32 `json:"calculated_flow"`
}

func main() {
	port := ":8080"
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
			return
		}

		var data OutputData
		if err := json.UnmarshalRead(r.Body, &data); err != nil {
			log.Printf("Error decoding JSON: %v", err)
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		fmt.Printf("Received: Sample=%d, Flow=%d, P=%d, T=%d, Calc=%d\n",
			data.SampleNumber, data.RawFlow, data.Pressure, data.Temperature, data.CalculatedFlow)

		w.WriteHeader(http.StatusOK)
	})

	fmt.Printf("HTTP Receiver listening on %s...\n", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal(err)
	}
}
