// Package ec2 — PoC tests for sdk2.go.
//
// These tests cover input validation only. The actual SDK-v2 call is
// exercised in integration tests run against LocalStack (out of scope
// for this PoC). The validation tests prove the wrapper is wired
// correctly without spinning up a network or mock HTTP server.
package ec2

import (
	"context"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

func TestRunInstancesSDK2_NilClient(t *testing.T) {
	_, err := RunInstancesSDK2(context.Background(), nil, RunInstancesSDK2Params{
		AMI:          "ami-test",
		InstanceType: "t3.micro",
	})
	if err == nil {
		t.Fatal("expected nil-client error, got nil")
	}
	if !strings.Contains(err.Error(), "client is nil") {
		t.Fatalf("expected 'client is nil' in error, got: %v", err)
	}
}

func TestRunInstancesSDK2_MissingAMI(t *testing.T) {
	_, err := RunInstancesSDK2(context.Background(), &ec2.Client{}, RunInstancesSDK2Params{
		InstanceType: "t3.micro",
	})
	if err == nil {
		t.Fatal("expected missing-AMI error, got nil")
	}
	if !strings.Contains(err.Error(), "AMI is required") {
		t.Fatalf("expected 'AMI is required' in error, got: %v", err)
	}
}

func TestRunInstancesSDK2_MissingInstanceType(t *testing.T) {
	_, err := RunInstancesSDK2(context.Background(), &ec2.Client{}, RunInstancesSDK2Params{
		AMI: "ami-test",
	})
	if err == nil {
		t.Fatal("expected missing-instance-type error, got nil")
	}
	if !strings.Contains(err.Error(), "instance type is required") {
		t.Fatalf("expected 'instance type is required' in error, got: %v", err)
	}
}

func TestDefaultInt(t *testing.T) {
	if defaultInt(0, 1) != 1 {
		t.Fatalf("defaultInt(0, 1) = %d, want 1", defaultInt(0, 1))
	}
	if defaultInt(5, 1) != 5 {
		t.Fatalf("defaultInt(5, 1) = %d, want 5", defaultInt(5, 1))
	}
	if defaultInt(-3, 1) != 1 {
		t.Fatalf("defaultInt(-3, 1) = %d, want 1", defaultInt(-3, 1))
	}
}

func TestTerminateInstancesSDK2_NilClient(t *testing.T) {
	err := TerminateInstancesSDK2(context.Background(), nil, []string{"i-test"})
	if err == nil {
		t.Fatal("expected nil-client error, got nil")
	}
	if !strings.Contains(err.Error(), "client is nil") {
		t.Fatalf("expected 'client is nil' in error, got: %v", err)
	}
}

func TestTerminateInstancesSDK2_Empty(t *testing.T) {
	err := TerminateInstancesSDK2(context.Background(), &ec2.Client{}, []string{})
	if err == nil {
		t.Fatal("expected empty-IDs error, got nil")
	}
	if !strings.Contains(err.Error(), "at least one instance ID") {
		t.Fatalf("expected 'at least one instance ID' in error, got: %v", err)
	}
}

func TestLoadSDKConfigFromEnv_NoCreds(t *testing.T) {
	// Calling LoadDefaultConfig without AWS_ACCESS_KEY_ID etc. will
	// return a config that uses anonymous credentials — this is the
	// "credentials will be loaded later" pattern. Verify it doesn't
	// panic and returns a valid Config struct.
	t.Setenv("AWS_REGION", "us-east-1")
	t.Setenv("AWS_ACCESS_KEY_ID", "")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "")
	cfg, err := LoadSDKConfigFromEnv(context.Background())
	if err != nil {
		t.Fatalf("LoadSDKConfigFromEnv failed: %v", err)
	}
	if cfg.Region != "us-east-1" {
		t.Fatalf("expected Region=us-east-1, got %q", cfg.Region)
	}
}
