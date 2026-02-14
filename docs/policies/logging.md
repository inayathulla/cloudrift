# CloudWatch Logs Policies

2 policies covering log management.

| ID | Name | Severity | Frameworks |
|----|------|----------|------------|
| [LOG-001](#log-001) | CloudWatch Log Group Encryption | <span class="severity-medium">MEDIUM</span> | HIPAA, PCI DSS, GDPR, SOC 2 |
| [LOG-002](#log-002) | CloudWatch Log Retention | <span class="severity-medium">MEDIUM</span> | HIPAA, GDPR, SOC 2, ISO 27001 |

---

## LOG-001

**CloudWatch Log Group Encryption** | <span class="severity-medium">MEDIUM</span>

**Frameworks:** HIPAA, PCI DSS, GDPR, SOC 2

CloudWatch Log Group should be encrypted with a KMS key. By default, log data is encrypted at rest using AWS-managed keys, but using a customer-managed KMS key provides additional control over access policies, key rotation, and audit logging of key usage.

**Remediation:**

```hcl
resource "aws_cloudwatch_log_group" "example" {
  name       = "/aws/lambda/example"
  kms_key_id = aws_kms_key.log_encryption.arn
}
```

**Resource type:** `aws_cloudwatch_log_group`

---

## LOG-002

**CloudWatch Log Retention** | <span class="severity-medium">MEDIUM</span>

**Frameworks:** HIPAA, GDPR, SOC 2, ISO 27001

CloudWatch Log Group does not have a retention policy configured or retention is 0. Without a retention policy, logs are retained indefinitely, increasing storage costs and potentially violating data retention regulations that require logs to be purged after a defined period.

**Remediation:**

```hcl
resource "aws_cloudwatch_log_group" "example" {
  name              = "/aws/lambda/example"
  retention_in_days = 90

  # Common values: 1, 3, 5, 7, 14, 30, 60, 90, 120, 150,
  # 180, 365, 400, 545, 731, 1096, 1827, 2192, 2557, 2922, 3653
}
```

**Resource type:** `aws_cloudwatch_log_group`
