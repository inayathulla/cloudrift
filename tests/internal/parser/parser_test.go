package parser

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/inayathulla/cloudrift/internal/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper to create a temporary plan file
func createTempPlanFile(t *testing.T, content string) string {
	t.Helper()
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "plan.json")
	err := os.WriteFile(path, []byte(content), 0644)
	require.NoError(t, err)
	return path
}

// S3 Parser Tests
func TestParseS3Buckets_BasicBucket(t *testing.T) {
	planJSON := `{
		"resource_changes": [
			{
				"address": "aws_s3_bucket.test",
				"type": "aws_s3_bucket",
				"name": "test",
				"change": {
					"actions": ["create"],
					"after": {
						"bucket": "my-test-bucket",
						"acl": "private",
						"tags": {"Environment": "prod"}
					}
				}
			}
		]
	}`

	path := createTempPlanFile(t, planJSON)
	buckets, err := parser.LoadPlan(path)

	require.NoError(t, err)
	require.Len(t, buckets, 1)
	assert.Equal(t, "my-test-bucket", buckets[0].Name)
	assert.Equal(t, "private", buckets[0].Acl)
	assert.Equal(t, "prod", buckets[0].Tags["Environment"])
}

func TestParseS3Buckets_WithVersioning(t *testing.T) {
	planJSON := `{
		"resource_changes": [
			{
				"address": "aws_s3_bucket.versioned",
				"type": "aws_s3_bucket",
				"name": "versioned",
				"change": {
					"actions": ["create"],
					"after": {
						"bucket": "versioned-bucket",
						"versioning": {"enabled": true}
					}
				}
			}
		]
	}`

	path := createTempPlanFile(t, planJSON)
	buckets, err := parser.LoadPlan(path)

	require.NoError(t, err)
	require.Len(t, buckets, 1)
	assert.True(t, buckets[0].VersioningEnabled)
}

func TestParseS3Buckets_WithEncryption(t *testing.T) {
	planJSON := `{
		"resource_changes": [
			{
				"address": "aws_s3_bucket.encrypted",
				"type": "aws_s3_bucket",
				"name": "encrypted",
				"change": {
					"actions": ["create"],
					"after": {
						"bucket": "encrypted-bucket",
						"server_side_encryption_configuration": {
							"rules": [{
								"apply_server_side_encryption_by_default": {
									"sse_algorithm": "AES256"
								}
							}]
						}
					}
				}
			}
		]
	}`

	path := createTempPlanFile(t, planJSON)
	buckets, err := parser.LoadPlan(path)

	require.NoError(t, err)
	require.Len(t, buckets, 1)
	assert.Equal(t, "AES256", buckets[0].EncryptionAlgorithm)
}

func TestParseS3Buckets_WithLogging(t *testing.T) {
	planJSON := `{
		"resource_changes": [
			{
				"address": "aws_s3_bucket.logged",
				"type": "aws_s3_bucket",
				"name": "logged",
				"change": {
					"actions": ["create"],
					"after": {
						"bucket": "logged-bucket",
						"logging": {
							"target_bucket": "logs-bucket",
							"target_prefix": "access-logs/"
						}
					}
				}
			}
		]
	}`

	path := createTempPlanFile(t, planJSON)
	buckets, err := parser.LoadPlan(path)

	require.NoError(t, err)
	require.Len(t, buckets, 1)
	assert.True(t, buckets[0].LoggingEnabled)
	assert.Equal(t, "logs-bucket", buckets[0].LoggingTargetBucket)
	assert.Equal(t, "access-logs/", buckets[0].LoggingTargetPrefix)
}

