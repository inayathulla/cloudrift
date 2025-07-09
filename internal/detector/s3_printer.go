package detector

import (
	"fmt"
	"os"
	"strings"

	"github.com/inayathulla/cloudrift/internal/models"
)

type S3DriftResultPrinter struct{}

func (p S3DriftResultPrinter) PrintDrift(results interface{}, plan, live interface{}) {
	planBuckets, _ := plan.([]models.S3Bucket)
	liveBuckets, _ := live.([]models.S3Bucket)
	s3Results, ok := results.([]DriftResult)
	if !ok {
		fmt.Fprintln(os.Stderr, "❌ Invalid drift result type for S3")
		return
	}
	if len(s3Results) == 0 {
		fmt.Println("✅ No drift detected!")
		return
	}
	fmt.Printf("⚠️ Drift in %d resource(s):\n", len(s3Results))
	for _, r := range s3Results {
		fmt.Printf("🪣 %s\n", r.BucketName)

		// Tags
		if len(r.TagDiffs) > 0 || len(r.ExtraTags) > 0 {
			fmt.Println("  🏷️ Tags:")
			// Separate out mismatches, missing, and extras
			var mismatches []string
			var missing []string

			for key, diff := range r.TagDiffs {
				expVal, liveVal := diff[0], diff[1]
				if liveVal == "" {
					// Key exists in plan but missing in live
					missing = append(missing, fmt.Sprintf("%s:%s", key, expVal))
				} else if expVal != liveVal {
					// Value mismatch
					mismatches = append(mismatches, fmt.Sprintf("%s:%s != %s:%s",
						key, expVal, key, liveVal))
				}
			}

			// 1) True mismatches
			if len(mismatches) > 0 {
				fmt.Println("    🔀 Mismatches:") // 4 spaces
				for _, m := range mismatches {
					fmt.Printf("        • %s\n", m) // 8 spaces
				}
			}

			// 2) Missing tags
			if len(missing) > 0 {
				fmt.Printf("    ⚠️ Missing:\n")
				for _, m := range missing {
					fmt.Printf("        • %s\n", m)
				}
			}

			// 3) Extra tags
			if len(r.ExtraTags) > 0 {
				fmt.Printf("    ➕")
				fmt.Printf("  Extra:\n")
				for key, val := range r.ExtraTags {
					fmt.Printf("        • %s:%s\n", key, val)
				}
			}
		}

		// Versioning
		if r.VersioningDiff {
			planVer := getBool(planBuckets, r.BucketName, func(b models.S3Bucket) bool { return b.VersioningEnabled })
			liveVer := getBoolLive(liveBuckets, r.BucketName, func(b models.S3Bucket) bool { return b.VersioningEnabled })
			fmt.Printf(
				"  🔄 Versioning mismatch:\n"+
					"    • expected → enabled: %t\n"+
					"    • actual   → enabled: %t\n",
				planVer, liveVer,
			)
		}

		// Encryption
		if r.EncryptionDiff {
			planEnc := getString(planBuckets, r.BucketName, func(b models.S3Bucket) string { return b.EncryptionAlgorithm })
			liveEnc := getStringLive(liveBuckets, r.BucketName, func(b models.S3Bucket) string { return b.EncryptionAlgorithm })
			fmt.Printf(
				"  🔐 Encryption mismatch:\n"+
					"    • expected → %q\n"+
					"    • actual   → %q\n",
				planEnc, liveEnc,
			)
		}

		// Logging
		if r.LoggingDiff {
			planLogEnabled := getBool(planBuckets, r.BucketName, func(b models.S3Bucket) bool { return b.LoggingEnabled })
			planLogBucket := getString(planBuckets, r.BucketName, func(b models.S3Bucket) string { return b.LoggingTargetBucket })
			planLogPrefix := getString(planBuckets, r.BucketName, func(b models.S3Bucket) string { return b.LoggingTargetPrefix })

			liveLogEnabled := getBoolLive(liveBuckets, r.BucketName, func(b models.S3Bucket) bool { return b.LoggingEnabled })
			liveLogBucket := getStringLive(liveBuckets, r.BucketName, func(b models.S3Bucket) string { return b.LoggingTargetBucket })
			liveLogPrefix := getStringLive(liveBuckets, r.BucketName, func(b models.S3Bucket) string { return b.LoggingTargetPrefix })

			// Build the “plan” summary
			planFields := []string{fmt.Sprintf("enabled=%t", planLogEnabled)}
			if planLogBucket != "" {
				planFields = append(planFields, fmt.Sprintf("bucket=%s", planLogBucket))
			}
			if planLogPrefix != "" {
				planFields = append(planFields, fmt.Sprintf("prefix=%s", planLogPrefix))
			}

			// Build the “live” summary (only differing fields)
			liveFields := []string{fmt.Sprintf("enabled=%t", liveLogEnabled)}
			if liveLogBucket != planLogBucket && liveLogBucket != "" {
				liveFields = append(liveFields, fmt.Sprintf("bucket=%s", liveLogBucket))
			}
			if liveLogPrefix != planLogPrefix && liveLogPrefix != "" {
				liveFields = append(liveFields, fmt.Sprintf("prefix=%s", liveLogPrefix))
			}

			fmt.Println("  📑 Logging:")
			fmt.Printf("    • plan → %s\n", strings.Join(planFields, ", "))
			fmt.Printf("    • live → %s\n", strings.Join(liveFields, ", "))
		}

		// Public access block
		if r.PublicAccessBlockDiff {
			planPAB := getPAB(planBuckets, r.BucketName)
			livePAB := getPABLive(liveBuckets, r.BucketName)

			// Build plan summary with all flags
			planFields := []string{
				fmt.Sprintf("BlockPublicAcls=%t", planPAB.BlockPublicAcls),
				fmt.Sprintf("IgnorePublicAcls=%t", planPAB.IgnorePublicAcls),
				fmt.Sprintf("BlockPublicPolicy=%t", planPAB.BlockPublicPolicy),
				fmt.Sprintf("RestrictPublicBuckets=%t", planPAB.RestrictPublicBuckets),
			}

			// Build live summary, only include flags that differ
			liveFields := []string{}
			if livePAB.BlockPublicAcls != planPAB.BlockPublicAcls {
				liveFields = append(liveFields, fmt.Sprintf("BlockPublicAcls=%t", livePAB.BlockPublicAcls))
			}
			if livePAB.IgnorePublicAcls != planPAB.IgnorePublicAcls {
				liveFields = append(liveFields, fmt.Sprintf("IgnorePublicAcls=%t", livePAB.IgnorePublicAcls))
			}
			if livePAB.BlockPublicPolicy != planPAB.BlockPublicPolicy {
				liveFields = append(liveFields, fmt.Sprintf("BlockPublicPolicy=%t", livePAB.BlockPublicPolicy))
			}
			if livePAB.RestrictPublicBuckets != planPAB.RestrictPublicBuckets {
				liveFields = append(liveFields, fmt.Sprintf("RestrictPublicBuckets=%t", livePAB.RestrictPublicBuckets))
			}

			fmt.Println("  🚫 Public Access Block differ:")
			fmt.Printf("    • plan → %s\n", strings.Join(planFields, ", "))
			fmt.Printf("    • live → %s\n", strings.Join(liveFields, ", "))
		}

		// Lifecycle
		if r.LifecycleDiff {
			planLC := getLifecycle(planBuckets, r.BucketName)
			liveLC := getLifecycleLive(liveBuckets, r.BucketName)

			// Maps for lookup
			planMap := map[string]models.LifecycleRuleSummary{}
			liveMap := map[string]models.LifecycleRuleSummary{}
			for _, pr := range planLC {
				planMap[pr.ID] = pr
			}
			for _, lr := range liveLC {
				liveMap[lr.ID] = lr
			}

			fmt.Println("  ⏳ Lifecycle rules:")

			// Mismatches
			var mismatches []string
			for id, pr := range planMap {
				if lr, ok := liveMap[id]; ok {
					if pr.ExpirationDays != lr.ExpirationDays || pr.Status != lr.Status || pr.Prefix != lr.Prefix {
						mismatches = append(mismatches, id)
					}
				}
			}
			if len(mismatches) > 0 {
				fmt.Println("    • Mismatched rules:")
				for _, id := range mismatches {
					pr := planMap[id]
					lr := liveMap[id]
					fmt.Printf("        – %s: plan=Expires %d days (%s), live=Expires %d days (%s)\n",
						id, pr.ExpirationDays, pr.Status, lr.ExpirationDays, lr.Status)
				}
			}

			// Deleted
			var deleted []string
			for id := range planMap {
				if _, ok := liveMap[id]; !ok {
					deleted = append(deleted, id)
				}
			}
			if len(deleted) > 0 {
				fmt.Println("    ⚠️ Deleted rules:")
				for _, id := range deleted {
					pr := planMap[id]
					fmt.Printf("        – %s:\n", id)
					fmt.Printf("            • Status         : %s\n", pr.Status)
					fmt.Printf("            • Expires after  : %d days\n", pr.ExpirationDays)
					if pr.Prefix != "" {
						fmt.Printf("            • Prefix         : %s\n", pr.Prefix)
					}
				}
			}

			// Extra
			var extras []string
			for id := range liveMap {
				if _, ok := planMap[id]; !ok {
					extras = append(extras, id)
				}
			}
			if len(extras) > 0 {
				fmt.Println("    • Extra rules:")
				for _, id := range extras {
					lr := liveMap[id]
					fmt.Printf("        – %s:\n", id)
					fmt.Printf("            • Status         : %s\n", lr.Status)
					fmt.Printf("            • Expires after  : %d days\n", lr.ExpirationDays)
					if lr.Prefix != "" {
						fmt.Printf("            • Prefix         : %s\n", lr.Prefix)
					}
				}
			}
		}
	}
}

