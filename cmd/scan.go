// cmd/scan.go
package cmd

import (
	"fmt"
	"os"

	"github.com/inayathulla/cloudrift/internal/common"
	"github.com/inayathulla/cloudrift/internal/detector"
	"github.com/inayathulla/cloudrift/internal/models"
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
		profile, region, planPath, err := common.LoadAppConfig(configPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Failed to load config: %v\n", err)
			os.Exit(1)
		}
		if planPath == "" {
			fmt.Fprintln(os.Stderr, "❌ 'plan_path' not found in config")
			os.Exit(1)
		}

		// 2) Load AWS config and validate credentials
		cfg, err := common.InitAWS(profile, region)
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Failed to load AWS config: %v\n", err)
			os.Exit(1)
		}
		if err := common.ValidateCredentials(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "❌ Invalid AWS credentials: %v\n", err)
			os.Exit(1)
		}

		// 3) Print AWS caller identity
		identity, err := common.GetCallerIdentity(cfg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Failed to retrieve AWS identity: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("🔐 Connected as: %s (%s)\n",
			*identity.Arn, *identity.Account)

		// 4) Load Terraform plan
		planResources, err := common.LoadPlan(planPath)
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
		var printer detector.DriftResultPrinter
		switch service {
		case "s3":
			printer = detector.S3DriftResultPrinter{}
			// Add more cases for other services as you implement them
		default:
			fmt.Fprintf(os.Stderr, "❌ Unsupported service: %s\n", service)
			os.Exit(1)
		}

		printer.PrintDrift(results, planResources, liveResources)
	},
}

func init() {
	scanCmd.Flags().StringVarP(&configPath, "config", "c", "cloudrift.yml", "Path to Cloudrift config file")
	scanCmd.Flags().StringVarP(&service, "service", "s", "s3", "AWS service to scan (e.g., s3)")
	rootCmd.AddCommand(scanCmd)
}
