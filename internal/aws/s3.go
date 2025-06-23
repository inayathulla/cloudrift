package aws

import (
	"context"
	"fmt"
	"github.com/inayathulla/cloudrift/internal/models"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// FetchS3Buckets returns the live state of all S3 buckets in the account.
func FetchS3Buckets() ([]models.S3Bucket, error) {
	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to load AWS config: %w", err)
	}

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

		liveState, err := FetchS3BucketState(*b.Name)
		if err != nil {
			fmt.Printf("⚠️ Warning: Could not fetch state for %s: %v\n", *b.Name, err)
			continue
		}
		buckets = append(buckets, *liveState)
	}

	return buckets, nil
}

// FetchS3BucketState returns ACL and Tags for a given bucket from AWS.
func FetchS3BucketState(bucketName string) (*models.S3Bucket, error) {
	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to load AWS config: %w", err)
	}

	s3Client := s3.NewFromConfig(cfg)

	// Get ACL
	aclResp, err := s3Client.GetBucketAcl(ctx, &s3.GetBucketAclInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get ACL for bucket %s: %w", bucketName, err)
	}
	acl := aclToString(aclResp.Grants)

	// Get Tags
	tagResp, err := s3Client.GetBucketTagging(ctx, &s3.GetBucketTaggingInput{
		Bucket: aws.String(bucketName),
	})
	tags := make(map[string]string)
	if err == nil {
		for _, tag := range tagResp.TagSet {
			tags[*tag.Key] = *tag.Value
		}
	}

	return &models.S3Bucket{
		Name: bucketName,
		Acl:  acl,
		Tags: tags,
	}, nil
}

func aclToString(grants []types.Grant) string {
	// Simplify to owner ACL only (expand logic later if needed)
	for _, g := range grants {
		if g.Grantee != nil && g.Grantee.Type == types.TypeCanonicalUser {
			if g.Permission == types.PermissionFullControl {
				return "private"
			}
		}
	}
	return "public-read"
}
