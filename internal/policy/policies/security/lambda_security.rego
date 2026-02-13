# Lambda Security Policy
# Security best practices for Lambda functions

package cloudrift.security.lambda

# Warn about missing X-Ray tracing
warn[result] {
	input.resource.type == "aws_lambda_function"

	planned := input.resource.planned
	not planned.tracing_config

	result := {
		"policy_id": "LAMBDA-001",
		"policy_name": "Lambda Tracing Enabled",
		"msg": sprintf("Lambda function '%s' should have X-Ray tracing enabled for observability", [input.resource.address]),
		"severity": "medium",
		"remediation": "Add tracing_config { mode = \"Active\" } to enable X-Ray tracing",
		"category": "security",
		"frameworks": ["soc2", "iso_27001"],
	}
}

# Warn about tracing configured but not active
warn[result] {
	input.resource.type == "aws_lambda_function"

	planned := input.resource.planned
	planned.tracing_config
	planned.tracing_config.mode != "Active"

	result := {
		"policy_id": "LAMBDA-001",
		"policy_name": "Lambda Tracing Enabled",
		"msg": sprintf("Lambda function '%s' has tracing configured but not set to Active mode", [input.resource.address]),
		"severity": "medium",
		"remediation": "Set tracing_config { mode = \"Active\" } to enable X-Ray tracing",
		"category": "security",
		"frameworks": ["soc2", "iso_27001"],
	}
}

# Warn about Lambda not in VPC
warn[result] {
	input.resource.type == "aws_lambda_function"

	planned := input.resource.planned
	not planned.vpc_config

	result := {
		"policy_id": "LAMBDA-002",
		"policy_name": "Lambda VPC Configuration",
		"msg": sprintf("Lambda function '%s' is not configured to run in a VPC. Consider VPC placement for network isolation", [input.resource.address]),
		"severity": "medium",
		"remediation": "Add vpc_config block with subnet_ids and security_group_ids for network isolation",
		"category": "security",
		"frameworks": ["hipaa", "pci_dss", "iso_27001"],
	}
}
