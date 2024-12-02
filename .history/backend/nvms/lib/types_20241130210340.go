// ScaleConfig defines the scaling configurations
type ScaleConfig struct {
	Min             uint32   `yaml:"min"`
	Max             uint32   `yaml:"max"`
	CPUThreshold    *float32 `yaml:"cpu_threshold,omitempty"`
	MemoryThreshold *float32 `yaml:"memory_threshold,omitempty"`
	QueueThreshold  *uint32  `yaml:"queue_threshold,omitempty"`
}

// ConfigValue represents a configuration value which can be a string, number, environment reference, or AWS reference
type ConfigValue struct {
	// Only one of the following fields should be non-nil
	String   *string
	Number   *float64
	EnvRef   *string
	AWSRef   *AWSRef
}

type AWSRef struct {
	Service  string
	Resource string
}
