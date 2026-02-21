package detector

import (
	"encoding/json"
	"fmt"
	"strings"

	sdkaws "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/inayathulla/cloudrift/internal/aws"
	"github.com/inayathulla/cloudrift/internal/models"
)

// IAMDriftResult captures the drift detection results for a single IAM resource.
type IAMDriftResult struct {
	// ResourceType identifies the IAM resource type: "role", "user", "policy", "group".
	ResourceType string

	// ResourceName is the name of the IAM resource.
	ResourceName string

	// TerraformAddress is the Terraform resource address.
	TerraformAddress string

	// Missing is true if the resource exists in the plan but not in AWS.
	Missing bool

	// AssumeRolePolicyDiff is true if the trust policy differs (roles only).
	AssumeRolePolicyDiff bool

	// MaxSessionDiff is true if the max session duration differs (roles only).
	MaxSessionDiff bool

	// DescriptionDiff is true if the description differs.
	DescriptionDiff bool

	// PolicyDocumentDiff is true if the policy document differs (policies only).
	PolicyDocumentDiff bool

	// PathDiff is true if the IAM path differs.
	PathDiff bool

	// AttachedPoliciesDiff is true if attached managed policies differ.
	AttachedPoliciesDiff bool

	// MembersDiff is true if group membership differs (groups only).
	MembersDiff bool

	// TagDiffs maps tag keys to [expected, actual] value pairs for mismatched tags.
	TagDiffs map[string][2]string

	// ExtraTags contains tags present in AWS but not in the plan.
	ExtraTags map[string]string
}

// HasAnyDrift returns true if any drift was detected for this resource.
func (r IAMDriftResult) HasAnyDrift() bool {
	return r.Missing ||
		r.AssumeRolePolicyDiff ||
		r.MaxSessionDiff ||
		r.DescriptionDiff ||
		r.PolicyDocumentDiff ||
		r.PathDiff ||
		r.AttachedPoliciesDiff ||
		r.MembersDiff ||
		len(r.TagDiffs) > 0 ||
		len(r.ExtraTags) > 0
}

// IAMDriftDetector implements drift detection for AWS IAM resources.
type IAMDriftDetector struct {
	cfg sdkaws.Config
}

// NewIAMDriftDetector creates a new IAM drift detector with the given AWS configuration.
func NewIAMDriftDetector(cfg sdkaws.Config) *IAMDriftDetector {
	return &IAMDriftDetector{cfg: cfg}
}

// FetchLiveState retrieves the current state of all IAM resources from AWS.
func (d *IAMDriftDetector) FetchLiveState() (interface{}, error) {
	return aws.FetchIAMResources(d.cfg)
}

// DetectDrift compares Terraform-planned IAM configurations against live AWS state.
func (d *IAMDriftDetector) DetectDrift(plan, live interface{}) ([]DriftResult, error) {
	plans, ok := plan.(*models.IAMPlanResources)
	if !ok {
		return nil, fmt.Errorf("plan type mismatch: expected *models.IAMPlanResources")
	}
	lives, ok := live.(*models.IAMLiveState)
	if !ok {
		return nil, fmt.Errorf("live type mismatch: expected *models.IAMLiveState")
	}

	iamResults := DetectAllIAMDrift(plans, lives)

	// Convert IAMDriftResult to generic DriftResult for compatibility
	results := make([]DriftResult, 0, len(iamResults))
	for _, r := range iamResults {
		dr := DriftResult{
			BucketName: r.ResourceName, // Reuse BucketName field for resource name
		}
		if r.Missing {
			dr.Missing = true
		}
		dr.TagDiffs = r.TagDiffs
		dr.ExtraTags = r.ExtraTags

		// Use AclDiff to indicate "other diffs exist" (same pattern as EC2)
		if r.AssumeRolePolicyDiff || r.MaxSessionDiff || r.DescriptionDiff ||
			r.PolicyDocumentDiff || r.PathDiff || r.AttachedPoliciesDiff || r.MembersDiff {
			dr.AclDiff = true
		}

		results = append(results, dr)
	}

	return results, nil
}

