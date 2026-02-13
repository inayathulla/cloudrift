# Security Group Policy
# Prevents overly permissive security group rules

package cloudrift.security.security_groups

# Deny unrestricted SSH ingress
deny[result] {
	input.resource.type == "aws_security_group_rule"

	rule := input.resource.planned
	rule.type == "ingress"
	rule.from_port <= 22
	rule.to_port >= 22
	rule.cidr_blocks[_] == "0.0.0.0/0"

	result := {
		"policy_id": "SG-001",
		"policy_name": "No Unrestricted SSH Access",
		"msg": sprintf("Security group rule '%s' allows SSH (port 22) from 0.0.0.0/0", [input.resource.address]),
		"severity": "critical",
		"remediation": "Restrict SSH access to specific IP addresses or CIDR blocks. Use a bastion host or VPN",
		"category": "security",
		"frameworks": ["pci_dss", "iso_27001", "soc2"],
	}
}

# Deny unrestricted RDP ingress
deny[result] {
	input.resource.type == "aws_security_group_rule"

	rule := input.resource.planned
	rule.type == "ingress"
	rule.from_port <= 3389
	rule.to_port >= 3389
	rule.cidr_blocks[_] == "0.0.0.0/0"

	result := {
		"policy_id": "SG-002",
		"policy_name": "No Unrestricted RDP Access",
		"msg": sprintf("Security group rule '%s' allows RDP (port 3389) from 0.0.0.0/0", [input.resource.address]),
		"severity": "critical",
		"remediation": "Restrict RDP access to specific IP addresses. Use a bastion host or VPN",
		"category": "security",
		"frameworks": ["pci_dss", "iso_27001", "soc2"],
	}
}

# Deny all traffic from anywhere
deny[result] {
	input.resource.type == "aws_security_group_rule"

	rule := input.resource.planned
	rule.type == "ingress"
	rule.from_port == 0
	rule.to_port == 65535
	rule.cidr_blocks[_] == "0.0.0.0/0"

	result := {
		"policy_id": "SG-003",
		"policy_name": "No Unrestricted All Ports Access",
		"msg": sprintf("Security group rule '%s' allows all ports from 0.0.0.0/0", [input.resource.address]),
		"severity": "critical",
		"remediation": "Define specific ports needed and restrict source IP ranges",
		"category": "security",
		"frameworks": ["pci_dss", "iso_27001", "soc2"],
	}
}

# Warn about database port exposure
warn[result] {
	input.resource.type == "aws_security_group_rule"

	rule := input.resource.planned
	rule.type == "ingress"

	# Common database ports
	db_ports := [3306, 5432, 1433, 1521, 27017, 6379, 11211]
	port := db_ports[_]

	rule.from_port <= port
	rule.to_port >= port
	rule.cidr_blocks[_] == "0.0.0.0/0"

	result := {
		"policy_id": "SG-004",
		"policy_name": "Database Port Public Exposure",
		"msg": sprintf("Security group rule '%s' exposes database port %d to 0.0.0.0/0", [input.resource.address, port]),
		"severity": "high",
		"remediation": "Never expose database ports to the internet. Use private subnets and VPN",
		"category": "security",
		"frameworks": ["hipaa", "pci_dss", "iso_27001", "soc2"],
	}
}

# Also check inline security group rules
deny[result] {
	input.resource.type == "aws_security_group"

	sg := input.resource.planned
	ingress := sg.ingress[_]
	ingress.from_port <= 22
	ingress.to_port >= 22
	ingress.cidr_blocks[_] == "0.0.0.0/0"

	result := {
		"policy_id": "SG-001",
		"policy_name": "No Unrestricted SSH Access",
		"msg": sprintf("Security group '%s' allows SSH (port 22) from 0.0.0.0/0", [input.resource.address]),
		"severity": "critical",
		"remediation": "Restrict SSH access to specific IP addresses or CIDR blocks",
		"category": "security",
		"frameworks": ["pci_dss", "iso_27001", "soc2"],
	}
}
