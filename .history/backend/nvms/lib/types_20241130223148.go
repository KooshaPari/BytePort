package lib

// NVMSError defines custom error types for NVMS
type NVMSError string

func (e NVMSError) Error() string {
	return string(e)
}

const (
	YamlError        NVMSError = "YAML parsing error: %s"
	MissingField     NVMSError = "Missing required field: %s"
	InvalidValueFmt  NVMSError = "Invalid value for field %s: %s"
	InvalidAWSFormat NVMSError = "Invalid AWS reference format"
)

// ResourceSize represents the size of a resource with its unit
type ResourceSize struct {
	Size uint32       `yaml:"size"`
	Unit ResourceUnit `yaml:"unit"`
}

// ResourceUnit defines the unit of the resource size
type ResourceUnit string

const (
	GB ResourceUnit = "GB"
	MB ResourceUnit = "MB"
	KB ResourceUnit = "KB"
)

// ScaleRange defines the minimum and maximum scaling limits
type ScaleRange struct {
	Min uint32 `yaml:"min"`
	Max uint32 `yaml:"max"`
}

// NVMS is the main configuration structure
type NVMS struct {
	From       string                    `yaml:"from"`
	Name       string                    `yaml:"name"`
	Services   map[string]ServiceConfig  `yaml:"services,omitempty"`
	
	AWS        *AWSConfig                `yaml:"aws,omitempty"`
	Network    *NetworkConfig            `yaml:"network,omitempty"`
	
}

// ServiceConfig defines the configuration for a single service
type ServiceConfig struct {
	Path      string                 `yaml:"path"`
	Build     BuildCommand           `yaml:"build"`
	Port      *uint16                `yaml:"port,omitempty"`
	Env       map[string]ConfigValue `yaml:"env,omitempty"`
	Resources *ResourceConfig        `yaml:"resources,omitempty"`
	Volumes  []string               `yaml:"volumes,omitempty"`
}

// BuildCommand represents either a single build command or multiple commands
type BuildCommand struct {
	Single   *string  `yaml:"single,omitempty"`
	Multiple []string `yaml:"multiple,omitempty"`
}

// ResourceConfig defines the resources allocated to a service or cluster
type ResourceConfig struct {
	CPU     uint32        `yaml:"cpu"`
	Memory  ResourceSize  `yaml:"memory"`
	Storage *ResourceSize `yaml:"storage,omitempty"`
	GPU     *bool         `yaml:"gpu,omitempty"`
}

// AWSConfig defines the AWS infrastructure configuration
type AWSConfig struct {
	Region   string        `yaml:"region"`
	Services []AWSService  `yaml:"services"`
}

// AWSService represents different AWS services configurations
type AWSService struct {
	
	/*ElastiCache *ElastiCacheConfig `yaml:"elasticache,omitempty"`
	SQS         *SQSConfig         `yaml:"sqs,omitempty"`
	S3          *S3Config          `yaml:"s3,omitempty"`
	DynamoDB    *DynamoDBConfig    `yaml:"dynamodb,omitempty"`
	OpenSearch  *OpenSearchConfig  `yaml:"opensearch,omitempty"`
	MSK         *MSKConfig         `yaml:"msk,omitempty"`
	Lambda      *LambdaConfig      `yaml:"lambda,omitempty"`
	SageMaker   *SageMakerConfig   `yaml:"sagemaker,omitempty"`*/
}

// NetworkConfig defines the network-related configurations
type NetworkConfig struct {
	Domain        *string            `yaml:"domain,omitempty"`
	SSL           bool               `yaml:"ssl"`

	Security      SecurityConfig     `yaml:"security"`
}


// SecurityConfig defines the security-related configurations
type SecurityConfig struct {
	VPC             bool            `yaml:"vpc"`
	PrivateSubnets  bool            `yaml:"private_subnets"`
	WAF             *bool           `yaml:"waf,omitempty"`
	DDOSProtection  *bool           `yaml:"ddos_protection,omitempty"`
	Rules           *SecurityRules  `yaml:"rules,omitempty"`
}

// SecurityRules defines inbound and outbound security rules
type SecurityRules struct {
	Inbound          []string `yaml:"inbound,omitempty"`
	Outbound         []string `yaml:"outbound,omitempty"`
	LatencyThreshold *string  `yaml:"latency_threshold,omitempty"`
}

// MonitoringConfig defines the monitoring configurations
type MonitoringConfig struct {
	Metrics []string      `yaml:"metrics"`
	Alerts  []AlertConfig `yaml:"alerts"`
}

// AlertConfig defines the configuration for alerts
type AlertConfig struct {
	AlertType string      `yaml:"alert_type"`
	Threshold ConfigValue `yaml:"threshold"`
	Window    string      `yaml:"window"`
}

// BackupConfig defines the backup configurations
type BackupConfig struct {
	Enabled   bool   `yaml:"enabled"`
	Retention string `yaml:"retention"`
	Schedule  string `yaml:"schedule"`
}
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
