# Instance Type Cost Policy
# Controls which instance types can be used

package cloudrift.cost.instances

# Allowed instance type families for different environments
dev_allowed_families := ["t3", "t3a", "t4g"]
staging_allowed_families := ["t3", "t3a", "t4g", "m5", "m6i", "c5", "c6i"]

# Warn about very large instance sizes
warn[result] {
	input.resource.type == "aws_instance"

	planned := input.resource.planned
	instance_type := planned.instance_type

	# Check for very large sizes
	contains(instance_type, "24xlarge")

	result := {
		"policy_id": "COST-002",
		"policy_name": "Very Large Instance Size",
		"msg": sprintf("EC2 instance '%s' uses very large size '%s'. Monthly cost may exceed $5,000", [input.resource.address, instance_type]),
		"severity": "medium",
		"remediation": "Verify this instance size is necessary. Consider auto-scaling instead of single large instances",
		"category": "cost",
		"frameworks": [],
	}
}

warn[result] {
	input.resource.type == "aws_instance"

	planned := input.resource.planned
	instance_type := planned.instance_type

	contains(instance_type, "16xlarge")

	result := {
		"policy_id": "COST-002",
		"policy_name": "Very Large Instance Size",
		"msg": sprintf("EC2 instance '%s' uses large size '%s'. Review for cost optimization", [input.resource.address, instance_type]),
		"severity": "low",
		"remediation": "Consider if this instance size is necessary. Review rightsizing recommendations",
		"category": "cost",
		"frameworks": [],
	}
}

# Warn about previous generation instances
warn[result] {
	input.resource.type == "aws_instance"

	planned := input.resource.planned
	instance_type := planned.instance_type
	parts := split(instance_type, ".")
	family := parts[0]

	# Previous generation families
	old_families := ["m4", "m3", "c4", "c3", "r4", "r3", "i3", "d2", "t2"]
	family == old_families[_]

	result := {
		"policy_id": "COST-003",
		"policy_name": "Previous Generation Instance",
		"msg": sprintf("EC2 instance '%s' uses previous generation '%s'. Newer types offer better price/performance", [input.resource.address, family]),
		"severity": "low",
		"remediation": sprintf("Consider upgrading %s to latest generation for better price/performance", [family]),
		"category": "cost",
		"frameworks": [],
	}
}
