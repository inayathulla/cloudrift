package aws

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	sdkconfig "github.com/aws/aws-sdk-go-v2/config"
)

// LoadAWSConfig loads the AWS configuration based on profile and region.
// Priority:
// 1. Explicit profile and region from config file (viper).
// 2. Environment variables (AWS_ACCESS_KEY_ID, etc.).
// 3. IAM role (when running in AWS EC2/ECS/Lambda).
// LoadAWSConfig loads AWS config using profile and region, with retry.
func LoadAWSConfig(profile, region string) (aws.Config, error) {
	ctx := context.Background()
	var opts []func(*sdkconfig.LoadOptions) error

	if profile != "" {
		opts = append(opts, sdkconfig.WithSharedConfigProfile(profile))
	}
	if region != "" {
		opts = append(opts, sdkconfig.WithRegion(region))
	}

	fmt.Printf("🔧 Using AWS Profile: %s | Region: %s\n", profile, region)

	var cfg aws.Config
	var err error

	// Retry logic for transient errors
	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		cfg, err = sdkconfig.LoadDefaultConfig(ctx, opts...)
		if err == nil {
			break
		}
		fmt.Printf("⚠️ Retry %d: failed to load AWS config: %v\n", i+1, err)
		time.Sleep(time.Duration(i+1) * time.Second)
	}

	if err != nil {
		return aws.Config{}, fmt.Errorf("failed to load AWS config after %d attempts: %w", maxRetries, err)
	}

	return cfg, nil
}

// ValidateAWSCredentials verifies that credentials are valid using STS GetCallerIdentity.
func ValidateAWSCredentials(cfg aws.Config) error {
	ctx := context.Background()
	stsClient := sts.NewFromConfig(cfg)

	_, err := stsClient.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return fmt.Errorf("invalid AWS credentials: %w", err)
	}
	return nil
}
