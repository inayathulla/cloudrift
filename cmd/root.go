// Package cmd implements the Cloudrift command-line interface.
//
// The CLI is built using Cobra and provides subcommands for different operations.
// Currently supported commands:
//   - scan: Detect drift between Terraform plans and live AWS state
//
// Usage:
//
//	cloudrift scan --config=config/cloudrift-s3.yml --service=s3
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "cloudrift",
	Short: "Cloudrift - Cloud Drift Detection Tool",
	Long: `Cloudrift helps detect configuration drift between Terraform-managed
AWS infrastructure and the actual deployed state.

You provide a Terraform plan in JSON format and Cloudrift compares it
with live AWS resources, identifying attribute-level differences before
you apply changes.

Example:
  cloudrift scan --config=config/cloudrift-s3.yml --service=s3`,
	SilenceUsage: true,
}

// Execute runs the root command and handles any errors.
// This is the main entry point called from main.go.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(scanCmd)
}