// Helper functions for S3 drift printing
func getBucket(plan []models.S3Bucket, name string) *models.S3Bucket {
	for _, b := range plan {
		if b.Name == name {
			return &b
		}
	}
	return nil
}

func getBool(plan []models.S3Bucket, name string, fn func(models.S3Bucket) bool) bool {
	if b := getBucket(plan, name); b != nil {
		return fn(*b)
	}
	return false
}

func getBoolLive(live []models.S3Bucket, name string, fn func(models.S3Bucket) bool) bool {
	return getBool(live, name, fn)
}

func getString(plan []models.S3Bucket, name string, fn func(models.S3Bucket) string) string {
	if b := getBucket(plan, name); b != nil {
		return fn(*b)
	}
	return ""
}

func getStringLive(live []models.S3Bucket, name string, fn func(models.S3Bucket) string) string {
	return getString(live, name, fn)
}

func getPAB(plan []models.S3Bucket, name string) models.PublicAccessBlockConfig {
	if b := getBucket(plan, name); b != nil {
		return b.PublicAccessBlock
	}
	return models.PublicAccessBlockConfig{}
}

func getPABLive(live []models.S3Bucket, name string) models.PublicAccessBlockConfig {
	return getPAB(live, name)
}

func getLifecycle(plan []models.S3Bucket, name string) []models.LifecycleRuleSummary {
	if b := getBucket(plan, name); b != nil {
		return b.LifecycleRules
	}
	return nil
}

func getLifecycleLive(live []models.S3Bucket, name string) []models.LifecycleRuleSummary {
	return getLifecycle(live, name)
}
