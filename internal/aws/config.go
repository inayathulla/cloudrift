// Package aws provides AWS SDK integration for fetching live infrastructure state.
//
// This package handles AWS configuration loading, credential validation, and
// API interactions for retrieving current resource configurations. It serves
// as the bridge between Cloudrift and AWS services.
//
// The package uses AWS SDK v2 and supports standard credential chains including
// environment variables, shared credentials files, and IAM roles.
package aws

import (
	"context"
	"fmt"
	"time"

	sdkaws "github.com/aws/aws-sdk-go-v2/aws"
	v2config "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

const (
	// maxRetries is the number of times to retry loading AWS config on failure.
	maxRetries = 3
)

// LoadAWSConfig initializes and returns an AWS SDK configuration.
//
// The function loads credentials using the standard AWS credential chain,
// optionally overriding the profile and region. It implements retry logic
// with exponential backoff for transient failures.
//
// Parameters:
//   - profile: AWS credentials profile name (empty string uses default)
//   - region: AWS region (empty string uses default from config/environment)
//
// Returns:
//   - aws.Config: configured AWS SDK client configuration
//   - error: if configuration cannot be loaded after retries
func LoadAWSConfig(profile, region string) (sdkaws.Config, error) {
	ctx := context.Background()
	var opts []func(*v2config.LoadOptions) error
	if profile != "" {
		opts = append(opts, v2config.WithSharedConfigProfile(profile))
	}
	if region != "" {
		opts = append(opts, v2config.WithRegion(region))
	}
	fmt.Printf("🔧 AWS Profile=%s Region=%s\n", profile, region)

	var cfg sdkaws.Config
	var err error
	for i := 1; i <= maxRetries; i++ {
		cfg, err = v2config.LoadDefaultConfig(ctx, opts...)
		if err == nil {
			return cfg, nil
		}
		fmt.Printf("⚠️ Retry %d: %v\n", i, err)
		time.Sleep(time.Duration(i) * time.Second)
	}
	return sdkaws.Config{}, fmt.Errorf("could not load AWS config: %w", err)
}

// ValidateAWSCredentials verifies that the configured credentials are valid.
//
// This function calls STS GetCallerIdentity to confirm that the credentials
// can successfully authenticate with AWS. This is a lightweight operation
// that validates credentials without making resource-specific API calls.
//
// Parameters:
//   - cfg: AWS SDK configuration to validate
//
// Returns:
//   - error: if credentials are invalid or the API call fails
func ValidateAWSCredentials(cfg sdkaws.Config) error {
	client := sts.NewFromConfig(cfg)
	_, err := client.GetCallerIdentity(context.Background(), &sts.GetCallerIdentityInput{})
	if err != nil {
		return fmt.Errorf("invalid AWS credentials: %w", err)
	}
	return nil
}
