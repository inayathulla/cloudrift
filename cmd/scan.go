package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
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
		startScan := time.Now()
		color.Cyan("🚀 Starting Cloudrift scan...")

		viper.SetConfigFile(configPath)
		if err := viper.ReadInConfig(); err != nil {
			color.Red("❌ Failed to read config file: %v", err)
			os.Exit(1)
		}
		profile, region, planPath, err := common.LoadAppConfig(configPath)
		if err != nil {
			color.Red("❌ Failed to load config: %v", err)
			os.Exit(1)
		}
		if planPath == "" {
			color.Red("❌ 'plan_path' not found in config")
			os.Exit(1)
		}

		s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
		s.Color("cyan")

		// 1. Loading AWS config
		s.Suffix = " Loading AWS config..."
		start := time.Now()
		s.Start()
		cfg, err := common.InitAWS(profile, region)
		s.Stop()
		if err != nil {
			color.Red("❌ Failed to load AWS config: %v", err)
			os.Exit(1)
		}
		color.Yellow("✔️  AWS config loaded in %s", time.Since(start).Round(time.Millisecond))

		// 2. Validating credentials
		s.Suffix = " Validating AWS credentials..."
		start = time.Now()
		s.Start()
		err = common.ValidateCredentials(cfg)
		s.Stop()
		if err != nil {
			color.Red("❌ Invalid AWS credentials: %v", err)
			os.Exit(1)
		}
		color.Yellow("✔️  Credentials valided in %s", time.Since(start).Round(time.Millisecond))

		// 3. Fetching AWS identity
		s.Suffix = " Fetching AWS identity..."
		start = time.Now()
		s.Start()
		identity, err := common.GetCallerIdentity(cfg)
		s.Stop()
		if err != nil {
			color.Red("❌ Failed to retrieve AWS identity: %v", err)
			os.Exit(1)
		}
		color.Green("🔐 Connected as: %s (%s) [%s] in %s", *identity.Arn, *identity.Account, region, time.Since(start).Round(time.Millisecond))

		// 4. Loading Terraform plan
		s.Suffix = " Loading Terraform plan..."
		start = time.Now()
		s.Start()
		planResources, err := common.LoadPlan(planPath)
		s.Stop()
		if err != nil {
			color.Red("❌ Failed to load plan: %v", err)
			os.Exit(1)
		}
		color.Yellow("📄 Plan loaded from json in %s", time.Since(start).Round(time.Millisecond))

		// 5. Select the appropriate detector and printer
		var det DriftDetector
		var printer detector.DriftResultPrinter
		var serviceName string
		switch service {
		case "s3":
			det = detector.NewS3DriftDetector(cfg)
			printer = detector.S3DriftResultPrinter{}
			serviceName = "S3"
		default:
			color.Red("❌ Unsupported service: %s", service)
			os.Exit(1)
		}

		// 6. Fetching live state
		s.Suffix = fmt.Sprintf(" Fetching live %s state...", serviceName)
		start = time.Now()
		s.Start()
		rawLive, err := det.FetchLiveState()
		s.Stop()
		if err != nil {
			color.Red("❌ Failed to fetch live state: %v", err)
			os.Exit(1)
		}
		color.Yellow("✔️  Live %s state fetched in %s", serviceName, time.Since(start).Round(time.Millisecond))

		// 7. Cast to concrete type so we can inspect
		liveResources, ok := rawLive.([]models.S3Bucket)
		if !ok {
			color.Red("❌ Unexpected live state type")
			os.Exit(1)
		}

		// 8. Detect drift
		results, err := det.DetectDrift(planResources, liveResources)
		if err != nil {
			color.Red("❌ Drift detection failed: %v", err)
			os.Exit(1)
		}
		color.Green("✔️  Scan completed in %s!", time.Since(startScan).Round(time.Millisecond))
		fmt.Println()

		// 9. Print drift results (all at once)
		printer.PrintDrift(results, planResources, liveResources)
	},
}

func init() {
	scanCmd.Flags().StringVarP(&configPath, "config", "c", "cloudrift.yml", "Path to Cloudrift config file")
	scanCmd.Flags().StringVarP(&service, "service", "s", "s3", "AWS service to scan (e.g., s3)")
	rootCmd.AddCommand(scanCmd)
}
