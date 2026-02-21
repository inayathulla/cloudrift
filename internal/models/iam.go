// Package models defines data structures for cloud resources.
package models

// IAMRole represents an AWS IAM role with its configuration.
//
// This model captures the key attributes compared for drift detection
// between Terraform plans and live AWS state.
type IAMRole struct {
	// TerraformAddress is the Terraform resource address (e.g., "aws_iam_role.my_role").
	TerraformAddress string `json:"terraform_address,omitempty"`

	// RoleName is the name of the IAM role.
	RoleName string `json:"role_name"`

	// Arn is the Amazon Resource Name for the role.
	Arn string `json:"arn,omitempty"`

	// Path is the IAM path for the role (e.g., "/", "/service-roles/").
	Path string `json:"path"`

	// AssumeRolePolicy is the JSON trust policy document that grants permission to assume the role.
	AssumeRolePolicy string `json:"assume_role_policy"`

	// MaxSessionDuration is the maximum session duration in seconds (default 3600).
	MaxSessionDuration int `json:"max_session_duration"`

	// Description is the description of the role.
	Description string `json:"description,omitempty"`

	// Tags contains the role's tag key-value pairs.
	Tags map[string]string `json:"tags"`

	// AttachedPolicies lists the ARNs of managed policies attached to the role.
	AttachedPolicies []string `json:"attached_policies"`
}

// Name returns the role name for display purposes.
func (r IAMRole) Name() string {
	if r.RoleName != "" {
		return r.RoleName
	}
	return r.Arn
}

// IAMUser represents an AWS IAM user with its configuration.
type IAMUser struct {
	// TerraformAddress is the Terraform resource address (e.g., "aws_iam_user.my_user").
	TerraformAddress string `json:"terraform_address,omitempty"`

	// UserName is the name of the IAM user.
	UserName string `json:"user_name"`

	// Arn is the Amazon Resource Name for the user.
	Arn string `json:"arn,omitempty"`

	// Path is the IAM path for the user.
	Path string `json:"path"`

	// Tags contains the user's tag key-value pairs.
	Tags map[string]string `json:"tags"`

	// AttachedPolicies lists the ARNs of managed policies attached to the user.
	AttachedPolicies []string `json:"attached_policies"`
}

// Name returns the user name for display purposes.
func (u IAMUser) Name() string {
	if u.UserName != "" {
		return u.UserName
	}
	return u.Arn
}

// IAMPolicy represents an AWS IAM managed policy.
type IAMPolicy struct {
	// TerraformAddress is the Terraform resource address (e.g., "aws_iam_policy.my_policy").
	TerraformAddress string `json:"terraform_address,omitempty"`

	// PolicyName is the name of the IAM policy.
	PolicyName string `json:"policy_name"`

	// Arn is the Amazon Resource Name for the policy.
	Arn string `json:"arn,omitempty"`

	// Path is the IAM path for the policy.
	Path string `json:"path"`

	// Description is the description of the policy.
	Description string `json:"description,omitempty"`

	// PolicyDocument is the JSON policy document.
	PolicyDocument string `json:"policy_document"`

	// Tags contains the policy's tag key-value pairs.
	Tags map[string]string `json:"tags"`
}

// Name returns the policy name for display purposes.
func (p IAMPolicy) Name() string {
	if p.PolicyName != "" {
		return p.PolicyName
	}
	return p.Arn
}

// IAMGroup represents an AWS IAM group.
type IAMGroup struct {
	// TerraformAddress is the Terraform resource address (e.g., "aws_iam_group.my_group").
	TerraformAddress string `json:"terraform_address,omitempty"`

	// GroupName is the name of the IAM group.
	GroupName string `json:"group_name"`

	// Arn is the Amazon Resource Name for the group.
	Arn string `json:"arn,omitempty"`

	// Path is the IAM path for the group.
	Path string `json:"path"`

	// AttachedPolicies lists the ARNs of managed policies attached to the group.
	AttachedPolicies []string `json:"attached_policies"`

	// Members lists the user names that belong to this group.
	Members []string `json:"members"`
}

// Name returns the group name for display purposes.
func (g IAMGroup) Name() string {
	if g.GroupName != "" {
		return g.GroupName
	}
	return g.Arn
}

// IAMPlanResources holds all IAM resources parsed from a Terraform plan.
type IAMPlanResources struct {
	Roles    []IAMRole   `json:"roles"`
	Users    []IAMUser   `json:"users"`
	Policies []IAMPolicy `json:"policies"`
	Groups   []IAMGroup  `json:"groups"`
}

// TotalCount returns the total number of IAM resources across all types.
func (p *IAMPlanResources) TotalCount() int {
	return len(p.Roles) + len(p.Users) + len(p.Policies) + len(p.Groups)
}

// IAMLiveState holds all IAM resources fetched from AWS.
type IAMLiveState struct {
	Roles    []IAMRole   `json:"roles"`
	Users    []IAMUser   `json:"users"`
	Policies []IAMPolicy `json:"policies"`
	Groups   []IAMGroup  `json:"groups"`
}
