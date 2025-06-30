package models

// S3Bucket represents a simplified structure of an AWS S3 bucket from a Terraform plan.
type S3Bucket struct {
	Id   string            `yaml:"id"`   // Terraform address (e.g., aws_s3_bucket.bucket_name)
	Name string            `yaml:"name"` // Actual bucket name
	Acl  string            `yaml:"acl"`  // Access Control List value (e.g., private, public-read)
	Tags map[string]string `yaml:"tags"` // Key-value tags

	// New metadata fields:
	VersioningEnabled   bool
	EncryptionAlgorithm string
	LoggingEnabled      bool
	LoggingTargetBucket string
	LoggingTargetPrefix string
	PublicAccessBlock   PublicAccessBlockConfig
	LifecycleRules      []LifecycleRuleSummary
}

// PublicAccessBlockConfig holds the Public Access Block settings for an S3 bucket.
type PublicAccessBlockConfig struct {
	BlockPublicAcls       bool
	IgnorePublicAcls      bool
	BlockPublicPolicy     bool
	RestrictPublicBuckets bool
}

// LifecycleRuleSummary is a simplified view of an S3 Lifecycle rule.
type LifecycleRuleSummary struct {
	ID             string // Rule identifier
	Status         string // "Enabled" or "Disabled"
	Prefix         string // Object key prefix
	ExpirationDays int    // Days until expiration
}
