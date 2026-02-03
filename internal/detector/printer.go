// Package detector implements drift detection logic for comparing Terraform plans
// against live AWS infrastructure state.
//
// The package provides service-specific detectors (e.g., S3DriftDetector) that
// compare planned resource configurations with their actual state in AWS,
// identifying attribute-level differences.
//
// Architecture:
//
//	┌─────────────────┐     ┌─────────────────┐
//	│  Terraform Plan │────▶│   Detector      │
//	│    (parsed)     │     │                 │
//	└─────────────────┘     │  Compare and    │
//	                        │  identify diffs │
//	┌─────────────────┐     │                 │
//	│   AWS Live      │────▶│                 │
//	│    State        │     └────────┬────────┘
//	└─────────────────┘              │
//	                                 ▼
//	                        ┌─────────────────┐
//	                        │  DriftResult[]  │
//	                        └─────────────────┘
package detector

// DriftResultPrinter defines the interface for outputting drift detection results.
//
// Implementations of this interface handle the presentation of drift results,
// whether as colorized CLI output, JSON, or other formats.
type DriftResultPrinter interface {
	// PrintDrift outputs the drift detection results.
	//
	// Parameters:
	//   - results: drift detection results (service-specific, e.g., []DriftResult)
	//   - plan: planned resource configurations from Terraform
	//   - live: actual resource configurations from AWS
	PrintDrift(results interface{}, plan, live interface{})
}
