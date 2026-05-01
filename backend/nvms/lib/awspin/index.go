// Package awspin provides AWS service abstractions for NVMS.
//
// # Subpackages
//
//	ec2 - EC2 instance management
//	network - VPC, ALB, and Route53 management
//	s3 - S3 bucket operations
package awspin

import "nvms/lib/awspin/s3"
import "nvms/lib/awspin/ec2"
import "nvms/lib/awspin/network"

// GetPayloadHash returns SHA256 hash of payload content.
func GetPayloadHash(payload []byte) string {
	return s3.GetPayloadHash(payload)
}

// Config contains AWS client configuration options.
type Config = s3.Config
