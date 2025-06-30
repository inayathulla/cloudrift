package aws

import (
	"context"
	"fmt"

	sdkaws "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/inayathulla/cloudrift/internal/models"
)

// FetchS3Buckets lists all buckets and their ACL, tags, and core metadata.
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

	// ACL
	aclResp, err := client.GetBucketAcl(ctx, &s3.GetBucketAclInput{Bucket: &name})
	if err != nil {
		return nil, fmt.Errorf("GetBucketAcl: %w", err)
	}

	// Tags
	tags := map[string]string{}
	tagResp, err := client.GetBucketTagging(ctx, &s3.GetBucketTaggingInput{Bucket: &name})
	if err == nil {
		for _, t := range tagResp.TagSet {
			tags[*t.Key] = *t.Value
		}
	}

	// Base bucket object
	bucket := &models.S3Bucket{
		Name: name,
		Acl:  aclToString(aclResp.Grants),
		Tags: tags,
	}

	// 1. Versioning
	if verResp, err := client.GetBucketVersioning(ctx, &s3.GetBucketVersioningInput{Bucket: &name}); err == nil &&
		verResp.Status == types.BucketVersioningStatusEnabled {
		bucket.VersioningEnabled = true
	}

	// 2. Server-Side Encryption
	if encResp, err := client.GetBucketEncryption(ctx, &s3.GetBucketEncryptionInput{Bucket: &name}); err == nil &&
		len(encResp.ServerSideEncryptionConfiguration.Rules) > 0 {
		rule := encResp.ServerSideEncryptionConfiguration.Rules[0]
		bucket.EncryptionAlgorithm = string(rule.ApplyServerSideEncryptionByDefault.SSEAlgorithm)
	}

	// 3. Access Logging
	if logResp, err := client.GetBucketLogging(ctx, &s3.GetBucketLoggingInput{Bucket: &name}); err == nil &&
		logResp.LoggingEnabled != nil {
		bucket.LoggingEnabled = true
		bucket.LoggingTargetBucket = *logResp.LoggingEnabled.TargetBucket
		bucket.LoggingTargetPrefix = *logResp.LoggingEnabled.TargetPrefix
	}

	// 4. Public Access Block
	if pabResp, err := client.GetPublicAccessBlock(ctx, &s3.GetPublicAccessBlockInput{Bucket: &name}); err == nil &&
		pabResp.PublicAccessBlockConfiguration != nil {
		cfg := pabResp.PublicAccessBlockConfiguration
		bucket.PublicAccessBlock = models.PublicAccessBlockConfig{
			BlockPublicAcls:       cfg.BlockPublicAcls != nil && *cfg.BlockPublicAcls,
			IgnorePublicAcls:      cfg.IgnorePublicAcls != nil && *cfg.IgnorePublicAcls,
			BlockPublicPolicy:     cfg.BlockPublicPolicy != nil && *cfg.BlockPublicPolicy,
			RestrictPublicBuckets: cfg.RestrictPublicBuckets != nil && *cfg.RestrictPublicBuckets,
		}
	}

	// 5. Lifecycle Rules
	if lcResp, err := client.GetBucketLifecycleConfiguration(ctx, &s3.GetBucketLifecycleConfigurationInput{Bucket: &name}); err == nil {
		for _, r := range lcResp.Rules {
			bucket.LifecycleRules = append(bucket.LifecycleRules, models.LifecycleRuleSummary{
				ID:     *r.ID,
				Status: string(r.Status),
			})
		}
	}

	return bucket, nil
}

func aclToString(grants []types.Grant) string {
	for _, g := range grants {
		if g.Permission == types.PermissionFullControl {
			return "private"
		}
	}
	return "public-read"
}
