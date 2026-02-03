package detector

import (
	"fmt"

	sdkaws "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/inayathulla/cloudrift/internal/aws"
	"github.com/inayathulla/cloudrift/internal/models"
)

// DriftResult captures the drift detection results for a single S3 bucket.
//
// Each field indicates whether a specific attribute differs between the
// Terraform plan and the live AWS state. The struct is designed to provide
// granular drift information for detailed reporting.
type DriftResult struct {
	// BucketName is the name of the S3 bucket being compared.
	BucketName string

	// Missing is true if the bucket exists in the plan but not in AWS.
	Missing bool

	// AclDiff is true if the bucket ACL differs.
	AclDiff bool

	// TagDiffs maps tag keys to [expected, actual] value pairs for mismatched tags.
	TagDiffs map[string][2]string

	// ExtraTags contains tags present in AWS but not in the plan.
	ExtraTags map[string]string

	// VersioningDiff is true if versioning configuration differs.
	VersioningDiff bool

	// EncryptionDiff is true if encryption algorithm differs.
	EncryptionDiff bool

	// LoggingDiff is true if access logging configuration differs.
	LoggingDiff bool

	// PublicAccessBlockDiff is true if public access block settings differ.
	PublicAccessBlockDiff bool

	// LifecycleDiff is true if lifecycle rules differ.
	LifecycleDiff bool
}

// S3DriftDetector implements drift detection for AWS S3 buckets.
//
// It fetches live bucket state from AWS and compares it against
// Terraform-planned configurations to identify configuration drift.
type S3DriftDetector struct {
	cfg sdkaws.Config
}

// NewS3DriftDetector creates a new S3 drift detector with the given AWS configuration.
func NewS3DriftDetector(cfg sdkaws.Config) *S3DriftDetector {
	return &S3DriftDetector{cfg: cfg}
}

// FetchLiveState retrieves the current state of all S3 buckets from AWS.
//
// Returns:
//   - interface{}: []models.S3Bucket containing all bucket configurations
//   - error: if the AWS API call fails
func (d *S3DriftDetector) FetchLiveState() (interface{}, error) {
	return aws.FetchS3Buckets(d.cfg)
}

// DetectDrift compares Terraform-planned bucket configurations against live AWS state.
//
// Parameters:
//   - plan: []models.S3Bucket from the Terraform plan
//   - live: []models.S3Bucket from the live AWS state
//
// Returns:
//   - []DriftResult: drift results for buckets with detected differences
//   - error: if type assertions fail
func (d *S3DriftDetector) DetectDrift(plan, live interface{}) ([]DriftResult, error) {
	plans, ok := plan.([]models.S3Bucket)
	if !ok {
		return nil, fmt.Errorf("plan type mismatch")
	}
	lives, ok := live.([]models.S3Bucket)
	if !ok {
		return nil, fmt.Errorf("live type mismatch")
	}
	return DetectAllS3Drift(plans, lives), nil
}

// DetectS3Drift compares a single planned bucket against its actual AWS state.
//
// The comparison checks the following attributes:
//   - ACL
//   - Tags (missing, mismatched, and extra)
//   - Versioning
//   - Encryption (only if plan specifies an algorithm)
//   - Logging configuration
//   - Public Access Block settings
//   - Lifecycle rules (only if plan defines rules)
//
// Parameters:
//   - plan: the Terraform-planned bucket configuration
//   - actual: the live bucket configuration (nil if bucket doesn't exist)
//
// Returns:
//   - DriftResult: detailed drift information for the bucket
func DetectS3Drift(plan models.S3Bucket, actual *models.S3Bucket) DriftResult {
	res := DriftResult{
		BucketName: plan.Name,
		TagDiffs:   make(map[string][2]string),
		ExtraTags:  make(map[string]string),
	}
	if actual == nil {
		res.Missing = true
		return res
	}

	// ACL diff
	if plan.Acl != actual.Acl {
		res.AclDiff = true
	}

	// Tag diffs
	for k, v := range plan.Tags {
		if av, ok := actual.Tags[k]; !ok || av != v {
			res.TagDiffs[k] = [2]string{v, av}
		}
	}
	// Extra tags
	for k, av := range actual.Tags {
		if _, ok := plan.Tags[k]; !ok {
			res.ExtraTags[k] = av
		}
	}

	// Versioning diff (always considered explicit)
	if plan.VersioningEnabled != actual.VersioningEnabled {
		res.VersioningDiff = true
	}

	// Encryption diff (only if plan specified an algorithm)
	if plan.EncryptionAlgorithm != "" && plan.EncryptionAlgorithm != actual.EncryptionAlgorithm {
		res.EncryptionDiff = true
	}

	// Logging diff — compare regardless of plan fields
	if plan.LoggingEnabled != actual.LoggingEnabled ||
		plan.LoggingTargetBucket != actual.LoggingTargetBucket ||
		plan.LoggingTargetPrefix != actual.LoggingTargetPrefix {
		res.LoggingDiff = true
	}

	// Public Access Block diff — compare unconditionally
	if !publicAccessBlockEqual(plan.PublicAccessBlock, actual.PublicAccessBlock) {
		res.PublicAccessBlockDiff = true
	}

	// Lifecycle rules diff (only if plan defined any rules)
	if len(plan.LifecycleRules) > 0 {
		if !lifecycleRulesEqual(plan.LifecycleRules, actual.LifecycleRules) {
			res.LifecycleDiff = true
		}
	}

	return res
}

// DetectAllS3Drift performs drift detection across all planned S3 buckets.
//
// This function builds a lookup map of live buckets for O(1) access, then
// compares each planned bucket against its live counterpart. Only buckets
// with actual drift are included in the results.
//
// Parameters:
//   - plans: slice of S3 bucket configurations from the Terraform plan
//   - lives: slice of S3 bucket configurations from live AWS state
//
// Returns:
//   - []DriftResult: drift results for buckets with detected differences
func DetectAllS3Drift(plans, lives []models.S3Bucket) []DriftResult {
	m := make(map[string]*models.S3Bucket, len(lives))
	for i := range lives {
		m[lives[i].Name] = &lives[i]
	}

	out := make([]DriftResult, 0, len(plans))
	for _, p := range plans {
		dr := DetectS3Drift(p, m[p.Name])
		if dr.Missing ||
			dr.AclDiff ||
			len(dr.TagDiffs) > 0 ||
			len(dr.ExtraTags) > 0 ||
			dr.VersioningDiff ||
			dr.EncryptionDiff ||
			dr.LoggingDiff ||
			dr.PublicAccessBlockDiff ||
			dr.LifecycleDiff {
			out = append(out, dr)
		}
	}
	return out
}

// publicAccessBlockEqual compares two PublicAccessBlockConfig structs.
func publicAccessBlockEqual(a, b models.PublicAccessBlockConfig) bool {
	return a.BlockPublicAcls == b.BlockPublicAcls &&
		a.IgnorePublicAcls == b.IgnorePublicAcls &&
		a.BlockPublicPolicy == b.BlockPublicPolicy &&
		a.RestrictPublicBuckets == b.RestrictPublicBuckets
}

// lifecycleRulesEqual compares two slices of LifecycleRuleSummary.
func lifecycleRulesEqual(a, b []models.LifecycleRuleSummary) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].ID != b[i].ID ||
			a[i].Status != b[i].Status ||
			a[i].Prefix != b[i].Prefix ||
			a[i].ExpirationDays != b[i].ExpirationDays {
			return false
		}
	}
	return true
}
