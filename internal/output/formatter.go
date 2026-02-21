// Package output provides formatters for drift detection results.
//
// Supported formats include:
//   - Console: Colorized CLI output (default)
//   - JSON: Machine-readable JSON format
//   - SARIF: Static Analysis Results Interchange Format for GitHub/GitLab integration
package output

import (
	"io"

	"github.com/inayathulla/cloudrift/internal/detector"
)

// PolicyOutput contains the results of policy evaluation in a JSON-friendly format.
type PolicyOutput struct {
	// Violations is the list of policy violations found.
	Violations []PolicyViolationOutput `json:"violations"`

	// Warnings contains non-blocking policy warnings.
	Warnings []PolicyViolationOutput `json:"warnings,omitempty"`

	// Passed indicates the number of policies that passed.
	Passed int `json:"passed"`

	// Failed indicates the number of policies that failed.
	Failed int `json:"failed"`

	// ComplianceResult contains compliance scoring data.
	ComplianceResult *ComplianceOutput `json:"compliance,omitempty"`
}

// ComplianceOutput contains compliance scoring results.
type ComplianceOutput struct {
	OverallPercentage float64                   `json:"overall_percentage"`
	TotalPolicies     int                       `json:"total_policies"`
	PassingPolicies   int                       `json:"passing_policies"`
	FailingPolicies   int                       `json:"failing_policies"`
	Categories        map[string]CategoryScore  `json:"categories"`
	Frameworks        map[string]FrameworkScore `json:"frameworks"`
	ActiveFrameworks  []string                  `json:"active_frameworks,omitempty"`
}

// CategoryScore holds pass/fail counts for a single policy category.
type CategoryScore struct {
	Percentage float64 `json:"percentage"`
	Passed     int     `json:"passed"`
	Failed     int     `json:"failed"`
	Total      int     `json:"total"`
}

// FrameworkScore holds pass/fail counts for a single compliance framework.
type FrameworkScore struct {
	Percentage float64 `json:"percentage"`
	Passed     int     `json:"passed"`
	Failed     int     `json:"failed"`
	Total      int     `json:"total"`
}

// PolicyViolationOutput represents a single policy violation in JSON output.
type PolicyViolationOutput struct {
	PolicyID        string   `json:"policy_id"`
	PolicyName      string   `json:"policy_name"`
	Message         string   `json:"message"`
	Severity        string   `json:"severity"`
	ResourceType    string   `json:"resource_type"`
	ResourceAddress string   `json:"resource_address"`
	Remediation     string   `json:"remediation,omitempty"`
	Category        string   `json:"category,omitempty"`
	Frameworks      []string `json:"frameworks,omitempty"`
}

// ScanResult contains the complete results of a drift scan.
type ScanResult struct {
	// Service is the AWS service that was scanned (e.g., "s3", "ec2").
	Service string `json:"service"`

	// AccountID is the AWS account that was scanned.
	AccountID string `json:"account_id,omitempty"`

	// Region is the AWS region that was scanned.
	Region string `json:"region,omitempty"`

	// TotalResources is the number of resources scanned.
	TotalResources int `json:"total_resources"`

	// DriftCount is the number of resources with drift.
	DriftCount int `json:"drift_count"`

	// Drifts contains detailed drift information for each resource.
	Drifts []detector.DriftInfo `json:"drifts"`

	// PolicyResult contains policy evaluation results (nil if policies were skipped).
	PolicyResult *PolicyOutput `json:"policy_result,omitempty"`

	// ScanDuration is how long the scan took in milliseconds.
	ScanDuration int64 `json:"scan_duration_ms"`

	// Timestamp is when the scan was performed (ISO 8601).
	Timestamp string `json:"timestamp"`
}

// Formatter defines the interface for output formatters.
//
// Each format (JSON, SARIF, Console) implements this interface
// to provide consistent output generation.
type Formatter interface {
	// Format writes the scan result to the provided writer.
	Format(w io.Writer, result ScanResult) error

	// Name returns the format name (e.g., "json", "sarif", "console").
	Name() string

	// FileExtension returns the recommended file extension (e.g., ".json", ".sarif").
	FileExtension() string
}

// FormatType represents the output format type.
type FormatType string

const (
	FormatConsole FormatType = "console"
	FormatJSON    FormatType = "json"
	FormatSARIF   FormatType = "sarif"
)

// registry holds registered formatters.
var registry = make(map[FormatType]Formatter)

// Register adds a formatter to the registry.
func Register(format FormatType, formatter Formatter) {
	registry[format] = formatter
}

// Get retrieves a formatter by type.
func Get(format FormatType) (Formatter, bool) {
	f, ok := registry[format]
	return f, ok
}

// List returns all registered format types.
func List() []FormatType {
	formats := make([]FormatType, 0, len(registry))
	for f := range registry {
		formats = append(formats, f)
	}
	return formats
}
