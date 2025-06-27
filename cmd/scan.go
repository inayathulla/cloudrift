package cmd

import (
	"fmt"
	"os"

	"github.com/inayathulla/cloudrift/internal/aws"
	"github.com/inayathulla/cloudrift/internal/detector"
	"github.com/inayathulla/cloudrift/internal/parser"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var configPath string

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan for infrastructure drift",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("🚀 Starting Cloudrift scan...")

		// Load YAML config
		viper.SetConfigFile(configPath)
		if err := viper.ReadInConfig(); err != nil {
			fmt.Fprintf(os.Stderr, "❌ Failed to read config file: %v\n", err)
			os.Exit(1)
		}

		// Extract config values
		profile := viper.GetString("aws_profile")
		region := viper.GetString("region")
		planPath := viper.GetString("plan_path")

		if planPath == "" {
			fmt.Fprintln(os.Stderr, "❌ 'plan_path' not found in config")
			os.Exit(1)
		}

		// Load AWS config
		cfg, err := aws.LoadAWSConfig(profile, region)
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Failed to load AWS config: %v\n", err)
			os.Exit(1)
		}

		// Validate credentials
		if err := aws.ValidateAWSCredentials(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "❌ Invalid AWS credentials: %v\n", err)
			os.Exit(1)
		}

		// Print connected identity
		identity, err := aws.GetCallerIdentity(cfg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Failed to retrieve AWS identity: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("🔐 Connected to AWS as: %s (%s)\n",
			aws.SafeString(identity.Arn),
			aws.SafeString(identity.Account))

		// Load plan
		plan, err := parser.LoadPlan(planPath)
		if err != nil {
			fmt.Printf("❌ Failed to load plan: %v\n", err)
			return
		}
		fmt.Printf("📄 Plan loaded: %+v\n", plan)

		// Fetch live AWS state
		liveBuckets, err := aws.FetchS3Buckets(cfg)
		if err != nil {
			fmt.Printf("❌ Failed to fetch live S3 state: %v\n", err)
			return
		}

		// Detect drift
		results := detector.DetectAllS3Drift(plan, liveBuckets)
		if len(results) == 0 {
			fmt.Println("✅ No S3 drift detected!")
		} else {
			fmt.Printf("⚠️ Drift detected in %d S3 bucket(s):\n", len(results))
			for _, r := range results {
				fmt.Printf("- Bucket: %s\n", r.BucketName)
				if r.Missing {
					fmt.Println("  ✖ Missing in AWS")
				}
				if r.AclDiff {
					fmt.Println("  ✖ ACL mismatch")
				}
				for k, diff := range r.TagDiffs {
					fmt.Printf("  ✖ Tag %s: expected=%s, actual=%s\n", k, diff[0], diff[1])
				}
				for k, v := range r.ExtraTags {
					fmt.Printf("  ✱ Extra tag in AWS: %s=%s\n", k, v)
				}
			}
		}
	},
}

func awsValue(ptr *string) string {
	if ptr != nil {
		return *ptr
	}
	return "unknown"
}

func init() {
	scanCmd.Flags().StringVarP(&configPath, "config", "c", "cloudrift.yml", "Path to Cloudrift config file")
	rootCmd.AddCommand(scanCmd)
}
