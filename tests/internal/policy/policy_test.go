package policy

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/inayathulla/cloudrift/internal/policy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test Violation struct
func TestViolation_Fields(t *testing.T) {
	v := policy.Violation{
		PolicyID:        "S3-001",
		PolicyName:      "S3 Encryption Required",
		Message:         "Bucket must have encryption",
		Severity:        policy.SeverityHigh,
		ResourceType:    "aws_s3_bucket",
		ResourceAddress: "aws_s3_bucket.test",
		Remediation:     "Enable encryption",
	}

	assert.Equal(t, "S3-001", v.PolicyID)
	assert.Equal(t, "S3 Encryption Required", v.PolicyName)
	assert.Equal(t, policy.SeverityHigh, v.Severity)
}

// Test EvaluationResult methods
func TestEvaluationResult_HasViolations(t *testing.T) {
	tests := []struct {
		name     string
		result   policy.EvaluationResult
		expected bool
	}{
		{
			name:     "no violations",
			result:   policy.EvaluationResult{},
			expected: false,
		},
		{
			name: "has violations",
			result: policy.EvaluationResult{
				Violations: []policy.Violation{{Message: "test"}},
			},
			expected: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.result.HasViolations())
		})
	}
}

func TestEvaluationResult_HasCriticalViolations(t *testing.T) {
	tests := []struct {
		name     string
		result   policy.EvaluationResult
		expected bool
	}{
		{
			name:     "no violations",
			result:   policy.EvaluationResult{},
			expected: false,
		},
		{
			name: "high severity only",
			result: policy.EvaluationResult{
				Violations: []policy.Violation{{Severity: policy.SeverityHigh}},
			},
			expected: false,
		},
		{
			name: "has critical",
			result: policy.EvaluationResult{
				Violations: []policy.Violation{{Severity: policy.SeverityCritical}},
			},
			expected: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.result.HasCriticalViolations())
		})
	}
}

func TestEvaluationResult_BySeverity(t *testing.T) {
	result := policy.EvaluationResult{
		Violations: []policy.Violation{
			{PolicyID: "1", Severity: policy.SeverityCritical},
			{PolicyID: "2", Severity: policy.SeverityHigh},
			{PolicyID: "3", Severity: policy.SeverityHigh},
			{PolicyID: "4", Severity: policy.SeverityMedium},
		},
	}

	critical := result.BySeverity(policy.SeverityCritical)
	assert.Len(t, critical, 1)
	assert.Equal(t, "1", critical[0].PolicyID)

	high := result.BySeverity(policy.SeverityHigh)
	assert.Len(t, high, 2)

	low := result.BySeverity(policy.SeverityLow)
	assert.Empty(t, low)
}

// Test PolicyInput
func TestNewPolicyInput(t *testing.T) {
	input := policy.NewPolicyInput("aws_s3_bucket", "module.storage.aws_s3_bucket.data")

	assert.Equal(t, "aws_s3_bucket", input.Resource.Type)
	assert.Equal(t, "module.storage.aws_s3_bucket.data", input.Resource.Address)
	assert.NotNil(t, input.Resource.Planned)
	assert.NotNil(t, input.Resource.Live)
	assert.NotNil(t, input.Metadata)
}

// Test Engine creation
func TestNewEngine_EmptyPath(t *testing.T) {
	// Create empty temp directory
	tmpDir := t.TempDir()

	engine, err := policy.NewEngine(tmpDir)
	require.NoError(t, err)
	assert.Equal(t, 0, engine.PolicyCount())
}

func TestNewEngine_InvalidPath(t *testing.T) {
	_, err := policy.NewEngine("/nonexistent/path")
	assert.Error(t, err)
}

func TestNewEngine_WithPolicy(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a simple policy
	policyContent := `
package test.simple

deny[msg] {
	input.resource.type == "aws_s3_bucket"
	not input.resource.planned.encryption_algorithm
	msg := "Bucket must have encryption"
}
`
	err := os.WriteFile(filepath.Join(tmpDir, "test.rego"), []byte(policyContent), 0644)
	require.NoError(t, err)

	engine, err := policy.NewEngine(tmpDir)
	require.NoError(t, err)
	assert.Equal(t, 1, engine.PolicyCount())
}

func TestNewEngine_InvalidPolicy(t *testing.T) {
	tmpDir := t.TempDir()

	// Create an invalid policy
	err := os.WriteFile(filepath.Join(tmpDir, "invalid.rego"), []byte("this is not valid rego"), 0644)
	require.NoError(t, err)

	_, err = policy.NewEngine(tmpDir)
	assert.Error(t, err)
}

// Test policy evaluation
func TestEngine_Evaluate_NoPolicies(t *testing.T) {
	tmpDir := t.TempDir()
	engine, err := policy.NewEngine(tmpDir)
	require.NoError(t, err)

	input := policy.NewPolicyInput("aws_s3_bucket", "test")
	result, err := engine.Evaluate(context.Background(), input)

	require.NoError(t, err)
	assert.Empty(t, result.Violations)
}

