package cmd

import (
	"context"
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
	"github.com/inayathulla/cloudrift/internal/policy"
)

// Command-line flags for the scan command.
var (
	configPath      string // Path to cloudrift.yml configuration file
	service         string // AWS service to scan (e.g., "s3", "ec2")
	outputFormat    string // Output format (console, json, sarif)
	outputFile      string // Output file path (optional)
	policyDir       string // Directory containing custom OPA policies
	failOnViolation bool   // Exit with non-zero code if policy violations found
	skipPolicies    bool   // Skip policy evaluation
	noEmoji         bool   // Use ASCII characters instead of emojis
)

// icons holds the characters used for status indicators (emoji or ASCII)
var icons struct {
	Rocket, Check, Cross, Lock, Doc, Warn, Gear, Pin, Msg string
}

func initIcons() {
	if noEmoji {
		icons.Rocket = "[*]"
		icons.Check = "[+]"
		icons.Cross = "[X]"
		icons.Lock = "[>]"
		icons.Doc = "[i]"
		icons.Warn = "[!]"
		icons.Gear = "[#]"
		icons.Pin = "[-]"
		icons.Msg = "[-]"
	} else {
		icons.Rocket = "🚀"
		icons.Check = "✔️ "
		icons.Cross = "❌"
		icons.Lock = "🔐"
		icons.Doc = "📄"
		icons.Warn = "⚠️ "
		icons.Gear = "🔧"
		icons.Pin = "📍"
		icons.Msg = "💬"
	}
}

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
	Short: "Scan for infrastructure drift and policy violations",
	Long: `Scan compares your Terraform plan against live AWS infrastructure
to detect configuration drift and evaluate against security policies.

The command reads a Terraform plan JSON file and fetches the current state
of corresponding resources from AWS, then reports any differences found.
Additionally, it evaluates resources against OPA policies to detect
security and compliance violations.

Flags:
  --config, -c         Path to cloudrift.yml configuration file
  --service, -s        AWS service to scan (supports: s3, ec2)
  --format, -f         Output format: console, json, sarif (default: console)
  --output, -o         Write output to file instead of stdout
  --policy-dir, -p     Directory containing custom OPA policies (.rego files)
  --fail-on-violation  Exit with non-zero code if policy violations are found
  --skip-policies      Skip policy evaluation (drift detection only)
  --no-emoji           Use ASCII characters instead of emojis

Example:
  cloudrift scan --config=config/cloudrift.yml --service=s3
  cloudrift scan --service=ec2 --format=json
  cloudrift scan --service=s3 --format=sarif --output=drift-report.sarif
  cloudrift scan --service=s3 --policy-dir=./my-policies --fail-on-violation`,
	Run: func(cmd *cobra.Command, args []string) {
		initIcons()
		startScan := time.Now()
		color.Cyan("%s Starting Cloudrift scan...", icons.Rocket)

		viper.SetConfigFile(configPath)
		if err := viper.ReadInConfig(); err != nil {
			color.Red("%s Failed to read config file: %v", icons.Cross, err)
			os.Exit(1)
		}
		profile, region, planPath, err := common.LoadAppConfig(configPath)
		if err != nil {
			color.Red("%s Failed to load config: %v", icons.Cross, err)
			os.Exit(1)
		}
		if planPath == "" {
			color.Red("%s 'plan_path' not found in config", icons.Cross)
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
			color.Red("%s Failed to load AWS config: %v", icons.Cross, err)
			os.Exit(1)
		}
		color.Yellow("%s AWS config loaded in %s", icons.Check, time.Since(start).Round(time.Millisecond))

		// 2. Validating credentials
		s.Suffix = " Validating AWS credentials..."
		start = time.Now()
		s.Start()
		err = common.ValidateCredentials(cfg)
		s.Stop()
		if err != nil {
			color.Red("%s Invalid AWS credentials: %v", icons.Cross, err)
			os.Exit(1)
		}
		color.Yellow("%s Credentials valided in %s", icons.Check, time.Since(start).Round(time.Millisecond))

		// 3. Fetching AWS identity
		s.Suffix = " Fetching AWS identity..."
		start = time.Now()
		s.Start()
		identity, err := common.GetCallerIdentity(cfg)
		s.Stop()
		if err != nil {
			color.Red("%s Failed to retrieve AWS identity: %v", icons.Cross, err)
			os.Exit(1)
		}
		color.Green("%s Connected as: %s (%s) [%s] in %s", icons.Lock, *identity.Arn, *identity.Account, region, time.Since(start).Round(time.Millisecond))

		// 4. Select the appropriate detector and printer, load plan
		var det DriftDetector
		var printer detector.DriftResultPrinter
		var serviceName string
		var planResources interface{}
		var liveResources interface{}
		var planCount int

		s.Suffix = " Loading Terraform plan..."
		start = time.Now()
		s.Start()

		switch service {
		case "s3":
			det = detector.NewS3DriftDetector(cfg)
			printer = detector.S3DriftResultPrinter{}
			serviceName = "S3"
			pr, err := common.LoadPlan(planPath)
			if err != nil {
				s.Stop()
				color.Red("%s Failed to load plan: %v", icons.Cross, err)
				os.Exit(1)
			}
			planResources = pr
			planCount = len(pr)

		case "ec2":
			det = detector.NewEC2DriftDetector(cfg)
			printer = detector.EC2DriftResultPrinter{}
			serviceName = "EC2"
			pr, err := common.LoadEC2Plan(planPath)
			if err != nil {
				s.Stop()
				color.Red("%s Failed to load plan: %v", icons.Cross, err)
				os.Exit(1)
			}
			planResources = pr
			planCount = len(pr)

		default:
			s.Stop()
			color.Red("%s Unsupported service: %s (supported: s3, ec2)", icons.Cross, service)
			os.Exit(1)
		}

		s.Stop()
		color.Yellow("%s Plan loaded from json in %s", icons.Doc, time.Since(start).Round(time.Millisecond))

		// 5. Fetching live state
		s.Suffix = fmt.Sprintf(" Fetching live %s state...", serviceName)
		start = time.Now()
		s.Start()
		rawLive, err := det.FetchLiveState()
		s.Stop()
		if err != nil {
			color.Red("%s Failed to fetch live state: %v", icons.Cross, err)
			os.Exit(1)
		}
		liveResources = rawLive
		color.Yellow("%s Live %s state fetched in %s", icons.Check, serviceName, time.Since(start).Round(time.Millisecond))

		// 6. Detect drift
		results, err := det.DetectDrift(planResources, liveResources)
		if err != nil {
			color.Red("%s Drift detection failed: %v", icons.Cross, err)
			os.Exit(1)
		}
		color.Green("%s Drift detection completed", icons.Check)

		// 7. Policy evaluation
		var policyResult *policy.EvaluationResult
		if !skipPolicies {
			s.Suffix = " Evaluating policies..."
			start = time.Now()
			s.Start()

			var engine *policy.Engine
			if policyDir != "" {
				// Load custom policies along with built-ins
				engine, err = policy.LoadPoliciesWithBuiltins(policyDir)
			} else {
				// Load only built-in policies
				engine, err = policy.LoadBuiltinPolicies()
			}
			s.Stop()

			if err != nil {
				color.Yellow("%s Policy engine initialization failed: %v", icons.Warn, err)
				// Continue without policies
			} else if engine.PolicyCount() > 0 {
				// Build policy inputs from plan resources
				inputs := buildPolicyInputs(service, planResources, liveResources, results)

				policyResult, err = engine.EvaluateAll(context.Background(), inputs)
				if err != nil {
					color.Yellow("%s Policy evaluation failed: %v", icons.Warn, err)
				} else {
					color.Yellow("%s Evaluated %d policies in %s", icons.Check, engine.PolicyCount(), time.Since(start).Round(time.Millisecond))
					if policyResult.HasViolations() {
						color.Red("%s Found %d policy violations", icons.Warn, len(policyResult.Violations))
					}
					if len(policyResult.Warnings) > 0 {
						color.Yellow("%s Found %d policy warnings", icons.Warn, len(policyResult.Warnings))
					}
				}
			}
		}

		scanDuration := time.Since(startScan)
		color.Green("%s Scan completed in %s!", icons.Check, scanDuration.Round(time.Millisecond))
		fmt.Println()

		// 8. Format and output results
		formatType := output.FormatType(strings.ToLower(outputFormat))
		formatter, ok := output.Get(formatType)
		if !ok {
			color.Red("%s Unsupported output format: %s (supported: console, json, sarif)", icons.Cross, outputFormat)
			os.Exit(1)
		}

		// Convert results to output.ScanResult
		scanResult := convertToScanResult(results, serviceName, *identity.Account, region, planCount, scanDuration)

		// Determine output writer
		var writer *os.File = os.Stdout
		if outputFile != "" {
			writer, err = os.Create(outputFile)
			if err != nil {
				color.Red("%s Failed to create output file: %v", icons.Cross, err)
				os.Exit(1)
			}
			defer writer.Close()
		}

		// For non-console formats, suppress the colorized output
		if formatType != output.FormatConsole {
			if err := formatter.Format(writer, scanResult); err != nil {
				color.Red("%s Failed to format output: %v", icons.Cross, err)
				os.Exit(1)
			}
			if outputFile != "" {
				color.Green("%s Output written to %s", icons.Doc, outputFile)
			}
		} else {
			// Use the legacy printer for console output (for now)
			printer.PrintDrift(results, planResources, liveResources)

			// Print policy violations if present
			if policyResult != nil && (len(policyResult.Violations) > 0 || len(policyResult.Warnings) > 0) {
				printPolicyResults(policyResult)
			}
		}

		// Exit with error if --fail-on-violation is set and violations exist
		if failOnViolation && policyResult != nil && policyResult.HasViolations() {
			os.Exit(2)
		}
	},
}

