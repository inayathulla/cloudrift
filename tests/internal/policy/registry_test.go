package policy

import (
	"testing"

	"github.com/inayathulla/cloudrift/internal/policy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadBuiltinRegistry_NotEmpty(t *testing.T) {
	reg := policy.LoadBuiltinRegistry()

	require.NotNil(t, reg)
	assert.Greater(t, reg.TotalPolicies, 0, "registry should contain policies")
	assert.Greater(t, len(reg.Policies), 0, "policy map should not be empty")
}

func TestLoadBuiltinRegistry_TotalMatchesPolicyMap(t *testing.T) {
	reg := policy.LoadBuiltinRegistry()

	assert.Equal(t, len(reg.Policies), reg.TotalPolicies,
		"TotalPolicies must equal len(Policies)")
}

func TestLoadBuiltinRegistry_CategoryTotalsConsistent(t *testing.T) {
	reg := policy.LoadBuiltinRegistry()

	// Sum of all category totals should equal total policies
	catSum := 0
	for _, count := range reg.CategoryTotals {
		catSum += count
	}
	assert.Equal(t, reg.TotalPolicies, catSum,
		"sum of category totals must equal total policies")
}

func TestLoadBuiltinRegistry_ExpectedCategories(t *testing.T) {
	reg := policy.LoadBuiltinRegistry()

	assert.Contains(t, reg.CategoryTotals, "security")
	assert.Contains(t, reg.CategoryTotals, "tagging")
	assert.Contains(t, reg.CategoryTotals, "cost")
	assert.Equal(t, 3, len(reg.CategoryTotals), "should have exactly 3 categories")
}

func TestLoadBuiltinRegistry_ExpectedFrameworks(t *testing.T) {
	reg := policy.LoadBuiltinRegistry()

	expectedFrameworks := []string{"hipaa", "pci_dss", "iso_27001", "gdpr", "soc2"}
	for _, fw := range expectedFrameworks {
		assert.Contains(t, reg.FrameworkTotals, fw, "missing framework: %s", fw)
		assert.Greater(t, reg.FrameworkTotals[fw], 0, "framework %s should have > 0 policies", fw)
	}
}

func TestLoadBuiltinRegistry_KnownPoliciesExist(t *testing.T) {
	reg := policy.LoadBuiltinRegistry()

	// Spot-check well-known policies
	knownPolicies := []struct {
		id       string
		category string
	}{
		{"S3-001", "security"},
		{"S3-002", "security"},
		{"EC2-001", "security"},
		{"EC2-002", "security"},
		{"EC2-005", "cost"},
		{"SG-001", "security"},
		{"TAG-001", "tagging"},
		{"TAG-002", "tagging"},
		{"RDS-001", "security"},
		{"RDS-002", "security"},
		{"IAM-001", "security"},
		{"CT-001", "security"},
		{"KMS-001", "security"},
		{"ELB-001", "security"},
		{"EBS-001", "security"},
		{"LAMBDA-001", "security"},
		{"LOG-001", "security"},
		{"VPC-001", "security"},
		{"SECRET-001", "security"},
		{"COST-002", "cost"},
		{"COST-003", "cost"},
	}

	for _, kp := range knownPolicies {
		p, exists := reg.Policies[kp.id]
		assert.True(t, exists, "policy %s should exist", kp.id)
		if exists {
			assert.Equal(t, kp.category, p.Category, "policy %s category mismatch", kp.id)
		}
	}
}

func TestLoadBuiltinRegistry_FrameworkMappings(t *testing.T) {
	reg := policy.LoadBuiltinRegistry()

	// S3-001 should map to all 5 frameworks
	s3001, exists := reg.Policies["S3-001"]
	require.True(t, exists)
	assert.Contains(t, s3001.Frameworks, "hipaa")
	assert.Contains(t, s3001.Frameworks, "pci_dss")
	assert.Contains(t, s3001.Frameworks, "iso_27001")
	assert.Contains(t, s3001.Frameworks, "gdpr")
	assert.Contains(t, s3001.Frameworks, "soc2")

	// Cost policies should have no frameworks
	cost002, exists := reg.Policies["COST-002"]
	require.True(t, exists)
	assert.Empty(t, cost002.Frameworks, "cost policies should have no frameworks")

	// TAG-001 maps only to soc2
	tag001, exists := reg.Policies["TAG-001"]
	require.True(t, exists)
	assert.Equal(t, []string{"soc2"}, tag001.Frameworks)
}

func TestLoadBuiltinRegistry_NoDuplicatePolicyIDs(t *testing.T) {
	reg := policy.LoadBuiltinRegistry()

	// Multi-rule policies (e.g., VPC-001 has 2 deny rules, S3-001 has 2 deny rules)
	// should still only appear once
	_, vpcExists := reg.Policies["VPC-001"]
	assert.True(t, vpcExists, "VPC-001 should exist")

	_, s3Exists := reg.Policies["S3-001"]
	assert.True(t, s3Exists, "S3-001 should exist")

	// Verify total count matches unique policies
	assert.Equal(t, len(reg.Policies), reg.TotalPolicies)
}

