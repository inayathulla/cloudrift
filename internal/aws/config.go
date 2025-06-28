package aws

import (
	"context"
	"fmt"
	"time"

	sdkaws "github.com/aws/aws-sdk-go-v2/aws"
	v2config "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

// LoadAWSConfig returns a v2 aws.Config (value), retrying on errors.
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
	for i := 1; i <= 3; i++ {
		cfg, err = v2config.LoadDefaultConfig(ctx, opts...)
		if err == nil {
			return cfg, nil
		}
		fmt.Printf("⚠️ Retry %d: %v\n", i, err)
		time.Sleep(time.Duration(i) * time.Second)
	}
	return sdkaws.Config{}, fmt.Errorf("could not load AWS config: %w", err)
}

// ValidateAWSCredentials calls STS GetCallerIdentity.
func ValidateAWSCredentials(cfg sdkaws.Config) error {
	client := sts.NewFromConfig(cfg)
	_, err := client.GetCallerIdentity(context.Background(), &sts.GetCallerIdentityInput{})
	if err != nil {
		return fmt.Errorf("invalid AWS credentials: %w", err)
	}
	return nil
}
