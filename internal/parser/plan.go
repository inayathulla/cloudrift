// Package parser provides functionality for reading and parsing Terraform plan files.
//
// Terraform plans exported as JSON (via `terraform show -json`) contain detailed
// information about intended infrastructure changes. This package extracts
// resource configurations from these plans for drift comparison against live state.
//
// Usage:
//
//	buckets, err := parser.LoadPlan("path/to/plan.json")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	// buckets now contains all S3 bucket configurations from the plan
package parser

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/inayathulla/cloudrift/internal/models"
)

// TerraformPlan represents the top-level structure of a Terraform JSON plan.
// This struct maps to the output of `terraform show -json <plan>`.
type TerraformPlan struct {
	// ResourceChanges contains all resources that will be created, updated, or deleted.
	ResourceChanges []ResourceChange `json:"resource_changes"`
}

// ResourceChange represents a single resource modification in the Terraform plan.
// Each change includes the resource's address, type, and the planned state.
type ResourceChange struct {
	// Address is the fully-qualified resource address (e.g., "aws_s3_bucket.my_bucket").
	Address string `json:"address"`

	// Type is the resource type (e.g., "aws_s3_bucket", "aws_instance").
	Type string `json:"type"`

	// Name is the local name of the resource within the Terraform configuration.
	Name string `json:"name"`

	// Change contains the planned actions and resulting state.
	Change Change `json:"change"`
}

// Change describes what Terraform will do to a resource and the resulting state.
type Change struct {
	// Actions lists the operations Terraform will perform (e.g., ["create"], ["update"], ["delete"]).
	Actions []string `json:"actions"`

	// After contains the planned attribute values after the change is applied.
	// For create/update actions, this represents the desired state.
	// For delete actions, this is null.
	After map[string]interface{} `json:"after"`
}

// LoadPlan reads a Terraform JSON plan file and extracts S3 bucket configurations.
//
// The function opens the specified file, decodes it as a Terraform plan,
// and extracts all aws_s3_bucket resources using ParseS3Buckets.
//
// Parameters:
//   - path: filesystem path to the Terraform plan JSON file
//
// Returns:
//   - []models.S3Bucket: slice of S3 bucket configurations from the plan
//   - error: if the file cannot be read or parsed
func LoadPlan(path string) ([]models.S3Bucket, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open plan file: %w", err)
	}
	defer file.Close()

	var plan TerraformPlan
	if err := json.NewDecoder(file).Decode(&plan); err != nil {
		return nil, fmt.Errorf("failed to decode JSON: %w", err)
	}

	return ParseS3Buckets(&plan), nil
}
