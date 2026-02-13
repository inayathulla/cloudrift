# ELB Security Policy
# Security best practices for Application Load Balancers

package cloudrift.security.elb

# Warn about missing access logging
warn[result] {
	input.resource.type == "aws_lb"

	planned := input.resource.planned
	not planned.access_logs

	result := {
		"policy_id": "ELB-001",
		"policy_name": "ALB Access Logging",
		"msg": sprintf("Load balancer '%s' should have access logging enabled for audit trails", [input.resource.address]),
		"severity": "medium",
		"remediation": "Add access_logs block with enabled = true and specify an S3 bucket for log storage",
		"category": "security",
		"frameworks": ["hipaa", "pci_dss", "iso_27001", "soc2"],
	}
}

# Warn about access logs configured but not enabled
warn[result] {
	input.resource.type == "aws_lb"

	planned := input.resource.planned
	planned.access_logs
	not planned.access_logs.enabled

	result := {
		"policy_id": "ELB-001",
		"policy_name": "ALB Access Logging",
		"msg": sprintf("Load balancer '%s' has access_logs configured but not enabled", [input.resource.address]),
		"severity": "medium",
		"remediation": "Set access_logs { enabled = true } to activate access logging",
		"category": "security",
		"frameworks": ["hipaa", "pci_dss", "iso_27001", "soc2"],
	}
}

# Deny non-HTTPS listeners
deny[result] {
	input.resource.type == "aws_lb_listener"

	planned := input.resource.planned
	planned.protocol != "HTTPS"
	planned.protocol != "TLS"

	result := {
		"policy_id": "ELB-002",
		"policy_name": "ALB HTTPS Listener Required",
		"msg": sprintf("Load balancer listener '%s' uses protocol '%s'. HTTPS or TLS is required for encryption in transit", [input.resource.address, planned.protocol]),
		"severity": "high",
		"remediation": "Change protocol to HTTPS and configure an SSL certificate",
		"category": "security",
		"frameworks": ["hipaa", "pci_dss", "iso_27001", "gdpr", "soc2"],
	}
}

# Warn about missing deletion protection
warn[result] {
	input.resource.type == "aws_lb"

	planned := input.resource.planned
	not planned.enable_deletion_protection

	result := {
		"policy_id": "ELB-003",
		"policy_name": "ALB Deletion Protection",
		"msg": sprintf("Load balancer '%s' does not have deletion protection enabled", [input.resource.address]),
		"severity": "medium",
		"remediation": "Set enable_deletion_protection = true to prevent accidental deletion",
		"category": "security",
		"frameworks": ["iso_27001", "soc2"],
	}
}
