package detector

import (
	"fmt"
	"sort"
	"strings"

	sdkaws "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/inayathulla/cloudrift/internal/aws"
	"github.com/inayathulla/cloudrift/internal/models"
)

// EC2DriftResult captures the drift detection results for a single EC2 instance.
type EC2DriftResult struct {
	// InstanceName is the name of the instance (from Name tag or instance ID).
	InstanceName string

	// InstanceID is the EC2 instance ID.
	InstanceID string

	// TerraformAddress is the Terraform resource address.
	TerraformAddress string

	// Missing is true if the instance exists in the plan but not in AWS.
	Missing bool

	// InstanceTypeDiff is true if instance type differs.
	InstanceTypeDiff bool

	// AMIDiff is true if the AMI differs.
	AMIDiff bool

	// SubnetDiff is true if the subnet differs.
	SubnetDiff bool

	// SecurityGroupsDiff is true if security groups differ.
	SecurityGroupsDiff bool

	// TagDiffs maps tag keys to [expected, actual] value pairs for mismatched tags.
	TagDiffs map[string][2]string

	// ExtraTags contains tags present in AWS but not in the plan.
	ExtraTags map[string]string

	// EBSOptimizedDiff is true if EBS optimization setting differs.
	EBSOptimizedDiff bool

	// MonitoringDiff is true if monitoring setting differs.
	MonitoringDiff bool

	// KeyNameDiff is true if the key pair differs.
	KeyNameDiff bool

	// IAMProfileDiff is true if the IAM instance profile differs.
	IAMProfileDiff bool

	// RootVolumeDiff is true if root volume configuration differs.
	RootVolumeDiff bool
}

// EC2DriftDetector implements drift detection for AWS EC2 instances.
type EC2DriftDetector struct {
	cfg sdkaws.Config
}

// NewEC2DriftDetector creates a new EC2 drift detector with the given AWS configuration.
func NewEC2DriftDetector(cfg sdkaws.Config) *EC2DriftDetector {
	return &EC2DriftDetector{cfg: cfg}
}

// FetchLiveState retrieves the current state of all EC2 instances from AWS.
func (d *EC2DriftDetector) FetchLiveState() (interface{}, error) {
	return aws.FetchEC2Instances(d.cfg)
}

// DetectDrift compares Terraform-planned instance configurations against live AWS state.
func (d *EC2DriftDetector) DetectDrift(plan, live interface{}) ([]DriftResult, error) {
	plans, ok := plan.([]models.EC2Instance)
	if !ok {
		return nil, fmt.Errorf("plan type mismatch: expected []models.EC2Instance")
	}
	lives, ok := live.([]models.EC2Instance)
	if !ok {
		return nil, fmt.Errorf("live type mismatch: expected []models.EC2Instance")
	}

	ec2Results := DetectAllEC2Drift(plans, lives)

	// Convert EC2DriftResult to generic DriftResult for compatibility
	results := make([]DriftResult, 0, len(ec2Results))
	for _, r := range ec2Results {
		// Create a generic DriftResult with bucket name field (reusing existing struct)
		// This is a temporary compatibility layer until we fully migrate to the new interface
		dr := DriftResult{
			BucketName: r.InstanceName, // Reuse BucketName field for resource name
		}
		if r.Missing {
			dr.Missing = true
		}
		// Store tag diffs
		dr.TagDiffs = r.TagDiffs
		dr.ExtraTags = r.ExtraTags

		// Use other diff fields to indicate drift (hacky but maintains compatibility)
		if r.InstanceTypeDiff || r.AMIDiff || r.SubnetDiff || r.SecurityGroupsDiff ||
			r.EBSOptimizedDiff || r.MonitoringDiff || r.KeyNameDiff || r.IAMProfileDiff ||
			r.RootVolumeDiff {
			dr.AclDiff = true // Indicates "other diffs exist"
		}

		results = append(results, dr)
	}

	return results, nil
}

