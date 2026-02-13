# KMS Security Policy
# Ensures proper key management practices

package cloudrift.security.kms

# Deny KMS keys without rotation enabled
deny[result] {
	input.resource.type == "aws_kms_key"

	planned := input.resource.planned
	not planned.enable_key_rotation

	result := {
		"policy_id": "KMS-001",
		"policy_name": "KMS Key Rotation Enabled",
		"msg": sprintf("KMS key '%s' must have automatic key rotation enabled", [input.resource.address]),
		"severity": "high",
		"remediation": "Set enable_key_rotation = true to automatically rotate the KMS key annually",
		"category": "security",
		"frameworks": ["hipaa", "pci_dss", "iso_27001", "soc2"],
	}
}

# Warn about short deletion window
warn[result] {
	input.resource.type == "aws_kms_key"

	planned := input.resource.planned
	window := object.get(planned, "deletion_window_in_days", 30)
	window < 14

	result := {
		"policy_id": "KMS-002",
		"policy_name": "KMS Key Deletion Window",
		"msg": sprintf("KMS key '%s' has a deletion window of %d days. Minimum 14 days recommended to prevent accidental key loss", [input.resource.address, window]),
		"severity": "medium",
		"remediation": "Set deletion_window_in_days >= 14 to allow sufficient time to recover from accidental deletion",
		"category": "security",
		"frameworks": ["iso_27001", "soc2"],
	}
}
