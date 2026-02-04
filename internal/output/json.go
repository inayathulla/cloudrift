package output

import (
	"encoding/json"
	"io"
)

// JSONFormatter outputs scan results as JSON.
//
// This format is ideal for CI/CD pipelines, scripting, and machine processing.
// The output is pretty-printed with 2-space indentation for readability.
type JSONFormatter struct{}

// NewJSONFormatter creates a new JSON formatter.
func NewJSONFormatter() *JSONFormatter {
	return &JSONFormatter{}
}

// Format writes the scan result as JSON to the provided writer.
func (f *JSONFormatter) Format(w io.Writer, result ScanResult) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(result)
}

// Name returns the format name.
func (f *JSONFormatter) Name() string {
	return "json"
}

// FileExtension returns the recommended file extension.
func (f *JSONFormatter) FileExtension() string {
	return ".json"
}

func init() {
	Register(FormatJSON, NewJSONFormatter())
}
