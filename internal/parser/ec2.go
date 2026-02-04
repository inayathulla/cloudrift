package parser

import (
	"github.com/inayathulla/cloudrift/internal/models"
)

// ParseEC2Instances extracts aws_instance resources from a Terraform plan.
//
// This function iterates through all resource changes in the plan and extracts
// EC2 instance configurations from resources with type "aws_instance". It parses
// the following attributes from each instance:
//
//   - Instance ID, type, and AMI
//   - Network configuration (subnet, VPC, security groups)
//   - Tags
//   - EBS optimization and monitoring settings
//   - Root block device configuration
//
// Resources being deleted (with nil "after" state) are skipped.
//
// Parameters:
//   - plan: pointer to a parsed TerraformPlan structure
//
// Returns:
//   - []models.EC2Instance: slice of EC2 instance configurations found in the plan
func ParseEC2Instances(plan *TerraformPlan) []models.EC2Instance {
	var instances []models.EC2Instance

	for _, rc := range plan.ResourceChanges {
		if rc.Type != "aws_instance" {
			continue
		}
		after := rc.Change.After
		if after == nil {
			continue
		}

		instance := models.EC2Instance{
			TerraformAddress: rc.Address,
			Tags:             make(map[string]string),
			SecurityGroupIDs: make([]string, 0),
		}

		// Basic attributes
		if v, ok := after["id"].(string); ok {
			instance.InstanceID = v
		}
		if v, ok := after["instance_type"].(string); ok {
			instance.InstanceType = v
		}
		if v, ok := after["ami"].(string); ok {
			instance.AMI = v
		}
		if v, ok := after["subnet_id"].(string); ok {
			instance.SubnetID = v
		}
		if v, ok := after["availability_zone"].(string); ok {
			instance.AvailabilityZone = v
		}
		if v, ok := after["private_ip"].(string); ok {
			instance.PrivateIP = v
		}
		if v, ok := after["public_ip"].(string); ok {
			instance.PublicIP = v
		}
		if v, ok := after["key_name"].(string); ok {
			instance.KeyName = v
		}
		if v, ok := after["iam_instance_profile"].(string); ok {
			instance.IAMInstanceProfile = v
		}
		if v, ok := after["ebs_optimized"].(bool); ok {
			instance.EBSOptimized = v
		}
		if v, ok := after["monitoring"].(bool); ok {
			instance.Monitoring = v
		}
		if v, ok := after["source_dest_check"].(bool); ok {
			instance.SourceDestCheck = v
		}

		// Security groups - can be vpc_security_group_ids or security_groups
		if sgs, ok := after["vpc_security_group_ids"].([]interface{}); ok {
			for _, sg := range sgs {
				if sgStr, ok := sg.(string); ok {
					instance.SecurityGroupIDs = append(instance.SecurityGroupIDs, sgStr)
				}
			}
		}
		if sgs, ok := after["security_groups"].([]interface{}); ok {
			for _, sg := range sgs {
				if sgStr, ok := sg.(string); ok {
					instance.SecurityGroupIDs = append(instance.SecurityGroupIDs, sgStr)
				}
			}
		}

		// Tags
		if tags, ok := after["tags"].(map[string]interface{}); ok {
			for k, v := range tags {
				if vStr, ok := v.(string); ok {
					instance.Tags[k] = vStr
				}
			}
		}
		// Also check tags_all for provider default tags
		if tags, ok := after["tags_all"].(map[string]interface{}); ok {
			for k, v := range tags {
				if vStr, ok := v.(string); ok {
					// Only add if not already present from tags
					if _, exists := instance.Tags[k]; !exists {
						instance.Tags[k] = vStr
					}
				}
			}
		}

		// Root block device
		if rbd, ok := after["root_block_device"].([]interface{}); ok && len(rbd) > 0 {
			if rbdMap, ok := rbd[0].(map[string]interface{}); ok {
				instance.RootBlockDevice = parseBlockDevice(rbdMap)
			}
		}

		instances = append(instances, instance)
	}

	return instances
}

// parseBlockDevice parses a block device configuration from Terraform plan.
func parseBlockDevice(bd map[string]interface{}) models.BlockDevice {
	device := models.BlockDevice{}

	if v, ok := bd["volume_type"].(string); ok {
		device.VolumeType = v
	}
	if v, ok := bd["volume_size"].(float64); ok {
		device.VolumeSize = int(v)
	}
	if v, ok := bd["delete_on_termination"].(bool); ok {
		device.DeleteOnTermination = v
	}
	if v, ok := bd["encrypted"].(bool); ok {
		device.Encrypted = v
	}
	if v, ok := bd["iops"].(float64); ok {
		device.IOPS = int(v)
	}
	if v, ok := bd["throughput"].(float64); ok {
		device.Throughput = int(v)
	}

	return device
}
