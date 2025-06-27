package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/inayathulla/cloudrift/internal/models"
)

// FetchS3Buckets returns the live state of all S3 buckets using the provided AWS config.
func FetchS3Buckets(cfg aws.Config) ([]models.S3Bucket, error) {
	ctx := context.Background()
	s3Client := s3.NewFromConfig(cfg)

	// List all buckets
	listOutput, err := s3Client.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list buckets: %w", err)
	}

	var buckets []models.S3Bucket
	for _, b := range listOutput.Buckets {
		if b.Name == nil {
			continue
		}

		liveState, err := FetchS3BucketState(*b.Name, cfg)
		if err != nil {
			fmt.Printf("⚠️ Warning: Could not fetch state for %s: %v\n", *b.Name, err)
			continue
		}
		buckets = append(buckets, *liveState)
	}

	return buckets, nil
}

// FetchS3BucketState returns ACL and Tags for a given bucket using the provided AWS config.
func FetchS3BucketState(bucketName string, cfg aws.Config) (*models.S3Bucket, error) {
	ctx := context.Background()
	s3Client := s3.NewFromConfig(cfg)

	// Get ACL
	aclResp, err := s3Client.GetBucketAcl(ctx, &s3.GetBucketAclInput{
		Bucket: &bucketName,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get ACL for bucket %s: %w", bucketName, err)
	}
	acl := aclToString(aclResp.Grants)

	// Get Tags
	tagResp, err := s3Client.GetBucketTagging(ctx, &s3.GetBucketTaggingInput{
		Bucket: &bucketName,
	})
	tags := make(map[string]string)
	if err == nil {
		for _, tag := range tagResp.TagSet {
			tags[*tag.Key] = *tag.Value
		}
	}
	fmt.Printf("🔍 Live bucket state for %s: tags=%+v acl=%s\n", bucketName, tags, acl)

	return &models.S3Bucket{
		Name: bucketName,
		Acl:  acl,
		Tags: tags,
	}, nil
}

// aclToString simplifies the ACL to either "private" or "public-read"
func aclToString(grants []types.Grant) string {
	for _, g := range grants {
		if g.Grantee != nil && g.Grantee.Type == types.TypeCanonicalUser {
			if g.Permission == types.PermissionFullControl {
				return "private"
			}
		}
	}
	return "public-read"
}
