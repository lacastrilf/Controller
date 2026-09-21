package techniques

import "controller/internal/metrics"

func PatternMatching(samples []metrics.Sample, window int) float64 {
	if len(samples) < window*2+1 {
		return recentAverage(samples, window)
	}

	current := recentSamples(samples, window)

	bestDifference := -1.0 // -1 means "no candidate compared yet"
	bestNextValue := current[len(current)-1].Value

	for start := 0; start+window < len(samples)-window; start++ {
		candidate := samples[start : start+window]

		difference := totalDifference(current, candidate)
		if bestDifference < 0 || difference < bestDifference {
			bestDifference = difference
			bestNextValue = samples[start+window].Value
		}
	}

	return bestNextValue
}

func totalDifference(a, b []metrics.Sample) float64 {
	total := 0.0
	for i := range a {
		diff := a[i].Value - b[i].Value
		if diff < 0 {
			diff = -diff
		}
		total += diff
	}
	return total
}

func recentAverage(samples []metrics.Sample, window int) float64 {
	recent := recentSamples(samples, window)
	if len(recent) == 0 {
		return 0
	}
	sum := 0.0
	for _, s := range recent {
		sum += s.Value
	}
	return sum / float64(len(recent))
}
