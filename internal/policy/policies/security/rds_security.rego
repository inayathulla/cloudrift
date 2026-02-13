# RDS Security Policy
# Security best practices for RDS database instances

package cloudrift.security.rds

# Deny unencrypted RDS instances
deny[result] {
	input.resource.type == "aws_db_instance"

	planned := input.resource.planned
	not planned.storage_encrypted

	result := {
		"policy_id": "RDS-001",
		"policy_name": "RDS Storage Encryption Required",
		"msg": sprintf("RDS instance '%s' must have storage encryption enabled", [input.resource.address]),
		"severity": "high",
		"remediation": "Set storage_encrypted = true in aws_db_instance resource",
		"category": "security",
		"frameworks": ["hipaa", "pci_dss", "iso_27001", "gdpr", "soc2"],
	}
}

# Deny publicly accessible RDS instances
deny[result] {
	input.resource.type == "aws_db_instance"

	planned := input.resource.planned
	planned.publicly_accessible == true

	result := {
		"policy_id": "RDS-002",
		"policy_name": "RDS No Public Access",
		"msg": sprintf("RDS instance '%s' must not be publicly accessible", [input.resource.address]),
		"severity": "critical",
		"remediation": "Set publicly_accessible = false in aws_db_instance resource",
		"category": "security",
		"frameworks": ["hipaa", "pci_dss", "iso_27001", "gdpr", "soc2"],
	}
}

# Warn about insufficient backup retention
warn[result] {
	input.resource.type == "aws_db_instance"

	planned := input.resource.planned
	retention := object.get(planned, "backup_retention_period", 0)
	retention < 7

	result := {
		"policy_id": "RDS-003",
		"policy_name": "RDS Backup Retention Period",
		"msg": sprintf("RDS instance '%s' has backup retention of %d days. Minimum 7 days recommended", [input.resource.address, retention]),
		"severity": "medium",
		"remediation": "Set backup_retention_period >= 7 for adequate backup coverage",
		"category": "security",
		"frameworks": ["hipaa", "iso_27001", "soc2"],
	}
}

# Warn about missing deletion protection
warn[result] {
	input.resource.type == "aws_db_instance"

	planned := input.resource.planned
	not planned.deletion_protection

	result := {
		"policy_id": "RDS-004",
		"policy_name": "RDS Deletion Protection",
		"msg": sprintf("RDS instance '%s' does not have deletion protection enabled", [input.resource.address]),
		"severity": "medium",
		"remediation": "Set deletion_protection = true to prevent accidental deletion",
		"category": "security",
		"frameworks": ["iso_27001", "soc2"],
	}
}

# Warn about single-AZ deployments
warn[result] {
	input.resource.type == "aws_db_instance"

	planned := input.resource.planned
	not planned.multi_az

	result := {
		"policy_id": "RDS-005",
		"policy_name": "RDS Multi-AZ Recommended",
		"msg": sprintf("RDS instance '%s' is not configured for Multi-AZ deployment", [input.resource.address]),
		"severity": "low",
		"remediation": "Set multi_az = true for high availability",
		"category": "security",
		"frameworks": ["hipaa", "iso_27001", "soc2"],
	}
}
