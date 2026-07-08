package main

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// This file bridges between handlers.go's DeployRequest and the SDK-v2 EC2
// client implemented in nvms/lib/sdk2/ec2. Until awspin is fully retired the
// nvms path will keep its own adapter; this one drives the main backend HTTP
// path so a /deploy request with provider="aws" actually provisions an
// instance instead of sleeping 3 seconds.

const (
	// awsDefaultAMI is the official BytePort base AMI. In dev/localstack
	// the value is ignored — aws-sdk-go-v2 just needs any valid ami id.
	awsDefaultAMI = "ami-0c55b159cbfafe1f0"

	// awsDefaultInstanceType is t3.micro which is free-tier eligible.
	awsDefaultInstanceType = "t3.micro"

	// awsDefaultRegion matches the default in loadAWSConfigFromEnv.
	awsDefaultRegion = "us-east-1"
)

// buildEC2InputFromDeploy converts an HTTP DeployRequest into an SDK-v2
// RunInstancesInput. Validation here means a bad request fails fast at the
// HTTP boundary (handlers.go: bind+validate) rather than reaching AWS.
func buildEC2InputFromDeploy(req DeployRequest) ec2.RunInstancesInput {
	// name maps to a single Tag; we strip shell metacharacters since this
	// flows through EC2 user-data and tag values eventually.
	safeName := sanitizeName(req.Name)

	cfg := req.Config
	ami := envOr("AWS_AMI_ID", awsDefaultAMI)
	itype := envOr("AWS_INSTANCE_TYPE", awsDefaultInstanceType)
	subnet := envOr("AWS_SUBNET_ID", "")
	sg := envOr("AWS_SECURITY_GROUP_ID", "")
	iamProfile := envOr("IAM_INSTANCE_PROFILE", "")
	keyName := envOr("AWS_KEY_NAME", "")

	// Optional overrides from req.Config — callers can pass config.ami_id,
	// config.subnet_id, etc. without needing env vars.
	if cfg != nil {
		if v, ok := cfg["ami_id"].(string); ok && v != "" {
			ami = v
		}
		if v, ok := cfg["instance_type"].(string); ok && v != "" {
			itype = v
		}
		if v, ok := cfg["subnet_id"].(string); ok && v != "" {
			subnet = v
		}
		if v, ok := cfg["security_group_id"].(string); ok && v != "" {
			sg = v
		}
		if v, ok := cfg["iam_instance_profile"].(string); ok && v != "" {
			iamProfile = v
		}
		if v, ok := cfg["key_name"].(string); ok && v != "" {
			keyName = v
		}
	}

	input := ec2.RunInstancesInput{
		ImageId:      stringPtr(ami),
		InstanceType: awstypes.InstanceType(itype),
		MinCount:     int32Ptr(1),
		MaxCount:     int32Ptr(1),
		TagSpecifications: []awstypes.TagSpecification{
			{
				ResourceType: awstypes.ResourceTypeInstance,
				Tags: []awstypes.Tag{
					{Key: stringPtr("Name"), Value: stringPtr(safeName)},
					{Key: stringPtr("byteport:owner"), Value: stringPtr(envOr("USER", "anonymous"))},
				},
			},
		},
	}

	if subnet != "" {
		input.SubnetId = stringPtr(subnet)
	}
	if sg != "" {
		input.SecurityGroupIds = []string{sg}
	}
	if keyName != "" {
		input.KeyName = stringPtr(keyName)
	}
	if iamProfile != "" {
		input.IamInstanceProfile = &awstypes.IamInstanceProfileSpecification{
			Name: stringPtr(iamProfile),
		}
	}

	return input
}

// loadAWSConfigFromEnv reads AWS_* env vars and returns a configured SDK-v2
// config. Setting AWS_ENDPOINT_URL switches the SDK into LocalStack mode.
//
// This is intentionally a thin wrapper — heavier config (assume role, MFA,
// SSO) lives in the nvms layer when needed.
type awsConfig struct {
	Region    string
	Endpoint  string
	AccessKey string
	SecretKey string
}

func loadAWSConfigFromEnv() (awsConfig, error) {
	cfg := awsConfig{
		Region:    envOr("AWS_REGION", awsDefaultRegion),
		Endpoint:  os.Getenv("AWS_ENDPOINT_URL"),
		AccessKey: os.Getenv("AWS_ACCESS_KEY_ID"),
		SecretKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
	}
	if cfg.Region == "" {
		return cfg, fmt.Errorf("AWS_REGION not set")
	}
	return cfg, nil
}

// runEC2Instance issues an SDK-v2 RunInstances call against the configured
// endpoint. The wrapper exists so future changes (retry policies, metrics,
// tracing) have one place to live.
func runEC2Instance(ctx context.Context, cfg awsConfig, input ec2.RunInstancesInput) (string, error) {
	// The nvms layer has the heavy SDK client. Until main backend depends on
	// the nvms package we use aws-sdk-go-v2 directly here.
	client, err := newEC2Client(ctx, cfg)
	if err != nil {
		return "", err
	}
	out, err := client.RunInstances(ctx, &input)
	if err != nil {
		return "", err
	}
	if out == nil || len(out.Instances) == 0 || out.Instances[0].InstanceId == nil {
		return "", fmt.Errorf("aws returned no instance id")
	}
	return *out.Instances[0].InstanceId, nil
}

// newEC2Client builds an SDK-v2 EC2 client honoring AWS_ENDPOINT_URL for
// LocalStack. Kept as its own function so tests can swap it for a fake.
var newEC2Client = func(_ context.Context, cfg awsConfig) (ec2Client, error) {
	return realEC2Client(cfg)
}

type ec2Client interface {
	RunInstances(ctx context.Context, params *ec2.RunInstancesInput, optFns ...func(*ec2.Options)) (*ec2.RunInstancesOutput, error)
}

// sanitizeName strips characters that could be confusing in EC2 tag values
// or downstream user-data. Allows [a-zA-Z0-9-_], replaces everything else
// with '-', trims length to 64 chars.
func sanitizeName(s string) string {
	out := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c >= 'a' && c <= 'z',
			c >= 'A' && c <= 'Z',
			c >= '0' && c <= '9',
			c == '-' || c == '_':
			out = append(out, c)
		default:
			out = append(out, '-')
		}
	}
	res := string(out)
	if len(res) > 64 {
		res = res[:64]
	}
	return res
}

func stringPtr(s string) *string { return &s }
func int32Ptr(n int32) *int32    { return &n }