// DetectAllIAMDrift performs drift detection across all planned IAM resources.
func DetectAllIAMDrift(plans *models.IAMPlanResources, lives *models.IAMLiveState) []IAMDriftResult {
	var results []IAMDriftResult

	// Detect role drift
	roleMap := make(map[string]*models.IAMRole, len(lives.Roles))
	for i := range lives.Roles {
		roleMap[lives.Roles[i].RoleName] = &lives.Roles[i]
	}
	for _, p := range plans.Roles {
		dr := DetectIAMRoleDrift(p, roleMap[p.RoleName])
		if dr.HasAnyDrift() {
			results = append(results, dr)
		}
	}

	// Detect user drift
	userMap := make(map[string]*models.IAMUser, len(lives.Users))
	for i := range lives.Users {
		userMap[lives.Users[i].UserName] = &lives.Users[i]
	}
	for _, p := range plans.Users {
		dr := DetectIAMUserDrift(p, userMap[p.UserName])
		if dr.HasAnyDrift() {
			results = append(results, dr)
		}
	}

	// Detect policy drift
	policyMap := make(map[string]*models.IAMPolicy, len(lives.Policies))
	for i := range lives.Policies {
		policyMap[lives.Policies[i].PolicyName] = &lives.Policies[i]
	}
	for _, p := range plans.Policies {
		dr := DetectIAMPolicyDrift(p, policyMap[p.PolicyName])
		if dr.HasAnyDrift() {
			results = append(results, dr)
		}
	}

	// Detect group drift
	groupMap := make(map[string]*models.IAMGroup, len(lives.Groups))
	for i := range lives.Groups {
		groupMap[lives.Groups[i].GroupName] = &lives.Groups[i]
	}
	for _, p := range plans.Groups {
		dr := DetectIAMGroupDrift(p, groupMap[p.GroupName])
		if dr.HasAnyDrift() {
			results = append(results, dr)
		}
	}

	return results
}

// DetectIAMRoleDrift compares a single planned role against its actual AWS state.
func DetectIAMRoleDrift(plan models.IAMRole, actual *models.IAMRole) IAMDriftResult {
	res := IAMDriftResult{
		ResourceType:     "role",
		ResourceName:     plan.RoleName,
		TerraformAddress: plan.TerraformAddress,
		TagDiffs:         make(map[string][2]string),
		ExtraTags:        make(map[string]string),
	}

	if actual == nil {
		res.Missing = true
		return res
	}

	// Trust policy comparison (normalize JSON for comparison)
	if plan.AssumeRolePolicy != "" && !jsonEqual(plan.AssumeRolePolicy, actual.AssumeRolePolicy) {
		res.AssumeRolePolicyDiff = true
	}

	// Max session duration
	if plan.MaxSessionDuration > 0 && plan.MaxSessionDuration != actual.MaxSessionDuration {
		res.MaxSessionDiff = true
	}

	// Description
	if plan.Description != "" && plan.Description != actual.Description {
		res.DescriptionDiff = true
	}

	// Path
	if plan.Path != "" && plan.Path != actual.Path {
		res.PathDiff = true
	}

	// Attached policies
	if len(plan.AttachedPolicies) > 0 && !stringSlicesEqual(plan.AttachedPolicies, actual.AttachedPolicies) {
		res.AttachedPoliciesDiff = true
	}

	// Tags
	compareTags(plan.Tags, actual.Tags, res.TagDiffs, res.ExtraTags)

	return res
}

