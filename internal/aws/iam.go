package aws

import (
	"context"
	"fmt"
	"net/url"

	sdkaws "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	"golang.org/x/sync/errgroup"

	"github.com/inayathulla/cloudrift/internal/models"
)

// FetchIAMResources retrieves all IAM resources (roles, users, policies, groups) from AWS.
//
// The four resource types are fetched in parallel using errgroup. AWS-managed roles,
// users, and policies (those with paths starting with /aws-service-role/ or /aws-reserved/)
// are excluded to focus on customer-managed resources.
//
// Parameters:
//   - cfg: AWS SDK configuration for API calls
//
// Returns:
//   - *models.IAMLiveState: all IAM resources
//   - error: if any AWS API call fails
func FetchIAMResources(cfg sdkaws.Config) (*models.IAMLiveState, error) {
	ctx := context.Background()
	client := iam.NewFromConfig(cfg)

	var (
		roles    []models.IAMRole
		users    []models.IAMUser
		policies []models.IAMPolicy
		groups   []models.IAMGroup
	)

	// Fetch all four IAM resource types in parallel
	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		var err error
		roles, err = fetchIAMRoles(ctx, client)
		return err
	})

	g.Go(func() error {
		var err error
		users, err = fetchIAMUsers(ctx, client)
		return err
	})

	g.Go(func() error {
		var err error
		policies, err = fetchIAMPolicies(ctx, client)
		return err
	})

	g.Go(func() error {
		var err error
		groups, err = fetchIAMGroups(ctx, client)
		return err
	})

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return &models.IAMLiveState{
		Roles:    roles,
		Users:    users,
		Policies: policies,
		Groups:   groups,
	}, nil
}

// fetchIAMRoles lists all customer-managed IAM roles with their trust policies and attached policies.
func fetchIAMRoles(ctx context.Context, client *iam.Client) ([]models.IAMRole, error) {
	var roles []models.IAMRole
	paginator := iam.NewListRolesPaginator(client, &iam.ListRolesInput{})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("ListRoles: %w", err)
		}

		for _, r := range page.Roles {
			// Skip AWS service-linked roles
			path := safeString(r.Path)
			if path == "/aws-service-role/" || len(path) > 1 && path[:len("/aws-service-role/")] == "/aws-service-role/" {
				continue
			}

			role := convertIAMRole(r)

			// Fetch attached managed policies
			attached, err := fetchAttachedRolePolicies(ctx, client, role.RoleName)
			if err != nil {
				fmt.Printf("  warning: listing attached policies for role %s: %v\n", role.RoleName, err)
			} else {
				role.AttachedPolicies = attached
			}

			roles = append(roles, role)
		}
	}

	return roles, nil
}

// convertIAMRole converts an AWS SDK IAM role to our model.
func convertIAMRole(r types.Role) models.IAMRole {
	role := models.IAMRole{
		RoleName:           safeString(r.RoleName),
		Arn:                safeString(r.Arn),
		Path:               safeString(r.Path),
		Tags:               make(map[string]string),
		AttachedPolicies:   make([]string, 0),
		MaxSessionDuration: 3600, // default
	}

	// Trust policy document is URL-encoded
	if r.AssumeRolePolicyDocument != nil {
		decoded, err := url.QueryUnescape(*r.AssumeRolePolicyDocument)
		if err == nil {
			role.AssumeRolePolicy = decoded
		} else {
			role.AssumeRolePolicy = *r.AssumeRolePolicyDocument
		}
	}

	if r.MaxSessionDuration != nil {
		role.MaxSessionDuration = int(*r.MaxSessionDuration)
	}

	if r.Description != nil {
		role.Description = *r.Description
	}

	for _, tag := range r.Tags {
		if tag.Key != nil && tag.Value != nil {
			role.Tags[*tag.Key] = *tag.Value
		}
	}

	return role
}

// fetchAttachedRolePolicies lists managed policy ARNs attached to a role.
func fetchAttachedRolePolicies(ctx context.Context, client *iam.Client, roleName string) ([]string, error) {
	var arns []string
	paginator := iam.NewListAttachedRolePoliciesPaginator(client, &iam.ListAttachedRolePoliciesInput{
		RoleName: &roleName,
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		for _, p := range page.AttachedPolicies {
			if p.PolicyArn != nil {
				arns = append(arns, *p.PolicyArn)
			}
		}
	}

	return arns, nil
}

// fetchIAMUsers lists all IAM users with their tags and attached policies.
func fetchIAMUsers(ctx context.Context, client *iam.Client) ([]models.IAMUser, error) {
	var users []models.IAMUser
	paginator := iam.NewListUsersPaginator(client, &iam.ListUsersInput{})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("ListUsers: %w", err)
		}

		for _, u := range page.Users {
			user := convertIAMUser(u)

			// Fetch tags
			tagResp, err := client.ListUserTags(ctx, &iam.ListUserTagsInput{
				UserName: u.UserName,
			})
			if err == nil {
				for _, tag := range tagResp.Tags {
					if tag.Key != nil && tag.Value != nil {
						user.Tags[*tag.Key] = *tag.Value
					}
				}
			}

			// Fetch attached managed policies
			attached, err := fetchAttachedUserPolicies(ctx, client, user.UserName)
			if err != nil {
				fmt.Printf("  warning: listing attached policies for user %s: %v\n", user.UserName, err)
			} else {
				user.AttachedPolicies = attached
			}

			users = append(users, user)
		}
	}

	return users, nil
}

