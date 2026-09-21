package techniques

import "controller/internal/collector"

func Autoregression(samples []collector.Sample, window int) float64 {
	recent := recentSamples(samples, window)
	n := len(recent)

	if n < 2 {
		return recent[n-1].Value
	}

	var sumX, sumY, sumXY, sumX2 float64
	for i, s := range recent {
		x := float64(i)
		y := s.Value

		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}

	nf := float64(n)
	slope := (nf*sumXY - sumX*sumY) / (nf*sumX2 - sumX*sumX)
	intercept := (sumY - slope*sumX) / nf

	nextX := float64(n)
	return slope*nextX + intercept
}
