package main

import (
	"fmt"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Config keys
const (
	ConfigEnvironment  = "byteport:environment"
	ConfigComputeTier  = "byteport:computeTier"
	ConfigProviders    = "byteport:providers"
)

// Compute tiers
const (
	TierLocal     = "local"
	TierFreemium  = "freemium"
	TierProduction = "production"
)

func main() {
pulumi.Run(func(ctx *pulumi.Context) error {
		// Get configuration
		environment, err := ctx.TryConfig("byteport:environment")
		if err != nil || environment == "" {
			environment = "dev"
		}

		computeTier, err := ctx.TryConfig("byteport:computeTier")
		if err != nil || computeTier == "" {
			computeTier = TierFreemium
		}

		ctx.Log.Info(fmt.Sprintf("Deploying BytePort environment=%s tier=%s", environment, computeTier), nil)

		// Deploy based on compute tier
		switch computeTier {
		case TierLocal:
			return deployLocal(ctx, environment)
		case TierFreemium:
			return deployFreemium(ctx, environment)
		case TierProduction:
			return deployProduction(ctx, environment)
		default:
			return fmt.Errorf("unknown compute tier: %s", computeTier)
		}
	})
}

// deployLocal deploys to local nvms infrastructure
func deployLocal(ctx *pulumi.Context, env string) error {
	ctx.Log.Info("Deploying to local nvms...", nil)
	// TODO: Implement nvms deployment
	return nil
}

// deployFreemium deploys to freemium SaaS tier
func deployFreemium(ctx *pulumi.Context, env string) error {
	ctx.Log.Info("Deploying to freemium SaaS tier...", nil)
	// TODO: Implement Vercel + Railway + Supabase deployment
	return nil
}

// deployProduction deploys to cloud providers
func deployProduction(ctx *pulumi.Context, env string) error {
	ctx.Log.Info("Deploying to production cloud...", nil)
	// TODO: Implement AWS/GCP/Azure deployment
	return nil
}
