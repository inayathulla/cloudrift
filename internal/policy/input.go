package policy

// PolicyInput is the input structure passed to OPA for policy evaluation.
type PolicyInput struct {
	// Resource contains information about the resource being evaluated.
	Resource ResourceInput `json:"resource"`

	// Metadata contains additional context about the evaluation.
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// ResourceInput represents a resource being evaluated.
type ResourceInput struct {
	// Type is the Terraform resource type (e.g., "aws_s3_bucket").
	Type string `json:"type"`

	// Address is the Terraform resource address.
	Address string `json:"address"`

	// Planned contains the planned (Terraform plan) attributes.
	Planned map[string]interface{} `json:"planned,omitempty"`

	// Live contains the live (AWS actual) attributes.
	Live map[string]interface{} `json:"live,omitempty"`

	// Drift contains information about detected drift.
	Drift *DriftInput `json:"drift,omitempty"`
}

// DriftInput contains drift detection information for a resource.
type DriftInput struct {
	// HasDrift indicates if drift was detected.
	HasDrift bool `json:"has_drift"`

	// Missing indicates if the resource is missing from AWS.
	Missing bool `json:"missing"`

	// Diffs maps attribute names to [expected, actual] values.
	Diffs map[string][2]interface{} `json:"diffs,omitempty"`

	// ExtraAttributes contains attributes in AWS but not in the plan.
	ExtraAttributes map[string]interface{} `json:"extra_attributes,omitempty"`
}

// NewPolicyInput creates a new PolicyInput with default values.
func NewPolicyInput(resourceType, address string) *PolicyInput {
	return &PolicyInput{
		Resource: ResourceInput{
			Type:    resourceType,
			Address: address,
			Planned: make(map[string]interface{}),
			Live:    make(map[string]interface{}),
		},
		Metadata: make(map[string]interface{}),
	}
}
