package aws

import (
	"context"
	"errors"
	"fmt"

	sdkaws "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
	"golang.org/x/sync/errgroup"

	"github.com/inayathulla/cloudrift/internal/models"
)

// FetchS3Buckets lists all buckets and their metadata in parallel.
func FetchS3Buckets(cfg sdkaws.Config) ([]models.S3Bucket, error) {
	ctx := context.Background()
	client := s3.NewFromConfig(cfg)

	lst, err := client.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		return nil, fmt.Errorf("ListBuckets: %w", err)
	}

	// Pre-allocate output slice
	out := make([]models.S3Bucket, 0, len(lst.Buckets))
	for _, b := range lst.Buckets {
		if b.Name == nil {
			continue
		}
		st, err := fetchBucketState(ctx, *b.Name, client)
		if err != nil {
			fmt.Printf("⚠️ bucket %s: %v\n", *b.Name, err)
			continue
		}
		out = append(out, *st)
	}
	return out, nil
}

// fetchBucketState retrieves all relevant metadata for one bucket in parallel.
func fetchBucketState(ctx context.Context, name string, client *s3.Client) (*models.S3Bucket, error) {
	var (
		aclResp *s3.GetBucketAclOutput
		tagResp *s3.GetBucketTaggingOutput
		verResp *s3.GetBucketVersioningOutput
		encResp *s3.GetBucketEncryptionOutput
		logResp *s3.GetBucketLoggingOutput
		pabResp *s3.GetPublicAccessBlockOutput
		lcResp  *s3.GetBucketLifecycleConfigurationOutput
	)

	g, ctx := errgroup.WithContext(ctx)

	// 1) ACL
	g.Go(func() error {
		var err error
		aclResp, err = client.GetBucketAcl(ctx, &s3.GetBucketAclInput{Bucket: &name})
		return err
	})

	// 2) Tags (ignore missing tag-set)
	g.Go(func() error {
		var err error
		tagResp, err = client.GetBucketTagging(ctx, &s3.GetBucketTaggingInput{Bucket: &name})
		var apiErr smithy.APIError
		if errors.As(err, &apiErr) && apiErr.ErrorCode() == "NoSuchTagSet" {
			return nil
		}
		return err
	})

	// 3) Versioning
	g.Go(func() error {
		var err error
		verResp, err = client.GetBucketVersioning(ctx, &s3.GetBucketVersioningInput{Bucket: &name})
		return err
	})

	// 4) Encryption (ignore if not configured)
	g.Go(func() error {
		var err error
		encResp, err = client.GetBucketEncryption(ctx, &s3.GetBucketEncryptionInput{Bucket: &name})
		var apiErr smithy.APIError
		if errors.As(err, &apiErr) && apiErr.ErrorCode() == "ServerSideEncryptionConfigurationNotFoundError" {
			return nil
		}
		return err
	})

	// 5) Logging
	g.Go(func() error {
		var err error
		logResp, err = client.GetBucketLogging(ctx, &s3.GetBucketLoggingInput{Bucket: &name})
		return err
	})

	// 6) Public Access Block (ignore if absent)
	g.Go(func() error {
		var err error
		pabResp, err = client.GetPublicAccessBlock(ctx, &s3.GetPublicAccessBlockInput{Bucket: &name})
		var apiErr smithy.APIError
		if errors.As(err, &apiErr) && apiErr.ErrorCode() == "NoSuchPublicAccessBlockConfiguration" {
			return nil
		}
		return err
	})

	// 7) Lifecycle Rules (ignore if absent)
	g.Go(func() error {
		var err error
		lcResp, err = client.GetBucketLifecycleConfiguration(ctx, &s3.GetBucketLifecycleConfigurationInput{Bucket: &name})
		var apiErr smithy.APIError
		if errors.As(err, &apiErr) && apiErr.ErrorCode() == "NoSuchLifecycleConfiguration" {
			return nil
		}
		return err
	})

	// Wait for all calls to complete
	if err := g.Wait(); err != nil {
		return nil, err
	}

	// Build tags map, if any
	tags := make(map[string]string)
	if tagResp != nil {
		tags = make(map[string]string, len(tagResp.TagSet))
		for _, t := range tagResp.TagSet {
			if t.Key != nil && t.Value != nil {
				tags[*t.Key] = *t.Value
			}
		}
	}

	// Assemble the bucket model
	bucket := &models.S3Bucket{
		Name: name,
		Acl:  aclToString(aclResp.Grants),
		Tags: tags,
	}

	// Versioning
	if verResp != nil && verResp.Status == types.BucketVersioningStatusEnabled {
		bucket.VersioningEnabled = true
	}

	// Encryption
	if encResp != nil && len(encResp.ServerSideEncryptionConfiguration.Rules) > 0 {
		rule := encResp.ServerSideEncryptionConfiguration.Rules[0]
		bucket.EncryptionAlgorithm = string(rule.ApplyServerSideEncryptionByDefault.SSEAlgorithm)
	}

	// Logging
	if logResp != nil && logResp.LoggingEnabled != nil {
		bucket.LoggingEnabled = true
		bucket.LoggingTargetBucket = *logResp.LoggingEnabled.TargetBucket
		bucket.LoggingTargetPrefix = *logResp.LoggingEnabled.TargetPrefix
	}

	// Public Access Block
	if pabResp != nil && pabResp.PublicAccessBlockConfiguration != nil {
		cfg := pabResp.PublicAccessBlockConfiguration
		bucket.PublicAccessBlock = models.PublicAccessBlockConfig{
			BlockPublicAcls:       cfg.BlockPublicAcls != nil && *cfg.BlockPublicAcls,
			IgnorePublicAcls:      cfg.IgnorePublicAcls != nil && *cfg.IgnorePublicAcls,
			BlockPublicPolicy:     cfg.BlockPublicPolicy != nil && *cfg.BlockPublicPolicy,
			RestrictPublicBuckets: cfg.RestrictPublicBuckets != nil && *cfg.RestrictPublicBuckets,
		}
	}

	// Lifecycle Rules
	if lcResp != nil {
		for _, r := range lcResp.Rules {
			if r.ID != nil {
				bucket.LifecycleRules = append(bucket.LifecycleRules, models.LifecycleRuleSummary{
					ID:     *r.ID,
					Status: string(r.Status),
				})
			}
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
