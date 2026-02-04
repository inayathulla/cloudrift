package detector

import (
	"testing"

	"github.com/inayathulla/cloudrift/internal/detector"
	"github.com/inayathulla/cloudrift/internal/models"
	"github.com/stretchr/testify/assert"
)

// Test missing instance
func TestDetectEC2Drift_Missing(t *testing.T) {
	plan := models.EC2Instance{
		InstanceID:   "i-12345",
		InstanceType: "t3.micro",
		Tags:         map[string]string{"Name": "test-instance"},
	}
	res := detector.DetectEC2Drift(plan, nil)
	assert.True(t, res.Missing)
	assert.Equal(t, "test-instance", res.InstanceName)
}

// Instance type drift
func TestDetectEC2Drift_InstanceType_Positive(t *testing.T) {
	plan := models.EC2Instance{
		InstanceID:   "i-12345",
		InstanceType: "t3.micro",
		Tags:         map[string]string{"Name": "test"},
	}
	actual := &models.EC2Instance{
		InstanceID:   "i-12345",
		InstanceType: "t3.small",
		Tags:         map[string]string{"Name": "test"},
	}
	res := detector.DetectEC2Drift(plan, actual)
	assert.True(t, res.InstanceTypeDiff)
}

func TestDetectEC2Drift_InstanceType_Negative(t *testing.T) {
	plan := models.EC2Instance{
		InstanceID:   "i-12345",
		InstanceType: "t3.micro",
		Tags:         map[string]string{"Name": "test"},
	}
	actual := &models.EC2Instance{
		InstanceID:   "i-12345",
		InstanceType: "t3.micro",
		Tags:         map[string]string{"Name": "test"},
	}
	res := detector.DetectEC2Drift(plan, actual)
	assert.False(t, res.InstanceTypeDiff)
}

// AMI drift
func TestDetectEC2Drift_AMI_Positive(t *testing.T) {
	plan := models.EC2Instance{
		InstanceID: "i-12345",
		AMI:        "ami-12345",
		Tags:       map[string]string{"Name": "test"},
	}
	actual := &models.EC2Instance{
		InstanceID: "i-12345",
		AMI:        "ami-67890",
		Tags:       map[string]string{"Name": "test"},
	}
	res := detector.DetectEC2Drift(plan, actual)
	assert.True(t, res.AMIDiff)
}

func TestDetectEC2Drift_AMI_Negative(t *testing.T) {
	plan := models.EC2Instance{
		InstanceID: "i-12345",
		AMI:        "ami-12345",
		Tags:       map[string]string{"Name": "test"},
	}
	actual := &models.EC2Instance{
		InstanceID: "i-12345",
		AMI:        "ami-12345",
		Tags:       map[string]string{"Name": "test"},
	}
	res := detector.DetectEC2Drift(plan, actual)
	assert.False(t, res.AMIDiff)
}

// Subnet drift
func TestDetectEC2Drift_Subnet_Positive(t *testing.T) {
	plan := models.EC2Instance{
		InstanceID: "i-12345",
		SubnetID:   "subnet-aaa",
		Tags:       map[string]string{"Name": "test"},
	}
	actual := &models.EC2Instance{
		InstanceID: "i-12345",
		SubnetID:   "subnet-bbb",
		Tags:       map[string]string{"Name": "test"},
	}
	res := detector.DetectEC2Drift(plan, actual)
	assert.True(t, res.SubnetDiff)
}

func TestDetectEC2Drift_Subnet_Negative(t *testing.T) {
	plan := models.EC2Instance{
		InstanceID: "i-12345",
		SubnetID:   "subnet-aaa",
		Tags:       map[string]string{"Name": "test"},
	}
	actual := &models.EC2Instance{
		InstanceID: "i-12345",
		SubnetID:   "subnet-aaa",
		Tags:       map[string]string{"Name": "test"},
	}
	res := detector.DetectEC2Drift(plan, actual)
	assert.False(t, res.SubnetDiff)
}

