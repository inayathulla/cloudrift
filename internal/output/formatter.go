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
