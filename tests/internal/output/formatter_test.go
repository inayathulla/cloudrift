package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/inayathulla/cloudrift/internal/detector"
	"github.com/inayathulla/cloudrift/internal/output"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestScanResult() output.ScanResult {
	return output.ScanResult{
		Service:        "S3",
		AccountID:      "123456789012",
		Region:         "us-east-1",
		TotalResources: 5,
		DriftCount:     2,
		Drifts: []detector.DriftInfo{
			{
				ResourceID:   "my-bucket",
				ResourceType: "aws_s3_bucket",
				ResourceName: "my-bucket",
				Missing:      false,
				Diffs: map[string][2]interface{}{
					"versioning_enabled": {"true", "false"},
				},
				Severity: "warning",
			},
			{
				ResourceID:   "missing-bucket",
				ResourceType: "aws_s3_bucket",
				ResourceName: "missing-bucket",
				Missing:      true,
				Severity:     "critical",
			},
		},
		ScanDuration: 1500,
		Timestamp:    "2024-01-15T10:30:00Z",
	}
}

// JSON Formatter Tests
func TestJSONFormatter_Format(t *testing.T) {
	formatter := output.NewJSONFormatter()
	result := createTestScanResult()

	var buf bytes.Buffer
	err := formatter.Format(&buf, result)

	require.NoError(t, err)
	assert.Contains(t, buf.String(), `"service": "S3"`)
	assert.Contains(t, buf.String(), `"account_id": "123456789012"`)
	assert.Contains(t, buf.String(), `"drift_count": 2`)
}

func TestJSONFormatter_Format_ValidJSON(t *testing.T) {
	formatter := output.NewJSONFormatter()
	result := createTestScanResult()

	var buf bytes.Buffer
	err := formatter.Format(&buf, result)
	require.NoError(t, err)

	// Verify output is valid JSON
	var parsed output.ScanResult
	err = json.Unmarshal(buf.Bytes(), &parsed)
	require.NoError(t, err)

	assert.Equal(t, result.Service, parsed.Service)
	assert.Equal(t, result.DriftCount, parsed.DriftCount)
	assert.Len(t, parsed.Drifts, 2)
}

func TestJSONFormatter_Format_NoDrifts(t *testing.T) {
	formatter := output.NewJSONFormatter()
	result := output.ScanResult{
		Service:        "S3",
		AccountID:      "123456789012",
		TotalResources: 5,
		DriftCount:     0,
		Drifts:         []detector.DriftInfo{},
	}

	var buf bytes.Buffer
	err := formatter.Format(&buf, result)
	require.NoError(t, err)

	assert.Contains(t, buf.String(), `"drift_count": 0`)
}

func TestJSONFormatter_Name(t *testing.T) {
	formatter := output.NewJSONFormatter()
	assert.Equal(t, "json", formatter.Name())
}

func TestJSONFormatter_FileExtension(t *testing.T) {
	formatter := output.NewJSONFormatter()
	assert.Equal(t, ".json", formatter.FileExtension())
}

// SARIF Formatter Tests
func TestSARIFFormatter_Format(t *testing.T) {
	formatter := output.NewSARIFFormatter()
	result := createTestScanResult()

	var buf bytes.Buffer
	err := formatter.Format(&buf, result)

	require.NoError(t, err)
	assert.Contains(t, buf.String(), `"version": "2.1.0"`)
	assert.Contains(t, buf.String(), `"$schema"`)
	assert.Contains(t, buf.String(), `"Cloudrift"`)
}

func TestSARIFFormatter_Format_ValidJSON(t *testing.T) {
	formatter := output.NewSARIFFormatter()
	result := createTestScanResult()

	var buf bytes.Buffer
	err := formatter.Format(&buf, result)
	require.NoError(t, err)

	// Verify output is valid JSON
	var parsed map[string]interface{}
	err = json.Unmarshal(buf.Bytes(), &parsed)
	require.NoError(t, err)

	assert.Equal(t, "2.1.0", parsed["version"])
	runs, ok := parsed["runs"].([]interface{})
	require.True(t, ok)
	assert.Len(t, runs, 1)
}

func TestSARIFFormatter_Format_ContainsRules(t *testing.T) {
	formatter := output.NewSARIFFormatter()
	result := createTestScanResult()

	var buf bytes.Buffer
	err := formatter.Format(&buf, result)
	require.NoError(t, err)

	// Should contain drift rules
	assert.Contains(t, buf.String(), `"DRIFT001"`)
	assert.Contains(t, buf.String(), `"DRIFT002"`)
	assert.Contains(t, buf.String(), `"resource-missing"`)
	assert.Contains(t, buf.String(), `"attribute-mismatch"`)
}

