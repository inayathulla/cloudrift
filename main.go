package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "cloudrift",
		Short: "Detect cloud drift between Terraform and real infrastructure",
		Long: `Cloudrift is an open-source tool that detects infrastructure drift — when your live cloud resources
no longer match what's defined in your Terraform IaC.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Use 'cloudrift scan' to run a drift detection scan.")
		},
	}

	rootCmd.AddCommand(scanCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Run a drift detection scan",
	Run: func(cmd *cobra.Command, args []string) {
		// Placeholder: Replace with actual scanning logic
		fmt.Println("🔍 Scanning for infrastructure drift...")
		// TODO: Call scanner.Scan() and display result
	},
}
