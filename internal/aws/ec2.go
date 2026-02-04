package aws

import (
	"context"
	"fmt"

	sdkaws "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"

	"github.com/inayathulla/cloudrift/internal/models"
)

// FetchEC2Instances retrieves all EC2 instances and their configurations from AWS.
//
// This function lists all instances in the account/region and fetches detailed metadata
// including instance type, security groups, tags, and block device mappings.
//
// Terminated instances are excluded from the results.
//
// Parameters:
//   - cfg: AWS SDK configuration for API calls
//
// Returns:
//   - []models.EC2Instance: slice of instance configurations
//   - error: if the DescribeInstances call fails
func FetchEC2Instances(cfg sdkaws.Config) ([]models.EC2Instance, error) {
	ctx := context.Background()
	client := ec2.NewFromConfig(cfg)

	var instances []models.EC2Instance
	paginator := ec2.NewDescribeInstancesPaginator(client, &ec2.DescribeInstancesInput{})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("DescribeInstances: %w", err)
		}

		for _, reservation := range page.Reservations {
			for _, inst := range reservation.Instances {
				// Skip terminated instances
				if inst.State != nil && inst.State.Name == types.InstanceStateNameTerminated {
					continue
				}

				instance := convertEC2Instance(inst)
				instances = append(instances, instance)
			}
		}
	}

	return instances, nil
}

// convertEC2Instance converts an AWS SDK EC2 instance to our model.
func convertEC2Instance(inst types.Instance) models.EC2Instance {
	instance := models.EC2Instance{
		InstanceID:       safeString(inst.InstanceId),
		InstanceType:     string(inst.InstanceType),
		AMI:              safeString(inst.ImageId),
		SubnetID:         safeString(inst.SubnetId),
		VpcID:            safeString(inst.VpcId),
		PrivateIP:        safeString(inst.PrivateIpAddress),
		PublicIP:         safeString(inst.PublicIpAddress),
		KeyName:          safeString(inst.KeyName),
		EBSOptimized:     safeBool(inst.EbsOptimized),
		SourceDestCheck:  safeBool(inst.SourceDestCheck),
		Tags:             make(map[string]string),
		SecurityGroupIDs: make([]string, 0),
	}

	// State
	if inst.State != nil {
		instance.State = string(inst.State.Name)
	}

	// Availability Zone
	if inst.Placement != nil {
		instance.AvailabilityZone = safeString(inst.Placement.AvailabilityZone)
	}

	// IAM Instance Profile
	if inst.IamInstanceProfile != nil {
		instance.IAMInstanceProfile = safeString(inst.IamInstanceProfile.Arn)
	}

	// Monitoring
	if inst.Monitoring != nil {
		instance.Monitoring = inst.Monitoring.State == types.MonitoringStateEnabled
	}

	// Security Groups
	for _, sg := range inst.SecurityGroups {
		if sg.GroupId != nil {
			instance.SecurityGroupIDs = append(instance.SecurityGroupIDs, *sg.GroupId)
		}
	}

	// Tags
	for _, tag := range inst.Tags {
		if tag.Key != nil && tag.Value != nil {
			instance.Tags[*tag.Key] = *tag.Value
		}
	}

	// Root Block Device
	for _, mapping := range inst.BlockDeviceMappings {
		if mapping.DeviceName != nil && *mapping.DeviceName == safeString(inst.RootDeviceName) {
			if mapping.Ebs != nil {
				instance.RootBlockDevice = models.BlockDevice{
					DeleteOnTermination: safeBool(mapping.Ebs.DeleteOnTermination),
				}
				// Note: Volume details like size, type, encryption require additional DescribeVolumes call
			}
			break
		}
	}

	return instance
}

// safeString safely dereferences a string pointer, returning empty string if nil.
func safeString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// safeBool safely dereferences a bool pointer, returning false if nil.
func safeBool(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}
