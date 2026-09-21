package techniques

type MovingAverageParams struct {
	Window int `yaml:"window"`
}

type AutoregressionParams struct {
	Lag int `yaml:"lag"`
}

type PatternMatchingParams struct {
	Window int `yaml:"window"`
}

type Config struct {
	MovingAverage            *MovingAverageParams   `yaml:"moving_average,omitempty"`
	Autoregression           *AutoregressionParams  `yaml:"autoregression,omitempty"`
	PatternMatching          *PatternMatchingParams `yaml:"pattern_matching,omitempty"`
	ProportionalThresholding bool                   `yaml:"proportional_thresholding,omitempty"`
}
