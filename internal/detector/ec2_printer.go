package detector

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/inayathulla/cloudrift/internal/models"
)

// EC2DriftResultPrinter handles console output for EC2 drift detection results.
type EC2DriftResultPrinter struct{}

// PrintDrift outputs EC2 drift detection results to the console.
func (p EC2DriftResultPrinter) PrintDrift(results interface{}, plan, live interface{}) {
	driftResults, ok := results.([]DriftResult)
	if !ok {
		color.Red("❌ Invalid results type for EC2 printer")
		return
	}

	planInstances, ok := plan.([]models.EC2Instance)
	if !ok {
		color.Red("❌ Invalid plan type for EC2 printer")
		return
	}

	liveInstances, ok := live.([]models.EC2Instance)
	if !ok {
		color.Red("❌ Invalid live type for EC2 printer")
		return
	}

	// Build lookup maps
	planMap := make(map[string]*models.EC2Instance, len(planInstances))
	for i := range planInstances {
		planMap[planInstances[i].Name()] = &planInstances[i]
	}

	liveMap := make(map[string]*models.EC2Instance, len(liveInstances))
	for i := range liveInstances {
		liveMap[liveInstances[i].Name()] = &liveInstances[i]
	}

	// Summary
	fmt.Println()
	color.Cyan("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	color.Cyan("               EC2 DRIFT DETECTION                 ")
	color.Cyan("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("📊 Planned instances: %d\n", len(planInstances))
	fmt.Printf("☁️  Live instances: %d\n", len(liveInstances))
	fmt.Printf("⚠️  Instances with drift: %d\n", len(driftResults))
	fmt.Println()

	if len(driftResults) == 0 {
		color.Green("✅ No drift detected! All planned instances match live state.")
		return
	}

	// Print each drift result
	for _, dr := range driftResults {
		instanceName := dr.BucketName // Reused field
		fmt.Println()
		color.Yellow("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Printf("🖥️  Instance: %s\n", color.CyanString(instanceName))

		plannedInst := planMap[instanceName]
		liveInst := liveMap[instanceName]

		if dr.Missing {
			color.Red("   ❌ MISSING - Instance not found in AWS")
			if plannedInst != nil {
				fmt.Printf("      Planned instance type: %s\n", plannedInst.InstanceType)
				fmt.Printf("      Planned AMI: %s\n", plannedInst.AMI)
			}
			continue
		}

		// Show attribute differences
		if dr.AclDiff && plannedInst != nil && liveInst != nil {
			color.Yellow("   📋 Attribute differences:")

			if plannedInst.InstanceType != liveInst.InstanceType {
				fmt.Printf("      • Instance Type:\n")
				fmt.Printf("        %s %s\n", color.RedString("- planned:"), plannedInst.InstanceType)
				fmt.Printf("        %s %s\n", color.GreenString("+ actual: "), liveInst.InstanceType)
			}

			if plannedInst.AMI != "" && plannedInst.AMI != liveInst.AMI {
				fmt.Printf("      • AMI:\n")
				fmt.Printf("        %s %s\n", color.RedString("- planned:"), plannedInst.AMI)
				fmt.Printf("        %s %s\n", color.GreenString("+ actual: "), liveInst.AMI)
			}

			if plannedInst.SubnetID != "" && plannedInst.SubnetID != liveInst.SubnetID {
				fmt.Printf("      • Subnet ID:\n")
				fmt.Printf("        %s %s\n", color.RedString("- planned:"), plannedInst.SubnetID)
				fmt.Printf("        %s %s\n", color.GreenString("+ actual: "), liveInst.SubnetID)
			}

			if !stringSlicesEqual(plannedInst.SecurityGroupIDs, liveInst.SecurityGroupIDs) {
				fmt.Printf("      • Security Groups:\n")
				fmt.Printf("        %s %v\n", color.RedString("- planned:"), plannedInst.SecurityGroupIDs)
				fmt.Printf("        %s %v\n", color.GreenString("+ actual: "), liveInst.SecurityGroupIDs)
			}

			if plannedInst.EBSOptimized != liveInst.EBSOptimized {
				fmt.Printf("      • EBS Optimized:\n")
				fmt.Printf("        %s %v\n", color.RedString("- planned:"), plannedInst.EBSOptimized)
				fmt.Printf("        %s %v\n", color.GreenString("+ actual: "), liveInst.EBSOptimized)
			}

			if plannedInst.Monitoring != liveInst.Monitoring {
				fmt.Printf("      • Detailed Monitoring:\n")
				fmt.Printf("        %s %v\n", color.RedString("- planned:"), plannedInst.Monitoring)
				fmt.Printf("        %s %v\n", color.GreenString("+ actual: "), liveInst.Monitoring)
			}
		}

		// Tag differences
		if len(dr.TagDiffs) > 0 {
			color.Yellow("   🏷️  Tag differences:")
			for k, v := range dr.TagDiffs {
				fmt.Printf("      • %s:\n", k)
				fmt.Printf("        %s %q\n", color.RedString("- planned:"), v[0])
				fmt.Printf("        %s %q\n", color.GreenString("+ actual: "), v[1])
			}
		}

		// Extra tags
		if len(dr.ExtraTags) > 0 {
			color.Blue("   🏷️  Extra tags in AWS:")
			for k, v := range dr.ExtraTags {
				fmt.Printf("      • %s: %q\n", k, v)
			}
		}
	}

	fmt.Println()
	color.Yellow("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
}
