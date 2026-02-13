# VPC Security Policy
# Network security best practices

package cloudrift.security.vpc

# Deny default security groups with permissive rules
deny[result] {
	input.resource.type == "aws_default_security_group"

	planned := input.resource.planned
	ingress := object.get(planned, "ingress", [])
	count(ingress) > 0

	result := {
		"policy_id": "VPC-001",
		"policy_name": "Default Security Group Restrict All",
		"msg": sprintf("Default security group '%s' must not have any ingress rules. Default SGs should block all traffic", [input.resource.address]),
		"severity": "high",
		"remediation": "Remove all ingress and egress rules from the default security group. Use custom security groups instead",
		"category": "security",
		"frameworks": ["pci_dss", "iso_27001", "soc2"],
	}
}

# Also deny default security groups with egress rules
deny[result] {
	input.resource.type == "aws_default_security_group"

	planned := input.resource.planned
	egress := object.get(planned, "egress", [])
	count(egress) > 0

	result := {
		"policy_id": "VPC-001",
		"policy_name": "Default Security Group Restrict All",
		"msg": sprintf("Default security group '%s' must not have any egress rules. Default SGs should block all traffic", [input.resource.address]),
		"severity": "high",
		"remediation": "Remove all ingress and egress rules from the default security group. Use custom security groups instead",
		"category": "security",
		"frameworks": ["pci_dss", "iso_27001", "soc2"],
	}
}

# Warn about subnets with auto-assign public IP
warn[result] {
	input.resource.type == "aws_subnet"

	planned := input.resource.planned
	planned.map_public_ip_on_launch == true

	result := {
		"policy_id": "VPC-002",
		"policy_name": "Subnet No Auto-Assign Public IP",
		"msg": sprintf("Subnet '%s' automatically assigns public IPs to launched instances. This may expose resources to the internet", [input.resource.address]),
		"severity": "medium",
		"remediation": "Set map_public_ip_on_launch = false. Use NAT gateways for outbound internet access from private subnets",
		"category": "security",
		"frameworks": ["pci_dss", "iso_27001"],
	}
}
