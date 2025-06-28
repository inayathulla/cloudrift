package aws

import (
	"context"
	"fmt"

	sdkaws "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/inayathulla/cloudrift/internal/models"
)

// FetchS3Buckets lists all buckets and their ACL/tags.
func FetchS3Buckets(cfg sdkaws.Config) ([]models.S3Bucket, error) {
	ctx := context.Background()
	client := s3.NewFromConfig(cfg)

	lst, err := client.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		return nil, fmt.Errorf("ListBuckets: %w", err)
	}

	var out []models.S3Bucket
	for _, b := range lst.Buckets {
		if b.Name == nil {
			continue
		}
		st, err := fetchBucketState(ctx, *b.Name, cfg)
		if err != nil {
			fmt.Printf("⚠️ bucket %s: %v\n", *b.Name, err)
			continue
		}
		out = append(out, *st)
	}
	return out, nil
}

func fetchBucketState(ctx context.Context, name string, cfg sdkaws.Config) (*models.S3Bucket, error) {
	client := s3.NewFromConfig(cfg)

	aclResp, err := client.GetBucketAcl(ctx, &s3.GetBucketAclInput{Bucket: &name})
	if err != nil {
		return nil, fmt.Errorf("GetBucketAcl: %w", err)
	}

	tags := map[string]string{}
	tagResp, err := client.GetBucketTagging(ctx, &s3.GetBucketTaggingInput{Bucket: &name})
	if err == nil {
		for _, t := range tagResp.TagSet {
			tags[*t.Key] = *t.Value
		}
	}

	return &models.S3Bucket{
		Name: name,
		Acl:  aclToString(aclResp.Grants),
		Tags: tags,
	}, nil
}

func aclToString(grants []types.Grant) string {
	for _, g := range grants {
		if g.Permission == types.PermissionFullControl {
			return "private"
		}
	}
	return "public-read"
}
