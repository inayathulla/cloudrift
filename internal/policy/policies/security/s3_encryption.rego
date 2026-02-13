# S3 Encryption Policy
# Ensures S3 buckets have server-side encryption enabled

package cloudrift.security.s3_encryption

deny[result] {
	input.resource.type == "aws_s3_bucket"

	# Check if encryption is not configured in planned state
	planned := input.resource.planned
	not planned.encryption_algorithm

	result := {
		"policy_id": "S3-001",
		"policy_name": "S3 Encryption Required",
		"msg": sprintf("S3 bucket '%s' must have server-side encryption enabled", [input.resource.address]),
		"severity": "high",
		"remediation": "Add server_side_encryption_configuration block with sse_algorithm set to 'AES256' or 'aws:kms'",
		"category": "security",
		"frameworks": ["hipaa", "pci_dss", "iso_27001", "gdpr", "soc2"],
	}
}

deny[result] {
	input.resource.type == "aws_s3_bucket"

	# Check live state - bucket exists but has no encryption
	live := input.resource.live
	count(live) > 0
	not live.encryption_algorithm

	result := {
		"policy_id": "S3-001",
		"policy_name": "S3 Encryption Required",
		"msg": sprintf("S3 bucket '%s' in AWS does not have server-side encryption enabled", [input.resource.address]),
		"severity": "critical",
		"remediation": "Enable server-side encryption on the existing S3 bucket in AWS console or via Terraform",
		"category": "security",
		"frameworks": ["hipaa", "pci_dss", "iso_27001", "gdpr", "soc2"],
	}
}

# Warn if using AES256 instead of KMS
warn[result] {
	input.resource.type == "aws_s3_bucket"

	planned := input.resource.planned
	planned.encryption_algorithm == "AES256"

	result := {
		"policy_id": "S3-002",
		"policy_name": "S3 KMS Encryption Recommended",
		"msg": sprintf("S3 bucket '%s' uses AES256 encryption. Consider using AWS KMS for better key management", [input.resource.address]),
		"severity": "low",
		"remediation": "Change sse_algorithm to 'aws:kms' and specify a KMS key",
		"category": "security",
		"frameworks": ["hipaa", "pci_dss", "soc2"],
	}
}
