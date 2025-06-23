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
