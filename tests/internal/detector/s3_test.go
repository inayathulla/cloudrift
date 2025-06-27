package detector

import (
	"testing"

	"github.com/inayathulla/cloudrift/internal/detector"
	"github.com/inayathulla/cloudrift/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestDetectS3Drift_MissingBucket(t *testing.T) {
	plan := models.S3Bucket{
		Name: "test-bucket",
		Acl:  "private",
		Tags: map[string]string{"env": "prod"},
	}
	var actual *models.S3Bucket = nil

	result := detector.DetectS3Drift(plan, actual)

	assert.Equal(t, "test-bucket", result.BucketName)
	assert.True(t, result.Missing)
	assert.False(t, result.AclDiff)
	assert.Empty(t, result.TagDiffs)
}

func TestDetectS3Drift_AclAndTagMismatch(t *testing.T) {
	plan := models.S3Bucket{
		Name: "example-bucket",
		Acl:  "private",
		Tags: map[string]string{"owner": "team-a", "env": "prod"},
	}
	actual := &models.S3Bucket{
		Name: "example-bucket",
		Acl:  "public-read",
		Tags: map[string]string{"owner": "team-b", "env": "prod"},
	}

	result := detector.DetectS3Drift(plan, actual)

	assert.Equal(t, "example-bucket", result.BucketName)
	assert.False(t, result.Missing)
	assert.True(t, result.AclDiff)
	assert.Len(t, result.TagDiffs, 1)
	assert.Equal(t, [2]string{"team-a", "team-b"}, result.TagDiffs["owner"])
}

func TestDetectAllS3Drift(t *testing.T) {
	planBuckets := []models.S3Bucket{
		{Name: "bucket-1", Acl: "private", Tags: map[string]string{"env": "prod"}},
		{Name: "bucket-2", Acl: "private", Tags: map[string]string{}},
	}
	actualBuckets := []models.S3Bucket{
		{Name: "bucket-1", Acl: "private", Tags: map[string]string{"env": "prod"}},
		{Name: "bucket-2", Acl: "public-read", Tags: map[string]string{}},
	}

	results := detector.DetectAllS3Drift(planBuckets, actualBuckets)

	assert.Len(t, results, 1)
	assert.Equal(t, "bucket-2", results[0].BucketName)
	assert.True(t, results[0].AclDiff)
}
