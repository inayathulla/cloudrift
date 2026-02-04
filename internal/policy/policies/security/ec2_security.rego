# EC2 Security Policy
# Security best practices for EC2 instances

package cloudrift.security.ec2

# Warn about instances without IMDSv2 enforcement
warn[result] {
	input.resource.type == "aws_instance"

	planned := input.resource.planned

	# Check if metadata options are not configured for IMDSv2
	not planned.metadata_options.http_tokens == "required"

	result := {
		"policy_id": "EC2-001",
		"policy_name": "EC2 IMDSv2 Required",
		"msg": sprintf("EC2 instance '%s' should require IMDSv2 (http_tokens = required)", [input.resource.address]),
		"severity": "medium",
		"remediation": "Add metadata_options block with http_tokens = 'required' to enforce IMDSv2",
	}
}

# Deny instances without encryption on root volume
deny[result] {
	input.resource.type == "aws_instance"

	planned := input.resource.planned
	rbd := planned.root_block_device

	# Root block device exists but not encrypted
	count(object.keys(rbd)) > 0
	not rbd.encrypted

	result := {
		"policy_id": "EC2-002",
		"policy_name": "EC2 Root Volume Encryption",
		"msg": sprintf("EC2 instance '%s' must have encrypted root volume", [input.resource.address]),
		"severity": "high",
		"remediation": "Set encrypted = true in root_block_device block",
	}
}

# Warn about public IP assignment
warn[result] {
	input.resource.type == "aws_instance"

	planned := input.resource.planned
	planned.associate_public_ip_address == true

	result := {
		"policy_id": "EC2-003",
		"policy_name": "EC2 Public IP Warning",
		"msg": sprintf("EC2 instance '%s' will have a public IP assigned. Ensure this is intentional", [input.resource.address]),
		"severity": "medium",
		"remediation": "If public access is not needed, set associate_public_ip_address = false",
	}
}

# Warn about extremely large instance types
warn[result] {
	input.resource.type == "aws_instance"

	planned := input.resource.planned

	# List of very expensive instance types
	expensive_types := [
		"x1.32xlarge", "x1e.32xlarge", "x2idn.32xlarge", "x2iedn.32xlarge",
		"p4d.24xlarge", "p4de.24xlarge", "p5.48xlarge",
		"u-6tb1.112xlarge", "u-9tb1.112xlarge", "u-12tb1.112xlarge",
	]

	planned.instance_type == expensive_types[_]

	result := {
		"policy_id": "EC2-005",
		"policy_name": "EC2 Large Instance Review",
		"msg": sprintf("EC2 instance '%s' uses very large instance type '%s'. Please review for cost optimization", [input.resource.address, planned.instance_type]),
		"severity": "medium",
		"remediation": "Ensure this instance size is necessary. Consider using smaller instances or Spot instances",
	}
}
