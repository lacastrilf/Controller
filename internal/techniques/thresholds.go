package techniques

func ExceedsHigh(value, thresholdHigh float64) bool {
	return value > thresholdHigh
}

func BelowLow(value, thresholdLow float64) bool {
	return value < thresholdLow
}

func ProportionalStep(currentCapacity int, percentage int) int {
	return max(1, (currentCapacity*percentage)/100)
}
