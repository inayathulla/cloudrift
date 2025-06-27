package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

// GetCallerIdentity fetches the current AWS IAM identity (ARN, Account ID).
func GetCallerIdentity(cfg aws.Config) (*sts.GetCallerIdentityOutput, error) {
	ctx := context.Background()
	client := sts.NewFromConfig(cfg)

	output, err := client.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to get caller identity: %w", err)
	}
	return output, nil
}

func SafeString(s *string) string {
	if s == nil {
		return "N/A"
	}
	return *s
}
