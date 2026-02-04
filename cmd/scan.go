package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/inayathulla/cloudrift/internal/common"
	"github.com/inayathulla/cloudrift/internal/detector"
	"github.com/inayathulla/cloudrift/internal/models"
	"github.com/inayathulla/cloudrift/internal/output"
)

// Command-line flags for the scan command.
var (
	configPath   string // Path to cloudrift.yml configuration file
	service      string // AWS service to scan (e.g., "s3", "ec2")
	outputFormat string // Output format (console, json, sarif)
	outputFile   string // Output file path (optional)
)

// DriftDetector defines the interface for service-specific drift detectors.
//
// Each supported AWS service (S3, EC2, etc.) implements this interface
// to provide consistent drift detection behavior across services.
type DriftDetector interface {
	// FetchLiveState retrieves the current state of resources from AWS.
	FetchLiveState() (interface{}, error)

	// DetectDrift compares planned state against live state and returns differences.
	DetectDrift(plan interface{}, live interface{}) ([]detector.DriftResult, error)
}

// scanCmd implements the "cloudrift scan" subcommand.
//
// The scan command performs the following steps:
//  1. Load configuration from cloudrift.yml
//  2. Initialize AWS SDK and validate credentials
//  3. Parse the Terraform plan JSON
//  4. Fetch live state from AWS
//  5. Compare plan vs live state
//  6. Output drift results
var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan for infrastructure drift",
	Long: `Scan compares your Terraform plan against live AWS infrastructure
to detect configuration drift.

The command reads a Terraform plan JSON file and fetches the current state
of corresponding resources from AWS, then reports any differences found.

Flags:
  --config, -c    Path to cloudrift.yml configuration file
  --service, -s   AWS service to scan (currently supports: s3)
  --format, -f    Output format: console, json, sarif (default: console)
  --output, -o    Write output to file instead of stdout

Example:
  cloudrift scan --config=config/cloudrift.yml --service=s3
  cloudrift scan --service=s3 --format=json
  cloudrift scan --service=s3 --format=sarif --output=drift-report.sarif`,
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
		scanDuration := time.Since(startScan)
		color.Green("✔️  Scan completed in %s!", scanDuration.Round(time.Millisecond))
		fmt.Println()

		// 9. Format and output results
		formatType := output.FormatType(strings.ToLower(outputFormat))
		formatter, ok := output.Get(formatType)
		if !ok {
			color.Red("❌ Unsupported output format: %s (supported: console, json, sarif)", outputFormat)
			os.Exit(1)
		}

		// Convert results to output.ScanResult
		scanResult := convertToScanResult(results, serviceName, *identity.Account, region, len(planResources), scanDuration)

		// Determine output writer
		var writer *os.File = os.Stdout
		if outputFile != "" {
			writer, err = os.Create(outputFile)
			if err != nil {
				color.Red("❌ Failed to create output file: %v", err)
				os.Exit(1)
			}
			defer writer.Close()
		}

		// For non-console formats, suppress the colorized output
		if formatType != output.FormatConsole {
			if err := formatter.Format(writer, scanResult); err != nil {
				color.Red("❌ Failed to format output: %v", err)
				os.Exit(1)
			}
			if outputFile != "" {
				color.Green("📄 Output written to %s", outputFile)
			}
		} else {
			// Use the legacy printer for console output (for now)
			printer.PrintDrift(results, planResources, liveResources)
		}
	},
}

// convertToScanResult converts legacy DriftResult to the new output.ScanResult format.
func convertToScanResult(results []detector.DriftResult, service, accountID, region string, totalResources int, duration time.Duration) output.ScanResult {
	drifts := make([]detector.DriftInfo, 0, len(results))

	for _, r := range results {
		info := detector.DriftInfo{
			ResourceID:   r.BucketName,
			ResourceType: "aws_s3_bucket",
			ResourceName: r.BucketName,
			Missing:      r.Missing,
			Diffs:        make(map[string][2]interface{}),
			ExtraAttributes: make(map[string]interface{}),
			Severity:     "warning",
		}

		if r.AclDiff {
			info.Diffs["acl"] = [2]interface{}{"<planned>", "<actual>"}
		}
		if r.VersioningDiff {
			info.Diffs["versioning_enabled"] = [2]interface{}{"<planned>", "<actual>"}
		}
		if r.EncryptionDiff {
			info.Diffs["encryption_algorithm"] = [2]interface{}{"<planned>", "<actual>"}
		}
		if r.LoggingDiff {
			info.Diffs["logging"] = [2]interface{}{"<planned>", "<actual>"}
		}
		if r.PublicAccessBlockDiff {
			info.Diffs["public_access_block"] = [2]interface{}{"<planned>", "<actual>"}
		}
		if r.LifecycleDiff {
			info.Diffs["lifecycle_rules"] = [2]interface{}{"<planned>", "<actual>"}
		}
		for k, v := range r.TagDiffs {
			info.Diffs["tags."+k] = [2]interface{}{v[0], v[1]}
		}
		for k, v := range r.ExtraTags {
			info.ExtraAttributes["tags."+k] = v
		}

		if r.Missing {
			info.Severity = "critical"
		}

		drifts = append(drifts, info)
	}

	driftCount := 0
	for _, d := range drifts {
		if d.HasDrift() {
			driftCount++
		}
	}

	return output.ScanResult{
		Service:        service,
		AccountID:      accountID,
		Region:         region,
		TotalResources: totalResources,
		DriftCount:     driftCount,
		Drifts:         drifts,
		ScanDuration:   duration.Milliseconds(),
		Timestamp:      time.Now().UTC().Format(time.RFC3339),
	}
}

func init() {
	scanCmd.Flags().StringVarP(&configPath, "config", "c", "cloudrift.yml", "Path to Cloudrift config file")
	scanCmd.Flags().StringVarP(&service, "service", "s", "s3", "AWS service to scan (e.g., s3)")
	scanCmd.Flags().StringVarP(&outputFormat, "format", "f", "console", "Output format: console, json, sarif")
	scanCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Write output to file instead of stdout")
	rootCmd.AddCommand(scanCmd)
}
