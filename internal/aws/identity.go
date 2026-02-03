package aws

import (
	"context"
	"fmt"

	sdkaws "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

// GetCallerIdentity retrieves the AWS IAM identity for the current credentials.
//
// This function returns details about the authenticated principal, including
// the ARN, AWS account ID, and user ID. It's useful for displaying connection
// information and verifying the correct account is being accessed.
//
// Parameters:
//   - cfg: AWS SDK configuration with valid credentials
//
// Returns:
//   - *sts.GetCallerIdentityOutput: identity details (ARN, Account, UserId)
//   - error: if the STS API call fails
func GetCallerIdentity(cfg sdkaws.Config) (*sts.GetCallerIdentityOutput, error) {
	client := sts.NewFromConfig(cfg)
	out, err := client.GetCallerIdentity(context.Background(), &sts.GetCallerIdentityInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to get caller identity: %w", err)
	}
	return out, nil
}
