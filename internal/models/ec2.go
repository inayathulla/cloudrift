// Package models defines data structures for cloud resources.
package models

// EC2Instance represents an AWS EC2 instance with its configuration.
//
// This model captures the key attributes that are compared for drift detection
// between Terraform plans and live AWS state.
type EC2Instance struct {
	// InstanceID is the unique EC2 instance identifier (e.g., "i-0123456789abcdef0").
	InstanceID string `json:"instance_id"`

	// InstanceType is the EC2 instance type (e.g., "t3.micro", "m5.large").
	InstanceType string `json:"instance_type"`

	// AMI is the Amazon Machine Image ID used to launch the instance.
	AMI string `json:"ami"`

	// SubnetID is the VPC subnet where the instance is launched.
	SubnetID string `json:"subnet_id"`

	// VpcID is the VPC where the instance resides.
	VpcID string `json:"vpc_id"`

	// AvailabilityZone is the AZ where the instance is running.
	AvailabilityZone string `json:"availability_zone"`

	// State is the current instance state (running, stopped, terminated, etc.).
	State string `json:"state"`

	// PrivateIP is the private IPv4 address assigned to the instance.
	PrivateIP string `json:"private_ip"`

	// PublicIP is the public IPv4 address (if assigned).
	PublicIP string `json:"public_ip,omitempty"`

	// SecurityGroupIDs lists the security groups attached to the instance.
	SecurityGroupIDs []string `json:"security_group_ids"`

	// KeyName is the name of the key pair used for SSH access.
	KeyName string `json:"key_name,omitempty"`

	// IAMInstanceProfile is the IAM role attached to the instance.
	IAMInstanceProfile string `json:"iam_instance_profile,omitempty"`

	// Tags contains the instance's tag key-value pairs.
	Tags map[string]string `json:"tags"`

	// EBSOptimized indicates if EBS optimization is enabled.
	EBSOptimized bool `json:"ebs_optimized"`

	// Monitoring indicates if detailed monitoring is enabled.
	Monitoring bool `json:"monitoring"`

	// RootBlockDevice contains root volume configuration.
	RootBlockDevice BlockDevice `json:"root_block_device,omitempty"`

	// SourceDestCheck indicates if source/destination checking is enabled.
	SourceDestCheck bool `json:"source_dest_check"`

	// TerraformAddress is the Terraform resource address (e.g., "aws_instance.web").
	// This is only populated when parsing from Terraform plans.
	TerraformAddress string `json:"terraform_address,omitempty"`
}

// BlockDevice represents an EBS block device configuration.
type BlockDevice struct {
	// VolumeType is the EBS volume type (gp2, gp3, io1, io2, etc.).
	VolumeType string `json:"volume_type"`

	// VolumeSize is the volume size in GiB.
	VolumeSize int `json:"volume_size"`

	// DeleteOnTermination indicates if the volume is deleted when instance terminates.
	DeleteOnTermination bool `json:"delete_on_termination"`

	// Encrypted indicates if the volume is encrypted.
	Encrypted bool `json:"encrypted"`

	// IOPS is the provisioned IOPS (for io1/io2/gp3 volumes).
	IOPS int `json:"iops,omitempty"`

	// Throughput is the provisioned throughput in MiB/s (for gp3 volumes).
	Throughput int `json:"throughput,omitempty"`
}

// Name returns the instance name from tags, or the instance ID if no name tag exists.
func (i EC2Instance) Name() string {
	if name, ok := i.Tags["Name"]; ok && name != "" {
		return name
	}
	return i.InstanceID
}
