// internal/detector/s3.go
package detector

import (
	"fmt"
	"reflect"

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

	// Logging diff (only if plan configured any logging settings)
	if plan.LoggingEnabled || plan.LoggingTargetBucket != "" || plan.LoggingTargetPrefix != "" {
		if plan.LoggingEnabled != actual.LoggingEnabled ||
			plan.LoggingTargetBucket != actual.LoggingTargetBucket ||
			plan.LoggingTargetPrefix != actual.LoggingTargetPrefix {
			res.LoggingDiff = true
		}
	}

	// Public Access Block diff (only if plan specified any PAB settings)
	defaultPAB := models.PublicAccessBlockConfig{}
	if plan.PublicAccessBlock != defaultPAB {
		if !reflect.DeepEqual(plan.PublicAccessBlock, actual.PublicAccessBlock) {
			res.PublicAccessBlockDiff = true
		}
	}

	// Lifecycle rules diff (only if plan defined any rules)
	if len(plan.LifecycleRules) > 0 {
		if !reflect.DeepEqual(plan.LifecycleRules, actual.LifecycleRules) {
			res.LifecycleDiff = true
		}
	}

	return res
}

// DetectAllS3Drift runs detection across all buckets.
func DetectAllS3Drift(plans, lives []models.S3Bucket) []DriftResult {
	m := map[string]*models.S3Bucket{}
	for i := range lives {
		b := lives[i]
		m[b.Name] = &b
	}

	var out []DriftResult
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
