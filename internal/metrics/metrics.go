package metrics

import (
	"context"
	"time"
)

type Sample struct {
	Timestamp time.Time
	Value     float64
}

type MetricsProvider interface {
	GetMetric(ctx context.Context, resourceIDs []string, metricName string) ([]Sample, error)
}