func TestEngine_Evaluate_DenyRule(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a policy that denies unencrypted buckets using classic OPA syntax
	policyContent := `
package test.encryption

deny[msg] {
	input.resource.type == "aws_s3_bucket"
	not input.resource.planned.encryption_algorithm
	msg := "S3 bucket must have encryption enabled"
}
`
	err := os.WriteFile(filepath.Join(tmpDir, "encryption.rego"), []byte(policyContent), 0644)
	require.NoError(t, err)

	engine, err := policy.NewEngine(tmpDir)
	require.NoError(t, err)

	// Test with unencrypted bucket
	input := policy.NewPolicyInput("aws_s3_bucket", "aws_s3_bucket.test")
	input.Resource.Planned = map[string]interface{}{
		"bucket": "my-bucket",
	}

	result, err := engine.Evaluate(context.Background(), input)
	require.NoError(t, err)
	require.Len(t, result.Violations, 1)
	assert.Contains(t, result.Violations[0].Message, "encryption")
}

func TestEngine_Evaluate_DenyRule_Pass(t *testing.T) {
	tmpDir := t.TempDir()

	policyContent := `
package test.encryption

deny[msg] {
	input.resource.type == "aws_s3_bucket"
	not input.resource.planned.encryption_algorithm
	msg := "S3 bucket must have encryption enabled"
}
`
	err := os.WriteFile(filepath.Join(tmpDir, "encryption.rego"), []byte(policyContent), 0644)
	require.NoError(t, err)

	engine, err := policy.NewEngine(tmpDir)
	require.NoError(t, err)

	// Test with encrypted bucket - should pass
	input := policy.NewPolicyInput("aws_s3_bucket", "aws_s3_bucket.test")
	input.Resource.Planned = map[string]interface{}{
		"bucket":               "my-bucket",
		"encryption_algorithm": "AES256",
	}

	result, err := engine.Evaluate(context.Background(), input)
	require.NoError(t, err)
	assert.Empty(t, result.Violations)
}

func TestEngine_Evaluate_WarnRule(t *testing.T) {
	tmpDir := t.TempDir()

	policyContent := `
package test.tags

warn[msg] {
	input.resource.type == "aws_s3_bucket"
	not input.resource.planned.tags.Environment
	msg := "Bucket should have Environment tag"
}
`
	err := os.WriteFile(filepath.Join(tmpDir, "tags.rego"), []byte(policyContent), 0644)
	require.NoError(t, err)

	engine, err := policy.NewEngine(tmpDir)
	require.NoError(t, err)

	input := policy.NewPolicyInput("aws_s3_bucket", "aws_s3_bucket.test")
	input.Resource.Planned = map[string]interface{}{
		"bucket": "my-bucket",
		"tags":   map[string]interface{}{},
	}

	result, err := engine.Evaluate(context.Background(), input)
	require.NoError(t, err)
	assert.Empty(t, result.Violations)
	assert.Len(t, result.Warnings, 1)
}

func TestEngine_Evaluate_RichViolation(t *testing.T) {
	tmpDir := t.TempDir()

	policyContent := `
package test.rich

deny[result] {
	input.resource.type == "aws_s3_bucket"
	result := {
		"policy_id": "S3-001",
		"policy_name": "Encryption Required",
		"msg": "Bucket must be encrypted",
		"severity": "critical",
		"remediation": "Add encryption configuration"
	}
}
`
	err := os.WriteFile(filepath.Join(tmpDir, "rich.rego"), []byte(policyContent), 0644)
	require.NoError(t, err)

	engine, err := policy.NewEngine(tmpDir)
	require.NoError(t, err)

	input := policy.NewPolicyInput("aws_s3_bucket", "aws_s3_bucket.test")
	result, err := engine.Evaluate(context.Background(), input)

	require.NoError(t, err)
	require.Len(t, result.Violations, 1)

	v := result.Violations[0]
	assert.Equal(t, "S3-001", v.PolicyID)
	assert.Equal(t, "Encryption Required", v.PolicyName)
	assert.Equal(t, "Bucket must be encrypted", v.Message)
	assert.Equal(t, policy.Severity("critical"), v.Severity)
	assert.Equal(t, "Add encryption configuration", v.Remediation)
}

func TestEngine_Evaluate_NonMatchingResourceType(t *testing.T) {
	tmpDir := t.TempDir()

	policyContent := `
package test.s3only

deny[msg] {
	input.resource.type == "aws_s3_bucket"
	msg := "S3 specific rule"
}
`
	err := os.WriteFile(filepath.Join(tmpDir, "s3only.rego"), []byte(policyContent), 0644)
	require.NoError(t, err)

	engine, err := policy.NewEngine(tmpDir)
	require.NoError(t, err)

	// Test with EC2 instance - should not trigger S3 rule
	input := policy.NewPolicyInput("aws_instance", "aws_instance.web")
	result, err := engine.Evaluate(context.Background(), input)

	require.NoError(t, err)
	assert.Empty(t, result.Violations)
}

