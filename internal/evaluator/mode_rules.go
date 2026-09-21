package evaluator

import (
	"context"
	"fmt"

	"controller/internal/collector"
	"controller/internal/config"
	"controller/internal/techniques"
)

type ModeRule struct {
	Config          config.ModeConfig
	MetricsProvider collector.Provider
}

func (rule *ModeRule) Evaluate(ctx context.Context, currentCapacity int, resourceIDs []string) (Decision, int, string, error) {
	metric := rule.Config.Metrics[0]

	samples, err := rule.MetricsProvider.GetMetric(ctx, resourceIDs, metric.Name)
	if err != nil {
		return MaintainCapacity, 0, "", err
	}

	value := applyTechniques(samples, rule.Config.Techniques)

	if techniques.ExceedsHigh(value, metric.ThresholdHigh) {
		step := techniques.ProportionalStep(currentCapacity, rule.Config.Techniques.ProportionalThresholding)
		justification := fmt.Sprintf("%s = %.2f exceeds high threshold %.2f", metric.Name, value, metric.ThresholdHigh)

		return IncreaseCapacity, step, justification, nil
	}

	if techniques.BelowLow(value, metric.ThresholdLow) {
		step := techniques.ProportionalStep(currentCapacity, rule.Config.Techniques.ProportionalThresholding)
		justification := fmt.Sprintf("%s = %.2f below low threshold %.2f", metric.Name, value, metric.ThresholdLow)

		return ReduceCapacity, step, justification, nil
	}

	justification := fmt.Sprintf("%s = %.2f within thresholds [%.2f, %.2f]", metric.Name, value, metric.ThresholdLow, metric.ThresholdHigh)

	return MaintainCapacity, 0, justification, nil
}

func applyTechniques(samples []collector.Sample, cfg techniques.Config) float64 {
	var results []float64

	if cfg.MovingAverage != nil {
		results = append(results, techniques.MovingAverage(samples, cfg.MovingAverage.Window))
	}
	if cfg.Autoregression != nil {
		results = append(results, techniques.Autoregression(samples, cfg.Autoregression.Window))
	}
	if cfg.PatternMatching != nil {
		results = append(results, techniques.PatternMatching(samples, cfg.PatternMatching.Window))
	}

	if len(results) == 0 {
		if len(samples) == 0 {
			return 0
		}
		return samples[len(samples)-1].Value
	}

	max := results[0]
	for _, v := range results[1:] {
		if v > max {
			max = v
		}
	}
	return max
}
