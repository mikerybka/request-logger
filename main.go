package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/mikerybka/util"
)

// LogEntry represents the structure of the logged request
type LogEntry struct {
	Timestamp time.Time   `json:"timestamp"`
	Method    string      `json:"method"`
	Header    http.Header `json:"header"`
	Host      string      `json:"host"`
	URL       string      `json:"url"`
	Body      string      `json:"body"`
}

func (e *LogEntry) Write(dir string) error {
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return err
	}
	path := filepath.Join(dir, strconv.FormatInt(e.Timestamp.UnixNano(), 10))
	b, _ := json.MarshalIndent(e, "", "  ")
	return os.WriteFile(path, b, os.ModePerm)
}

func main() {
	logDir := util.RequireEnvVar("LOG_DIR")
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Read the request body
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		// Restore the request body for further use
		r.Body = io.NopCloser(http.MaxBytesReader(w, io.NopCloser(http.NoBody), int64(len(bodyBytes))))

		logEntry := LogEntry{
			Timestamp: time.Now(),
			Method:    r.Method,
			Header:    r.Header,
			Host:      r.Host,
			URL:       r.URL.String(),
			Body:      string(bodyBytes),
		}

		// Log the request as JSON
		err = logEntry.Write(logDir)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	addr := ":" + util.RequireEnvVar("PORT")
	log.Fatal(http.ListenAndServe(addr, nil))
}
