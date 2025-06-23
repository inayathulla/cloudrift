package parser

import (
	"encoding/json"
	"fmt"
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

// LoadTerraformPlan loads the Terraform plan from a JSON file and unmarshals it into a TerraformPlan struct.
func LoadTerraformPlan(path string) (*TerraformPlan, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open plan file: %w", err)
	}
	defer file.Close()

	var plan TerraformPlan
	if err := json.NewDecoder(file).Decode(&plan); err != nil {
		return nil, fmt.Errorf("failed to decode plan JSON: %w", err)
	}

	return &plan, nil
}
