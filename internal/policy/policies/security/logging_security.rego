# Logging Security Policy
# Ensures CloudWatch log groups are properly configured

package cloudrift.security.logging

# Warn about unencrypted log groups
warn[result] {
	input.resource.type == "aws_cloudwatch_log_group"

	planned := input.resource.planned
	not planned.kms_key_id

	result := {
		"policy_id": "LOG-001",
		"policy_name": "CloudWatch Log Group Encryption",
		"msg": sprintf("CloudWatch log group '%s' should be encrypted with a KMS key", [input.resource.address]),
		"severity": "medium",
		"remediation": "Set kms_key_id to a KMS key ARN to encrypt log data at rest",
		"category": "security",
		"frameworks": ["hipaa", "pci_dss", "gdpr", "soc2"],
	}
}

# Warn about missing retention policy
warn[result] {
	input.resource.type == "aws_cloudwatch_log_group"

	planned := input.resource.planned
	not planned.retention_in_days

	result := {
		"policy_id": "LOG-002",
		"policy_name": "CloudWatch Log Retention",
		"msg": sprintf("CloudWatch log group '%s' does not have a retention policy configured. Logs will be retained indefinitely", [input.resource.address]),
		"severity": "medium",
		"remediation": "Set retention_in_days to an appropriate value (e.g., 90, 365) to manage storage costs and comply with data retention policies",
		"category": "security",
		"frameworks": ["hipaa", "gdpr", "soc2", "iso_27001"],
	}
}

# Warn about zero-day retention (effectively no retention)
warn[result] {
	input.resource.type == "aws_cloudwatch_log_group"

	planned := input.resource.planned
	planned.retention_in_days == 0

	result := {
		"policy_id": "LOG-002",
		"policy_name": "CloudWatch Log Retention",
		"msg": sprintf("CloudWatch log group '%s' has retention set to 0 (never expire). Set an explicit retention period", [input.resource.address]),
		"severity": "medium",
		"remediation": "Set retention_in_days to an appropriate value (e.g., 90, 365) for compliance with data retention policies",
		"category": "security",
		"frameworks": ["hipaa", "gdpr", "soc2", "iso_27001"],
	}
}