// DetectEC2Drift compares a single planned instance against its actual AWS state.
func DetectEC2Drift(plan models.EC2Instance, actual *models.EC2Instance) EC2DriftResult {
	res := EC2DriftResult{
		InstanceName:     plan.Name(),
		InstanceID:       plan.InstanceID,
		TerraformAddress: plan.TerraformAddress,
		TagDiffs:         make(map[string][2]string),
		ExtraTags:        make(map[string]string),
	}

	if actual == nil {
		res.Missing = true
		return res
	}

	// Instance type
	if plan.InstanceType != "" && plan.InstanceType != actual.InstanceType {
		res.InstanceTypeDiff = true
	}

	// AMI
	if plan.AMI != "" && plan.AMI != actual.AMI {
		res.AMIDiff = true
	}

	// Subnet
	if plan.SubnetID != "" && plan.SubnetID != actual.SubnetID {
		res.SubnetDiff = true
	}

	// Security groups (compare as sets)
	if !stringSlicesEqual(plan.SecurityGroupIDs, actual.SecurityGroupIDs) {
		res.SecurityGroupsDiff = true
	}

	// EBS optimized
	if plan.EBSOptimized != actual.EBSOptimized {
		res.EBSOptimizedDiff = true
	}

	// Monitoring
	if plan.Monitoring != actual.Monitoring {
		res.MonitoringDiff = true
	}

	// Key name
	if plan.KeyName != "" && plan.KeyName != actual.KeyName {
		res.KeyNameDiff = true
	}

	// IAM instance profile
	if plan.IAMInstanceProfile != "" && !strings.Contains(actual.IAMInstanceProfile, plan.IAMInstanceProfile) {
		res.IAMProfileDiff = true
	}

	// Root block device
	if plan.RootBlockDevice.VolumeType != "" {
		if plan.RootBlockDevice.VolumeType != actual.RootBlockDevice.VolumeType ||
			(plan.RootBlockDevice.VolumeSize > 0 && plan.RootBlockDevice.VolumeSize != actual.RootBlockDevice.VolumeSize) ||
			plan.RootBlockDevice.Encrypted != actual.RootBlockDevice.Encrypted {
			res.RootVolumeDiff = true
		}
	}

	// Tag diffs
	for k, v := range plan.Tags {
		if av, ok := actual.Tags[k]; !ok || av != v {
			res.TagDiffs[k] = [2]string{v, av}
		}
	}

	// Extra tags in AWS
	for k, av := range actual.Tags {
		if _, ok := plan.Tags[k]; !ok {
			res.ExtraTags[k] = av
		}
	}

	return res
}

// DetectAllEC2Drift performs drift detection across all planned EC2 instances.
func DetectAllEC2Drift(plans, lives []models.EC2Instance) []EC2DriftResult {
	// Build map of live instances by ID
	byID := make(map[string]*models.EC2Instance, len(lives))
	for i := range lives {
		byID[lives[i].InstanceID] = &lives[i]
	}

	// Also build map by Name tag for matching
	byName := make(map[string]*models.EC2Instance, len(lives))
	for i := range lives {
		if name := lives[i].Name(); name != lives[i].InstanceID {
			byName[name] = &lives[i]
		}
	}

	out := make([]EC2DriftResult, 0, len(plans))
	for _, p := range plans {
		// Try to find matching live instance by ID first, then by name
		var live *models.EC2Instance
		if p.InstanceID != "" {
			live = byID[p.InstanceID]
		}
		if live == nil {
			// For new instances, try matching by Name tag
			if name, ok := p.Tags["Name"]; ok {
				live = byName[name]
			}
		}

		dr := DetectEC2Drift(p, live)

		// Only include results with actual drift
		if dr.Missing ||
			dr.InstanceTypeDiff ||
			dr.AMIDiff ||
			dr.SubnetDiff ||
			dr.SecurityGroupsDiff ||
			len(dr.TagDiffs) > 0 ||
			len(dr.ExtraTags) > 0 ||
			dr.EBSOptimizedDiff ||
			dr.MonitoringDiff ||
			dr.KeyNameDiff ||
			dr.IAMProfileDiff ||
			dr.RootVolumeDiff {
			out = append(out, dr)
		}
	}
	return out
}

// stringSlicesEqual compares two string slices as sets (order-independent).
func stringSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	aCopy := make([]string, len(a))
	bCopy := make([]string, len(b))
	copy(aCopy, a)
	copy(bCopy, b)
	sort.Strings(aCopy)
	sort.Strings(bCopy)
	for i := range aCopy {
		if aCopy[i] != bCopy[i] {
			return false
		}
	}
	return true
}
