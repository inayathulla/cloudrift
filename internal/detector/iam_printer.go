package detector

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/inayathulla/cloudrift/internal/models"
)

// IAMDriftResultPrinter handles console output for IAM drift detection results.
type IAMDriftResultPrinter struct{}

// PrintDrift outputs IAM drift detection results to the console.
func (p IAMDriftResultPrinter) PrintDrift(results interface{}, plan, live interface{}) {
	driftResults, ok := results.([]DriftResult)
	if !ok {
		color.Red("Invalid results type for IAM printer")
		return
	}

	planResources, ok := plan.(*models.IAMPlanResources)
	if !ok {
		color.Red("Invalid plan type for IAM printer")
		return
	}

	liveState, ok := live.(*models.IAMLiveState)
	if !ok {
		color.Red("Invalid live type for IAM printer")
		return
	}

	// Build lookup maps for plan resources
	planRoles := make(map[string]*models.IAMRole, len(planResources.Roles))
	for i := range planResources.Roles {
		planRoles[planResources.Roles[i].RoleName] = &planResources.Roles[i]
	}
	planUsers := make(map[string]*models.IAMUser, len(planResources.Users))
	for i := range planResources.Users {
		planUsers[planResources.Users[i].UserName] = &planResources.Users[i]
	}
	planPolicies := make(map[string]*models.IAMPolicy, len(planResources.Policies))
	for i := range planResources.Policies {
		planPolicies[planResources.Policies[i].PolicyName] = &planResources.Policies[i]
	}
	planGroups := make(map[string]*models.IAMGroup, len(planResources.Groups))
	for i := range planResources.Groups {
		planGroups[planResources.Groups[i].GroupName] = &planResources.Groups[i]
	}

	// Build lookup maps for live resources
	liveRoles := make(map[string]*models.IAMRole, len(liveState.Roles))
	for i := range liveState.Roles {
		liveRoles[liveState.Roles[i].RoleName] = &liveState.Roles[i]
	}
	liveUsers := make(map[string]*models.IAMUser, len(liveState.Users))
	for i := range liveState.Users {
		liveUsers[liveState.Users[i].UserName] = &liveState.Users[i]
	}
	livePolicies := make(map[string]*models.IAMPolicy, len(liveState.Policies))
	for i := range liveState.Policies {
		livePolicies[liveState.Policies[i].PolicyName] = &liveState.Policies[i]
	}
	liveGroups := make(map[string]*models.IAMGroup, len(liveState.Groups))
	for i := range liveState.Groups {
		liveGroups[liveState.Groups[i].GroupName] = &liveState.Groups[i]
	}

	totalPlanned := planResources.TotalCount()

	// Summary banner
	fmt.Println()
	color.Cyan("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	color.Cyan("              IAM DRIFT DETECTION                  ")
	color.Cyan("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("  Planned: %d roles, %d users, %d policies, %d groups (%d total)\n",
		len(planResources.Roles), len(planResources.Users),
		len(planResources.Policies), len(planResources.Groups), totalPlanned)
	fmt.Printf("  Live:    %d roles, %d users, %d policies, %d groups\n",
		len(liveState.Roles), len(liveState.Users),
		len(liveState.Policies), len(liveState.Groups))
	fmt.Printf("  Drifted: %d resources\n", len(driftResults))
	fmt.Println()

	if len(driftResults) == 0 {
		color.Green("  No drift detected! All planned IAM resources match live state.")
		return
	}

	// Print each drift result
	for _, dr := range driftResults {
		resourceName := dr.BucketName // Reused field
		fmt.Println()
		color.Yellow("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

		// Determine resource type icon and label
		icon := "  "
		if _, ok := planRoles[resourceName]; ok {
			icon = "  Role: "
		} else if _, ok := planUsers[resourceName]; ok {
			icon = "  User: "
		} else if _, ok := planPolicies[resourceName]; ok {
			icon = "  Policy: "
		} else if _, ok := planGroups[resourceName]; ok {
			icon = "  Group: "
		}
		fmt.Printf("%s%s\n", icon, color.CyanString(resourceName))

		if dr.Missing {
			color.Red("   MISSING - Resource not found in AWS")
			continue
		}

		// Show attribute differences for roles
		if planRole, ok := planRoles[resourceName]; ok {
			if liveRole, ok := liveRoles[resourceName]; ok && dr.AclDiff {
				printIAMRoleDiffs(planRole, liveRole)
			}
		}

		// Show attribute differences for users
		if planUser, ok := planUsers[resourceName]; ok {
			if liveUser, ok := liveUsers[resourceName]; ok && dr.AclDiff {
				printIAMUserDiffs(planUser, liveUser)
			}
		}

		// Show attribute differences for policies
		if planPolicy, ok := planPolicies[resourceName]; ok {
			if livePolicy, ok := livePolicies[resourceName]; ok && dr.AclDiff {
				printIAMPolicyDiffs(planPolicy, livePolicy)
			}
		}

		// Show attribute differences for groups
		if planGroup, ok := planGroups[resourceName]; ok {
			if liveGroup, ok := liveGroups[resourceName]; ok && dr.AclDiff {
				printIAMGroupDiffs(planGroup, liveGroup)
			}
		}

		// Tag differences
		if len(dr.TagDiffs) > 0 {
			color.Yellow("   Tag differences:")
			for k, v := range dr.TagDiffs {
				fmt.Printf("      %s:\n", k)
				fmt.Printf("        %s %q\n", color.RedString("- planned:"), v[0])
				fmt.Printf("        %s %q\n", color.GreenString("+ actual: "), v[1])
			}
		}

		// Extra tags
		if len(dr.ExtraTags) > 0 {
			color.Blue("   Extra tags in AWS:")
			for k, v := range dr.ExtraTags {
				fmt.Printf("      %s: %q\n", k, v)
			}
		}
	}

	fmt.Println()
	color.Yellow("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
}

// printIAMRoleDiffs prints attribute-level diffs for an IAM role.
func printIAMRoleDiffs(plan, live *models.IAMRole) {
	color.Yellow("   Attribute differences:")

	if plan.AssumeRolePolicy != "" && !jsonEqual(plan.AssumeRolePolicy, live.AssumeRolePolicy) {
		fmt.Printf("      Assume Role Policy:\n")
		fmt.Printf("        %s (trust policy changed)\n", color.RedString("- planned differs from actual"))
	}

	if plan.MaxSessionDuration > 0 && plan.MaxSessionDuration != live.MaxSessionDuration {
		fmt.Printf("      Max Session Duration:\n")
		fmt.Printf("        %s %d\n", color.RedString("- planned:"), plan.MaxSessionDuration)
		fmt.Printf("        %s %d\n", color.GreenString("+ actual: "), live.MaxSessionDuration)
	}

	if plan.Description != "" && plan.Description != live.Description {
		fmt.Printf("      Description:\n")
		fmt.Printf("        %s %q\n", color.RedString("- planned:"), plan.Description)
		fmt.Printf("        %s %q\n", color.GreenString("+ actual: "), live.Description)
	}

	if plan.Path != "" && plan.Path != live.Path {
		fmt.Printf("      Path:\n")
		fmt.Printf("        %s %s\n", color.RedString("- planned:"), plan.Path)
		fmt.Printf("        %s %s\n", color.GreenString("+ actual: "), live.Path)
	}

	if len(plan.AttachedPolicies) > 0 && !stringSlicesEqual(plan.AttachedPolicies, live.AttachedPolicies) {
		fmt.Printf("      Attached Policies:\n")
		fmt.Printf("        %s %v\n", color.RedString("- planned:"), plan.AttachedPolicies)
		fmt.Printf("        %s %v\n", color.GreenString("+ actual: "), live.AttachedPolicies)
	}
}

// printIAMUserDiffs prints attribute-level diffs for an IAM user.
func printIAMUserDiffs(plan, live *models.IAMUser) {
	color.Yellow("   Attribute differences:")

	if plan.Path != "" && plan.Path != live.Path {
		fmt.Printf("      Path:\n")
		fmt.Printf("        %s %s\n", color.RedString("- planned:"), plan.Path)
		fmt.Printf("        %s %s\n", color.GreenString("+ actual: "), live.Path)
	}

	if len(plan.AttachedPolicies) > 0 && !stringSlicesEqual(plan.AttachedPolicies, live.AttachedPolicies) {
		fmt.Printf("      Attached Policies:\n")
		fmt.Printf("        %s %v\n", color.RedString("- planned:"), plan.AttachedPolicies)
		fmt.Printf("        %s %v\n", color.GreenString("+ actual: "), live.AttachedPolicies)
	}
}

// printIAMPolicyDiffs prints attribute-level diffs for an IAM policy.
func printIAMPolicyDiffs(plan, live *models.IAMPolicy) {
	color.Yellow("   Attribute differences:")

	if plan.PolicyDocument != "" && !jsonEqual(plan.PolicyDocument, live.PolicyDocument) {
		fmt.Printf("      Policy Document:\n")
		fmt.Printf("        %s (policy document changed)\n", color.RedString("- planned differs from actual"))
	}

	if plan.Description != "" && plan.Description != live.Description {
		fmt.Printf("      Description:\n")
		fmt.Printf("        %s %q\n", color.RedString("- planned:"), plan.Description)
		fmt.Printf("        %s %q\n", color.GreenString("+ actual: "), live.Description)
	}

	if plan.Path != "" && plan.Path != live.Path {
		fmt.Printf("      Path:\n")
		fmt.Printf("        %s %s\n", color.RedString("- planned:"), plan.Path)
		fmt.Printf("        %s %s\n", color.GreenString("+ actual: "), live.Path)
	}
}

// printIAMGroupDiffs prints attribute-level diffs for an IAM group.
func printIAMGroupDiffs(plan, live *models.IAMGroup) {
	color.Yellow("   Attribute differences:")

	if plan.Path != "" && plan.Path != live.Path {
		fmt.Printf("      Path:\n")
		fmt.Printf("        %s %s\n", color.RedString("- planned:"), plan.Path)
		fmt.Printf("        %s %s\n", color.GreenString("+ actual: "), live.Path)
	}

	if len(plan.AttachedPolicies) > 0 && !stringSlicesEqual(plan.AttachedPolicies, live.AttachedPolicies) {
		fmt.Printf("      Attached Policies:\n")
		fmt.Printf("        %s %v\n", color.RedString("- planned:"), plan.AttachedPolicies)
		fmt.Printf("        %s %v\n", color.GreenString("+ actual: "), live.AttachedPolicies)
	}

	if len(plan.Members) > 0 && !stringSlicesEqual(plan.Members, live.Members) {
		fmt.Printf("      Members:\n")
		fmt.Printf("        %s %v\n", color.RedString("- planned:"), plan.Members)
		fmt.Printf("        %s %v\n", color.GreenString("+ actual: "), live.Members)
	}
}
