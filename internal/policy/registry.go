package policy

import (
	"io/fs"
	"regexp"
	"strings"
)

// PolicyInfo holds metadata for a single built-in policy rule.
type PolicyInfo struct {
	ID         string   `json:"id"`
	Category   string   `json:"category"`
	Frameworks []string `json:"frameworks"`
}

// PolicyRegistry provides aggregated metadata about all built-in policies.
// Totals are computed dynamically from embedded .rego files — never hardcoded.
type PolicyRegistry struct {
	Policies       map[string]PolicyInfo `json:"policies"`
	TotalPolicies  int                   `json:"total_policies"`
	CategoryTotals map[string]int        `json:"category_totals"`
	FrameworkTotals map[string]int       `json:"framework_totals"`
}

// Regex patterns for extracting policy metadata from .rego files.
var (
	policyIDRe     = regexp.MustCompile(`"policy_id"\s*:\s*"([^"]+)"`)
	categoryRe     = regexp.MustCompile(`"category"\s*:\s*"([^"]+)"`)
	frameworksRe   = regexp.MustCompile(`"frameworks"\s*:\s*\[([^\]]*)\]`)
	quotedStringRe = regexp.MustCompile(`"([^"]+)"`)
)

// LoadBuiltinRegistry scans all embedded .rego files and extracts policy
// metadata (policy_id, category, frameworks). Duplicate policy IDs across
// multiple rules (e.g., VPC-001 with two deny rules) are deduplicated.
func LoadBuiltinRegistry() *PolicyRegistry {
	reg := &PolicyRegistry{
		Policies:        make(map[string]PolicyInfo),
		CategoryTotals:  make(map[string]int),
		FrameworkTotals: make(map[string]int),
	}

	fs.WalkDir(BuiltinPolicies, "policies", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".rego") {
			return nil
		}

		content, readErr := BuiltinPolicies.ReadFile(path)
		if readErr != nil {
			return nil
		}

		dirCategory := inferCategoryFromPath(path)
		extractPolicyMetadata(reg, string(content), dirCategory)
		return nil
	})

	// Compute aggregated totals from the deduplicated policy map.
	reg.TotalPolicies = len(reg.Policies)
	for _, p := range reg.Policies {
		if p.Category != "" {
			reg.CategoryTotals[p.Category]++
		}
		for _, fw := range p.Frameworks {
			reg.FrameworkTotals[fw]++
		}
	}

	return reg
}

// inferCategoryFromPath derives the policy category from its directory path.
func inferCategoryFromPath(path string) string {
	switch {
	case strings.Contains(path, "security/"):
		return "security"
	case strings.Contains(path, "cost/"):
		return "cost"
	case strings.Contains(path, "tagging/"):
		return "tagging"
	default:
		return ""
	}
}

// extractPolicyMetadata finds all policy_id occurrences in a .rego file and
// extracts category and frameworks from the surrounding region. This approach
// avoids block-boundary issues when result objects contain braces in string
// literals (e.g., remediation messages with embedded JSON examples).
// Policies with the same ID (multi-rule policies) are only registered once.
func extractPolicyMetadata(reg *PolicyRegistry, content, dirCategory string) {
	idMatches := policyIDRe.FindAllStringSubmatchIndex(content, -1)

	for i, idxs := range idMatches {
		policyID := content[idxs[2]:idxs[3]]

		// Skip if already registered (dedup for multi-rule policies like VPC-001)
		if _, exists := reg.Policies[policyID]; exists {
			continue
		}

		// Search region: from this policy_id to the next (or EOF).
		// Category and frameworks always appear after policy_id in the same block.
		regionStart := idxs[0]
		regionEnd := len(content)
		if i+1 < len(idMatches) {
			regionEnd = idMatches[i+1][0]
		}
		region := content[regionStart:regionEnd]

		// Extract category; fall back to directory-based inference
		category := dirCategory
		if catMatch := categoryRe.FindStringSubmatch(region); len(catMatch) > 1 {
			category = catMatch[1]
		}

		// Extract frameworks
		var frameworks []string
		if fwMatch := frameworksRe.FindStringSubmatch(region); len(fwMatch) > 1 {
			for _, qMatch := range quotedStringRe.FindAllStringSubmatch(fwMatch[1], -1) {
				frameworks = append(frameworks, qMatch[1])
			}
		}

		reg.Policies[policyID] = PolicyInfo{
			ID:         policyID,
			Category:   category,
			Frameworks: frameworks,
		}
	}
}
