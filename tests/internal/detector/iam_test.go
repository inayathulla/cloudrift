package detector

import (
	"testing"

	"github.com/inayathulla/cloudrift/internal/detector"
	"github.com/inayathulla/cloudrift/internal/models"
	"github.com/stretchr/testify/assert"
)

// ──────────────────────────────────────────────────────────────────────────────
// IAM Role Drift Detection
// ──────────────────────────────────────────────────────────────────────────────

func TestDetectIAMRoleDrift_Missing(t *testing.T) {
	plan := models.IAMRole{RoleName: "my-role", Tags: map[string]string{}}
	res := detector.DetectIAMRoleDrift(plan, nil)
	assert.True(t, res.Missing)
	assert.True(t, res.HasAnyDrift())
}

func TestDetectIAMRoleDrift_NoDrift(t *testing.T) {
	plan := models.IAMRole{
		RoleName:           "my-role",
		AssumeRolePolicy:   `{"Version":"2012-10-17","Statement":[]}`,
		MaxSessionDuration: 3600,
		Path:               "/",
		Tags:               map[string]string{"env": "prod"},
	}
	actual := &models.IAMRole{
		RoleName:           "my-role",
		AssumeRolePolicy:   `{"Version":"2012-10-17","Statement":[]}`,
		MaxSessionDuration: 3600,
		Path:               "/",
		Tags:               map[string]string{"env": "prod"},
	}
	res := detector.DetectIAMRoleDrift(plan, actual)
	assert.False(t, res.HasAnyDrift())
}

func TestDetectIAMRoleDrift_TrustPolicyDiff(t *testing.T) {
	plan := models.IAMRole{
		RoleName:         "my-role",
		AssumeRolePolicy: `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":{"Service":"lambda.amazonaws.com"},"Action":"sts:AssumeRole"}]}`,
		Tags:             map[string]string{},
	}
	actual := &models.IAMRole{
		RoleName:         "my-role",
		AssumeRolePolicy: `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":{"Service":"ec2.amazonaws.com"},"Action":"sts:AssumeRole"}]}`,
		Tags:             map[string]string{},
	}
	res := detector.DetectIAMRoleDrift(plan, actual)
	assert.True(t, res.AssumeRolePolicyDiff)
	assert.True(t, res.HasAnyDrift())
}

func TestDetectIAMRoleDrift_TrustPolicyNormalized(t *testing.T) {
	// Same JSON with different formatting should not be considered drift
	plan := models.IAMRole{
		RoleName:         "my-role",
		AssumeRolePolicy: `{"Version":"2012-10-17","Statement":[]}`,
		Tags:             map[string]string{},
	}
	actual := &models.IAMRole{
		RoleName:         "my-role",
		AssumeRolePolicy: `{ "Version" : "2012-10-17", "Statement" : [] }`,
		Tags:             map[string]string{},
	}
	res := detector.DetectIAMRoleDrift(plan, actual)
	assert.False(t, res.AssumeRolePolicyDiff)
}

func TestDetectIAMRoleDrift_MaxSessionDiff(t *testing.T) {
	plan := models.IAMRole{
		RoleName:           "my-role",
		MaxSessionDuration: 3600,
		Tags:               map[string]string{},
	}
	actual := &models.IAMRole{
		RoleName:           "my-role",
		MaxSessionDuration: 7200,
		Tags:               map[string]string{},
	}
	res := detector.DetectIAMRoleDrift(plan, actual)
	assert.True(t, res.MaxSessionDiff)
}

func TestDetectIAMRoleDrift_TagDiff(t *testing.T) {
	plan := models.IAMRole{
		RoleName: "my-role",
		Tags:     map[string]string{"env": "prod"},
	}
	actual := &models.IAMRole{
		RoleName: "my-role",
		Tags:     map[string]string{"env": "staging"},
	}
	res := detector.DetectIAMRoleDrift(plan, actual)
	assert.Len(t, res.TagDiffs, 1)
	assert.Equal(t, [2]string{"prod", "staging"}, res.TagDiffs["env"])
}

func TestDetectIAMRoleDrift_ExtraTags(t *testing.T) {
	plan := models.IAMRole{
		RoleName: "my-role",
		Tags:     map[string]string{},
	}
	actual := &models.IAMRole{
		RoleName: "my-role",
		Tags:     map[string]string{"extra": "val"},
	}
	res := detector.DetectIAMRoleDrift(plan, actual)
	assert.Len(t, res.ExtraTags, 1)
	assert.Equal(t, "val", res.ExtraTags["extra"])
}