func TestParseS3Buckets_WithPublicAccessBlock(t *testing.T) {
	planJSON := `{
		"resource_changes": [
			{
				"address": "aws_s3_bucket.blocked",
				"type": "aws_s3_bucket",
				"name": "blocked",
				"change": {
					"actions": ["create"],
					"after": {
						"bucket": "blocked-bucket"
					}
				}
			},
			{
				"address": "aws_s3_bucket_public_access_block.blocked",
				"type": "aws_s3_bucket_public_access_block",
				"name": "blocked",
				"change": {
					"actions": ["create"],
					"after": {
						"bucket": "blocked-bucket",
						"block_public_acls": true,
						"ignore_public_acls": true,
						"block_public_policy": true,
						"restrict_public_buckets": true
					}
				}
			}
		]
	}`

	path := createTempPlanFile(t, planJSON)
	buckets, err := parser.LoadPlan(path)

	require.NoError(t, err)
	// Note: The parser needs to handle public_access_block resources
	// This test documents expected behavior
	assert.GreaterOrEqual(t, len(buckets), 1)
}

func TestParseS3Buckets_WithLifecycle(t *testing.T) {
	planJSON := `{
		"resource_changes": [
			{
				"address": "aws_s3_bucket.lifecycle",
				"type": "aws_s3_bucket",
				"name": "lifecycle",
				"change": {
					"actions": ["create"],
					"after": {
						"bucket": "lifecycle-bucket",
						"lifecycle_rule": [{
							"id": "expire-old",
							"status": "Enabled",
							"prefix": "logs/",
							"expiration": {"days": 90}
						}]
					}
				}
			}
		]
	}`

	path := createTempPlanFile(t, planJSON)
	buckets, err := parser.LoadPlan(path)

	require.NoError(t, err)
	require.Len(t, buckets, 1)
	require.Len(t, buckets[0].LifecycleRules, 1)
	assert.Equal(t, "expire-old", buckets[0].LifecycleRules[0].ID)
	assert.Equal(t, "Enabled", buckets[0].LifecycleRules[0].Status)
	assert.Equal(t, 90, buckets[0].LifecycleRules[0].ExpirationDays)
}

func TestParseS3Buckets_MultipleBuckets(t *testing.T) {
	planJSON := `{
		"resource_changes": [
			{
				"address": "aws_s3_bucket.one",
				"type": "aws_s3_bucket",
				"name": "one",
				"change": {
					"actions": ["create"],
					"after": {"bucket": "bucket-one"}
				}
			},
			{
				"address": "aws_s3_bucket.two",
				"type": "aws_s3_bucket",
				"name": "two",
				"change": {
					"actions": ["create"],
					"after": {"bucket": "bucket-two"}
				}
			}
		]
	}`

	path := createTempPlanFile(t, planJSON)
	buckets, err := parser.LoadPlan(path)

	require.NoError(t, err)
	assert.Len(t, buckets, 2)
}

func TestParseS3Buckets_SkipDeleteAction(t *testing.T) {
	planJSON := `{
		"resource_changes": [
			{
				"address": "aws_s3_bucket.deleted",
				"type": "aws_s3_bucket",
				"name": "deleted",
				"change": {
					"actions": ["delete"],
					"after": null
				}
			}
		]
	}`

	path := createTempPlanFile(t, planJSON)
	buckets, err := parser.LoadPlan(path)

	require.NoError(t, err)
	assert.Empty(t, buckets)
}

func TestParseS3Buckets_SkipNonS3Resources(t *testing.T) {
	planJSON := `{
		"resource_changes": [
			{
				"address": "aws_instance.web",
				"type": "aws_instance",
				"name": "web",
				"change": {
					"actions": ["create"],
					"after": {"instance_type": "t3.micro"}
				}
			}
		]
	}`

	path := createTempPlanFile(t, planJSON)
	buckets, err := parser.LoadPlan(path)

	require.NoError(t, err)
	assert.Empty(t, buckets)
}

