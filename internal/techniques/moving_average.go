package techniques

import "controller/internal/metrics"

func MovingAverage(samples []metrics.Sample, window int) float64 {
	recent := recentSamples(samples, window)

	sum := 0.0
	for _, s := range recent {
		sum += s.Value
	}
	return sum / float64(window)
}
