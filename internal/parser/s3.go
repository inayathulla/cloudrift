package parser

import (
	"github.com/inayathulla/cloudrift/internal/models"
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
			Id:   rc.Address,
			Tags: make(map[string]string),
		}

		// 1) Name
		if name, ok := after["bucket"].(string); ok {
			bucket.Name = name
		}

		// 2) ACL
		if acl, ok := after["acl"].(string); ok {
			bucket.Acl = acl
		}

		// 3) Tags
		if tagsRaw, ok := after["tags"].(map[string]interface{}); ok {
			for k, v := range tagsRaw {
				if s, ok := v.(string); ok {
					bucket.Tags[k] = s
				}
			}
		}

		// 4) Versioning
		if verRaw, ok := after["versioning"].(map[string]interface{}); ok {
			if enabled, ok := verRaw["enabled"].(bool); ok {
				bucket.VersioningEnabled = enabled
			}
		}

		// 5) Encryption
		if encRaw, ok := after["server_side_encryption_configuration"].(map[string]interface{}); ok {
			if rules, ok := encRaw["rules"].([]interface{}); ok && len(rules) > 0 {
				if rule0, ok := rules[0].(map[string]interface{}); ok {
					if apply, ok := rule0["apply_server_side_encryption_by_default"].(map[string]interface{}); ok {
						if algo, ok := apply["sse_algorithm"].(string); ok {
							bucket.EncryptionAlgorithm = algo
						}
					}
				}
			}
		}

		// 6) Access Logging
		if logRaw, ok := after["logging"].(map[string]interface{}); ok {
			if tb, ok := logRaw["target_bucket"].(string); ok {
				bucket.LoggingEnabled = true
				bucket.LoggingTargetBucket = tb
			}
			if tp, ok := logRaw["target_prefix"].(string); ok {
				bucket.LoggingTargetPrefix = tp
			}
		}

		// 7) Public Access Block
		if pabRaw, ok := after["public_access_block"].(map[string]interface{}); ok {
			var cfg models.PublicAccessBlockConfig
			if v, ok := pabRaw["block_public_acls"].(bool); ok {
				cfg.BlockPublicAcls = v
			}
			if v, ok := pabRaw["ignore_public_acls"].(bool); ok {
				cfg.IgnorePublicAcls = v
			}
			if v, ok := pabRaw["block_public_policy"].(bool); ok {
				cfg.BlockPublicPolicy = v
			}
			if v, ok := pabRaw["restrict_public_buckets"].(bool); ok {
				cfg.RestrictPublicBuckets = v
			}
			bucket.PublicAccessBlock = cfg
		}

		// 8) Lifecycle Rules
		// 8) Lifecycle Rules
		if lcRaw, ok := after["lifecycle_rule"].([]interface{}); ok {
			for _, raw := range lcRaw {
				if m, ok := raw.(map[string]interface{}); ok {
					var rule models.LifecycleRuleSummary

					// ID
					if id, ok := m["id"].(string); ok {
						rule.ID = id
					}

					// Status
					if st, ok := m["status"].(string); ok {
						rule.Status = st
					}

					// Prefix (if provided)
					if prefix, ok := m["prefix"].(string); ok {
						rule.Prefix = prefix
					}

					// Expiration days
					if expRaw, ok := m["expiration"].(map[string]interface{}); ok {
						if daysF, ok := expRaw["days"].(float64); ok {
							rule.ExpirationDays = int(daysF)
						}
					}

					bucket.LifecycleRules = append(bucket.LifecycleRules, rule)
				}
			}
		}

		buckets = append(buckets, bucket)
	}

	return buckets
}