// convertIAMUser converts an AWS SDK IAM user to our model.
func convertIAMUser(u types.User) models.IAMUser {
	return models.IAMUser{
		UserName:         safeString(u.UserName),
		Arn:              safeString(u.Arn),
		Path:             safeString(u.Path),
		Tags:             make(map[string]string),
		AttachedPolicies: make([]string, 0),
	}
}

// fetchAttachedUserPolicies lists managed policy ARNs attached to a user.
func fetchAttachedUserPolicies(ctx context.Context, client *iam.Client, userName string) ([]string, error) {
	var arns []string
	paginator := iam.NewListAttachedUserPoliciesPaginator(client, &iam.ListAttachedUserPoliciesInput{
		UserName: &userName,
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		for _, p := range page.AttachedPolicies {
			if p.PolicyArn != nil {
				arns = append(arns, *p.PolicyArn)
			}
		}
	}

	return arns, nil
}

// fetchIAMPolicies lists all customer-managed IAM policies with their policy documents.
func fetchIAMPolicies(ctx context.Context, client *iam.Client) ([]models.IAMPolicy, error) {
	var policies []models.IAMPolicy
	paginator := iam.NewListPoliciesPaginator(client, &iam.ListPoliciesInput{
		Scope: types.PolicyScopeTypeLocal, // Customer-managed only
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("ListPolicies: %w", err)
		}

		for _, p := range page.Policies {
			pol := convertIAMPolicy(p)

			// Fetch the policy document from the default version
			if p.DefaultVersionId != nil && p.Arn != nil {
				versionResp, err := client.GetPolicyVersion(ctx, &iam.GetPolicyVersionInput{
					PolicyArn: p.Arn,
					VersionId: p.DefaultVersionId,
				})
				if err == nil && versionResp.PolicyVersion != nil && versionResp.PolicyVersion.Document != nil {
					decoded, err := url.QueryUnescape(*versionResp.PolicyVersion.Document)
					if err == nil {
						pol.PolicyDocument = decoded
					} else {
						pol.PolicyDocument = *versionResp.PolicyVersion.Document
					}
				}
			}

			// Fetch tags
			if p.Arn != nil {
				tagResp, err := client.ListPolicyTags(ctx, &iam.ListPolicyTagsInput{
					PolicyArn: p.Arn,
				})
				if err == nil {
					for _, tag := range tagResp.Tags {
						if tag.Key != nil && tag.Value != nil {
							pol.Tags[*tag.Key] = *tag.Value
						}
					}
				}
			}

			policies = append(policies, pol)
		}
	}

	return policies, nil
}

// convertIAMPolicy converts an AWS SDK IAM policy to our model.
func convertIAMPolicy(p types.Policy) models.IAMPolicy {
	pol := models.IAMPolicy{
		PolicyName: safeString(p.PolicyName),
		Arn:        safeString(p.Arn),
		Path:       safeString(p.Path),
		Tags:       make(map[string]string),
	}

	if p.Description != nil {
		pol.Description = *p.Description
	}

	return pol
}

// fetchIAMGroups lists all IAM groups with their attached policies and members.
func fetchIAMGroups(ctx context.Context, client *iam.Client) ([]models.IAMGroup, error) {
	var groups []models.IAMGroup
	paginator := iam.NewListGroupsPaginator(client, &iam.ListGroupsInput{})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("ListGroups: %w", err)
		}

		for _, g := range page.Groups {
			group := convertIAMGroup(g)

			// Fetch attached managed policies
			attachedPaginator := iam.NewListAttachedGroupPoliciesPaginator(client, &iam.ListAttachedGroupPoliciesInput{
				GroupName: g.GroupName,
			})
			for attachedPaginator.HasMorePages() {
				ap, err := attachedPaginator.NextPage(ctx)
				if err != nil {
					fmt.Printf("  warning: listing attached policies for group %s: %v\n", group.GroupName, err)
					break
				}
				for _, p := range ap.AttachedPolicies {
					if p.PolicyArn != nil {
						group.AttachedPolicies = append(group.AttachedPolicies, *p.PolicyArn)
					}
				}
			}

			// Fetch group members via GetGroup
			groupResp, err := client.GetGroup(ctx, &iam.GetGroupInput{
				GroupName: g.GroupName,
			})
			if err == nil {
				for _, u := range groupResp.Users {
					if u.UserName != nil {
						group.Members = append(group.Members, *u.UserName)
					}
				}
			}

			groups = append(groups, group)
		}
	}

	return groups, nil
}

// convertIAMGroup converts an AWS SDK IAM group to our model.
func convertIAMGroup(g types.Group) models.IAMGroup {
	return models.IAMGroup{
		GroupName:        safeString(g.GroupName),
		Arn:              safeString(g.Arn),
		Path:             safeString(g.Path),
		AttachedPolicies: make([]string, 0),
		Members:          make([]string, 0),
	}
}
