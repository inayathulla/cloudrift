package output

import (
	"encoding/json"
	"fmt"
	"io"
)

// SARIFFormatter outputs scan results in SARIF 2.1.0 format.
//
// SARIF (Static Analysis Results Interchange Format) is an OASIS standard
// supported by GitHub Code Scanning, GitLab SAST, and other security tools.
// This enables drift results to appear in GitHub's Security tab.
//
// Specification: https://docs.oasis-open.org/sarif/sarif/v2.1.0/sarif-v2.1.0.html
type SARIFFormatter struct{}

// NewSARIFFormatter creates a new SARIF formatter.
func NewSARIFFormatter() *SARIFFormatter {
	return &SARIFFormatter{}
}

// SARIF root document structure
type sarifDocument struct {
	Schema  string     `json:"$schema"`
	Version string     `json:"version"`
	Runs    []sarifRun `json:"runs"`
}

type sarifRun struct {
	Tool    sarifTool     `json:"tool"`
	Results []sarifResult `json:"results"`
}

type sarifTool struct {
	Driver sarifDriver `json:"driver"`
}

type sarifDriver struct {
	Name            string      `json:"name"`
	Version         string      `json:"version"`
	InformationURI  string      `json:"informationUri"`
	Rules           []sarifRule `json:"rules"`
	SemanticVersion string      `json:"semanticVersion"`
}

type sarifRule struct {
	ID               string              `json:"id"`
	Name             string              `json:"name"`
	ShortDescription sarifMessage        `json:"shortDescription"`
	FullDescription  sarifMessage        `json:"fullDescription,omitempty"`
	Help             sarifMessage        `json:"help,omitempty"`
	DefaultConfig    sarifDefaultConfig  `json:"defaultConfiguration"`
	Properties       map[string][]string `json:"properties,omitempty"`
}

type sarifDefaultConfig struct {
	Level string `json:"level"`
}

type sarifMessage struct {
	Text string `json:"text"`
}

type sarifResult struct {
	RuleID    string               `json:"ruleId"`
	RuleIndex int                  `json:"ruleIndex"`
	Level     string               `json:"level"`
	Message   sarifMessage         `json:"message"`
	Locations []sarifLocation      `json:"locations,omitempty"`
	Fixes     []sarifFix           `json:"fixes,omitempty"`
	Properties map[string]interface{} `json:"properties,omitempty"`
}

type sarifLocation struct {
	PhysicalLocation sarifPhysicalLocation `json:"physicalLocation"`
	LogicalLocations []sarifLogicalLocation `json:"logicalLocations,omitempty"`
}

type sarifPhysicalLocation struct {
	ArtifactLocation sarifArtifactLocation `json:"artifactLocation"`
	Region           *sarifRegion          `json:"region,omitempty"`
}

type sarifArtifactLocation struct {
	URI string `json:"uri"`
}

type sarifRegion struct {
	StartLine   int `json:"startLine,omitempty"`
	StartColumn int `json:"startColumn,omitempty"`
}

type sarifLogicalLocation struct {
	Name               string `json:"name"`
	FullyQualifiedName string `json:"fullyQualifiedName"`
	Kind               string `json:"kind"`
}

type sarifFix struct {
	Description sarifMessage `json:"description"`
}

// Format writes the scan result as SARIF to the provided writer.
func (f *SARIFFormatter) Format(w io.Writer, result ScanResult) error {
	doc := f.buildDocument(result)

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(doc)
}

func (f *SARIFFormatter) buildDocument(result ScanResult) sarifDocument {
	rules := f.buildRules()
	results := f.buildResults(result)

	return sarifDocument{
		Schema:  "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json",
		Version: "2.1.0",
		Runs: []sarifRun{
			{
				Tool: sarifTool{
					Driver: sarifDriver{
						Name:            "Cloudrift",
						Version:         "1.0.0",
						SemanticVersion: "1.0.0",
						InformationURI:  "https://github.com/inayathulla/cloudrift",
						Rules:           rules,
					},
				},
				Results: results,
			},
		},
	}
}

