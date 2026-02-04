# Cloudrift Usage Guide

Complete documentation for using Cloudrift to detect infrastructure drift and policy violations.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Configuration](#configuration)
- [Input Files](#input-files)
- [Commands](#commands)
- [Output Formats](#output-formats)
- [Policy Engine](#policy-engine)
- [CI/CD Integration](#cicd-integration)
- [Examples](#examples)
- [Troubleshooting](#troubleshooting)

---

## Prerequisites

### Required
- **AWS CLI** configured with valid credentials
- **Terraform** (to generate plan files)
- **Go 1.21+** (if building from source)

### AWS Permissions

Cloudrift requires read-only access to AWS resources. Minimum IAM permissions:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "s3:ListAllMyBuckets",
        "s3:GetBucketLocation",
        "s3:GetBucketVersioning",
        "s3:GetBucketEncryption",
        "s3:GetBucketLogging",
        "s3:GetBucketTagging",
        "s3:GetBucketPublicAccessBlock",
        "s3:GetLifecycleConfiguration",
        "s3:GetBucketAcl",
        "ec2:DescribeInstances",
        "ec2:DescribeTags",
        "sts:GetCallerIdentity"
      ],
      "Resource": "*"
    }
  ]
}
```

---

## Installation

### Option 1: Go Install
```bash
go install github.com/inayathulla/cloudrift@latest
```

### Option 2: Build from Source
```bash
git clone https://github.com/inayathulla/cloudrift.git
cd cloudrift
go build -o cloudrift .
```

### Option 3: Docker
```bash
docker pull inayathulla/cloudrift
docker run -v ~/.aws:/root/.aws -v $(pwd):/work inayathulla/cloudrift scan --config=/work/cloudrift.yml
```

---

## Configuration

### Configuration File (cloudrift.yml)

Create a `cloudrift.yml` file in your project root:

```yaml
# AWS credentials profile (from ~/.aws/credentials)
aws_profile: default

# AWS region to scan
region: us-east-1

# Path to Terraform plan JSON file
plan_path: ./plan.json
```

### Configuration Options

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `aws_profile` | string | Yes | - | AWS credentials profile name |
| `region` | string | Yes | - | AWS region (e.g., us-east-1, eu-west-1) |
| `plan_path` | string | Yes | - | Path to Terraform plan JSON file |

### Environment Variables

You can also use AWS environment variables:

```bash
export AWS_ACCESS_KEY_ID="your-access-key"
export AWS_SECRET_ACCESS_KEY="your-secret-key"
export AWS_DEFAULT_REGION="us-east-1"
```

---

## Input Files

### Terraform Plan JSON

Cloudrift requires a Terraform plan in JSON format. Generate it using:

```bash
# Initialize Terraform
terraform init

# Create binary plan
terraform plan -out=tfplan.binary

# Convert to JSON
terraform show -json tfplan.binary > plan.json
```

### Plan JSON Structure

Cloudrift parses the `resource_changes` array from the plan:

```json
{
  "resource_changes": [
    {
      "address": "aws_s3_bucket.example",
      "type": "aws_s3_bucket",
      "name": "example",
      "change": {
        "actions": ["create"],
        "after": {
          "bucket": "my-bucket-name",
          "acl": "private",
          "tags": {
            "Environment": "production"
          },
          "versioning": {
            "enabled": true
          }
        }
      }
    }
  ]
}
```

### Supported Resource Types

| Resource Type | Service Flag | Attributes Checked |
|---------------|--------------|-------------------|
| `aws_s3_bucket` | `--service=s3` | bucket, acl, tags, versioning, encryption, logging, public_access_block, lifecycle_rules |
| `aws_instance` | `--service=ec2` | instance_type, ami, subnet_id, security_groups, tags, ebs_optimized, monitoring, key_name, iam_instance_profile, root_block_device |

---

## Commands

### Main Command

```bash
cloudrift [command] [flags]
```

### Scan Command

```bash
cloudrift scan [flags]
```

#### Flags

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--config` | `-c` | string | `cloudrift.yml` | Path to configuration file |
| `--service` | `-s` | string | `s3` | AWS service to scan (s3, ec2) |
| `--format` | `-f` | string | `console` | Output format (console, json, sarif) |
| `--output` | `-o` | string | - | Write output to file |
| `--policy-dir` | `-p` | string | - | Directory with custom OPA policies |
| `--fail-on-violation` | - | bool | `false` | Exit non-zero on policy violations |
| `--skip-policies` | - | bool | `false` | Skip policy evaluation |
| `--no-emoji` | - | bool | `false` | Use ASCII instead of emojis |

#### Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success, no violations |
| 1 | Error (config, AWS, etc.) |
| 2 | Policy violations found (with `--fail-on-violation`) |

---

## Output Formats

### Console (Default)

Human-readable, colorized output with emojis:

```bash
cloudrift scan --service=s3
```

**Sample Output:**
```
🚀 Starting Cloudrift scan...
✔️  AWS config loaded in 45ms
✔️  Credentials validated in 234ms
🔐 Connected as: arn:aws:iam::123456789012:user/dev (123456789012) [us-east-1] in 156ms
📄 Plan loaded from json in 12ms
✔️  Live S3 state fetched in 1.234s
✔️  Drift detection completed
✔️  Evaluated 7 policies in 23ms
⚠️  Found 2 policy violations
⚠️  Found 4 policy warnings
✔️  Scan completed in 2.345s!

⚠️  Drift detected!

🪣 my-bucket
  🏷️  Tags:
    ⚠️  Missing:
        • Environment:production
  🔐 Encryption mismatch:
    • expected → "AES256"
    • actual   → ""

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Summary:
  S3 Buckets scanned: 1
  Buckets with drift: 1
  Buckets without drift: 0
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

### JSON

Machine-readable format for CI/CD pipelines:

```bash
cloudrift scan --service=s3 --format=json
```

**Sample Output:**
```json
{
  "service": "S3",
  "account_id": "123456789012",
  "region": "us-east-1",
  "total_resources": 3,
  "drift_count": 1,
  "drifts": [
    {
      "resource_id": "my-bucket",
      "resource_type": "aws_s3_bucket",
      "resource_name": "my-bucket",
      "missing": false,
      "diffs": {
        "versioning_enabled": ["true", "false"],
        "tags.Environment": ["production", ""]
      },
      "extra_attributes": {
        "tags.CreatedBy": "manual"
      },
      "severity": "warning"
    }
  ],
  "scan_duration_ms": 2345,
  "timestamp": "2024-01-15T10:30:00Z"
}
```

### SARIF

Static Analysis Results Interchange Format for GitHub/GitLab Security integration:

```bash
cloudrift scan --service=s3 --format=sarif --output=results.sarif
```

**Sample Output:**
```json
{
  "$schema": "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json",
  "version": "2.1.0",
  "runs": [
    {
      "tool": {
        "driver": {
          "name": "Cloudrift",
          "version": "1.0.0",
          "informationUri": "https://github.com/inayathulla/cloudrift",
          "rules": [
            {
              "id": "DRIFT001",
              "name": "resource-missing",
              "shortDescription": {
                "text": "Resource exists in Terraform plan but not in AWS"
              }
            },
            {
              "id": "DRIFT002",
              "name": "attribute-mismatch",
              "shortDescription": {
                "text": "Resource attribute differs between Terraform plan and AWS"
              }
            }
          ]
        }
      },
      "results": [
        {
          "ruleId": "DRIFT002",
          "level": "warning",
          "message": {
            "text": "Attribute 'versioning_enabled' of my-bucket (aws_s3_bucket) differs: expected true, got false"
          }
        }
      ]
    }
  ]
}
```

---

## Policy Engine

Cloudrift includes a built-in OPA (Open Policy Agent) policy engine with 7 security policies.

### Built-in Policies

| Policy ID | Name | Severity | Description |
|-----------|------|----------|-------------|
| S3-001 | S3 Encryption Required | high | S3 buckets must have server-side encryption |
| S3-002 | S3 KMS Encryption Recommended | low | Recommends KMS over AES256 |
| S3-003 | S3 Block Public ACLs | high | block_public_acls must be enabled |
| S3-004 | S3 Block Public Policy | high | block_public_policy must be enabled |
| S3-005 | S3 Ignore Public ACLs | high | ignore_public_acls must be enabled |
| S3-006 | S3 Restrict Public Buckets | high | restrict_public_buckets must be enabled |
| S3-007 | S3 No Public Read ACL | critical | public-read ACL not allowed |
| S3-008 | S3 No Public Read-Write ACL | critical | public-read-write ACL not allowed |
| TAG-001 | Environment Tag Required | medium | Resources must have Environment tag |
| TAG-002 | Owner Tag Recommended | low | Resources should have Owner tag |
| TAG-003 | Project Tag Recommended | low | Resources should have Project tag |
| TAG-004 | Name Tag Recommended | low | Resources should have Name tag |

### Custom Policies

Create custom OPA policies in Rego format:

```rego
# my-policies/custom.rego
package cloudrift.custom

deny[result] {
    input.resource.type == "aws_s3_bucket"
    input.resource.planned.tags.CostCenter == ""

    result := {
        "policy_id": "CUSTOM-001",
        "policy_name": "CostCenter Tag Required",
        "msg": sprintf("S3 bucket '%s' must have CostCenter tag", [input.resource.address]),
        "severity": "medium",
        "remediation": "Add CostCenter tag for billing allocation"
    }
}
```

Use custom policies:

```bash
cloudrift scan --service=s3 --policy-dir=./my-policies
```

### Policy Input Schema

Policies receive this input structure:

```json
{
  "resource": {
    "type": "aws_s3_bucket",
    "address": "aws_s3_bucket.example",
    "planned": {
      "bucket": "my-bucket",
      "tags": {"Environment": "prod"},
      "versioning_enabled": true,
      "encryption_algorithm": "AES256",
      "public_access_block": {
        "block_public_acls": true,
        "block_public_policy": true,
        "ignore_public_acls": true,
        "restrict_public_buckets": true
      }
    },
    "live": {
      "bucket": "my-bucket",
      "versioning_enabled": false
    },
    "drift": {
      "has_drift": true,
      "missing": false
    }
  }
}
```

### CI/CD Pipeline Gate

Use `--fail-on-violation` to block pipelines on policy violations:

```bash
cloudrift scan --service=s3 --fail-on-violation
# Exit code 2 if violations found
```

---

## CI/CD Integration

### GitHub Actions

```yaml
name: Infrastructure Drift Check

on:
  pull_request:
    branches: [main]
  schedule:
    - cron: '0 8 * * *'  # Daily at 8am

jobs:
  drift-scan:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Setup Terraform
        uses: hashicorp/setup-terraform@v3

      - name: Terraform Init & Plan
        run: |
          terraform init
          terraform plan -out=tfplan.binary
          terraform show -json tfplan.binary > plan.json
        env:
          AWS_ACCESS_KEY_ID: ${{ secrets.AWS_ACCESS_KEY_ID }}
          AWS_SECRET_ACCESS_KEY: ${{ secrets.AWS_SECRET_ACCESS_KEY }}

      - name: Install Cloudrift
        run: go install github.com/inayathulla/cloudrift@latest

      - name: Run Drift Scan
        run: |
          cloudrift scan \
            --config=cloudrift.yml \
            --service=s3 \
            --format=sarif \
            --output=drift-results.sarif \
            --fail-on-violation
        env:
          AWS_ACCESS_KEY_ID: ${{ secrets.AWS_ACCESS_KEY_ID }}
          AWS_SECRET_ACCESS_KEY: ${{ secrets.AWS_SECRET_ACCESS_KEY }}

      - name: Upload SARIF to GitHub Security
        uses: github/codeql-action/upload-sarif@v2
        if: always()
        with:
          sarif_file: drift-results.sarif
```

### GitLab CI

```yaml
stages:
  - validate

drift-scan:
  stage: validate
  image: golang:1.21
  before_script:
    - go install github.com/inayathulla/cloudrift@latest
  script:
    - terraform init
    - terraform plan -out=tfplan.binary
    - terraform show -json tfplan.binary > plan.json
    - cloudrift scan --service=s3 --format=json --output=drift-report.json --fail-on-violation
  artifacts:
    reports:
      sast: drift-report.json
    when: always
  variables:
    AWS_ACCESS_KEY_ID: $AWS_ACCESS_KEY_ID
    AWS_SECRET_ACCESS_KEY: $AWS_SECRET_ACCESS_KEY
```

### Jenkins Pipeline

```groovy
pipeline {
    agent any

    environment {
        AWS_ACCESS_KEY_ID = credentials('aws-access-key')
        AWS_SECRET_ACCESS_KEY = credentials('aws-secret-key')
    }

    stages {
        stage('Terraform Plan') {
            steps {
                sh 'terraform init'
                sh 'terraform plan -out=tfplan.binary'
                sh 'terraform show -json tfplan.binary > plan.json'
            }
        }

        stage('Drift Scan') {
            steps {
                sh 'go install github.com/inayathulla/cloudrift@latest'
                sh 'cloudrift scan --service=s3 --format=json --fail-on-violation'
            }
        }
    }
}
```

---

## Examples

### Basic S3 Scan

```bash
# Create config
cat > cloudrift.yml << EOF
aws_profile: default
region: us-east-1
plan_path: ./plan.json
EOF

# Generate Terraform plan
terraform plan -out=tfplan.binary
terraform show -json tfplan.binary > plan.json

# Run scan
cloudrift scan --service=s3
```

### EC2 Scan with JSON Output

```bash
cloudrift scan --service=ec2 --format=json --output=ec2-drift.json
```

### Policy-Only Scan (Skip Drift Detection)

```bash
# Just evaluate policies against planned resources
cloudrift scan --service=s3 --skip-policies=false
```

### Drift-Only Scan (Skip Policies)

```bash
cloudrift scan --service=s3 --skip-policies
```

### Custom Policies with Strict Mode

```bash
cloudrift scan \
  --service=s3 \
  --policy-dir=./company-policies \
  --fail-on-violation
```

### ASCII Output (No Emojis)

```bash
cloudrift scan --service=s3 --no-emoji
```

---

## Troubleshooting

### Common Issues

#### "Failed to load AWS config"
- Verify AWS credentials are configured: `aws sts get-caller-identity`
- Check `aws_profile` in cloudrift.yml matches a profile in `~/.aws/credentials`

#### "Plan type mismatch"
- Ensure plan.json was generated with `terraform show -json`
- Verify the plan contains `resource_changes` array

#### "No drift detected" but expecting drift
- Cloudrift only compares resources in the plan against live state
- Resources not in the plan won't be scanned
- Verify the correct plan file is being used

#### "Policy evaluation failed"
- Check policy syntax with `opa check your-policy.rego`
- Ensure policies use correct package name: `package cloudrift.*`

### Debug Mode

For verbose output, check the scan timing messages:
```
✔️  AWS config loaded in 45ms
✔️  Credentials validated in 234ms
🔐 Connected as: arn:aws:iam::123456789012:user/dev
📄 Plan loaded from json in 12ms
✔️  Live S3 state fetched in 1.234s
```

### Getting Help

- GitHub Issues: https://github.com/inayathulla/cloudrift/issues
- Email: inayathulla2020@gmail.com