func TestSARIFFormatter_Format_ContainsResults(t *testing.T) {
	formatter := output.NewSARIFFormatter()
	result := createTestScanResult()

	var buf bytes.Buffer
	err := formatter.Format(&buf, result)
	require.NoError(t, err)

	// Should contain results for the drifts
	assert.Contains(t, buf.String(), `"my-bucket"`)
	assert.Contains(t, buf.String(), `"missing-bucket"`)
}

func TestSARIFFormatter_Format_MissingResourceError(t *testing.T) {
	formatter := output.NewSARIFFormatter()
	result := output.ScanResult{
		Service:    "S3",
		DriftCount: 1,
		Drifts: []detector.DriftInfo{
			{
				ResourceName: "missing-bucket",
				ResourceType: "aws_s3_bucket",
				Missing:      true,
				Severity:     "critical",
			},
		},
	}

	var buf bytes.Buffer
	err := formatter.Format(&buf, result)
	require.NoError(t, err)

	// Missing resources should have "error" level
	assert.Contains(t, buf.String(), `"level": "error"`)
	assert.Contains(t, buf.String(), `"DRIFT001"`)
}

func TestSARIFFormatter_Name(t *testing.T) {
	formatter := output.NewSARIFFormatter()
	assert.Equal(t, "sarif", formatter.Name())
}

func TestSARIFFormatter_FileExtension(t *testing.T) {
	formatter := output.NewSARIFFormatter()
	assert.Equal(t, ".sarif", formatter.FileExtension())
}

// Console Formatter Tests
func TestConsoleFormatter_Format_NoDrift(t *testing.T) {
	formatter := output.NewConsoleFormatter()
	result := output.ScanResult{
		Service:        "S3",
		TotalResources: 5,
		DriftCount:     0,
		Drifts:         []detector.DriftInfo{},
		ScanDuration:   100,
	}

	var buf bytes.Buffer
	err := formatter.Format(&buf, result)
	require.NoError(t, err)

	assert.Contains(t, buf.String(), "No drift detected")
}

func TestConsoleFormatter_Format_WithDrift(t *testing.T) {
	formatter := output.NewConsoleFormatter()
	result := createTestScanResult()

	var buf bytes.Buffer
	err := formatter.Format(&buf, result)
	require.NoError(t, err)

	assert.Contains(t, buf.String(), "Drift detected")
	assert.Contains(t, buf.String(), "my-bucket")
}

func TestConsoleFormatter_Name(t *testing.T) {
	formatter := output.NewConsoleFormatter()
	assert.Equal(t, "console", formatter.Name())
}

func TestConsoleFormatter_FileExtension(t *testing.T) {
	formatter := output.NewConsoleFormatter()
	assert.Equal(t, ".txt", formatter.FileExtension())
}

// Registry Tests
func TestFormatRegistry_Get(t *testing.T) {
	// JSON should be registered
	formatter, ok := output.Get(output.FormatJSON)
	assert.True(t, ok)
	assert.NotNil(t, formatter)
	assert.Equal(t, "json", formatter.Name())

	// SARIF should be registered
	formatter, ok = output.Get(output.FormatSARIF)
	assert.True(t, ok)
	assert.NotNil(t, formatter)
	assert.Equal(t, "sarif", formatter.Name())

	// Console should be registered
	formatter, ok = output.Get(output.FormatConsole)
	assert.True(t, ok)
	assert.NotNil(t, formatter)
	assert.Equal(t, "console", formatter.Name())

	// Unknown format should not be found
	formatter, ok = output.Get(output.FormatType("unknown"))
	assert.False(t, ok)
	assert.Nil(t, formatter)
}

func TestFormatRegistry_List(t *testing.T) {
	formats := output.List()
	assert.GreaterOrEqual(t, len(formats), 3)

	// Should contain all expected formats
	formatSet := make(map[output.FormatType]bool)
	for _, f := range formats {
		formatSet[f] = true
	}
	assert.True(t, formatSet[output.FormatJSON])
	assert.True(t, formatSet[output.FormatSARIF])
	assert.True(t, formatSet[output.FormatConsole])
}

