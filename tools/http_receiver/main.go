package main

import (
    "encoding/csv"
    "fmt"
    "log"
    "net/http"
    "os"
    "strconv"
    "sync"

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
    csvFile := "rcv_out.csv"

    // Create or open the CSV file
    file, err := os.OpenFile(csvFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
    if err != nil {
        log.Fatalf("Failed to open CSV file: %v", err)
    }
    defer file.Close()

    writer := csv.NewWriter(file)
    defer writer.Flush()

    // Write Header
    header := []string{"sample_number",
                       "raw_flow",
                       "pressure",
                       "temperature",
                       "calculated_flow"}
    if err := writer.Write(header); err != nil {
        log.Fatalf("Failed to write CSV header: %v", err)
    }
    writer.Flush()

    var mu sync.Mutex

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

        // Write to CSV
        mu.Lock()
        record := []string{
            strconv.FormatInt(data.SampleNumber, 10),
            strconv.FormatInt(int64(data.RawFlow), 10),
            strconv.FormatInt(int64(data.Pressure), 10),
            strconv.FormatInt(int64(data.Temperature), 10),
            strconv.FormatInt(int64(data.CalculatedFlow), 10),
        }
        if err := writer.Write(record); err != nil {
            log.Printf("Error writing to CSV: %v", err)
        }
        writer.Flush()
        mu.Unlock()

        fmt.Printf("Received & Logged: " +
            "Sample=%d, Flow=%d, P=%d, T=%d, Calc=%d\n",
            data.SampleNumber,
                   data.RawFlow,
                   data.Pressure,
                   data.Temperature,
                   data.CalculatedFlow)

        w.WriteHeader(http.StatusOK)
    })

    fmt.Printf("HTTP Receiver listening on %s (logging to %s)...\n",
               port,
               csvFile)
    if err := http.ListenAndServe(port, nil); err != nil {
        log.Fatal(err)
    }
}

