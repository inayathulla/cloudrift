# EBS Security Policy
# Ensures EBS volumes and snapshots are encrypted

package cloudrift.security.ebs

# Deny unencrypted EBS volumes
deny[result] {
	input.resource.type == "aws_ebs_volume"

	planned := input.resource.planned
	not planned.encrypted

	result := {
		"policy_id": "EBS-001",
		"policy_name": "EBS Volume Encryption",
		"msg": sprintf("EBS volume '%s' must have encryption enabled", [input.resource.address]),
		"severity": "high",
		"remediation": "Set encrypted = true in aws_ebs_volume resource",
		"category": "security",
		"frameworks": ["hipaa", "pci_dss", "iso_27001", "gdpr", "soc2"],
	}
}

# Deny unencrypted EBS snapshot copies
deny[result] {
	input.resource.type == "aws_ebs_snapshot_copy"

	planned := input.resource.planned
	not planned.encrypted

	result := {
		"policy_id": "EBS-002",
		"policy_name": "EBS Snapshot Encryption",
		"msg": sprintf("EBS snapshot copy '%s' must have encryption enabled", [input.resource.address]),
		"severity": "high",
		"remediation": "Set encrypted = true in aws_ebs_snapshot_copy resource",
		"category": "security",
		"frameworks": ["hipaa", "pci_dss", "iso_27001", "gdpr"],
	}
}
