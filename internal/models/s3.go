package models

// S3Bucket represents a simplified structure of an AWS S3 bucket from a Terraform plan.
type S3Bucket struct {
	Id   string            `yaml:"id"`   // Terraform address (e.g., aws_s3_bucket.bucket_name)
	Name string            `yaml:"name"` // Actual bucket name
	Acl  string            `yaml:"acl"`  // Access Control List value (e.g., private, public-read)
	Tags map[string]string `yaml:"tags"` // Key-value tags
}