func TestDetectIAMRoleDrift_DescriptionDiff(t *testing.T) {
	plan := models.IAMRole{
		RoleName:    "my-role",
		Description: "original desc",
		Tags:        map[string]string{},
	}
	actual := &models.IAMRole{
		RoleName:    "my-role",
		Description: "changed desc",
		Tags:        map[string]string{},
	}
	res := detector.DetectIAMRoleDrift(plan, actual)
	assert.True(t, res.DescriptionDiff)
}

// ──────────────────────────────────────────────────────────────────────────────
// IAM User Drift Detection
// ──────────────────────────────────────────────────────────────────────────────

func TestDetectIAMUserDrift_Missing(t *testing.T) {
	plan := models.IAMUser{UserName: "my-user", Tags: map[string]string{}}
	res := detector.DetectIAMUserDrift(plan, nil)
	assert.True(t, res.Missing)
	assert.True(t, res.HasAnyDrift())
}

func TestDetectIAMUserDrift_NoDrift(t *testing.T) {
	plan := models.IAMUser{
		UserName: "my-user",
		Path:     "/",
		Tags:     map[string]string{"env": "prod"},
	}
	actual := &models.IAMUser{
		UserName: "my-user",
		Path:     "/",
		Tags:     map[string]string{"env": "prod"},
	}
	res := detector.DetectIAMUserDrift(plan, actual)
	assert.False(t, res.HasAnyDrift())
}

func TestDetectIAMUserDrift_PathDiff(t *testing.T) {
	plan := models.IAMUser{
		UserName: "my-user",
		Path:     "/ci/",
		Tags:     map[string]string{},
	}
	actual := &models.IAMUser{
		UserName: "my-user",
		Path:     "/",
		Tags:     map[string]string{},
	}
	res := detector.DetectIAMUserDrift(plan, actual)
	assert.True(t, res.PathDiff)
}

func TestDetectIAMUserDrift_TagDiff(t *testing.T) {
	plan := models.IAMUser{
		UserName: "my-user",
		Tags:     map[string]string{"team": "platform"},
	}
	actual := &models.IAMUser{
		UserName: "my-user",
		Tags:     map[string]string{"team": "backend"},
	}
	res := detector.DetectIAMUserDrift(plan, actual)
	assert.Len(t, res.TagDiffs, 1)
}

// ──────────────────────────────────────────────────────────────────────────────
// IAM Policy Drift Detection
// ──────────────────────────────────────────────────────────────────────────────

func TestDetectIAMPolicyDrift_Missing(t *testing.T) {
	plan := models.IAMPolicy{PolicyName: "my-policy", Tags: map[string]string{}}
	res := detector.DetectIAMPolicyDrift(plan, nil)
	assert.True(t, res.Missing)
	assert.True(t, res.HasAnyDrift())
}

func TestDetectIAMPolicyDrift_NoDrift(t *testing.T) {
	doc := `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Action":"s3:GetObject","Resource":"*"}]}`
	plan := models.IAMPolicy{
		PolicyName:     "my-policy",
		PolicyDocument: doc,
		Tags:           map[string]string{},
	}
	actual := &models.IAMPolicy{
		PolicyName:     "my-policy",
		PolicyDocument: doc,
		Tags:           map[string]string{},
	}
	res := detector.DetectIAMPolicyDrift(plan, actual)
	assert.False(t, res.HasAnyDrift())
}

func TestDetectIAMPolicyDrift_DocumentDiff(t *testing.T) {
	plan := models.IAMPolicy{
		PolicyName:     "my-policy",
		PolicyDocument: `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Action":"s3:GetObject","Resource":"*"}]}`,
		Tags:           map[string]string{},
	}
	actual := &models.IAMPolicy{
		PolicyName:     "my-policy",
		PolicyDocument: `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Action":"*","Resource":"*"}]}`,
		Tags:           map[string]string{},
	}
	res := detector.DetectIAMPolicyDrift(plan, actual)
	assert.True(t, res.PolicyDocumentDiff)
	assert.True(t, res.HasAnyDrift())
}

func TestDetectIAMPolicyDrift_DescriptionDiff(t *testing.T) {
	plan := models.IAMPolicy{
		PolicyName:  "my-policy",
		Description: "original",
		Tags:        map[string]string{},
	}
	actual := &models.IAMPolicy{
		PolicyName:  "my-policy",
		Description: "changed",
		Tags:        map[string]string{},
	}
	res := detector.DetectIAMPolicyDrift(plan, actual)
	assert.True(t, res.DescriptionDiff)
}

