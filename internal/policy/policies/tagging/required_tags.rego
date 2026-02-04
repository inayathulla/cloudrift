# Required Tags Policy
# Ensures resources have required tags for cost allocation and management

package cloudrift.tagging.required

# Default required tags - can be overridden by configuration
default_required_tags := ["Environment", "Owner", "Project"]

# Resources that should be tagged
taggable_resources := [
	"aws_s3_bucket",
	"aws_instance",
	"aws_security_group",
	"aws_db_instance",
	"aws_rds_cluster",
	"aws_lambda_function",
	"aws_ecs_cluster",
	"aws_eks_cluster",
	"aws_lb",
	"aws_vpc",
]

# Check for missing Environment tag
deny[result] {
	input.resource.type == taggable_resources[_]

	planned := input.resource.planned
	tags := object.get(planned, "tags", {})

	not tags.Environment
	not tags.environment
	not tags.env
	not tags.Env

	result := {
		"policy_id": "TAG-001",
		"policy_name": "Environment Tag Required",
		"msg": sprintf("Resource '%s' is missing required 'Environment' tag", [input.resource.address]),
		"severity": "medium",
		"remediation": "Add tags = { Environment = \"dev|staging|production\" } to the resource",
	}
}

# Check for missing Owner tag
warn[result] {
	input.resource.type == taggable_resources[_]

	planned := input.resource.planned
	tags := object.get(planned, "tags", {})

	not tags.Owner
	not tags.owner

	result := {
		"policy_id": "TAG-002",
		"policy_name": "Owner Tag Recommended",
		"msg": sprintf("Resource '%s' is missing 'Owner' tag for accountability", [input.resource.address]),
		"severity": "low",
		"remediation": "Add Owner tag with team or individual responsible for the resource",
	}
}

# Check for missing Project tag
warn[result] {
	input.resource.type == taggable_resources[_]

	planned := input.resource.planned
	tags := object.get(planned, "tags", {})

	not tags.Project
	not tags.project

	result := {
		"policy_id": "TAG-003",
		"policy_name": "Project Tag Recommended",
		"msg": sprintf("Resource '%s' is missing 'Project' tag for cost allocation", [input.resource.address]),
		"severity": "low",
		"remediation": "Add Project tag to enable cost allocation and tracking",
	}
}

# Check for missing Name tag (very common oversight)
warn[result] {
	input.resource.type == taggable_resources[_]

	planned := input.resource.planned
	tags := object.get(planned, "tags", {})

	not tags.Name
	not tags.name

	result := {
		"policy_id": "TAG-004",
		"policy_name": "Name Tag Recommended",
		"msg": sprintf("Resource '%s' is missing 'Name' tag", [input.resource.address]),
		"severity": "low",
		"remediation": "Add Name tag for easy identification in AWS console",
	}
}