func (f *SARIFFormatter) buildRules() []sarifRule {
	return []sarifRule{
		{
			ID:   "DRIFT001",
			Name: "resource-missing",
			ShortDescription: sarifMessage{
				Text: "Resource exists in Terraform plan but not in AWS",
			},
			FullDescription: sarifMessage{
				Text: "A resource defined in your Terraform plan does not exist in AWS. This may indicate the resource was deleted outside of Terraform.",
			},
			Help: sarifMessage{
				Text: "Run 'terraform apply' to create the resource, or remove it from your Terraform configuration if it's no longer needed.",
			},
			DefaultConfig: sarifDefaultConfig{Level: "error"},
			Properties: map[string][]string{
				"tags": {"drift", "infrastructure", "terraform"},
			},
		},
		{
			ID:   "DRIFT002",
			Name: "attribute-mismatch",
			ShortDescription: sarifMessage{
				Text: "Resource attribute differs between Terraform plan and AWS",
			},
			FullDescription: sarifMessage{
				Text: "One or more attributes of a resource differ between your Terraform plan and the actual AWS state. This indicates configuration drift.",
			},
			Help: sarifMessage{
				Text: "Review the drift details and either update your Terraform configuration to match AWS, or run 'terraform apply' to enforce your planned state.",
			},
			DefaultConfig: sarifDefaultConfig{Level: "warning"},
			Properties: map[string][]string{
				"tags": {"drift", "infrastructure", "terraform"},
			},
		},
		{
			ID:   "DRIFT003",
			Name: "extra-attribute",
			ShortDescription: sarifMessage{
				Text: "Resource has attributes in AWS not defined in Terraform",
			},
			FullDescription: sarifMessage{
				Text: "The resource in AWS has attributes that are not defined in your Terraform configuration. This may indicate manual changes.",
			},
			Help: sarifMessage{
				Text: "Consider importing the extra attributes into your Terraform configuration to prevent them from being overwritten.",
			},
			DefaultConfig: sarifDefaultConfig{Level: "note"},
			Properties: map[string][]string{
				"tags": {"drift", "infrastructure", "terraform"},
			},
		},
	}
}

func (f *SARIFFormatter) buildResults(scanResult ScanResult) []sarifResult {
	var results []sarifResult

	for _, drift := range scanResult.Drifts {
		if drift.Missing {
			results = append(results, sarifResult{
				RuleID:    "DRIFT001",
				RuleIndex: 0,
				Level:     "error",
				Message: sarifMessage{
					Text: fmt.Sprintf("Resource %s (%s) exists in Terraform plan but not in AWS", drift.ResourceName, drift.ResourceType),
				},
				Locations: []sarifLocation{
					{
						PhysicalLocation: sarifPhysicalLocation{
							ArtifactLocation: sarifArtifactLocation{
								URI: "terraform.tfstate",
							},
						},
						LogicalLocations: []sarifLogicalLocation{
							{
								Name:               drift.ResourceName,
								FullyQualifiedName: fmt.Sprintf("%s.%s", drift.ResourceType, drift.ResourceName),
								Kind:               "resource",
							},
						},
					},
				},
				Properties: map[string]interface{}{
					"resourceType": drift.ResourceType,
					"resourceName": drift.ResourceName,
					"service":      scanResult.Service,
				},
			})
		}

		for attr, values := range drift.Diffs {
			results = append(results, sarifResult{
				RuleID:    "DRIFT002",
				RuleIndex: 1,
				Level:     f.severityToLevel(drift.Severity),
				Message: sarifMessage{
					Text: fmt.Sprintf("Attribute '%s' of %s (%s) differs: expected %v, got %v",
						attr, drift.ResourceName, drift.ResourceType, values[0], values[1]),
				},
				Locations: []sarifLocation{
					{
						PhysicalLocation: sarifPhysicalLocation{
							ArtifactLocation: sarifArtifactLocation{
								URI: "terraform.tfstate",
							},
						},
						LogicalLocations: []sarifLogicalLocation{
							{
								Name:               drift.ResourceName,
								FullyQualifiedName: fmt.Sprintf("%s.%s.%s", drift.ResourceType, drift.ResourceName, attr),
								Kind:               "attribute",
							},
						},
					},
				},
				Properties: map[string]interface{}{
					"attribute":    attr,
					"expected":     values[0],
					"actual":       values[1],
					"resourceType": drift.ResourceType,
					"resourceName": drift.ResourceName,
					"service":      scanResult.Service,
				},
			})
		}

		for attr, value := range drift.ExtraAttributes {
			results = append(results, sarifResult{
				RuleID:    "DRIFT003",
				RuleIndex: 2,
				Level:     "note",
				Message: sarifMessage{
					Text: fmt.Sprintf("Resource %s (%s) has extra attribute '%s' in AWS: %v",
						drift.ResourceName, drift.ResourceType, attr, value),
				},
				Locations: []sarifLocation{
					{
						PhysicalLocation: sarifPhysicalLocation{
							ArtifactLocation: sarifArtifactLocation{
								URI: "terraform.tfstate",
							},
						},
						LogicalLocations: []sarifLogicalLocation{
							{
								Name:               drift.ResourceName,
								FullyQualifiedName: fmt.Sprintf("%s.%s.%s", drift.ResourceType, drift.ResourceName, attr),
								Kind:               "attribute",
							},
						},
					},
				},
				Properties: map[string]interface{}{
					"attribute":    attr,
					"value":        value,
					"resourceType": drift.ResourceType,
					"resourceName": drift.ResourceName,
					"service":      scanResult.Service,
				},
			})
		}
	}

	return results
}

func (f *SARIFFormatter) severityToLevel(severity string) string {
	switch severity {
	case "critical":
		return "error"
	case "warning":
		return "warning"
	case "info":
		return "note"
	default:
		return "warning"
	}
}

// Name returns the format name.
func (f *SARIFFormatter) Name() string {
	return "sarif"
}

// FileExtension returns the recommended file extension.
func (f *SARIFFormatter) FileExtension() string {
	return ".sarif"
}

func init() {
	Register(FormatSARIF, NewSARIFFormatter())
}
