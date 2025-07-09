package common

import (
	"github.com/inayathulla/cloudrift/internal/aws"
	"github.com/inayathulla/cloudrift/internal/models"
	"github.com/inayathulla/cloudrift/internal/parser"
	"github.com/spf13/viper"

	sdkaws "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

// LoadAppConfig loads the config file and returns profile, region, planPath
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

// InitAWS loads AWS config
func InitAWS(profile, region string) (cfg sdkaws.Config, err error) {
	return aws.LoadAWSConfig(profile, region)
}

// ValidateCredentials validates AWS credentials
func ValidateCredentials(cfg sdkaws.Config) error {
	return aws.ValidateAWSCredentials(cfg)
}

// GetCallerIdentity returns AWS caller identity
func GetCallerIdentity(cfg sdkaws.Config) (*sts.GetCallerIdentityOutput, error) {
	return aws.GetCallerIdentity(cfg)
}

// LoadPlan loads the Terraform plan
func LoadPlan(planPath string) ([]models.S3Bucket, error) {
	return parser.LoadPlan(planPath)
}