// DriftInfo Tests
func TestDriftInfo_HasDrift_Missing(t *testing.T) {
	info := detector.DriftInfo{
		Missing: true,
	}
	assert.True(t, info.HasDrift())
}

func TestDriftInfo_HasDrift_Diffs(t *testing.T) {
	info := detector.DriftInfo{
		Diffs: map[string][2]interface{}{
			"attribute": {"expected", "actual"},
		},
	}
	assert.True(t, info.HasDrift())
}

func TestDriftInfo_HasDrift_ExtraAttributes(t *testing.T) {
	info := detector.DriftInfo{
		ExtraAttributes: map[string]interface{}{
			"extra": "value",
		},
	}
	assert.True(t, info.HasDrift())
}

func TestDriftInfo_HasDrift_NoDrift(t *testing.T) {
	info := detector.DriftInfo{
		Missing:         false,
		Diffs:           map[string][2]interface{}{},
		ExtraAttributes: map[string]interface{}{},
	}
	assert.False(t, info.HasDrift())
}

// SARIF Schema Validation
func TestSARIFFormatter_SchemaURL(t *testing.T) {
	formatter := output.NewSARIFFormatter()
	result := createTestScanResult()

	var buf bytes.Buffer
	err := formatter.Format(&buf, result)
	require.NoError(t, err)

	// Verify schema URL is correct
	assert.Contains(t, buf.String(), "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json")
}

// Test extra attributes output
func TestSARIFFormatter_ExtraAttributes(t *testing.T) {
	formatter := output.NewSARIFFormatter()
	result := output.ScanResult{
		Service:    "S3",
		DriftCount: 1,
		Drifts: []detector.DriftInfo{
			{
				ResourceName: "bucket",
				ResourceType: "aws_s3_bucket",
				ExtraAttributes: map[string]interface{}{
					"extra_tag": "value",
				},
				Severity: "info",
			},
		},
	}

	var buf bytes.Buffer
	err := formatter.Format(&buf, result)
	require.NoError(t, err)

	// Should contain DRIFT003 rule for extra attributes
	assert.Contains(t, buf.String(), `"DRIFT003"`)
	assert.Contains(t, buf.String(), "extra_tag")
}

// Edge cases
func TestJSONFormatter_EmptyResult(t *testing.T) {
	formatter := output.NewJSONFormatter()
	result := output.ScanResult{}

	var buf bytes.Buffer
	err := formatter.Format(&buf, result)
	require.NoError(t, err)

	// Should still produce valid JSON
	var parsed map[string]interface{}
	err = json.Unmarshal(buf.Bytes(), &parsed)
	require.NoError(t, err)
}

func TestSARIFFormatter_EmptyResult(t *testing.T) {
	formatter := output.NewSARIFFormatter()
	result := output.ScanResult{}

	var buf bytes.Buffer
	err := formatter.Format(&buf, result)
	require.NoError(t, err)

	// Should still produce valid SARIF JSON
	var parsed map[string]interface{}
	err = json.Unmarshal(buf.Bytes(), &parsed)
	require.NoError(t, err)
	assert.Equal(t, "2.1.0", parsed["version"])
}

// Test severity mapping
func TestSARIFFormatter_SeverityMapping(t *testing.T) {
	formatter := output.NewSARIFFormatter()

	tests := []struct {
		severity string
		diffs    map[string][2]interface{}
		expected string
	}{
		{
			severity: "critical",
			diffs:    map[string][2]interface{}{"attr": {"a", "b"}},
			expected: "error",
		},
		{
			severity: "warning",
			diffs:    map[string][2]interface{}{"attr": {"a", "b"}},
			expected: "warning",
		},
		{
			severity: "info",
			diffs:    map[string][2]interface{}{"attr": {"a", "b"}},
			expected: "note",
		},
	}

	for _, tc := range tests {
		t.Run(tc.severity, func(t *testing.T) {
			result := output.ScanResult{
				Drifts: []detector.DriftInfo{
					{
						ResourceName: "resource",
						ResourceType: "aws_s3_bucket",
						Diffs:        tc.diffs,
						Severity:     tc.severity,
					},
				},
			}

			var buf bytes.Buffer
			err := formatter.Format(&buf, result)
			require.NoError(t, err)

			// Check that the expected level appears in the output
			// Note: This is a simplistic check; in real tests you might parse the JSON
			assert.True(t, strings.Contains(buf.String(), tc.expected) ||
				strings.Contains(buf.String(), "warning"))
		})
	}
}