// DetectIAMUserDrift compares a single planned user against its actual AWS state.
func DetectIAMUserDrift(plan models.IAMUser, actual *models.IAMUser) IAMDriftResult {
	res := IAMDriftResult{
		ResourceType:     "user",
		ResourceName:     plan.UserName,
		TerraformAddress: plan.TerraformAddress,
		TagDiffs:         make(map[string][2]string),
		ExtraTags:        make(map[string]string),
	}

	if actual == nil {
		res.Missing = true
		return res
	}

	// Path
	if plan.Path != "" && plan.Path != actual.Path {
		res.PathDiff = true
	}

	// Attached policies
	if len(plan.AttachedPolicies) > 0 && !stringSlicesEqual(plan.AttachedPolicies, actual.AttachedPolicies) {
		res.AttachedPoliciesDiff = true
	}

	// Tags
	compareTags(plan.Tags, actual.Tags, res.TagDiffs, res.ExtraTags)

	return res
}

// DetectIAMPolicyDrift compares a single planned policy against its actual AWS state.
func DetectIAMPolicyDrift(plan models.IAMPolicy, actual *models.IAMPolicy) IAMDriftResult {
	res := IAMDriftResult{
		ResourceType:     "policy",
		ResourceName:     plan.PolicyName,
		TerraformAddress: plan.TerraformAddress,
		TagDiffs:         make(map[string][2]string),
		ExtraTags:        make(map[string]string),
	}

	if actual == nil {
		res.Missing = true
		return res
	}

	// Policy document comparison (normalize JSON)
	if plan.PolicyDocument != "" && !jsonEqual(plan.PolicyDocument, actual.PolicyDocument) {
		res.PolicyDocumentDiff = true
	}

	// Description
	if plan.Description != "" && plan.Description != actual.Description {
		res.DescriptionDiff = true
	}

	// Path
	if plan.Path != "" && plan.Path != actual.Path {
		res.PathDiff = true
	}

	// Tags
	compareTags(plan.Tags, actual.Tags, res.TagDiffs, res.ExtraTags)

	return res
}

// DetectIAMGroupDrift compares a single planned group against its actual AWS state.
func DetectIAMGroupDrift(plan models.IAMGroup, actual *models.IAMGroup) IAMDriftResult {
	res := IAMDriftResult{
		ResourceType:     "group",
		ResourceName:     plan.GroupName,
		TerraformAddress: plan.TerraformAddress,
		TagDiffs:         make(map[string][2]string),
		ExtraTags:        make(map[string]string),
	}

	if actual == nil {
		res.Missing = true
		return res
	}

	// Path
	if plan.Path != "" && plan.Path != actual.Path {
		res.PathDiff = true
	}

	// Attached policies
	if len(plan.AttachedPolicies) > 0 && !stringSlicesEqual(plan.AttachedPolicies, actual.AttachedPolicies) {
		res.AttachedPoliciesDiff = true
	}

	// Members
	if len(plan.Members) > 0 && !stringSlicesEqual(plan.Members, actual.Members) {
		res.MembersDiff = true
	}

	return res
}

// compareTags detects tag differences and extra tags between planned and actual state.
func compareTags(plan, actual map[string]string, diffs map[string][2]string, extras map[string]string) {
	for k, v := range plan {
		if av, ok := actual[k]; !ok || av != v {
			diffs[k] = [2]string{v, av}
		}
	}
	for k, av := range actual {
		if _, ok := plan[k]; !ok {
			extras[k] = av
		}
	}
}

// jsonEqual compares two JSON strings after normalization.
// Returns true if both represent the same JSON structure.
func jsonEqual(a, b string) bool {
	a = strings.TrimSpace(a)
	b = strings.TrimSpace(b)

	if a == b {
		return true
	}

	var objA, objB interface{}
	if err := json.Unmarshal([]byte(a), &objA); err != nil {
		return a == b
	}
	if err := json.Unmarshal([]byte(b), &objB); err != nil {
		return a == b
	}

	normA, err := json.Marshal(objA)
	if err != nil {
		return false
	}
	normB, err := json.Marshal(objB)
	if err != nil {
		return false
	}

	return string(normA) == string(normB)
}
