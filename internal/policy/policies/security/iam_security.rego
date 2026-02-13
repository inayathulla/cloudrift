# IAM Security Policy
# Prevents overly permissive IAM configurations

package cloudrift.security.iam

# Deny wildcard actions in IAM policies (Action as array)
deny[result] {
	input.resource.type == "aws_iam_policy"

	planned := input.resource.planned
	doc := json.unmarshal(planned.policy)
	statement := doc.Statement[_]
	statement.Effect == "Allow"
	statement.Action[_] == "*"

	result := {
		"policy_id": "IAM-001",
		"policy_name": "No Wildcard IAM Actions",
		"msg": sprintf("IAM policy '%s' contains a wildcard (*) action with Allow effect. This grants unrestricted permissions", [input.resource.address]),
		"severity": "critical",
		"remediation": "Replace wildcard Action with specific actions following least-privilege principle",
		"category": "security",
		"frameworks": ["hipaa", "pci_dss", "iso_27001", "gdpr", "soc2"],
	}
}

# Deny wildcard actions in IAM policies (Action as string)
deny[result] {
	input.resource.type == "aws_iam_policy"

	planned := input.resource.planned
	doc := json.unmarshal(planned.policy)
	statement := doc.Statement[_]
	statement.Effect == "Allow"
	statement.Action == "*"

	result := {
		"policy_id": "IAM-001",
		"policy_name": "No Wildcard IAM Actions",
		"msg": sprintf("IAM policy '%s' contains a wildcard (*) action with Allow effect. This grants unrestricted permissions", [input.resource.address]),
		"severity": "critical",
		"remediation": "Replace wildcard Action with specific actions following least-privilege principle",
		"category": "security",
		"frameworks": ["hipaa", "pci_dss", "iso_27001", "gdpr", "soc2"],
	}
}

# Warn about inline policies on IAM users
warn[result] {
	input.resource.type == "aws_iam_user_policy"

	result := {
		"policy_id": "IAM-002",
		"policy_name": "No Inline Policies on Users",
		"msg": sprintf("IAM user policy '%s' uses inline policy. Use managed policies instead for better governance", [input.resource.address]),
		"severity": "medium",
		"remediation": "Convert inline policies to managed policies attached via aws_iam_user_policy_attachment",
		"category": "security",
		"frameworks": ["pci_dss", "iso_27001", "soc2"],
	}
}

# Warn about overly broad IAM role trust policies (Principal as string)
warn[result] {
	input.resource.type == "aws_iam_role"

	planned := input.resource.planned
	doc := json.unmarshal(planned.assume_role_policy)
	statement := doc.Statement[_]
	statement.Effect == "Allow"
	statement.Principal == "*"

	result := {
		"policy_id": "IAM-003",
		"policy_name": "IAM Role Trust Too Broad",
		"msg": sprintf("IAM role '%s' has overly broad trust policy allowing any principal to assume the role", [input.resource.address]),
		"severity": "high",
		"remediation": "Restrict Principal to specific AWS accounts, services, or IAM entities",
		"category": "security",
		"frameworks": ["pci_dss", "iso_27001", "soc2"],
	}
}

# Warn about overly broad IAM role trust policies (Principal.AWS as string)
warn[result] {
	input.resource.type == "aws_iam_role"

	planned := input.resource.planned
	doc := json.unmarshal(planned.assume_role_policy)
	statement := doc.Statement[_]
	statement.Effect == "Allow"
	statement.Principal.AWS == "*"

	result := {
		"policy_id": "IAM-003",
		"policy_name": "IAM Role Trust Too Broad",
		"msg": sprintf("IAM role '%s' has overly broad trust policy allowing any AWS principal to assume the role", [input.resource.address]),
		"severity": "high",
		"remediation": "Restrict Principal.AWS to specific AWS account ARNs or IAM role ARNs",
		"category": "security",
		"frameworks": ["pci_dss", "iso_27001", "soc2"],
	}
}