// Test EvaluateAll
func TestEngine_EvaluateAll(t *testing.T) {
	tmpDir := t.TempDir()

	policyContent := `
package test.all

deny[msg] {
	input.resource.type == "aws_s3_bucket"
	not input.resource.planned.encryption_algorithm
	msg := sprintf("Bucket %s needs encryption", [input.resource.address])
}
`
	err := os.WriteFile(filepath.Join(tmpDir, "all.rego"), []byte(policyContent), 0644)
	require.NoError(t, err)

	engine, err := policy.NewEngine(tmpDir)
	require.NoError(t, err)

	inputs := []*policy.PolicyInput{
		{
			Resource: policy.ResourceInput{
				Type:    "aws_s3_bucket",
				Address: "bucket1",
				Planned: map[string]interface{}{"bucket": "b1"},
			},
		},
		{
			Resource: policy.ResourceInput{
				Type:    "aws_s3_bucket",
				Address: "bucket2",
				Planned: map[string]interface{}{"bucket": "b2", "encryption_algorithm": "AES256"},
			},
		},
		{
			Resource: policy.ResourceInput{
				Type:    "aws_s3_bucket",
				Address: "bucket3",
				Planned: map[string]interface{}{"bucket": "b3"},
			},
		},
	}

	result, err := engine.EvaluateAll(context.Background(), inputs)
	require.NoError(t, err)
	assert.Len(t, result.Violations, 2) // bucket1 and bucket3 fail
}

// Test multiple policies
func TestEngine_MultiplePolicies(t *testing.T) {
	tmpDir := t.TempDir()

	// Policy 1
	policy1 := `
package test.policy1

deny[msg] {
	input.resource.type == "aws_s3_bucket"
	not input.resource.planned.encryption_algorithm
	msg := "No encryption"
}
`
	err := os.WriteFile(filepath.Join(tmpDir, "policy1.rego"), []byte(policy1), 0644)
	require.NoError(t, err)

	// Policy 2
	policy2 := `
package test.policy2

deny[msg] {
	input.resource.type == "aws_s3_bucket"
	not input.resource.planned.versioning_enabled
	msg := "No versioning"
}
`
	err = os.WriteFile(filepath.Join(tmpDir, "policy2.rego"), []byte(policy2), 0644)
	require.NoError(t, err)

	engine, err := policy.NewEngine(tmpDir)
	require.NoError(t, err)
	assert.Equal(t, 2, engine.PolicyCount())

	input := policy.NewPolicyInput("aws_s3_bucket", "test")
	input.Resource.Planned = map[string]interface{}{
		"bucket": "my-bucket",
	}

	result, err := engine.Evaluate(context.Background(), input)
	require.NoError(t, err)
	assert.Len(t, result.Violations, 2)
}

// Test nested policy directories
func TestEngine_NestedDirectories(t *testing.T) {
	tmpDir := t.TempDir()

	// Create nested structure
	securityDir := filepath.Join(tmpDir, "security")
	err := os.MkdirAll(securityDir, 0755)
	require.NoError(t, err)

	policyContent := `
package security.s3

deny[msg] {
	input.resource.type == "aws_s3_bucket"
	msg := "Test violation"
}
`
	err = os.WriteFile(filepath.Join(securityDir, "s3.rego"), []byte(policyContent), 0644)
	require.NoError(t, err)

	engine, err := policy.NewEngine(tmpDir)
	require.NoError(t, err)
	assert.Equal(t, 1, engine.PolicyCount())
}

// Test severity constants
func TestSeverityConstants(t *testing.T) {
	assert.Equal(t, policy.Severity("critical"), policy.SeverityCritical)
	assert.Equal(t, policy.Severity("high"), policy.SeverityHigh)
	assert.Equal(t, policy.Severity("medium"), policy.SeverityMedium)
	assert.Equal(t, policy.Severity("low"), policy.SeverityLow)
	assert.Equal(t, policy.Severity("info"), policy.SeverityInfo)
}

// Test DriftInput
func TestDriftInput(t *testing.T) {
	input := policy.NewPolicyInput("aws_s3_bucket", "test")
	input.Resource.Drift = &policy.DriftInput{
		HasDrift: true,
		Missing:  false,
		Diffs: map[string][2]interface{}{
			"encryption": {"AES256", nil},
		},
	}

	assert.True(t, input.Resource.Drift.HasDrift)
	assert.False(t, input.Resource.Drift.Missing)
	assert.Len(t, input.Resource.Drift.Diffs, 1)
}
