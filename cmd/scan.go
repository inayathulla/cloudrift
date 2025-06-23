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

		// 1. Read the YAML config using Viper
		viper.SetConfigFile(configPath)
		if err := viper.ReadInConfig(); err != nil {
			fmt.Fprintf(os.Stderr, "❌ Failed to read config file: %v\n", err)
			os.Exit(1)
		}

		planPath := viper.GetString("plan_path")
		if planPath == "" {
			fmt.Fprintln(os.Stderr, "❌ 'plan_path' not found in config")
			os.Exit(1)
		}

		// 2. Load plan
		plan, err := parser.LoadPlan(planPath)
		if err != nil {
			fmt.Printf("❌ Failed to load plan: %v\n", err)
			return
		}

		// 3. Fetch live state from AWS
		liveBuckets, err := aws.FetchS3Buckets()
		if err != nil {
			fmt.Printf("❌ Failed to fetch live S3 state: %v\n", err)
			return
		}

		// 4. Detect drift
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
			}
		}
	},
}

func init() {
	scanCmd.Flags().StringVarP(&configPath, "config", "c", "cloudrift.yml", "Path to Cloudrift config file")
	rootCmd.AddCommand(scanCmd)
}
