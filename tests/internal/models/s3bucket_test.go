package models

import (
	"reflect"
	"testing"

	"github.com/inayathulla/cloudrift/internal/models"
)

func EqualS3Bucket(a, b models.S3Bucket) bool {
	if a.Name != b.Name || a.Acl != b.Acl {
		return false
	}
	return reflect.DeepEqual(a.Tags, b.Tags)
}

func TestS3Bucket_Equal(t *testing.T) {
	bucket1 := models.S3Bucket{
		Name: "tests-bucket",
		Acl:  "private",
		Tags: map[string]string{
			"env":  "dev",
			"team": "backend",
		},
	}

	bucket2 := models.S3Bucket{
		Name: "tests-bucket",
		Acl:  "private",
		Tags: map[string]string{
			"env":  "dev",
			"team": "backend",
		},
	}

	if !EqualS3Bucket(bucket1, bucket2) {
		t.Errorf("Expected buckets to be equal")
	}
}

func TestS3Bucket_NotEqual_Name(t *testing.T) {
	bucket1 := models.S3Bucket{Name: "bucket-a", Acl: "private"}
	bucket2 := models.S3Bucket{Name: "bucket-b", Acl: "private"}

	if EqualS3Bucket(bucket1, bucket2) {
		t.Errorf("Expected buckets with different names to be unequal")
	}
}

func TestS3Bucket_NotEqual_Tags(t *testing.T) {
	bucket1 := models.S3Bucket{
		Name: "bucket-a",
		Acl:  "private",
		Tags: map[string]string{"env": "prod"},
	}
	bucket2 := models.S3Bucket{
		Name: "bucket-a",
		Acl:  "private",
		Tags: map[string]string{"env": "dev"},
	}

	if EqualS3Bucket(bucket1, bucket2) {
		t.Errorf("Expected buckets with different tags to be unequal")
	}
}
