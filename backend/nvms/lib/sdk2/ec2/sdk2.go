// Package ec2 — aws-sdk-go-v2 backed RunInstances wrapper (PoC).
//
// This file is a proof-of-concept showing how to swap the hand-rolled
// HTTP+SigV4 client for the official
// `aws-sdk-go-v2/service/ec2`. It deliberately does NOT depend on the
// Fermyon Spin SDK (`spinhttp.Send`) so it can be unit-tested with the
// standard Go toolchain (the Spin SDK currently breaks on Go ≥1.25 due
// to CGo type incompatibilities — see worklog/2026-07-06-byteport-poc.md).
//
// Path: `backend/nvms/lib/sdk2/ec2/`. Lives outside the existing
// `awspin/` tree because that tree transitively imports Spin (via
// `awspin/s3` and `awspin/network`). The SDK-v2 wrapper has no Spin
// dependency and can be built/tested with vanilla `go test ./...`.
//
// Usage:
//
//	awsCfg, _ := sdk2ec2.LoadSDKConfigFromEnv(ctx)
//	client := ec2.NewFromConfig(awsCfg)
//	out, err := sdk2ec2.RunInstances(ctx, client, sdk2ec2.RunInstancesParams{
//	    AMI:          "ami-0abcdef1234567890",
//	    InstanceType: "t3.micro",
//	    SubnetID:     "subnet-0123",
//	    Name:         "byteport-svc",
//	})
//
// To switch the existing awspin path to use this wrapper, set
// `USE_SDK_V2=true` in the Spin app config and route `DeployEC2` to
// `sdk2ec2.RunInstances` in `aws.go`. The hand-rolled client is kept as
// a fallback (LocalStack path uses `awspin` query protocol, SDK-v2 uses
// JSON — they cannot be used interchangeably without testing the
// endpoint compatibility).
package ec2

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// RunInstancesSDK2Params is the SDK-v2 wrapper input. It mirrors the
// subset of EC2 RunInstances arguments that BytePort actually uses:
// AMI, instance type, subnet, count, and a single Name tag. The full
// AWS API surface (IAM instance profile, security groups, user data,
// EBS, etc.) is intentionally NOT modeled here — callers should
// compose extended wrappers as needed.
type RunInstancesSDK2Params struct {
	AMI          string
	InstanceType string
	SubnetID     string
	MinCount     int32
	MaxCount     int32
	Name         string
	// IAMInstanceProfile is the EC2 IamInstanceProfileSpecification.ARN
	// value. Setting this is the secure replacement for embedding AWS
	// access keys in EC2 UserData.
	IAMInstanceProfile string
	// UserData is base64-encoded bootstrap script for the instance.
	UserData string
}

// RunInstancesSDK2Result is the SDK-v2 wrapper output. It extracts the
// fields BytePort stores in EC2InstanceInfo so callers don't need to
// walk the SDK's nested types.
type RunInstancesSDK2Result struct {
	InstanceID string
	State      string
	PrivateIP  string
	PublicIP   string
	Region     string
}

// RunInstancesSDK2 launches one or more EC2 instances using the official
// aws-sdk-go-v2 client. It is the SDK-v2 equivalent of the existing
// `(*Client).RunInstances` hand-rolled method.
func RunInstancesSDK2(ctx context.Context, client *ec2.Client, params RunInstancesSDK2Params) (*RunInstancesSDK2Result, error) {
	if client == nil {
		return nil, errors.New("ec2 client is nil")
	}
	if params.AMI == "" {
		return nil, errors.New("AMI is required")
	}
	if params.InstanceType == "" {
		return nil, errors.New("instance type is required")
	}

	input := &ec2.RunInstancesInput{
		ImageId:      aws.String(params.AMI),
		InstanceType: types.InstanceType(params.InstanceType),
		MinCount:     aws.Int32(defaultInt(params.MinCount, 1)),
		MaxCount:     aws.Int32(defaultInt(params.MaxCount, 1)),
	}

	if params.SubnetID != "" {
		input.SubnetId = aws.String(params.SubnetID)
	}

	if params.IAMInstanceProfile != "" {
		input.IamInstanceProfile = &types.IamInstanceProfileSpecification{
			Arn: aws.String(params.IAMInstanceProfile),
		}
	}

	if params.UserData != "" {
		input.UserData = aws.String(params.UserData)
	}

	if params.Name != "" {
		input.TagSpecifications = []types.TagSpecification{
			{
				ResourceType: types.ResourceTypeInstance,
				Tags: []types.Tag{
					{Key: aws.String("Name"), Value: aws.String(params.Name)},
					{Key: aws.String("ManagedBy"), Value: aws.String("byteport")},
				},
			},
		}
	}

	out, err := client.RunInstances(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("ec2 RunInstances: %w", err)
	}
	if out == nil || len(out.Instances) == 0 {
		return nil, errors.New("ec2 RunInstances returned no instances")
	}

	inst := out.Instances[0]
	res := &RunInstancesSDK2Result{}
	if inst.InstanceId != nil {
		res.InstanceID = *inst.InstanceId
	}
	if inst.State != nil && inst.State.Name != "" {
		res.State = string(inst.State.Name)
	}
	if inst.PrivateIpAddress != nil {
		res.PrivateIP = *inst.PrivateIpAddress
	}
	if inst.PublicIpAddress != nil {
		res.PublicIP = *inst.PublicIpAddress
	}
	return res, nil
}

// TerminateInstancesSDK2 terminates EC2 instances via the official SDK.
// Mirrors `(*Client).TerminateInstances` but with the SDK-v2 idiomatic
// input struct + error handling.
func TerminateInstancesSDK2(ctx context.Context, client *ec2.Client, instanceIDs []string) error {
	if client == nil {
		return errors.New("ec2 client is nil")
	}
	if len(instanceIDs) == 0 {
		return errors.New("at least one instance ID is required")
	}

	ids := make([]string, len(instanceIDs))
	copy(ids, instanceIDs)

	_, err := client.TerminateInstances(ctx, &ec2.TerminateInstancesInput{
		InstanceIds: ids,
	})
	if err != nil {
		return fmt.Errorf("ec2 TerminateInstances: %w", err)
	}
	return nil
}

func defaultInt(v, fallback int32) int32 {
	if v <= 0 {
		return fallback
	}
	return v
}