// Security groups drift
func TestDetectEC2Drift_SecurityGroups_Positive(t *testing.T) {
	plan := models.EC2Instance{
		InstanceID:       "i-12345",
		SecurityGroupIDs: []string{"sg-111", "sg-222"},
		Tags:             map[string]string{"Name": "test"},
	}
	actual := &models.EC2Instance{
		InstanceID:       "i-12345",
		SecurityGroupIDs: []string{"sg-111", "sg-333"},
		Tags:             map[string]string{"Name": "test"},
	}
	res := detector.DetectEC2Drift(plan, actual)
	assert.True(t, res.SecurityGroupsDiff)
}

func TestDetectEC2Drift_SecurityGroups_Negative(t *testing.T) {
	plan := models.EC2Instance{
		InstanceID:       "i-12345",
		SecurityGroupIDs: []string{"sg-111", "sg-222"},
		Tags:             map[string]string{"Name": "test"},
	}
	actual := &models.EC2Instance{
		InstanceID:       "i-12345",
		SecurityGroupIDs: []string{"sg-222", "sg-111"}, // Different order, same content
		Tags:             map[string]string{"Name": "test"},
	}
	res := detector.DetectEC2Drift(plan, actual)
	assert.False(t, res.SecurityGroupsDiff)
}

// Tag drift
func TestDetectEC2Drift_TagDiff_Positive(t *testing.T) {
	plan := models.EC2Instance{
		InstanceID: "i-12345",
		Tags:       map[string]string{"Name": "test", "Env": "prod"},
	}
	actual := &models.EC2Instance{
		InstanceID: "i-12345",
		Tags:       map[string]string{"Name": "test", "Env": "dev"},
	}
	res := detector.DetectEC2Drift(plan, actual)
	assert.Len(t, res.TagDiffs, 1)
	assert.Contains(t, res.TagDiffs, "Env")
}

func TestDetectEC2Drift_TagDiff_Negative(t *testing.T) {
	tags := map[string]string{"Name": "test", "Env": "prod"}
	plan := models.EC2Instance{
		InstanceID: "i-12345",
		Tags:       tags,
	}
	actual := &models.EC2Instance{
		InstanceID: "i-12345",
		Tags:       tags,
	}
	res := detector.DetectEC2Drift(plan, actual)
	assert.Empty(t, res.TagDiffs)
}

// Extra tags in AWS
func TestDetectEC2Drift_ExtraTags(t *testing.T) {
	plan := models.EC2Instance{
		InstanceID: "i-12345",
		Tags:       map[string]string{"Name": "test"},
	}
	actual := &models.EC2Instance{
		InstanceID: "i-12345",
		Tags:       map[string]string{"Name": "test", "aws:autoscaling:groupName": "asg-1"},
	}
	res := detector.DetectEC2Drift(plan, actual)
	assert.Len(t, res.ExtraTags, 1)
	assert.Contains(t, res.ExtraTags, "aws:autoscaling:groupName")
}

// EBS optimized drift
func TestDetectEC2Drift_EBSOptimized_Positive(t *testing.T) {
	plan := models.EC2Instance{
		InstanceID:   "i-12345",
		EBSOptimized: true,
		Tags:         map[string]string{"Name": "test"},
	}
	actual := &models.EC2Instance{
		InstanceID:   "i-12345",
		EBSOptimized: false,
		Tags:         map[string]string{"Name": "test"},
	}
	res := detector.DetectEC2Drift(plan, actual)
	assert.True(t, res.EBSOptimizedDiff)
}

func TestDetectEC2Drift_EBSOptimized_Negative(t *testing.T) {
	plan := models.EC2Instance{
		InstanceID:   "i-12345",
		EBSOptimized: true,
		Tags:         map[string]string{"Name": "test"},
	}
	actual := &models.EC2Instance{
		InstanceID:   "i-12345",
		EBSOptimized: true,
		Tags:         map[string]string{"Name": "test"},
	}
	res := detector.DetectEC2Drift(plan, actual)
	assert.False(t, res.EBSOptimizedDiff)
}

