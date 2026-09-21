package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"time"
)

type Decision int

const (
	MaintainCapacity Decision = iota
	IncreaseCapacity
	ReduceCapacity
)

func (d Decision) String() string {
	switch d {
	case IncreaseCapacity:
		return "INCREASE_CAPACITY"
	case ReduceCapacity:
		return "REDUCE_CAPACITY"
	default:
		return "MAINTAIN_CAPACITY"
	}
}

type MetricObservation struct {
	Name  string  `json:"name"`
	Value float64 `json:"value"`
}

type Record struct {
	Timestamp        time.Time           `json:"timestamp"`
	Mode             string              `json:"mode"`
	Metrics          []MetricObservation `json:"metrics"`
	ExistingCapacity int                 `json:"existing_capacity"`
	Decision         Decision            `json:"decision"`
	Justification    string              `json:"justification"`
	ActionResult     string              `json:"action_result"`
}

type Logger struct {
	mu      sync.Mutex
	console io.Writer
	file    io.Writer
}

func New(console io.Writer, file io.Writer) *Logger {
	return &Logger{console: console, file: file}
}

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
