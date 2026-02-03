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

// FetchS3Buckets retrieves all S3 buckets and their configurations from AWS.
//
// This function lists all buckets in the account and fetches detailed metadata
// for each bucket including ACL, tags, versioning, encryption, logging,
// public access block settings, and lifecycle rules.
//
// Buckets that fail to fetch (e.g., due to permissions) are logged and skipped
// rather than causing the entire operation to fail.
//
// Parameters:
//   - cfg: AWS SDK configuration for API calls
//
// Returns:
//   - []models.S3Bucket: slice of bucket configurations
//   - error: if the ListBuckets call fails
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

// fetchBucketState retrieves all configuration attributes for a single S3 bucket.
//
// This function makes 7 parallel API calls to fetch:
//   - ACL (GetBucketAcl)
//   - Tags (GetBucketTagging)
//   - Versioning (GetBucketVersioning)
//   - Encryption (GetBucketEncryption)
//   - Logging (GetBucketLogging)
//   - Public Access Block (GetPublicAccessBlock)
//   - Lifecycle Rules (GetBucketLifecycleConfiguration)
//
// Expected "not found" errors (e.g., NoSuchTagSet, NoSuchLifecycleConfiguration)
// are gracefully handled and don't cause failures.
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
	var tags map[string]string
	if tagResp != nil && len(tagResp.TagSet) > 0 {
		tags = make(map[string]string, len(tagResp.TagSet))
		for _, t := range tagResp.TagSet {
			if t.Key != nil && t.Value != nil {
				tags[*t.Key] = *t.Value
			}
		}
	} else {
		tags = make(map[string]string)
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

// aclToString converts S3 grants to a simplified ACL string.
//
// This is a simplified interpretation: if any grant has FULL_CONTROL permission,
// the bucket is considered "private"; otherwise, it's "public-read".
// This heuristic works for common cases but may not capture all ACL nuances.
func aclToString(grants []types.Grant) string {
	for _, g := range grants {
		if g.Permission == types.PermissionFullControl {
			return "private"
		}
	}
	return "public-read"
}
