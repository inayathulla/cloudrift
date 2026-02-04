// Package detector provides drift detection capabilities for AWS resources.
//
// The package uses a plugin architecture where each AWS service (S3, EC2, IAM, etc.)
// implements the Detector interface to provide consistent drift detection behavior.
package detector

import (
	sdkaws "github.com/aws/aws-sdk-go-v2/aws"
)

// Resource represents a generic cloud resource that can be compared for drift.
// Each service-specific implementation provides concrete types that satisfy this interface.
type Resource interface {
	// ResourceID returns a unique identifier for the resource.
	ResourceID() string

	// ResourceType returns the Terraform resource type (e.g., "aws_s3_bucket", "aws_instance").
	ResourceType() string

	// ResourceName returns a human-readable name for the resource.
	ResourceName() string

	// Attributes returns a map of resource attributes for comparison and policy evaluation.
	Attributes() map[string]interface{}
}

// DriftInfo captures drift information for a single resource.
type DriftInfo struct {
	// ResourceID is the unique identifier of the resource.
	ResourceID string `json:"resource_id"`

	// ResourceType is the Terraform resource type.
	ResourceType string `json:"resource_type"`

	// ResourceName is the human-readable name.
	ResourceName string `json:"resource_name"`

	// Missing is true if the resource exists in the plan but not in AWS.
	Missing bool `json:"missing"`

	// Diffs contains attribute-level differences.
	// Key is the attribute name, value is [expected, actual].
	Diffs map[string][2]interface{} `json:"diffs,omitempty"`

	// ExtraAttributes contains attributes present in AWS but not in the plan.
	ExtraAttributes map[string]interface{} `json:"extra_attributes,omitempty"`

	// Severity indicates the importance of this drift (info, warning, critical).
	Severity string `json:"severity"`
}

// HasDrift returns true if any drift was detected.
func (d DriftInfo) HasDrift() bool {
	return d.Missing || len(d.Diffs) > 0 || len(d.ExtraAttributes) > 0
}

// Detector defines the interface for service-specific drift detectors.
//
// Each supported AWS service (S3, EC2, IAM, etc.) implements this interface
// to provide consistent drift detection behavior across services.
type Detector interface {
	// ServiceName returns the name of the AWS service (e.g., "s3", "ec2", "iam").
	ServiceName() string

	// TerraformTypes returns the Terraform resource types this detector handles.
	// For example, S3 detector handles "aws_s3_bucket", "aws_s3_bucket_versioning", etc.
	TerraformTypes() []string

	// FetchLiveState retrieves the current state of resources from AWS.
	FetchLiveState(cfg sdkaws.Config) ([]Resource, error)

	// ParsePlanResources extracts resources from a Terraform plan's resource_changes.
	// The input is a slice of resource change maps from the plan JSON.
	ParsePlanResources(resourceChanges []map[string]interface{}) ([]Resource, error)

	// DetectDrift compares planned resources against live resources and returns drift info.
	DetectDrift(planned, live []Resource) ([]DriftInfo, error)
}

// DetectorFactory is a function that creates a new detector instance.
type DetectorFactory func() Detector
