package parser

import (
	"encoding/json"
	"fmt"
	"github.com/inayathulla/cloudrift/internal/models"
	"os"
)

// TerraformPlan represents the top-level structure of a Terraform JSON plan.
type TerraformPlan struct {
	ResourceChanges []ResourceChange `json:"resource_changes"`
}

// ResourceChange represents a change to a resource in the plan.
type ResourceChange struct {
	Address string `json:"address"`
	Type    string `json:"type"`
	Name    string `json:"name"`
	Change  Change `json:"change"`
}

// Change contains the actions Terraform will take and the resulting state.
type Change struct {
	Actions []string               `json:"actions"`
	After   map[string]interface{} `json:"after"`
}

// LoadPlan extracts only S3 bucket resources from the Terraform JSON plan.
func LoadPlan(path string) ([]models.S3Bucket, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open plan file: %w", err)
	}
	defer file.Close()

	var raw struct {
		PlannedValues struct {
			RootModule struct {
				Resources []struct {
					Type   string `json:"type"`
					Name   string `json:"name"`
					Values struct {
						Acl  string            `json:"acl"`
						Tags map[string]string `json:"tags"`
					} `json:"values"`
				} `json:"resources"`
			} `json:"root_module"`
		} `json:"planned_values"`
	}

	if err := json.NewDecoder(file).Decode(&raw); err != nil {
		return nil, fmt.Errorf("failed to decode JSON: %w", err)
	}

	var buckets []models.S3Bucket
	for _, r := range raw.PlannedValues.RootModule.Resources {
		if r.Type == "aws_s3_bucket" {
			buckets = append(buckets, models.S3Bucket{
				Name: r.Name,
				Acl:  r.Values.Acl,
				Tags: r.Values.Tags,
			})
		}
	}

	return buckets, nil
}
