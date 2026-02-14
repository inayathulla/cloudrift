# Secrets Manager Policies

2 policies covering secret management.

| ID | Name | Severity | Frameworks |
|----|------|----------|------------|
| [SECRET-001](#secret-001) | Secrets Manager KMS Encryption | <span class="severity-medium">MEDIUM</span> | HIPAA, PCI DSS, GDPR, SOC 2 |
| [SECRET-002](#secret-002) | Secrets Rotation Enabled | <span class="severity-medium">MEDIUM</span> | PCI DSS, ISO 27001, SOC 2 |

---

## SECRET-001

**Secrets Manager KMS Encryption** | <span class="severity-medium">MEDIUM</span>

**Frameworks:** HIPAA, PCI DSS, GDPR, SOC 2

Secret should use a customer-managed KMS key. By default, Secrets Manager encrypts secrets with an AWS-managed key. Using a customer-managed KMS key provides fine-grained access control, key rotation policies, and detailed audit logging through CloudTrail.

**Remediation:**

```hcl
resource "aws_secretsmanager_secret" "example" {
  name       = "example-secret"
  kms_key_id = aws_kms_key.secrets.arn
}
```

**Resource type:** `aws_secretsmanager_secret`

---

## SECRET-002

**Secrets Rotation Enabled** | <span class="severity-medium">MEDIUM</span>

**Frameworks:** PCI DSS, ISO 27001, SOC 2

Secret does not have automatic rotation configured. Automatic rotation reduces the risk of compromised credentials by periodically replacing secret values without manual intervention, meeting compliance requirements for credential lifecycle management.

**Remediation:**

```hcl
resource "aws_secretsmanager_secret" "example" {
  name = "example-secret"
}

resource "aws_secretsmanager_secret_rotation" "example" {
  secret_id           = aws_secretsmanager_secret.example.id
  rotation_lambda_arn = aws_lambda_function.rotation.arn

  rotation_rules {
    automatically_after_days = 30
  }
}
```

**Resource type:** `aws_secretsmanager_secret`
