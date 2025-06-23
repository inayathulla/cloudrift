package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "cloudrift",
	Short: "Cloudrift - Cloud Drift Detection Tool",
	Long: `Cloudrift helps detect configuration drift between Terraform-managed AWS infrastructure and the actual deployed state.

You provide a Terraform plan in JSON format and Cloudrift compares it with live AWS resources.`,
	SilenceUsage: true,
}

// Execute starts the CLI
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	// Add all subcommands here
	rootCmd.AddCommand(scanCmd)
}
