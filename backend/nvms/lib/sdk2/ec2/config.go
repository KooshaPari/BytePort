// Package ec2 — SDK-v2 config loader (kept separate from sdk2.go so
// tests can avoid pulling in the heavy config loader).
package ec2

import (
	"context"

	awsv2 "github.com/aws/aws-sdk-go-v2/aws"
	awscfg "github.com/aws/aws-sdk-go-v2/config"
)

// LoadSDKConfigFromEnv loads the default AWS config from environment.
// AWS_REGION, AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY, AWS_PROFILE,
// AWS_ENDPOINT_URL (LocalStack) are all honored by the SDK.
//
// Exposed (not internal) so Spin apps — which can't import the SDK
// config package directly because of the Spin CGo constraint — can
// call this from a non-Spin helper.
func LoadSDKConfigFromEnv(ctx context.Context) (awsv2.Config, error) {
	return awscfg.LoadDefaultConfig(ctx)
}
