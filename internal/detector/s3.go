package detector

import (
	"fmt"

	sdkaws "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/inayathulla/cloudrift/internal/aws"
	"github.com/inayathulla/cloudrift/internal/models"
)

// DriftResult holds drift info for one bucket.
type DriftResult struct {
	BucketName            string
	Missing               bool
	AclDiff               bool
	TagDiffs              map[string][2]string
	ExtraTags             map[string]string
	VersioningDiff        bool
	EncryptionDiff        bool
	LoggingDiff           bool
	PublicAccessBlockDiff bool
	LifecycleDiff         bool
}

// S3DriftDetector implements DriftDetector.
type S3DriftDetector struct {
	cfg sdkaws.Config
}

// NewS3DriftDetector constructs a detector.
func NewS3DriftDetector(cfg sdkaws.Config) *S3DriftDetector {
	return &S3DriftDetector{cfg: cfg}
}

// FetchLiveState returns []models.S3Bucket as interface{}.
func (d *S3DriftDetector) FetchLiveState() (interface{}, error) {
	return aws.FetchS3Buckets(d.cfg)
}

// DetectDrift compares plan vs live.
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

// DetectS3Drift compares one bucket.
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

// DetectAllS3Drift runs detection across all buckets.
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
