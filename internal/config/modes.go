package config

import (
	"controller/internal/techniques"
)

type ModesConfig struct {
	Modes map[string]ModeConfig `yaml:"modes"`
}

type ModeConfig struct {
	Metrics         []MetricConfig    `yaml:"metrics"`
	IntervalSeconds int               `yaml:"interval_seconds"`
	MinCapacity     int               `yaml:"min_capacity"`
	MaxCapacity     int               `yaml:"max_capacity"`
	Techniques      techniques.Config `yaml:"techniques"`
}

type MetricConfig struct {
	Name          string  `yaml:"name"`
	ThresholdHigh float64 `yaml:"threshold_high"`
	ThresholdLow  float64 `yaml:"threshold_low"`
}