// ──────────────────────────────────────────────────────────────────────────────
// IAM Group Drift Detection
// ──────────────────────────────────────────────────────────────────────────────

func TestDetectIAMGroupDrift_Missing(t *testing.T) {
	plan := models.IAMGroup{GroupName: "my-group"}
	res := detector.DetectIAMGroupDrift(plan, nil)
	assert.True(t, res.Missing)
	assert.True(t, res.HasAnyDrift())
}

func TestDetectIAMGroupDrift_NoDrift(t *testing.T) {
	plan := models.IAMGroup{
		GroupName: "my-group",
		Path:      "/",
	}
	actual := &models.IAMGroup{
		GroupName: "my-group",
		Path:      "/",
	}
	res := detector.DetectIAMGroupDrift(plan, actual)
	assert.False(t, res.HasAnyDrift())
}

func TestDetectIAMGroupDrift_PathDiff(t *testing.T) {
	plan := models.IAMGroup{
		GroupName: "my-group",
		Path:      "/admin/",
	}
	actual := &models.IAMGroup{
		GroupName: "my-group",
		Path:      "/",
	}
	res := detector.DetectIAMGroupDrift(plan, actual)
	assert.True(t, res.PathDiff)
}

func TestDetectIAMGroupDrift_MembersDiff(t *testing.T) {
	plan := models.IAMGroup{
		GroupName: "my-group",
		Members:   []string{"alice", "bob"},
	}
	actual := &models.IAMGroup{
		GroupName: "my-group",
		Members:   []string{"alice"},
	}
	res := detector.DetectIAMGroupDrift(plan, actual)
	assert.True(t, res.MembersDiff)
}

func TestDetectIAMGroupDrift_AttachedPoliciesDiff(t *testing.T) {
	plan := models.IAMGroup{
		GroupName:        "my-group",
		AttachedPolicies: []string{"arn:aws:iam::123:policy/PolicyA"},
	}
	actual := &models.IAMGroup{
		GroupName:        "my-group",
		AttachedPolicies: []string{"arn:aws:iam::123:policy/PolicyB"},
	}
	res := detector.DetectIAMGroupDrift(plan, actual)
	assert.True(t, res.AttachedPoliciesDiff)
}

// ──────────────────────────────────────────────────────────────────────────────
// DetectAllIAMDrift
// ──────────────────────────────────────────────────────────────────────────────

func TestDetectAllIAMDrift_MixedResults(t *testing.T) {
	plans := &models.IAMPlanResources{
		Roles: []models.IAMRole{
			{RoleName: "role-ok", Path: "/", Tags: map[string]string{}},
			{RoleName: "role-missing", Tags: map[string]string{}},
		},
		Users: []models.IAMUser{
			{UserName: "user-drifted", Path: "/ci/", Tags: map[string]string{}},
		},
		Policies: []models.IAMPolicy{},
		Groups:   []models.IAMGroup{},
	}

	lives := &models.IAMLiveState{
		Roles: []models.IAMRole{
			{RoleName: "role-ok", Path: "/", Tags: map[string]string{}},
			// role-missing is absent
		},
		Users: []models.IAMUser{
			{UserName: "user-drifted", Path: "/", Tags: map[string]string{}},
		},
		Policies: []models.IAMPolicy{},
		Groups:   []models.IAMGroup{},
	}

	results := detector.DetectAllIAMDrift(plans, lives)
	// role-ok has no drift, role-missing is missing, user-drifted has path diff
	assert.Len(t, results, 2)
}

func TestDetectAllIAMDrift_NoDrift(t *testing.T) {
	plans := &models.IAMPlanResources{
		Roles: []models.IAMRole{
			{RoleName: "role-a", Path: "/", Tags: map[string]string{}},
		},
		Users:    []models.IAMUser{},
		Policies: []models.IAMPolicy{},
		Groups:   []models.IAMGroup{},
	}

	lives := &models.IAMLiveState{
		Roles: []models.IAMRole{
			{RoleName: "role-a", Path: "/", Tags: map[string]string{}},
		},
		Users:    []models.IAMUser{},
		Policies: []models.IAMPolicy{},
		Groups:   []models.IAMGroup{},
	}

	results := detector.DetectAllIAMDrift(plans, lives)
	assert.Empty(t, results)
}
