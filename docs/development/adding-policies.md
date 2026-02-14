# Adding a New Policy

This guide explains how to write and register a new OPA policy in Cloudrift.

## Policy Structure

Each `.rego` file contains one or more policy rules. Every rule returns a result object with required metadata fields.

### Template

Create a new `.rego` file in the appropriate category directory:

```
internal/policy/policies/
├── security/     ← Encryption, access control, network, IAM, audit
├── tagging/      ← Resource tagging governance
└── cost/         ← Instance sizing, generation optimization
```

### Example: New S3 Policy

Create `internal/policy/policies/security/s3_lifecycle.rego`:

```rego
package cloudrift.security.s3_lifecycle

# S3 Lifecycle Policy Required
deny[result] {
    input.resource.type == "aws_s3_bucket"
    planned := input.resource.planned

    # Check if lifecycle rules are missing or empty
    not planned.lifecycle_rules

    result := {
        "policy_id": "S3-010",
        "policy_name": "S3 Lifecycle Policy Required",
        "msg": sprintf("S3 bucket '%s' does not have lifecycle rules configured", [input.resource.address]),
        "severity": "medium",
        "remediation": "Add aws_s3_bucket_lifecycle_configuration with appropriate transition and expiration rules",
        "category": "security",
        "frameworks": ["iso_27001", "soc2"],
    }
}
```

---

## Required Fields

Every policy result **must** include:

| Field | Type | Required | Example |
|-------|------|----------|---------|
| `policy_id` | string | yes | `"S3-010"` |
| `policy_name` | string | yes | `"S3 Lifecycle Policy Required"` |
| `msg` | string | yes | `"S3 bucket 'my-bucket' does not have..."` |
| `severity` | string | yes | `"critical"`, `"high"`, `"medium"`, `"low"` |
| `remediation` | string | recommended | Fix guidance |
| `category` | string | recommended | `"security"`, `"tagging"`, `"cost"` |
| `frameworks` | array | recommended | `["hipaa", "pci_dss", "iso_27001", "gdpr", "soc2"]` |

---

## Rule Types

### `deny` Rules (Violations)

Blocking findings that count as policy failures:

```rego
deny[result] {
    # condition
    result := { ... }
}
```

### `warn` Rules (Warnings)

Advisory findings that don't affect compliance scoring:

```rego
warn[result] {
    # condition
    result := { ... }
}
```

---

## Available Frameworks

| Key | Framework |
|-----|-----------|
| `hipaa` | HIPAA |
| `gdpr` | GDPR |
| `iso_27001` | ISO 27001 |
| `pci_dss` | PCI DSS |
| `soc2` | SOC 2 Type II |

---

## Policy ID Conventions

| Prefix | Service |
|--------|---------|
| `S3-` | S3 buckets |
| `EC2-` | EC2 instances |
| `SG-` | Security groups |
| `RDS-` | RDS databases |
| `IAM-` | IAM policies and roles |
| `CT-` | CloudTrail |
| `KMS-` | KMS keys |
| `ELB-` | Load balancers |
| `EBS-` | EBS volumes |
| `LAMBDA-` | Lambda functions |
| `LOG-` | CloudWatch logs |
| `VPC-` | VPC resources |
| `SECRET-` | Secrets Manager |
| `TAG-` | Tagging rules |
| `COST-` | Cost optimization |

---

## Input Structure

Policies receive this input structure:

```json
{
  "resource": {
    "type": "aws_s3_bucket",
    "address": "aws_s3_bucket.my_bucket",
    "planned": {
      "bucket": "my-bucket",
      "acl": "private",
      "tags": { "Environment": "prod" },
      "versioning_enabled": true,
      "encryption_algorithm": "AES256"
    },
    "drift": {
      "has_drift": false,
      "missing": false
    }
  }
}
```

---

## Testing Your Policy

### With Custom Policy Directory

```bash
cloudrift scan --service=s3 --policy-dir=./my-policies
```

### Verify Registration

After adding a built-in policy, verify the registry picks it up:

```bash
go test ./tests/internal/policy/... -v -run TestLoadBuiltinRegistry
```

The dynamic registry automatically detects new `.rego` files and updates policy counts.

---

## Multi-Rule Policies

A single policy ID can have multiple `deny` rules (e.g., VPC-001 checks both ingress and egress). The registry deduplicates by policy ID — the policy is only counted once in compliance totals.

```rego
# Rule 1: Check ingress
deny[result] {
    # ...
    result := { "policy_id": "VPC-001", ... }
}

# Rule 2: Check egress
deny[result] {
    # ...
    result := { "policy_id": "VPC-001", ... }
}
```
