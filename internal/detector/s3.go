package detector

import (
	"github.com/inayathulla/cloudrift/internal/models"
)

// DriftResult holds the differences between plan and actual state.
type DriftResult struct {
	BucketName string
	Missing    bool
	AclDiff    bool
	TagDiffs   map[string][2]string // [planned, actual]
}

// DetectS3Drift compares the Terraform plan bucket against the actual AWS bucket.
func DetectS3Drift(plan models.S3Bucket, actual *models.S3Bucket) DriftResult {
	result := DriftResult{
		BucketName: plan.Name,
		TagDiffs:   make(map[string][2]string),
	}

	if actual == nil {
		result.Missing = true
		return result
	}

	if plan.Acl != actual.Acl {
		result.AclDiff = true
	}

	// Compare tags (shallow diff)
	for k, planVal := range plan.Tags {
		if actualVal, ok := actual.Tags[k]; !ok || actualVal != planVal {
			result.TagDiffs[k] = [2]string{planVal, actualVal}
		}
	}

	return result
}

// DetectAllS3Drift processes multiple S3 buckets and returns all detected drifts.
func DetectAllS3Drift(planBuckets []models.S3Bucket, actualBuckets []models.S3Bucket) []DriftResult {
	results := []DriftResult{}

	// Map actual buckets by name for quick lookup
	actualMap := make(map[string]*models.S3Bucket)
	for _, b := range actualBuckets {
		b := b // avoid pointer capture issue in loop
		actualMap[b.Name] = &b
	}

	for _, plan := range planBuckets {
		var actual *models.S3Bucket
		if a, ok := actualMap[plan.Name]; ok {
			actual = a
		}
		result := DetectS3Drift(plan, actual)
		if result.Missing || result.AclDiff || len(result.TagDiffs) > 0 {
			results = append(results, result)
		}
	}

	return results
}
