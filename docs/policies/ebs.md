# EBS Policies

2 policies covering storage encryption.

| ID | Name | Severity | Frameworks |
|----|------|----------|------------|
| [EBS-001](#ebs-001) | EBS Volume Encryption | <span class="severity-high">HIGH</span> | HIPAA, PCI DSS, ISO 27001, GDPR, SOC 2 |
| [EBS-002](#ebs-002) | EBS Snapshot Encryption | <span class="severity-high">HIGH</span> | HIPAA, PCI DSS, ISO 27001, GDPR |

---

## EBS-001

**EBS Volume Encryption** | <span class="severity-high">HIGH</span>

**Frameworks:** HIPAA, PCI DSS, ISO 27001, GDPR, SOC 2

EBS volume must have encryption enabled. Unencrypted EBS volumes expose data at rest to unauthorized access if the underlying storage media is compromised. Enabling encryption ensures that data, snapshots, and disk I/O are all protected using AES-256 encryption.

**Remediation:**

```hcl
resource "aws_ebs_volume" "example" {
  availability_zone = "us-east-1a"
  size              = 100
  encrypted         = true
}
```

**Resource type:** `aws_ebs_volume`

---

## EBS-002

**EBS Snapshot Encryption** | <span class="severity-high">HIGH</span>

**Frameworks:** HIPAA, PCI DSS, ISO 27001, GDPR

EBS snapshot copy must have encryption enabled. When copying snapshots across regions or accounts, encryption must be explicitly enabled on the copy to ensure data remains protected in transit and at rest at the destination.

**Remediation:**

```hcl
resource "aws_ebs_snapshot_copy" "example" {
  source_snapshot_id = aws_ebs_snapshot.source.id
  source_region      = "us-east-1"
  encrypted          = true
}
```

**Resource type:** `aws_ebs_snapshot_copy`
