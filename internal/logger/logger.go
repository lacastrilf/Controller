package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"time"

	"controller/internal/evaluator"
)

type MetricObservation struct {
	Name  string  `json:"name"`
	Value float64 `json:"value"`
}
type Record struct {
	Timestamp        time.Time           `json:"timestamp"`
	Mode             string              `json:"mode"`
	Metrics          []MetricObservation `json:"metrics"`
	ExistingCapacity int                 `json:"existing_capacity"`
	Decision         evaluator.Decision  `json:"decision"`
	Justification    string              `json:"justification"`
	ActionResult     string              `json:"action_result"`
}

// Logger writes Records to one or more destinations (e.g. the terminal
// and a file), safely across multiple goroutines.
type Logger struct {
	mu      sync.Mutex
	console io.Writer
	file    io.Writer
}

// New creates a Logger that writes human-readable lines to `console`
// and JSON Lines to `file`.
func New(console io.Writer, file io.Writer) *Logger {
	return &Logger{console: console, file: file}
}

// Log writes a Record to both destinations. It's safe to call from
// multiple goroutines at once.
func (l *Logger) Log(r Record) {
	l.mu.Lock()
	defer l.mu.Unlock()

	fmt.Fprintf(l.console, "[%s] mode=%s | %s | capacity=%d | %s | %s\n",
		r.Timestamp.Format(time.RFC3339), r.Mode, r.Decision, r.ExistingCapacity,
		r.Justification, r.ActionResult)

	if l.file != nil {
		data, err := json.Marshal(r)
		if err == nil {
			fmt.Fprintln(l.file, string(data))
		}
	}
}
