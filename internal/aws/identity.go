package aws

import (
	"context"
	"fmt"

	sdkaws "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

// GetCallerIdentity fetches the current AWS IAM identity (ARN, Account ID).
func GetCallerIdentity(cfg sdkaws.Config) (*sts.GetCallerIdentityOutput, error) {
	client := sts.NewFromConfig(cfg)
	out, err := client.GetCallerIdentity(context.Background(), &sts.GetCallerIdentityInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to get caller identity: %w", err)
	}
	return out, nil
}

// SafeString returns a non-nil string.
func SafeString(s *string) string {
	if s == nil {
		return "N/A"
	}
	return *s
}
