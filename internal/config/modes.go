package config

import (
	"controller/internal/techniques"
)

type ModesConfig struct {
	Modes map[string]ModeConfig `yaml:"modes"`
}

type ModeConfig struct {
	Tier            string            `yaml:"tier"`
	Metrics         []MetricConfig    `yaml:"metrics"`
	IntervalSeconds int               `yaml:"interval_seconds"`
	MinCapacity     int               `yaml:"min_capacity"`
	MaxCapacity     int               `yaml:"max_capacity"`
	Techniques      techniques.Config `yaml:"techniques"`

	// CooldownSeconds is how long the runner waits after a scaling action
	// before it is allowed to scale again, so freshly-launched instances
	// (which report no CloudWatch data for a few minutes) don't get read
	// as "0" and immediately trigger a reversal. Defaults to 2x the
	// interval when left at 0 — see main.go.
	CooldownSeconds int `yaml:"cooldown_seconds,omitempty"`
}

type MetricConfig struct {
	Name          string  `yaml:"name"`
	ThresholdHigh float64 `yaml:"threshold_high"`
	ThresholdLow  float64 `yaml:"threshold_low"`
}
