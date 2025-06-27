package parser

import (
	"github.com/inayathulla/cloudrift/internal/models"
	"strings"
)

// ParseS3Buckets extracts aws_s3_bucket resources from a Terraform plan.
func ParseS3Buckets(plan *TerraformPlan) []models.S3Bucket {
	var buckets []models.S3Bucket

	for _, rc := range plan.ResourceChanges {
		if rc.Type != "aws_s3_bucket" {
			continue
		}

		after := rc.Change.After
		if after == nil {
			continue
		}

		bucket := models.S3Bucket{
			Id: rc.Address,
		}

		if name, ok := after["bucket"].(string); ok {
			bucket.Name = name
		}
		if acl, ok := after["acl"].(string); ok {
			bucket.Acl = acl
		}
		if tags, ok := after["tags"].(map[string]interface{}); ok {
			bucket.Tags = make(map[string]string)
			for k, v := range tags {
				if strVal, ok := v.(string); ok {
					bucket.Tags[k] = strVal
				}
			}
		}

		buckets = append(buckets, bucket)
	}

	return buckets
}

func contains(arr []string, target string) bool {
	for _, s := range arr {
		if strings.ToLower(s) == strings.ToLower(target) {
			return true
		}
	}
	return false
}
