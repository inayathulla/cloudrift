# KMS Policies

2 policies covering key management.

| ID | Name | Severity | Frameworks |
|----|------|----------|------------|
| [KMS-001](#kms-001) | KMS Key Rotation Enabled | <span class="severity-high">HIGH</span> | HIPAA, PCI DSS, ISO 27001, SOC 2 |
| [KMS-002](#kms-002) | KMS Key Deletion Window | <span class="severity-medium">MEDIUM</span> | ISO 27001, SOC 2 |

---

## KMS-001

**KMS Key Rotation Enabled** | <span class="severity-high">HIGH</span>

**Frameworks:** HIPAA, PCI DSS, ISO 27001, SOC 2

KMS key must have automatic key rotation enabled. Automatic annual rotation of KMS keys reduces the risk of key compromise by limiting the amount of data encrypted under a single key version and satisfying regulatory requirements for cryptographic key lifecycle management.

**Remediation:**

```hcl
resource "aws_kms_key" "example" {
  description         = "Example KMS key"
  enable_key_rotation = true
}
```

**Resource type:** `aws_kms_key`

---

## KMS-002

**KMS Key Deletion Window** | <span class="severity-medium">MEDIUM</span>

**Frameworks:** ISO 27001, SOC 2

KMS key has deletion window less than 14 days. A short deletion window increases the risk of accidental permanent key loss. Setting a minimum of 14 days provides adequate time to detect and cancel unintended key deletions before encrypted data becomes permanently inaccessible.

**Remediation:**

```hcl
resource "aws_kms_key" "example" {
  description             = "Example KMS key"
  deletion_window_in_days = 14
}
```

**Resource type:** `aws_kms_key`
