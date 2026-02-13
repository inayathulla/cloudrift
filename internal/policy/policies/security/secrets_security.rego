# Secrets Manager Security Policy
# Ensures secrets are properly encrypted and rotated

package cloudrift.security.secrets

# Warn about secrets without KMS encryption
warn[result] {
	input.resource.type == "aws_secretsmanager_secret"

	planned := input.resource.planned
	not planned.kms_key_id

	result := {
		"policy_id": "SECRET-001",
		"policy_name": "Secrets Manager KMS Encryption",
		"msg": sprintf("Secret '%s' should use a customer-managed KMS key instead of the default AWS-managed key", [input.resource.address]),
		"severity": "medium",
		"remediation": "Set kms_key_id to a customer-managed KMS key ARN for better key control and audit",
		"category": "security",
		"frameworks": ["hipaa", "pci_dss", "gdpr", "soc2"],
	}
}

# Warn about secrets without rotation
warn[result] {
	input.resource.type == "aws_secretsmanager_secret"

	planned := input.resource.planned
	not planned.rotation_lambda_arn

	result := {
		"policy_id": "SECRET-002",
		"policy_name": "Secrets Rotation Enabled",
		"msg": sprintf("Secret '%s' does not have automatic rotation configured", [input.resource.address]),
		"severity": "medium",
		"remediation": "Configure automatic rotation with a Lambda function using rotation_lambda_arn and rotation_rules",
		"category": "security",
		"frameworks": ["pci_dss", "iso_27001", "soc2"],
	}
}
