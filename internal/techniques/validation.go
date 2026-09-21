package techniques

import "controller/internal/metrics"

func recentSamples(samples []metrics.Sample, window int) []metrics.Sample {
	if len(samples) == 0 {
		return samples
	}
	if window > len(samples) {
		window = len(samples)
	}
	return samples[len(samples)-window:]
}
