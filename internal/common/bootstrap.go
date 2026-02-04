// Package common provides shared utilities and convenience wrappers for Cloudrift.
//
// This package serves as a facade over the aws and parser packages, providing
// simplified functions for common operations like loading configuration,
// initializing AWS, and parsing Terraform plans.
package common

import (
	sdkaws "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/spf13/viper"

	"github.com/inayathulla/cloudrift/internal/aws"
	"github.com/inayathulla/cloudrift/internal/models"
	"github.com/inayathulla/cloudrift/internal/parser"
)

// LoadAppConfig reads the Cloudrift configuration file and extracts settings.
//
// The configuration file (YAML format) should contain:
//   - aws_profile: AWS credentials profile name
//   - region: AWS region for API calls
//   - plan_path: path to the Terraform plan JSON file
//
// Parameters:
//   - configPath: filesystem path to the cloudrift.yml configuration file
//
// Returns:
//   - profile: AWS profile name (may be empty for default)
//   - region: AWS region
//   - planPath: path to Terraform plan JSON
//   - err: if configuration cannot be read
func LoadAppConfig(configPath string) (profile, region, planPath string, err error) {
	viper.SetConfigFile(configPath)
	if err = viper.ReadInConfig(); err != nil {
		return
	}
	profile = viper.GetString("aws_profile")
	region = viper.GetString("region")
	planPath = viper.GetString("plan_path")
	return
}

// InitAWS initializes and returns an AWS SDK configuration.
// This is a convenience wrapper around aws.LoadAWSConfig.
func InitAWS(profile, region string) (cfg sdkaws.Config, err error) {
	return aws.LoadAWSConfig(profile, region)
}

// ValidateCredentials verifies AWS credentials are valid.
// This is a convenience wrapper around aws.ValidateAWSCredentials.
func ValidateCredentials(cfg sdkaws.Config) error {
	return aws.ValidateAWSCredentials(cfg)
}

// GetCallerIdentity retrieves the current AWS IAM identity.
// This is a convenience wrapper around aws.GetCallerIdentity.
func GetCallerIdentity(cfg sdkaws.Config) (*sts.GetCallerIdentityOutput, error) {
	return aws.GetCallerIdentity(cfg)
}

// LoadPlan reads and parses a Terraform plan JSON file for S3 buckets.
// This is a convenience wrapper around parser.LoadPlan.
func LoadPlan(planPath string) ([]models.S3Bucket, error) {
	return parser.LoadPlan(planPath)
}

// LoadEC2Plan reads and parses a Terraform plan JSON file for EC2 instances.
func LoadEC2Plan(planPath string) ([]models.EC2Instance, error) {
	return parser.LoadEC2Plan(planPath)
}
