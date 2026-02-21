package models

import (
	"testing"

	"github.com/inayathulla/cloudrift/internal/models"
	"github.com/stretchr/testify/assert"
)

// ──────────────────────────────────────────────────────────────────────────────
// IAMRole
// ──────────────────────────────────────────────────────────────────────────────

func TestIAMRole_Name_WithRoleName(t *testing.T) {
	role := models.IAMRole{RoleName: "my-role", Arn: "arn:aws:iam::123:role/my-role"}
	assert.Equal(t, "my-role", role.Name())
}

func TestIAMRole_Name_FallbackToArn(t *testing.T) {
	role := models.IAMRole{RoleName: "", Arn: "arn:aws:iam::123:role/my-role"}
	assert.Equal(t, "arn:aws:iam::123:role/my-role", role.Name())
}

// ──────────────────────────────────────────────────────────────────────────────
// IAMUser
// ──────────────────────────────────────────────────────────────────────────────

func TestIAMUser_Name_WithUserName(t *testing.T) {
	user := models.IAMUser{UserName: "deploy-user", Arn: "arn:aws:iam::123:user/deploy-user"}
	assert.Equal(t, "deploy-user", user.Name())
}

func TestIAMUser_Name_FallbackToArn(t *testing.T) {
	user := models.IAMUser{UserName: "", Arn: "arn:aws:iam::123:user/deploy-user"}
	assert.Equal(t, "arn:aws:iam::123:user/deploy-user", user.Name())
}

// ──────────────────────────────────────────────────────────────────────────────
// IAMPolicy
// ──────────────────────────────────────────────────────────────────────────────

func TestIAMPolicy_Name_WithPolicyName(t *testing.T) {
	pol := models.IAMPolicy{PolicyName: "admin-policy", Arn: "arn:aws:iam::123:policy/admin-policy"}
	assert.Equal(t, "admin-policy", pol.Name())
}

func TestIAMPolicy_Name_FallbackToArn(t *testing.T) {
	pol := models.IAMPolicy{PolicyName: "", Arn: "arn:aws:iam::123:policy/admin-policy"}
	assert.Equal(t, "arn:aws:iam::123:policy/admin-policy", pol.Name())
}

// ──────────────────────────────────────────────────────────────────────────────
// IAMGroup
// ──────────────────────────────────────────────────────────────────────────────

func TestIAMGroup_Name_WithGroupName(t *testing.T) {
	grp := models.IAMGroup{GroupName: "developers", Arn: "arn:aws:iam::123:group/developers"}
	assert.Equal(t, "developers", grp.Name())
}

func TestIAMGroup_Name_FallbackToArn(t *testing.T) {
	grp := models.IAMGroup{GroupName: "", Arn: "arn:aws:iam::123:group/developers"}
	assert.Equal(t, "arn:aws:iam::123:group/developers", grp.Name())
}

// ──────────────────────────────────────────────────────────────────────────────
// IAMPlanResources
// ──────────────────────────────────────────────────────────────────────────────

func TestIAMPlanResources_TotalCount(t *testing.T) {
	pr := models.IAMPlanResources{
		Roles:    []models.IAMRole{{RoleName: "r1"}, {RoleName: "r2"}},
		Users:    []models.IAMUser{{UserName: "u1"}},
		Policies: []models.IAMPolicy{{PolicyName: "p1"}, {PolicyName: "p2"}, {PolicyName: "p3"}},
		Groups:   []models.IAMGroup{{GroupName: "g1"}},
	}
	assert.Equal(t, 7, pr.TotalCount())
}

func TestIAMPlanResources_TotalCount_Empty(t *testing.T) {
	pr := models.IAMPlanResources{}
	assert.Equal(t, 0, pr.TotalCount())
}
