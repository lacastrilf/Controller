package techniques

import "controller/internal/collector"

func recentSamples(samples []collector.Sample, window int) []collector.Sample {
	if len(samples) == 0 {
		return samples
	}
	if window > len(samples) {
		window = len(samples)
	}
	return samples[len(samples)-window:]
}
