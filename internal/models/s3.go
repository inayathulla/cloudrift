// Package models defines the data structures used throughout Cloudrift
// for representing AWS resources and their configurations.
//
// These models serve as the common language between the Terraform plan parser,
// AWS API fetchers, and drift detection logic.
package models

// S3Bucket represents the configuration state of an AWS S3 bucket.
// It captures both Terraform-planned attributes and live AWS state,
// enabling attribute-level drift comparison.
type S3Bucket struct {
	// Id is the Terraform resource address (e.g., "aws_s3_bucket.my_bucket").
	Id string `yaml:"id"`

	// Name is the actual S3 bucket name in AWS.
	Name string `yaml:"name"`

	// Acl is the canned ACL applied to the bucket (e.g., "private", "public-read").
	Acl string `yaml:"acl"`

	// Tags contains the key-value metadata tags associated with the bucket.
	Tags map[string]string `yaml:"tags"`

	// VersioningEnabled indicates whether object versioning is enabled.
	VersioningEnabled bool

	// EncryptionAlgorithm specifies the server-side encryption algorithm
	// (e.g., "AES256", "aws:kms"). Empty string means no encryption configured.
	EncryptionAlgorithm string

	// LoggingEnabled indicates whether server access logging is enabled.
	LoggingEnabled bool

	// LoggingTargetBucket is the destination bucket for access logs.
	LoggingTargetBucket string

	// LoggingTargetPrefix is the key prefix for log objects.
	LoggingTargetPrefix string

	// PublicAccessBlock contains the S3 Block Public Access settings.
	PublicAccessBlock PublicAccessBlockConfig

	// LifecycleRules contains the object lifecycle management rules.
	LifecycleRules []LifecycleRuleSummary
}

// PublicAccessBlockConfig represents the S3 Block Public Access configuration.
// These settings help prevent accidental public exposure of bucket contents.
type PublicAccessBlockConfig struct {
	// BlockPublicAcls blocks public access granted by new ACLs.
	BlockPublicAcls bool

	// IgnorePublicAcls ignores all public ACLs on the bucket and its objects.
	IgnorePublicAcls bool

	// BlockPublicPolicy blocks public access granted by new bucket policies.
	BlockPublicPolicy bool

	// RestrictPublicBuckets restricts access to principals with access
	// granted by bucket policies.
	RestrictPublicBuckets bool
}

// LifecycleRuleSummary represents a simplified view of an S3 lifecycle rule.
// It captures the essential attributes needed for drift detection.
type LifecycleRuleSummary struct {
	// ID is the unique identifier for the lifecycle rule.
	ID string

	// Status indicates whether the rule is "Enabled" or "Disabled".
	Status string

	// Prefix is the object key prefix that identifies objects subject to this rule.
	Prefix string

	// ExpirationDays is the number of days after creation when objects expire.
	ExpirationDays int
}