// Monitoring drift
func TestDetectEC2Drift_Monitoring_Positive(t *testing.T) {
	plan := models.EC2Instance{
		InstanceID: "i-12345",
		Monitoring: true,
		Tags:       map[string]string{"Name": "test"},
	}
	actual := &models.EC2Instance{
		InstanceID: "i-12345",
		Monitoring: false,
		Tags:       map[string]string{"Name": "test"},
	}
	res := detector.DetectEC2Drift(plan, actual)
	assert.True(t, res.MonitoringDiff)
}

func TestDetectEC2Drift_Monitoring_Negative(t *testing.T) {
	plan := models.EC2Instance{
		InstanceID: "i-12345",
		Monitoring: false,
		Tags:       map[string]string{"Name": "test"},
	}
	actual := &models.EC2Instance{
		InstanceID: "i-12345",
		Monitoring: false,
		Tags:       map[string]string{"Name": "test"},
	}
	res := detector.DetectEC2Drift(plan, actual)
	assert.False(t, res.MonitoringDiff)
}

// Key name drift
func TestDetectEC2Drift_KeyName_Positive(t *testing.T) {
	plan := models.EC2Instance{
		InstanceID: "i-12345",
		KeyName:    "key-1",
		Tags:       map[string]string{"Name": "test"},
	}
	actual := &models.EC2Instance{
		InstanceID: "i-12345",
		KeyName:    "key-2",
		Tags:       map[string]string{"Name": "test"},
	}
	res := detector.DetectEC2Drift(plan, actual)
	assert.True(t, res.KeyNameDiff)
}

func TestDetectEC2Drift_KeyName_Negative(t *testing.T) {
	plan := models.EC2Instance{
		InstanceID: "i-12345",
		KeyName:    "key-1",
		Tags:       map[string]string{"Name": "test"},
	}
	actual := &models.EC2Instance{
		InstanceID: "i-12345",
		KeyName:    "key-1",
		Tags:       map[string]string{"Name": "test"},
	}
	res := detector.DetectEC2Drift(plan, actual)
	assert.False(t, res.KeyNameDiff)
}

// DetectAllEC2Drift tests
func TestDetectAllEC2Drift_MultipleInstances(t *testing.T) {
	plans := []models.EC2Instance{
		{InstanceID: "i-111", InstanceType: "t3.micro", Tags: map[string]string{"Name": "inst-1"}},
		{InstanceID: "i-222", InstanceType: "t3.small", Tags: map[string]string{"Name": "inst-2"}},
		{InstanceID: "i-333", InstanceType: "t3.medium", Tags: map[string]string{"Name": "inst-3"}},
	}
	lives := []models.EC2Instance{
		{InstanceID: "i-111", InstanceType: "t3.micro", Tags: map[string]string{"Name": "inst-1"}},  // No drift
		{InstanceID: "i-222", InstanceType: "t3.large", Tags: map[string]string{"Name": "inst-2"}}, // Type drift
		// i-333 missing
	}
	results := detector.DetectAllEC2Drift(plans, lives)

	// Should have 2 results: type drift and missing
	assert.Len(t, results, 2)
}

func TestDetectAllEC2Drift_NoDrift(t *testing.T) {
	plans := []models.EC2Instance{
		{InstanceID: "i-111", InstanceType: "t3.micro", Tags: map[string]string{"Name": "inst-1"}},
	}
	lives := []models.EC2Instance{
		{InstanceID: "i-111", InstanceType: "t3.micro", Tags: map[string]string{"Name": "inst-1"}},
	}
	results := detector.DetectAllEC2Drift(plans, lives)
	assert.Empty(t, results)
}

// Test Name() method on EC2Instance
func TestEC2Instance_Name(t *testing.T) {
	// With Name tag
	inst := models.EC2Instance{
		InstanceID: "i-12345",
		Tags:       map[string]string{"Name": "my-instance"},
	}
	assert.Equal(t, "my-instance", inst.Name())

	// Without Name tag
	inst2 := models.EC2Instance{
		InstanceID: "i-67890",
		Tags:       map[string]string{},
	}
	assert.Equal(t, "i-67890", inst2.Name())

	// With empty Name tag
	inst3 := models.EC2Instance{
		InstanceID: "i-abcde",
		Tags:       map[string]string{"Name": ""},
	}
	assert.Equal(t, "i-abcde", inst3.Name())
}