func TestLoadBuiltinRegistry_SecurityIsDominantCategory(t *testing.T) {
	reg := policy.LoadBuiltinRegistry()

	// Security should have the most policies
	assert.Greater(t, reg.CategoryTotals["security"], reg.CategoryTotals["tagging"])
	assert.Greater(t, reg.CategoryTotals["security"], reg.CategoryTotals["cost"])
}

func TestPolicyRegistry_FilterByFrameworks_Single(t *testing.T) {
	reg := policy.LoadBuiltinRegistry()
	filtered := reg.FilterByFrameworks([]string{"hipaa"})

	// Every policy in the filtered set must map to hipaa
	for id, p := range filtered.Policies {
		assert.Contains(t, p.Frameworks, "hipaa",
			"policy %s should map to hipaa", id)
	}

	// Filtered total should equal the hipaa total from the full registry
	assert.Equal(t, reg.FrameworkTotals["hipaa"], filtered.TotalPolicies)
	assert.Equal(t, 1, len(filtered.FrameworkTotals), "only hipaa should appear in framework totals")
	assert.Contains(t, filtered.FrameworkTotals, "hipaa")
}

func TestPolicyRegistry_FilterByFrameworks_Multiple(t *testing.T) {
	reg := policy.LoadBuiltinRegistry()
	filtered := reg.FilterByFrameworks([]string{"hipaa", "gdpr"})

	// Every policy must map to at least one of hipaa or gdpr
	for id, p := range filtered.Policies {
		hasHipaa := false
		hasGdpr := false
		for _, fw := range p.Frameworks {
			if fw == "hipaa" {
				hasHipaa = true
			}
			if fw == "gdpr" {
				hasGdpr = true
			}
		}
		assert.True(t, hasHipaa || hasGdpr,
			"policy %s should map to hipaa or gdpr", id)
	}

	// Union should be >= either individual count
	assert.GreaterOrEqual(t, filtered.TotalPolicies, reg.FrameworkTotals["hipaa"])
	assert.GreaterOrEqual(t, filtered.TotalPolicies, reg.FrameworkTotals["gdpr"])
	// Only hipaa and gdpr should appear in framework totals
	assert.Equal(t, 2, len(filtered.FrameworkTotals))
}

func TestPolicyRegistry_FilterByFrameworks_Empty(t *testing.T) {
	reg := policy.LoadBuiltinRegistry()
	filtered := reg.FilterByFrameworks([]string{})

	// Empty filter returns full registry copy
	assert.Equal(t, reg.TotalPolicies, filtered.TotalPolicies)
	assert.Equal(t, len(reg.Policies), len(filtered.Policies))
	assert.Equal(t, reg.CategoryTotals, filtered.CategoryTotals)
	assert.Equal(t, reg.FrameworkTotals, filtered.FrameworkTotals)
}

func TestPolicyRegistry_FilterByFrameworks_DoesNotMutateOriginal(t *testing.T) {
	reg := policy.LoadBuiltinRegistry()
	originalTotal := reg.TotalPolicies
	originalPolicyCount := len(reg.Policies)

	_ = reg.FilterByFrameworks([]string{"hipaa"})

	assert.Equal(t, originalTotal, reg.TotalPolicies, "original should not be mutated")
	assert.Equal(t, originalPolicyCount, len(reg.Policies), "original policy map should not be mutated")
}

func TestPolicyRegistry_FilterByFrameworks_CategoryTotalsConsistent(t *testing.T) {
	reg := policy.LoadBuiltinRegistry()
	filtered := reg.FilterByFrameworks([]string{"soc2"})

	// Sum of filtered category totals should equal filtered total
	catSum := 0
	for _, count := range filtered.CategoryTotals {
		catSum += count
	}
	assert.Equal(t, filtered.TotalPolicies, catSum,
		"sum of filtered category totals must equal filtered total policies")
}

func TestPolicyRegistry_KnownFrameworks(t *testing.T) {
	reg := policy.LoadBuiltinRegistry()
	known := reg.KnownFrameworks()

	assert.Equal(t, 5, len(known))
	// Should be sorted
	for i := 1; i < len(known); i++ {
		assert.True(t, known[i-1] < known[i], "frameworks should be sorted")
	}
}

func TestLoadBuiltinRegistry_Idempotent(t *testing.T) {
	reg1 := policy.LoadBuiltinRegistry()
	reg2 := policy.LoadBuiltinRegistry()

	assert.Equal(t, reg1.TotalPolicies, reg2.TotalPolicies)
	assert.Equal(t, len(reg1.Policies), len(reg2.Policies))
	assert.Equal(t, reg1.CategoryTotals, reg2.CategoryTotals)
	assert.Equal(t, reg1.FrameworkTotals, reg2.FrameworkTotals)
}
