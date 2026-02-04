package policy

// Severity indicates the importance level of a policy violation.
type Severity string

const (
	SeverityCritical Severity = "critical"
	SeverityHigh     Severity = "high"
	SeverityMedium   Severity = "medium"
	SeverityLow      Severity = "low"
	SeverityInfo     Severity = "info"
)

// Violation represents a policy violation detected during evaluation.
type Violation struct {
	// PolicyID is the unique identifier for the policy that was violated.
	PolicyID string `json:"policy_id"`

	// PolicyName is a human-readable name for the policy.
	PolicyName string `json:"policy_name"`

	// Message describes the violation in detail.
	Message string `json:"message"`

	// Severity indicates how critical the violation is.
	Severity Severity `json:"severity"`

	// ResourceType is the type of resource that violated the policy (e.g., "aws_s3_bucket").
	ResourceType string `json:"resource_type"`

	// ResourceAddress is the Terraform address of the resource.
	ResourceAddress string `json:"resource_address"`

	// Remediation provides guidance on how to fix the violation.
	Remediation string `json:"remediation,omitempty"`

	// Metadata contains additional context about the violation.
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// EvaluationResult contains the results of policy evaluation.
type EvaluationResult struct {
	// Violations is the list of policy violations found.
	Violations []Violation `json:"violations"`

	// Passed indicates the number of policies that passed.
	Passed int `json:"passed"`

	// Failed indicates the number of policies that failed.
	Failed int `json:"failed"`

	// Warnings contains non-blocking policy warnings.
	Warnings []Violation `json:"warnings,omitempty"`
}

// HasViolations returns true if there are any violations.
func (r *EvaluationResult) HasViolations() bool {
	return len(r.Violations) > 0
}

// HasCriticalViolations returns true if any violations are critical.
func (r *EvaluationResult) HasCriticalViolations() bool {
	for _, v := range r.Violations {
		if v.Severity == SeverityCritical {
			return true
		}
	}
	return false
}

// BySeverity returns violations filtered by severity.
func (r *EvaluationResult) BySeverity(severity Severity) []Violation {
	var filtered []Violation
	for _, v := range r.Violations {
		if v.Severity == severity {
			filtered = append(filtered, v)
		}
	}
	return filtered
}
