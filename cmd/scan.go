package cmd

import (
	"fmt"
	"os"

	"github.com/inayathulla/cloudrift/internal/aws"
	"github.com/inayathulla/cloudrift/internal/detector"
	"github.com/inayathulla/cloudrift/internal/models"
	"github.com/inayathulla/cloudrift/internal/parser"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
		fmt.Printf("📄 Plan loaded: %+v\n", planResources)

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
		for _, res := range liveResources {
			fmt.Printf("🔍 Live state for %s: tags=%v acl=%s\n",
				res.Name, res.Tags, res.Acl)
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
			fmt.Printf("- %s\n", r.BucketName)
			for key, diff := range r.TagDiffs {
				fmt.Printf("  ✖ Tag %s: expected=%s, actual=%s\n",
					key, diff[0], diff[1])
			}
			for key, val := range r.ExtraTags {
				fmt.Printf("  ✱ Extra tag %s=%s\n", key, val)
			}
		}
	},
}

func init() {
	scanCmd.Flags().StringVarP(&configPath, "config", "c", "cloudrift.yml", "Path to Cloudrift config file")
	scanCmd.Flags().StringVarP(&service, "service", "s", "s3", "AWS service to scan (e.g., s3)")
	rootCmd.AddCommand(scanCmd)
}
