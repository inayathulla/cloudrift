package parser

import (
	"testing"

	"github.com/inayathulla/cloudrift/internal/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadIAMPlan(t *testing.T) {
	result, err := parser.LoadIAMPlan("../../../examples/iam-plan.json")
	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify roles
	assert.Len(t, result.Roles, 2)
	assert.Equal(t, "lambda-exec-role", result.Roles[0].RoleName)
	assert.Equal(t, "aws_iam_role.lambda_exec", result.Roles[0].TerraformAddress)
	assert.Equal(t, "/", result.Roles[0].Path)
	assert.Equal(t, "IAM role for Lambda execution", result.Roles[0].Description)
	assert.Equal(t, 3600, result.Roles[0].MaxSessionDuration)
	assert.Contains(t, result.Roles[0].AssumeRolePolicy, "lambda.amazonaws.com")
	assert.Equal(t, "production", result.Roles[0].Tags["Environment"])
	assert.Equal(t, "platform", result.Roles[0].Tags["Team"])

	// Second role
	assert.Equal(t, "ecs-task-role", result.Roles[1].RoleName)

	// Verify users
	assert.Len(t, result.Users, 2)
	assert.Equal(t, "deploy-user", result.Users[0].UserName)
	assert.Equal(t, "/ci/", result.Users[0].Path)
	assert.Equal(t, "terraform", result.Users[0].Tags["ManagedBy"])

	assert.Equal(t, "readonly-user", result.Users[1].UserName)
	assert.Equal(t, "/", result.Users[1].Path)

	// Verify policies
	assert.Len(t, result.Policies, 2)
	assert.Equal(t, "deploy-policy", result.Policies[0].PolicyName)
	assert.Equal(t, "Policy for CI/CD deployments", result.Policies[0].Description)
	assert.Contains(t, result.Policies[0].PolicyDocument, "s3:GetObject")

	assert.Equal(t, "admin-policy", result.Policies[1].PolicyName)
	assert.Contains(t, result.Policies[1].PolicyDocument, `"Action":"*"`)

	// Verify groups
	assert.Len(t, result.Groups, 2)
	assert.Equal(t, "developers", result.Groups[0].GroupName)
	assert.Equal(t, "/", result.Groups[0].Path)
	assert.Equal(t, "admins", result.Groups[1].GroupName)

	// Total count
	assert.Equal(t, 8, result.TotalCount())
}

func TestLoadIAMPlan_FileNotFound(t *testing.T) {
	_, err := parser.LoadIAMPlan("nonexistent.json")
	assert.Error(t, err)
}

func TestLoadIAMPlan_SkipsDeletedResources(t *testing.T) {
	// The example plan only has "create" actions, so all should be parsed
	result, err := parser.LoadIAMPlan("../../../examples/iam-plan.json")
	require.NoError(t, err)
	assert.Equal(t, 8, result.TotalCount())
}