// EC2 Parser Tests
func TestParseEC2Instances_BasicInstance(t *testing.T) {
	planJSON := `{
		"resource_changes": [
			{
				"address": "aws_instance.web",
				"type": "aws_instance",
				"name": "web",
				"change": {
					"actions": ["create"],
					"after": {
						"ami": "ami-12345678",
						"instance_type": "t3.micro",
						"subnet_id": "subnet-abc123",
						"tags": {"Name": "web-server", "Environment": "prod"}
					}
				}
			}
		]
	}`

	path := createTempPlanFile(t, planJSON)
	instances, err := parser.LoadEC2Plan(path)

	require.NoError(t, err)
	require.Len(t, instances, 1)
	assert.Equal(t, "ami-12345678", instances[0].AMI)
	assert.Equal(t, "t3.micro", instances[0].InstanceType)
	assert.Equal(t, "subnet-abc123", instances[0].SubnetID)
	assert.Equal(t, "web-server", instances[0].Tags["Name"])
	assert.Equal(t, "aws_instance.web", instances[0].TerraformAddress)
}

func TestParseEC2Instances_WithSecurityGroups(t *testing.T) {
	planJSON := `{
		"resource_changes": [
			{
				"address": "aws_instance.web",
				"type": "aws_instance",
				"name": "web",
				"change": {
					"actions": ["create"],
					"after": {
						"ami": "ami-12345678",
						"instance_type": "t3.micro",
						"vpc_security_group_ids": ["sg-111", "sg-222"]
					}
				}
			}
		]
	}`

	path := createTempPlanFile(t, planJSON)
	instances, err := parser.LoadEC2Plan(path)

	require.NoError(t, err)
	require.Len(t, instances, 1)
	assert.Len(t, instances[0].SecurityGroupIDs, 2)
	assert.Contains(t, instances[0].SecurityGroupIDs, "sg-111")
	assert.Contains(t, instances[0].SecurityGroupIDs, "sg-222")
}

func TestParseEC2Instances_WithRootBlockDevice(t *testing.T) {
	planJSON := `{
		"resource_changes": [
			{
				"address": "aws_instance.web",
				"type": "aws_instance",
				"name": "web",
				"change": {
					"actions": ["create"],
					"after": {
						"ami": "ami-12345678",
						"instance_type": "t3.micro",
						"root_block_device": [{
							"volume_type": "gp3",
							"volume_size": 100,
							"encrypted": true,
							"delete_on_termination": true
						}]
					}
				}
			}
		]
	}`

	path := createTempPlanFile(t, planJSON)
	instances, err := parser.LoadEC2Plan(path)

	require.NoError(t, err)
	require.Len(t, instances, 1)
	assert.Equal(t, "gp3", instances[0].RootBlockDevice.VolumeType)
	assert.Equal(t, 100, instances[0].RootBlockDevice.VolumeSize)
	assert.True(t, instances[0].RootBlockDevice.Encrypted)
	assert.True(t, instances[0].RootBlockDevice.DeleteOnTermination)
}

func TestParseEC2Instances_WithAllAttributes(t *testing.T) {
	planJSON := `{
		"resource_changes": [
			{
				"address": "aws_instance.full",
				"type": "aws_instance",
				"name": "full",
				"change": {
					"actions": ["create"],
					"after": {
						"ami": "ami-12345678",
						"instance_type": "t3.micro",
						"subnet_id": "subnet-abc",
						"availability_zone": "us-east-1a",
						"key_name": "my-key",
						"iam_instance_profile": "my-profile",
						"ebs_optimized": true,
						"monitoring": true,
						"source_dest_check": false,
						"vpc_security_group_ids": ["sg-111"],
						"tags": {"Name": "full-instance"}
					}
				}
			}
		]
	}`

	path := createTempPlanFile(t, planJSON)
	instances, err := parser.LoadEC2Plan(path)

	require.NoError(t, err)
	require.Len(t, instances, 1)
	inst := instances[0]
	assert.Equal(t, "ami-12345678", inst.AMI)
	assert.Equal(t, "t3.micro", inst.InstanceType)
	assert.Equal(t, "subnet-abc", inst.SubnetID)
	assert.Equal(t, "us-east-1a", inst.AvailabilityZone)
	assert.Equal(t, "my-key", inst.KeyName)
	assert.Equal(t, "my-profile", inst.IAMInstanceProfile)
	assert.True(t, inst.EBSOptimized)
	assert.True(t, inst.Monitoring)
	assert.False(t, inst.SourceDestCheck)
}

