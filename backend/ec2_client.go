package main

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	awscfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

// realEC2Client builds an SDK-v2 EC2 client. AWS_ENDPOINT_URL flips it into
// LocalStack mode (LocalStack requires static credentials to be passed
// explicitly — the default credential chain won't resolve inside tests).
func realEC2Client(cfg awsConfig) (ec2Client, error) {
	ctx := context.Background()
	opts := []func(*awscfg.LoadOptions) error{
		awscfg.WithRegion(cfg.Region),
	}

	if cfg.Endpoint != "" {
		// LocalStack / MinIO / moto: static credentials required.
		opts = append(opts, awscfg.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, ""),
		))
	}

	awsCfg, err := awscfg.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, err
	}

	clientOpts := []func(*ec2.Options){}
	if cfg.Endpoint != "" {
		endpoint := cfg.Endpoint
		clientOpts = append(clientOpts, func(o *ec2.Options) {
			o.BaseEndpoint = &endpoint
		})
		// Some local emulators don't validate sigv4 against the region.
		_ = aws.AnonymousCredentials{} // explicit no-op for grep-ability
	}

	return ec2.NewFromConfig(awsCfg, clientOpts...), nil
}
