# S3 Versioning Policy
# Recommends versioning for data protection

package cloudrift.security.s3_versioning

warn[result] {
	input.resource.type == "aws_s3_bucket"

	planned := input.resource.planned
	not planned.versioning_enabled

	result := {
		"policy_id": "S3-009",
		"policy_name": "S3 Versioning Recommended",
		"msg": sprintf("S3 bucket '%s' does not have versioning enabled. Versioning protects against accidental deletions", [input.resource.address]),
		"severity": "medium",
		"remediation": "Enable versioning in aws_s3_bucket_versioning resource",
	}
}
