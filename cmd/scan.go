// cmd/scan.go
package cmd

import (
	"fmt"
	"github.com/inayathulla/cloudrift/internal/aws"
	"github.com/inayathulla/cloudrift/internal/detector"
	"github.com/inayathulla/cloudrift/internal/models"
	"github.com/inayathulla/cloudrift/internal/parser"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"strings"
)

var configPath string
var service string // e.g. "s3", "ec2"

// DriftDetector defines the interface for any service-specific detector.
type DriftDetector interface {
	FetchLiveState() (interface{}, error)
	DetectDrift(plan interface{}, live interface{}) ([]detector.DriftResult, error)
}

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan for infrastructure drift",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("🚀 Starting Cloudrift scan...")

		// 1) Load YAML config
		viper.SetConfigFile(configPath)
		if err := viper.ReadInConfig(); err != nil {
			fmt.Fprintf(os.Stderr, "❌ Failed to read config file: %v\n", err)
			os.Exit(1)
		}
		profile := viper.GetString("aws_profile")
		region := viper.GetString("region")
		planPath := viper.GetString("plan_path")
		if planPath == "" {
			fmt.Fprintln(os.Stderr, "❌ 'plan_path' not found in config")
			os.Exit(1)
		}

		// 2) Load AWS config and validate credentials
		cfg, err := aws.LoadAWSConfig(profile, region)
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Failed to load AWS config: %v\n", err)
			os.Exit(1)
		}
		if err := aws.ValidateAWSCredentials(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "❌ Invalid AWS credentials: %v\n", err)
			os.Exit(1)
		}

		// 3) Print AWS caller identity
		identity, err := aws.GetCallerIdentity(cfg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Failed to retrieve AWS identity: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("🔐 Connected as: %s (%s)\n",
			*identity.Arn, *identity.Account)

		// 4) Load Terraform plan
		planResources, err := parser.LoadPlan(planPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Failed to load plan: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("📄 Plan loaded from json\n")

		// 5) Select the appropriate detector
		var det DriftDetector
		switch service {
		case "s3":
			det = detector.NewS3DriftDetector(cfg)
		default:
			fmt.Fprintf(os.Stderr, "❌ Unsupported service: %s\n", service)
			os.Exit(1)
		}

		// 6) Fetch live state
		rawLive, err := det.FetchLiveState()
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Failed to fetch live state: %v\n", err)
			os.Exit(1)
		}

		// 7) Cast to concrete type so we can inspect
		liveResources, ok := rawLive.([]models.S3Bucket)
		if !ok {
			fmt.Fprintf(os.Stderr, "❌ Unexpected live state type\n")
			os.Exit(1)
		}

		// 8) Detect drift
		results, err := det.DetectDrift(planResources, liveResources)
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Drift detection failed: %v\n", err)
			os.Exit(1)
		}

		// 9) Print drift results
		if len(results) == 0 {
			fmt.Println("✅ No drift detected!")
			return
		}
		fmt.Printf("⚠️ Drift in %d resource(s):\n", len(results))
		for _, r := range results {
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
				planVer := getBool(planResources, r.BucketName, func(b models.S3Bucket) bool { return b.VersioningEnabled })
				liveVer := getBoolLive(liveResources, r.BucketName, func(b models.S3Bucket) bool { return b.VersioningEnabled })
				fmt.Printf(
					"  🔄 Versioning mismatch:\n"+
						"    • expected → enabled: %t\n"+
						"    • actual   → enabled: %t\n",
					planVer, liveVer,
				)
			}

			// Encryption
			if r.EncryptionDiff {
				planEnc := getString(planResources, r.BucketName, func(b models.S3Bucket) string { return b.EncryptionAlgorithm })
				liveEnc := getStringLive(liveResources, r.BucketName, func(b models.S3Bucket) string { return b.EncryptionAlgorithm })
				fmt.Printf(
					"  🔐 Encryption mismatch:\n"+
						"    • expected → %q\n"+
						"    • actual   → %q\n",
					planEnc, liveEnc,
				)
			}

			// Logging
			if r.LoggingDiff {
				planLogEnabled := getBool(planResources, r.BucketName, func(b models.S3Bucket) bool { return b.LoggingEnabled })
				planLogBucket := getString(planResources, r.BucketName, func(b models.S3Bucket) string { return b.LoggingTargetBucket })
				planLogPrefix := getString(planResources, r.BucketName, func(b models.S3Bucket) string { return b.LoggingTargetPrefix })

				liveLogEnabled := getBoolLive(liveResources, r.BucketName, func(b models.S3Bucket) bool { return b.LoggingEnabled })
				liveLogBucket := getStringLive(liveResources, r.BucketName, func(b models.S3Bucket) string { return b.LoggingTargetBucket })
				liveLogPrefix := getStringLive(liveResources, r.BucketName, func(b models.S3Bucket) string { return b.LoggingTargetPrefix })

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
				planPAB := getPAB(planResources, r.BucketName)
				livePAB := getPABLive(liveResources, r.BucketName)

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
				planLC := getLifecycle(planResources, r.BucketName)
				liveLC := getLifecycleLive(liveResources, r.BucketName)

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
	},
}

func init() {
	scanCmd.Flags().StringVarP(&configPath, "config", "c", "cloudrift.yml", "Path to Cloudrift config file")
	scanCmd.Flags().StringVarP(&service, "service", "s", "s3", "AWS service to scan (e.g., s3)")
	rootCmd.AddCommand(scanCmd)
}

// helper functions to extract plan vs live values by bucket name

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
