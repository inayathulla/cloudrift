package detector

import (
	"testing"

	"github.com/inayathulla/cloudrift/internal/detector"
	"github.com/inayathulla/cloudrift/internal/models"
	"github.com/stretchr/testify/assert"
)

// ACL positive and negative
func TestDetectS3Drift_ACL_Positive(t *testing.T) {
	plan := models.S3Bucket{Name: "b-acl", Acl: "private"}
	actual := &models.S3Bucket{Name: "b-acl", Acl: "public-read"}
	res := detector.DetectS3Drift(plan, actual)
	assert.True(t, res.AclDiff)
}

func TestDetectS3Drift_ACL_Negative(t *testing.T) {
	plan := models.S3Bucket{Name: "b-acl", Acl: "private"}
	actual := &models.S3Bucket{Name: "b-acl", Acl: "private"}
	res := detector.DetectS3Drift(plan, actual)
	assert.False(t, res.AclDiff)
}

// TagDiff positive and negative
func TestDetectS3Drift_TagDiff_Positive(t *testing.T) {
	plan := models.S3Bucket{Name: "b-tag", Tags: map[string]string{"k": "v1"}}
	actual := &models.S3Bucket{Name: "b-tag", Tags: map[string]string{"k": "v2"}}
	res := detector.DetectS3Drift(plan, actual)
	assert.Len(t, res.TagDiffs, 1)
}

func TestDetectS3Drift_TagDiff_Negative(t *testing.T) {
	plan := models.S3Bucket{Name: "b-tag", Tags: map[string]string{"k": "v"}}
	actual := &models.S3Bucket{Name: "b-tag", Tags: map[string]string{"k": "v"}}
	res := detector.DetectS3Drift(plan, actual)
	assert.Empty(t, res.TagDiffs)
}

// Versioning positive and negative
func TestDetectS3Drift_Versioning_Positive(t *testing.T) {
	plan := models.S3Bucket{Name: "b-ver", VersioningEnabled: false}
	actual := &models.S3Bucket{Name: "b-ver", VersioningEnabled: true}
	res := detector.DetectS3Drift(plan, actual)
	assert.True(t, res.VersioningDiff)
}

func TestDetectS3Drift_Versioning_Negative(t *testing.T) {
	plan := models.S3Bucket{Name: "b-ver", VersioningEnabled: true}
	actual := &models.S3Bucket{Name: "b-ver", VersioningEnabled: true}
	res := detector.DetectS3Drift(plan, actual)
	assert.False(t, res.VersioningDiff)
}

// Encryption positive and negative
func TestDetectS3Drift_Encryption_Positive(t *testing.T) {
	plan := models.S3Bucket{Name: "b-enc", EncryptionAlgorithm: "AES256"}
	actual := &models.S3Bucket{Name: "b-enc", EncryptionAlgorithm: "aws:kms"}
	res := detector.DetectS3Drift(plan, actual)
	assert.True(t, res.EncryptionDiff)
}

func TestDetectS3Drift_Encryption_Negative(t *testing.T) {
	plan := models.S3Bucket{Name: "b-enc", EncryptionAlgorithm: "AES256"}
	actual := &models.S3Bucket{Name: "b-enc", EncryptionAlgorithm: "AES256"}
	res := detector.DetectS3Drift(plan, actual)
	assert.False(t, res.EncryptionDiff)
}

// Logging positive and negative
func TestDetectS3Drift_Logging_Positive(t *testing.T) {
	plan := models.S3Bucket{Name: "b-log", LoggingEnabled: false}
	actual := &models.S3Bucket{
		Name:                "b-log",
		LoggingEnabled:      true,
		LoggingTargetBucket: "lb",
		LoggingTargetPrefix: "p/",
	}
	res := detector.DetectS3Drift(plan, actual)
	assert.True(t, res.LoggingDiff)
}

func TestDetectS3Drift_Logging_Negative(t *testing.T) {
	plan := models.S3Bucket{Name: "b-log", LoggingEnabled: false}
	actual := &models.S3Bucket{Name: "b-log", LoggingEnabled: false}
	res := detector.DetectS3Drift(plan, actual)
	assert.False(t, res.LoggingDiff)
}

// PublicAccessBlock positive and negative
func TestDetectS3Drift_PublicAccessBlock_Positive(t *testing.T) {
	plan := models.S3Bucket{
		Name: "b-pab",
		PublicAccessBlock: models.PublicAccessBlockConfig{
			BlockPublicAcls:       false,
			IgnorePublicAcls:      false,
			BlockPublicPolicy:     false,
			RestrictPublicBuckets: false,
		},
	}
	actual := &models.S3Bucket{
		Name: "b-pab",
		PublicAccessBlock: models.PublicAccessBlockConfig{
			BlockPublicAcls:       true,
			IgnorePublicAcls:      false,
			BlockPublicPolicy:     false,
			RestrictPublicBuckets: false,
		},
	}
	res := detector.DetectS3Drift(plan, actual)
	assert.True(t, res.PublicAccessBlockDiff)
}

func TestDetectS3Drift_PublicAccessBlock_Negative(t *testing.T) {
	cfg := models.PublicAccessBlockConfig{
		BlockPublicAcls:       true,
		IgnorePublicAcls:      true,
		BlockPublicPolicy:     true,
		RestrictPublicBuckets: true,
	}
	plan := models.S3Bucket{Name: "b-pab", PublicAccessBlock: cfg}
	actual := &models.S3Bucket{Name: "b-pab", PublicAccessBlock: cfg}
	res := detector.DetectS3Drift(plan, actual)
	assert.False(t, res.PublicAccessBlockDiff)
}

// LifecycleRules positive and negative
func TestDetectS3Drift_Lifecycle_Positive(t *testing.T) {
	plan := models.S3Bucket{
		Name: "b-lc",
		LifecycleRules: []models.LifecycleRuleSummary{
			{ID: "r1", Status: "Enabled", Prefix: "logs/", ExpirationDays: 90},
		},
	}
	actual := &models.S3Bucket{
		Name: "b-lc",
		LifecycleRules: []models.LifecycleRuleSummary{
			{ID: "r1", Status: "Enabled", Prefix: "logs/", ExpirationDays: 80},
		},
	}
	res := detector.DetectS3Drift(plan, actual)
	assert.True(t, res.LifecycleDiff)
}

func TestDetectS3Drift_Lifecycle_Negative(t *testing.T) {
	rules := []models.LifecycleRuleSummary{
		{ID: "r1", Status: "Enabled", Prefix: "", ExpirationDays: 30},
	}
	plan := models.S3Bucket{Name: "b-lc", LifecycleRules: rules}
	actual := &models.S3Bucket{Name: "b-lc", LifecycleRules: rules}
	res := detector.DetectS3Drift(plan, actual)
	assert.False(t, res.LifecycleDiff)
}