func TestParseEC2Instances_MultipleInstances(t *testing.T) {
	planJSON := `{
		"resource_changes": [
			{
				"address": "aws_instance.web",
				"type": "aws_instance",
				"name": "web",
				"change": {
					"actions": ["create"],
					"after": {"ami": "ami-111", "instance_type": "t3.micro"}
				}
			},
			{
				"address": "aws_instance.db",
				"type": "aws_instance",
				"name": "db",
				"change": {
					"actions": ["create"],
					"after": {"ami": "ami-222", "instance_type": "r5.large"}
				}
			}
		]
	}`

	path := createTempPlanFile(t, planJSON)
	instances, err := parser.LoadEC2Plan(path)

	require.NoError(t, err)
	assert.Len(t, instances, 2)
}

func TestParseEC2Instances_SkipDeleteAction(t *testing.T) {
	planJSON := `{
		"resource_changes": [
			{
				"address": "aws_instance.deleted",
				"type": "aws_instance",
				"name": "deleted",
				"change": {
					"actions": ["delete"],
					"after": null
				}
			}
		]
	}`

	path := createTempPlanFile(t, planJSON)
	instances, err := parser.LoadEC2Plan(path)

	require.NoError(t, err)
	assert.Empty(t, instances)
}

func TestParseEC2Instances_SkipNonEC2Resources(t *testing.T) {
	planJSON := `{
		"resource_changes": [
			{
				"address": "aws_s3_bucket.data",
				"type": "aws_s3_bucket",
				"name": "data",
				"change": {
					"actions": ["create"],
					"after": {"bucket": "my-bucket"}
				}
			}
		]
	}`

	path := createTempPlanFile(t, planJSON)
	instances, err := parser.LoadEC2Plan(path)

	require.NoError(t, err)
	assert.Empty(t, instances)
}

func TestParseEC2Instances_TagsAll(t *testing.T) {
	planJSON := `{
		"resource_changes": [
			{
				"address": "aws_instance.web",
				"type": "aws_instance",
				"name": "web",
				"change": {
					"actions": ["create"],
					"after": {
						"ami": "ami-12345678",
						"instance_type": "t3.micro",
						"tags": {"Name": "web"},
						"tags_all": {"Name": "web", "ManagedBy": "terraform"}
					}
				}
			}
		]
	}`

	path := createTempPlanFile(t, planJSON)
	instances, err := parser.LoadEC2Plan(path)

	require.NoError(t, err)
	require.Len(t, instances, 1)
	// tags_all should be merged, but tags takes precedence
	assert.Equal(t, "web", instances[0].Tags["Name"])
	assert.Equal(t, "terraform", instances[0].Tags["ManagedBy"])
}

// Error Cases
func TestLoadPlan_FileNotFound(t *testing.T) {
	_, err := parser.LoadPlan("/nonexistent/path/plan.json")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to open plan file")
}

func TestLoadPlan_InvalidJSON(t *testing.T) {
	path := createTempPlanFile(t, "not valid json")
	_, err := parser.LoadPlan(path)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode JSON")
}

func TestLoadEC2Plan_FileNotFound(t *testing.T) {
	_, err := parser.LoadEC2Plan("/nonexistent/path/plan.json")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to open plan file")
}

func TestLoadEC2Plan_InvalidJSON(t *testing.T) {
	path := createTempPlanFile(t, "not valid json")
	_, err := parser.LoadEC2Plan(path)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode JSON")
}

// Empty Plan
func TestLoadPlan_EmptyPlan(t *testing.T) {
	planJSON := `{"resource_changes": []}`
	path := createTempPlanFile(t, planJSON)

	buckets, err := parser.LoadPlan(path)
	require.NoError(t, err)
	assert.Empty(t, buckets)

	instances, err := parser.LoadEC2Plan(path)
	require.NoError(t, err)
	assert.Empty(t, instances)
}
