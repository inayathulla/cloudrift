# IAM Policies

Cloudrift includes **3 built-in policies** for AWS Identity and Access Management, covering least-privilege enforcement, policy management best practices, and trust relationship controls.

## Summary

| ID | Policy Name | Severity | Frameworks |
|----|------------|----------|------------|
| [IAM-001](#iam-001) | No Wildcard IAM Actions | <span class="severity-critical">CRITICAL</span> | HIPAA, PCI DSS, ISO 27001, GDPR, SOC 2 |
| [IAM-002](#iam-002) | No Inline Policies on Users | <span class="severity-medium">MEDIUM</span> | PCI DSS, ISO 27001, SOC 2 |
| [IAM-003](#iam-003) | IAM Role Trust Too Broad | <span class="severity-high">HIGH</span> | PCI DSS, ISO 27001, SOC 2 |

---

## IAM-001

### No Wildcard IAM Actions

<span class="severity-critical">CRITICAL</span>

**Frameworks:** HIPAA, PCI DSS, ISO 27001, GDPR, SOC 2

**Resource type:** `aws_iam_policy`

**Description:**
IAM policy contains a wildcard (`*`) action with `Allow` effect. Wildcard actions grant unrestricted permissions across all AWS services and API calls, violating the principle of least privilege. A compromised principal with `Action: "*"` has full administrative access to the AWS account, enabling data exfiltration, resource destruction, and privilege escalation.

**Remediation:**

Replace wildcard `Action` with specific actions following the least-privilege principle. Identify the exact API calls required by the workload and grant only those permissions.

```hcl
resource "aws_iam_policy" "example" {
  name        = "application-policy"
  description = "Least-privilege policy for application workload"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "AllowS3ReadAccess"
        Effect = "Allow"
        Action = [
          "s3:GetObject",
          "s3:ListBucket"
        ]
        Resource = [
          "arn:aws:s3:::my-app-bucket",
          "arn:aws:s3:::my-app-bucket/*"
        ]
      },
      {
        Sid    = "AllowDynamoDBAccess"
        Effect = "Allow"
        Action = [
          "dynamodb:GetItem",
          "dynamodb:PutItem",
          "dynamodb:Query"
        ]
        Resource = "arn:aws:dynamodb:us-east-1:123456789012:table/my-app-table"
      }
    ]
  })
}
```

---

## IAM-002

### No Inline Policies on Users

<span class="severity-medium">MEDIUM</span>

**Frameworks:** PCI DSS, ISO 27001, SOC 2

**Resource type:** `aws_iam_user_policy`

**Description:**
IAM user policy uses an inline policy. Inline policies are embedded directly in a single IAM user and cannot be reused, versioned, or centrally managed. This makes auditing permissions difficult and increases the risk of policy sprawl. AWS best practice recommends using managed policies attached to groups or roles instead of inline policies on individual users.

**Remediation:**

Convert inline policies to managed policies and attach them to IAM groups or roles using `aws_iam_user_policy_attachment`. Users should inherit permissions through group membership rather than direct policy attachment.

```hcl
# Instead of an inline policy:
#
# resource "aws_iam_user_policy" "inline" {
#   name   = "user-inline-policy"
#   user   = aws_iam_user.example.name
#   policy = jsonencode({ ... })
# }

# Use a managed policy attached via group membership:

resource "aws_iam_policy" "app_read" {
  name        = "app-read-policy"
  description = "Read access for application resources"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "s3:GetObject",
          "s3:ListBucket"
        ]
        Resource = [
          "arn:aws:s3:::my-app-bucket",
          "arn:aws:s3:::my-app-bucket/*"
        ]
      }
    ]
  })
}

resource "aws_iam_group" "developers" {
  name = "developers"
}

resource "aws_iam_group_policy_attachment" "developers_app_read" {
  group      = aws_iam_group.developers.name
  policy_arn = aws_iam_policy.app_read.arn
}

resource "aws_iam_user_group_membership" "example" {
  user   = aws_iam_user.example.name
  groups = [aws_iam_group.developers.name]
}
```

---

## IAM-003

### IAM Role Trust Too Broad

<span class="severity-high">HIGH</span>

**Frameworks:** PCI DSS, ISO 27001, SOC 2

**Resource type:** `aws_iam_role`

**Description:**
IAM role has an overly broad trust policy allowing any principal. A trust policy with `Principal: "*"` or `Principal: {"AWS": "*"}` allows any AWS account, user, or service to assume the role. This effectively makes the role's permissions available to the entire internet (any AWS account) and can lead to unauthorized cross-account access and privilege escalation.

**Remediation:**

Restrict the `Principal` in the trust policy to specific AWS accounts, services, or IAM entities that legitimately need to assume the role.

```hcl
resource "aws_iam_role" "example" {
  name = "application-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "AllowEC2Assume"
        Effect = "Allow"
        Principal = {
          Service = "ec2.amazonaws.com"
        }
        Action = "sts:AssumeRole"
      },
      {
        Sid    = "AllowCrossAccountAccess"
        Effect = "Allow"
        Principal = {
          AWS = "arn:aws:iam::123456789012:root"
        }
        Action = "sts:AssumeRole"
        Condition = {
          StringEquals = {
            "sts:ExternalId" = "unique-external-id"
          }
        }
      }
    ]
  })
}
```
