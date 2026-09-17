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

# --------------------------------------------------------------------------
# Helpers
#
# All helpers guard their inputs. Terraform plan JSON is not a trusted schema:
# instance_type may be absent (AMI-derived, or set by a launch template) or an
# unresolved value, so every rule must stay undefined rather than error on it.
# --------------------------------------------------------------------------

# The declared instance type, only when it is a usable string.
instance_type_of(planned) := instance_type {
	instance_type := planned.instance_type
	is_string(instance_type)
	instance_type != ""
}

# "m5.2xlarge" -> "m5"
instance_family(instance_type) := family {
	parts := split(instance_type, ".")
	family := parts[0]
}

# "m5.2xlarge" -> "2xlarge"
instance_size(instance_type) := size {
	parts := split(instance_type, ".")
	count(parts) == 2
	size := parts[1]
}

# Relative ordering of EC2 sizes, used to tell an upsize from a downsize.
size_rank := {
	"nano": 1,
	"micro": 2,
	"small": 3,
	"medium": 4,
	"large": 5,
	"xlarge": 6,
	"2xlarge": 7,
	"3xlarge": 8,
	"4xlarge": 9,
	"6xlarge": 10,
	"8xlarge": 11,
	"9xlarge": 12,
	"10xlarge": 13,
	"12xlarge": 14,
	"16xlarge": 15,
	"18xlarge": 16,
	"24xlarge": 17,
	"32xlarge": 18,
	"48xlarge": 19,
	"112xlarge": 20,
	"metal": 21,
}

burstable_families := ["t2", "t3", "t3a", "t4g"]

# credit_specification is a Terraform block: a list of objects in plan JSON,
# but a bare object in some state shapes. Accept either.
unlimited_credits(planned) {
	planned.credit_specification[_].cpu_credits == "unlimited"
}

unlimited_credits(planned) {
	planned.credit_specification.cpu_credits == "unlimited"
}

# --------------------------------------------------------------------------
# COST-004: Accelerated computing (GPU / ML accelerator) instances
# --------------------------------------------------------------------------
warn[result] {
	input.resource.type == "aws_instance"

	instance_type := instance_type_of(input.resource.planned)
	family := instance_family(instance_type)

	accelerated_families := ["p2", "p3", "p3dn", "p4d", "p4de", "p5", "p5e", "p5en", "g3", "g3s", "g4ad", "g4dn", "g5", "g5g", "g6", "g6e", "gr6", "inf1", "inf2", "trn1", "trn1n", "trn2", "dl1", "dl2q"]
	family == accelerated_families[_]

	result := {
		"policy_id": "COST-004",
		"policy_name": "Accelerated Computing Instance",
		"msg": sprintf("EC2 instance '%s' uses accelerator type '%s'. GPU families are the most expensive on EC2 and bill at the same rate while idle", [input.resource.address, instance_type]),
		"severity": "high",
		"remediation": "Confirm the workload saturates the accelerator. Run training on Spot capacity, stop instances between jobs, and prefer a managed endpoint (SageMaker, Bedrock) for intermittent inference",
		"category": "cost",
		"frameworks": [],
	}
}

# --------------------------------------------------------------------------
# COST-005: Burstable instance running in unlimited credit mode
# --------------------------------------------------------------------------
warn[result] {
	input.resource.type == "aws_instance"

	planned := input.resource.planned
	instance_type := instance_type_of(planned)
	instance_family(instance_type) == burstable_families[_]

	unlimited_credits(planned)

	result := {
		"policy_id": "COST-005",
		"policy_name": "Unlimited Burst Credits",
		"msg": sprintf("EC2 instance '%s' runs '%s' with cpu_credits = \"unlimited\". CPU above the baseline bills as surplus credits with no ceiling", [input.resource.address, instance_type]),
		"severity": "medium",
		"remediation": "Set cpu_credits = \"standard\" to cap spend at the baseline, or move a consistently busy workload to a fixed-performance family (m6i, c6i) where the bill is predictable",
		"category": "cost",
		"frameworks": [],
	}
}

# --------------------------------------------------------------------------
# COST-006: Dedicated / host tenancy
# --------------------------------------------------------------------------
warn[result] {
	input.resource.type == "aws_instance"

	tenancy := object.get(input.resource.planned, "tenancy", "default")

	dedicated_tenancies := ["dedicated", "host"]
	tenancy == dedicated_tenancies[_]

	result := {
		"policy_id": "COST-006",
		"policy_name": "Dedicated Tenancy Instance",
		"msg": sprintf("EC2 instance '%s' uses '%s' tenancy, which carries a large premium over shared tenancy", [input.resource.address, tenancy]),
		"severity": "medium",
		"remediation": "Use tenancy = \"default\" unless BYOL licensing or a single-tenant compliance requirement applies. Dedicated hosts also bill per host whether or not instances run on them",
		"category": "cost",
		"frameworks": [],
	}
}

# --------------------------------------------------------------------------
# COST-007: Instance resized outside of Terraform
#
# Evaluated against live AWS state rather than the plan alone. An unmanaged
# upsize is unbudgeted spend that also reverts on the next apply, stopping and
# starting the instance under the load that prompted the resize.
# --------------------------------------------------------------------------
warn[result] {
	[declared, live] := instance_type_drift

	upsized(declared, live)

	result := {
		"policy_id": "COST-007",
		"policy_name": "Instance Resized Outside Terraform",
		"msg": sprintf("EC2 instance '%s' was resized up from '%s' to '%s' outside of Terraform. This spend is unbudgeted and reverts on the next apply", [input.resource.address, declared, live]),
		"severity": "high",
		"remediation": sprintf("Decide which size is correct: set instance_type = \"%s\" in Terraform to keep the larger instance, or re-apply to restore \"%s\". Schedule the apply deliberately, since the resize stops and starts the instance", [live, declared]),
		"category": "cost",
		"frameworks": [],
	}
}

warn[result] {
	[declared, live] := instance_type_drift

	not upsized(declared, live)

	result := {
		"policy_id": "COST-007",
		"policy_name": "Instance Resized Outside Terraform",
		"msg": sprintf("EC2 instance '%s' is running '%s' but Terraform declares '%s'. The live instance was changed outside of Terraform", [input.resource.address, live, declared]),
		"severity": "low",
		"remediation": "Reconcile the change: update instance_type in Terraform to match the live instance, or re-apply to restore the declared type",
		"category": "cost",
		"frameworks": [],
	}
}

# [declared, live] instance types, only for an existing instance whose type
# actually drifted. A missing instance is a separate finding, not a resize.
instance_type_drift := [declared, live] {
	input.resource.type == "aws_instance"

	drift := input.resource.drift
	not drift.missing

	diff := drift.diffs.instance_type
	declared := diff[0]
	live := diff[1]

	is_string(declared)
	is_string(live)
	declared != live
}

# True when the live instance type is a larger size than the declared one.
# Ranks are comparable across families, so c5.large -> r5.2xlarge is an upsize.
upsized(declared, live) {
	declared_rank := size_rank[instance_size(declared)]
	live_rank := size_rank[instance_size(live)]
	live_rank > declared_rank
}
