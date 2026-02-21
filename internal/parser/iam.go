package parser

import (
	"github.com/inayathulla/cloudrift/internal/models"
)

// ParseIAMRoles extracts aws_iam_role resources from a Terraform plan.
//
// Parses the following attributes from each role:
//   - name, path, description
//   - assume_role_policy (JSON trust policy)
//   - max_session_duration
//   - tags
//
// Resources being deleted (with nil "after" state) are skipped.
func ParseIAMRoles(plan *TerraformPlan) []models.IAMRole {
	var roles []models.IAMRole

	for _, rc := range plan.ResourceChanges {
		if rc.Type != "aws_iam_role" {
			continue
		}
		after := rc.Change.After
		if after == nil {
			continue
		}

		role := models.IAMRole{
			TerraformAddress: rc.Address,
			Tags:             make(map[string]string),
			AttachedPolicies: make([]string, 0),
		}

		if v, ok := after["name"].(string); ok {
			role.RoleName = v
		}
		if v, ok := after["path"].(string); ok {
			role.Path = v
		}
		if v, ok := after["description"].(string); ok {
			role.Description = v
		}
		if v, ok := after["assume_role_policy"].(string); ok {
			role.AssumeRolePolicy = v
		}
		if v, ok := after["max_session_duration"].(float64); ok {
			role.MaxSessionDuration = int(v)
		}

		// Tags
		if tags, ok := after["tags"].(map[string]interface{}); ok {
			for k, v := range tags {
				if vStr, ok := v.(string); ok {
					role.Tags[k] = vStr
				}
			}
		}
		if tags, ok := after["tags_all"].(map[string]interface{}); ok {
			for k, v := range tags {
				if vStr, ok := v.(string); ok {
					if _, exists := role.Tags[k]; !exists {
						role.Tags[k] = vStr
					}
				}
			}
		}

		roles = append(roles, role)
	}

	return roles
}

// ParseIAMUsers extracts aws_iam_user resources from a Terraform plan.
//
// Parses the following attributes from each user:
//   - name, path
//   - tags
//
// Resources being deleted (with nil "after" state) are skipped.
func ParseIAMUsers(plan *TerraformPlan) []models.IAMUser {
	var users []models.IAMUser

	for _, rc := range plan.ResourceChanges {
		if rc.Type != "aws_iam_user" {
			continue
		}
		after := rc.Change.After
		if after == nil {
			continue
		}

		user := models.IAMUser{
			TerraformAddress: rc.Address,
			Tags:             make(map[string]string),
			AttachedPolicies: make([]string, 0),
		}

		if v, ok := after["name"].(string); ok {
			user.UserName = v
		}
		if v, ok := after["path"].(string); ok {
			user.Path = v
		}

		// Tags
		if tags, ok := after["tags"].(map[string]interface{}); ok {
			for k, v := range tags {
				if vStr, ok := v.(string); ok {
					user.Tags[k] = vStr
				}
			}
		}
		if tags, ok := after["tags_all"].(map[string]interface{}); ok {
			for k, v := range tags {
				if vStr, ok := v.(string); ok {
					if _, exists := user.Tags[k]; !exists {
						user.Tags[k] = vStr
					}
				}
			}
		}

		users = append(users, user)
	}

	return users
}

// ParseIAMPolicies extracts aws_iam_policy resources from a Terraform plan.
//
// Parses the following attributes from each policy:
//   - name, path, description
//   - policy (JSON policy document)
//   - tags
//
// Resources being deleted (with nil "after" state) are skipped.
func ParseIAMPolicies(plan *TerraformPlan) []models.IAMPolicy {
	var policies []models.IAMPolicy

	for _, rc := range plan.ResourceChanges {
		if rc.Type != "aws_iam_policy" {
			continue
		}
		after := rc.Change.After
		if after == nil {
			continue
		}

		policy := models.IAMPolicy{
			TerraformAddress: rc.Address,
			Tags:             make(map[string]string),
		}

		if v, ok := after["name"].(string); ok {
			policy.PolicyName = v
		}
		if v, ok := after["path"].(string); ok {
			policy.Path = v
		}
		if v, ok := after["description"].(string); ok {
			policy.Description = v
		}
		if v, ok := after["policy"].(string); ok {
			policy.PolicyDocument = v
		}

		// Tags
		if tags, ok := after["tags"].(map[string]interface{}); ok {
			for k, v := range tags {
				if vStr, ok := v.(string); ok {
					policy.Tags[k] = vStr
				}
			}
		}
		if tags, ok := after["tags_all"].(map[string]interface{}); ok {
			for k, v := range tags {
				if vStr, ok := v.(string); ok {
					if _, exists := policy.Tags[k]; !exists {
						policy.Tags[k] = vStr
					}
				}
			}
		}

		policies = append(policies, policy)
	}

	return policies
}

// ParseIAMGroups extracts aws_iam_group resources from a Terraform plan.
//
// Parses the following attributes from each group:
//   - name, path
//
// Resources being deleted (with nil "after" state) are skipped.
func ParseIAMGroups(plan *TerraformPlan) []models.IAMGroup {
	var groups []models.IAMGroup

	for _, rc := range plan.ResourceChanges {
		if rc.Type != "aws_iam_group" {
			continue
		}
		after := rc.Change.After
		if after == nil {
			continue
		}

		group := models.IAMGroup{
			TerraformAddress: rc.Address,
			AttachedPolicies: make([]string, 0),
			Members:          make([]string, 0),
		}

		if v, ok := after["name"].(string); ok {
			group.GroupName = v
		}
		if v, ok := after["path"].(string); ok {
			group.Path = v
		}

		groups = append(groups, group)
	}

	return groups
}

// ParseAllIAMResources extracts all IAM resources from a Terraform plan.
// This is a convenience function that calls all four IAM parsers.
func ParseAllIAMResources(plan *TerraformPlan) *models.IAMPlanResources {
	return &models.IAMPlanResources{
		Roles:    ParseIAMRoles(plan),
		Users:    ParseIAMUsers(plan),
		Policies: ParseIAMPolicies(plan),
		Groups:   ParseIAMGroups(plan),
	}
}
