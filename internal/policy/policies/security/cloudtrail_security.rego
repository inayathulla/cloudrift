# CloudTrail Security Policy
# Ensures audit logging is properly configured

package cloudrift.security.cloudtrail

# Deny CloudTrail without KMS encryption
deny[result] {
	input.resource.type == "aws_cloudtrail"

	planned := input.resource.planned
	not planned.kms_key_id

	result := {
		"policy_id": "CT-001",
		"policy_name": "CloudTrail KMS Encryption",
		"msg": sprintf("CloudTrail '%s' must be encrypted with a KMS key", [input.resource.address]),
		"severity": "high",
		"remediation": "Set kms_key_id to a KMS key ARN to encrypt CloudTrail logs",
		"category": "security",
		"frameworks": ["hipaa", "pci_dss", "iso_27001", "gdpr", "soc2"],
	}
}

# Warn about missing log file validation
warn[result] {
	input.resource.type == "aws_cloudtrail"

	planned := input.resource.planned
	not planned.enable_log_file_validation

	result := {
		"policy_id": "CT-002",
		"policy_name": "CloudTrail Log File Validation",
		"msg": sprintf("CloudTrail '%s' should enable log file validation to detect tampering", [input.resource.address]),
		"severity": "medium",
		"remediation": "Set enable_log_file_validation = true to ensure log integrity",
		"category": "security",
		"frameworks": ["pci_dss", "iso_27001", "soc2"],
	}
}

# Warn about single-region CloudTrail
warn[result] {
	input.resource.type == "aws_cloudtrail"

	planned := input.resource.planned
	not planned.is_multi_region_trail

	result := {
		"policy_id": "CT-003",
		"policy_name": "CloudTrail Multi-Region",
		"msg": sprintf("CloudTrail '%s' should be configured as multi-region trail for complete audit coverage", [input.resource.address]),
		"severity": "medium",
		"remediation": "Set is_multi_region_trail = true to capture events across all AWS regions",
		"category": "security",
		"frameworks": ["hipaa", "pci_dss", "iso_27001", "soc2"],
	}
}
