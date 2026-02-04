# S3 Public Access Policy
# Ensures S3 buckets block public access

package cloudrift.security.s3_public_access

deny[result] {
	input.resource.type == "aws_s3_bucket"

	# Check if public access block is not configured
	planned := input.resource.planned
	pab := planned.public_access_block

	# Block is not configured or not all settings are enabled
	not pab.block_public_acls

	result := {
		"policy_id": "S3-003",
		"policy_name": "S3 Block Public ACLs",
		"msg": sprintf("S3 bucket '%s' must have block_public_acls enabled", [input.resource.address]),
		"severity": "high",
		"remediation": "Set block_public_acls = true in aws_s3_bucket_public_access_block resource",
	}
}

deny[result] {
	input.resource.type == "aws_s3_bucket"

	planned := input.resource.planned
	pab := planned.public_access_block

	not pab.block_public_policy

	result := {
		"policy_id": "S3-004",
		"policy_name": "S3 Block Public Policy",
		"msg": sprintf("S3 bucket '%s' must have block_public_policy enabled", [input.resource.address]),
		"severity": "high",
		"remediation": "Set block_public_policy = true in aws_s3_bucket_public_access_block resource",
	}
}

deny[result] {
	input.resource.type == "aws_s3_bucket"

	planned := input.resource.planned
	pab := planned.public_access_block

	not pab.ignore_public_acls

	result := {
		"policy_id": "S3-005",
		"policy_name": "S3 Ignore Public ACLs",
		"msg": sprintf("S3 bucket '%s' must have ignore_public_acls enabled", [input.resource.address]),
		"severity": "high",
		"remediation": "Set ignore_public_acls = true in aws_s3_bucket_public_access_block resource",
	}
}

deny[result] {
	input.resource.type == "aws_s3_bucket"

	planned := input.resource.planned
	pab := planned.public_access_block

	not pab.restrict_public_buckets

	result := {
		"policy_id": "S3-006",
		"policy_name": "S3 Restrict Public Buckets",
		"msg": sprintf("S3 bucket '%s' must have restrict_public_buckets enabled", [input.resource.address]),
		"severity": "high",
		"remediation": "Set restrict_public_buckets = true in aws_s3_bucket_public_access_block resource",
	}
}

# Deny public ACL on bucket
deny[result] {
	input.resource.type == "aws_s3_bucket"

	planned := input.resource.planned
	planned.acl == "public-read"

	result := {
		"policy_id": "S3-007",
		"policy_name": "S3 No Public Read ACL",
		"msg": sprintf("S3 bucket '%s' has public-read ACL which is not allowed", [input.resource.address]),
		"severity": "critical",
		"remediation": "Change ACL to 'private' or use bucket policies for controlled access",
	}
}

deny[result] {
	input.resource.type == "aws_s3_bucket"

	planned := input.resource.planned
	planned.acl == "public-read-write"

	result := {
		"policy_id": "S3-008",
		"policy_name": "S3 No Public Read-Write ACL",
		"msg": sprintf("S3 bucket '%s' has public-read-write ACL which is extremely dangerous", [input.resource.address]),
		"severity": "critical",
		"remediation": "Immediately change ACL to 'private'. Public write access is a serious security risk",
	}
}