// printPolicyResults outputs policy evaluation results to console.
func printPolicyResults(result *policy.EvaluationResult) {
	fmt.Println()
	color.Cyan("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	color.Cyan("              POLICY EVALUATION                   ")
	color.Cyan("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	if len(result.Violations) > 0 {
		fmt.Println()
		color.Red("%s VIOLATIONS (%d)", icons.Cross, len(result.Violations))
		for _, v := range result.Violations {
			fmt.Println()
			color.Red("  [%s] %s", v.Severity, v.PolicyID)
			fmt.Printf("  %s Resource: %s\n", icons.Pin, color.CyanString(v.ResourceAddress))
			fmt.Printf("  %s %s\n", icons.Msg, v.Message)
			if v.Remediation != "" {
				fmt.Printf("  %s %s\n", icons.Gear, color.YellowString(v.Remediation))
			}
		}
	}

	if len(result.Warnings) > 0 {
		fmt.Println()
		color.Yellow("%s WARNINGS (%d)", icons.Warn, len(result.Warnings))
		for _, w := range result.Warnings {
			fmt.Println()
			color.Yellow("  [%s] %s", w.Severity, w.PolicyID)
			fmt.Printf("  %s Resource: %s\n", icons.Pin, color.CyanString(w.ResourceAddress))
			fmt.Printf("  %s %s\n", icons.Msg, w.Message)
			if w.Remediation != "" {
				fmt.Printf("  %s %s\n", icons.Gear, color.YellowString(w.Remediation))
			}
		}
	}

	fmt.Println()
	color.Cyan("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
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

// buildPolicyInputs creates policy inputs from scan resources.
func buildPolicyInputs(service string, planResources, liveResources interface{}, results []detector.DriftResult) []*policy.PolicyInput {
	var inputs []*policy.PolicyInput

	// Build a map of drift results by resource name for quick lookup
	driftMap := make(map[string]detector.DriftResult)
	for _, r := range results {
		driftMap[r.BucketName] = r
	}

	switch service {
	case "s3":
		if buckets, ok := planResources.([]models.S3Bucket); ok {
			for _, b := range buckets {
				input := policy.NewPolicyInput("aws_s3_bucket", b.Id)
				input.Resource.Planned = map[string]interface{}{
					"bucket":               b.Name,
					"acl":                  b.Acl,
					"tags":                 b.Tags,
					"versioning_enabled":   b.VersioningEnabled,
					"encryption_algorithm": b.EncryptionAlgorithm,
					"logging_enabled":      b.LoggingEnabled,
					"public_access_block": map[string]interface{}{
						"block_public_acls":       b.PublicAccessBlock.BlockPublicAcls,
						"block_public_policy":     b.PublicAccessBlock.BlockPublicPolicy,
						"ignore_public_acls":      b.PublicAccessBlock.IgnorePublicAcls,
						"restrict_public_buckets": b.PublicAccessBlock.RestrictPublicBuckets,
					},
				}

				// Add drift info if present
				if dr, ok := driftMap[b.Name]; ok {
					input.Resource.Drift = &policy.DriftInput{
						HasDrift: true,
						Missing:  dr.Missing,
					}
				}

				inputs = append(inputs, input)
			}
		}

	case "ec2":
		if instances, ok := planResources.([]models.EC2Instance); ok {
			for _, inst := range instances {
				input := policy.NewPolicyInput("aws_instance", inst.TerraformAddress)
				input.Resource.Planned = map[string]interface{}{
					"instance_type":        inst.InstanceType,
					"ami":                  inst.AMI,
					"subnet_id":            inst.SubnetID,
					"tags":                 inst.Tags,
					"ebs_optimized":        inst.EBSOptimized,
					"monitoring":           inst.Monitoring,
					"key_name":             inst.KeyName,
					"iam_instance_profile": inst.IAMInstanceProfile,
					"root_block_device": map[string]interface{}{
						"volume_type": inst.RootBlockDevice.VolumeType,
						"volume_size": inst.RootBlockDevice.VolumeSize,
						"encrypted":   inst.RootBlockDevice.Encrypted,
					},
				}

				// Add drift info if present
				if dr, ok := driftMap[inst.Name()]; ok {
					input.Resource.Drift = &policy.DriftInput{
						HasDrift: true,
						Missing:  dr.Missing,
					}
				}

				inputs = append(inputs, input)
			}
		}
	}

	return inputs
}

func init() {
	scanCmd.Flags().StringVarP(&configPath, "config", "c", "cloudrift.yml", "Path to Cloudrift config file")
	scanCmd.Flags().StringVarP(&service, "service", "s", "s3", "AWS service to scan (e.g., s3)")
	scanCmd.Flags().StringVarP(&outputFormat, "format", "f", "console", "Output format: console, json, sarif")
	scanCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Write output to file instead of stdout")
	scanCmd.Flags().StringVarP(&policyDir, "policy-dir", "p", "", "Directory containing custom OPA policies")
	scanCmd.Flags().BoolVar(&failOnViolation, "fail-on-violation", false, "Exit with non-zero code if policy violations found")
	scanCmd.Flags().BoolVar(&skipPolicies, "skip-policies", false, "Skip policy evaluation")
	scanCmd.Flags().BoolVar(&noEmoji, "no-emoji", false, "Use ASCII characters instead of emojis")
	rootCmd.AddCommand(scanCmd)
}
