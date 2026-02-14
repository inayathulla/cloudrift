# CloudTrail Policies

3 policies covering audit trail security.

| ID | Name | Severity | Frameworks |
|----|------|----------|------------|
| [CT-001](#ct-001) | CloudTrail KMS Encryption | <span class="severity-high">HIGH</span> | HIPAA, PCI DSS, ISO 27001, GDPR, SOC 2 |
| [CT-002](#ct-002) | CloudTrail Log File Validation | <span class="severity-medium">MEDIUM</span> | PCI DSS, ISO 27001, SOC 2 |
| [CT-003](#ct-003) | CloudTrail Multi-Region | <span class="severity-medium">MEDIUM</span> | HIPAA, PCI DSS, ISO 27001, SOC 2 |

---

## CT-001

**CloudTrail KMS Encryption** | <span class="severity-high">HIGH</span>

**Frameworks:** HIPAA, PCI DSS, ISO 27001, GDPR, SOC 2

CloudTrail must be encrypted with a KMS key. Encrypting CloudTrail logs with a customer-managed KMS key provides additional protection for sensitive audit data and enables fine-grained access control over who can read the log files.

**Remediation:**

```hcl
resource "aws_cloudtrail" "example" {
  name                          = "example-trail"
  s3_bucket_name                = aws_s3_bucket.trail.id
  kms_key_id                    = aws_kms_key.cloudtrail.arn
  enable_log_file_validation    = true
  is_multi_region_trail         = true
}
```

**Resource type:** `aws_cloudtrail`

---

## CT-002

**CloudTrail Log File Validation** | <span class="severity-medium">MEDIUM</span>

**Frameworks:** PCI DSS, ISO 27001, SOC 2

CloudTrail should enable log file validation to detect tampering. Log file validation creates a digitally signed digest file containing a hash of each log file, allowing you to determine whether a log file was modified or deleted after CloudTrail delivered it.

**Remediation:**

```hcl
resource "aws_cloudtrail" "example" {
  name                          = "example-trail"
  s3_bucket_name                = aws_s3_bucket.trail.id
  enable_log_file_validation    = true
}
```

**Resource type:** `aws_cloudtrail`

---

## CT-003

**CloudTrail Multi-Region** | <span class="severity-medium">MEDIUM</span>

**Frameworks:** HIPAA, PCI DSS, ISO 27001, SOC 2

CloudTrail should be configured as multi-region trail. A multi-region trail ensures that API activity across all AWS regions is captured in a single trail, preventing gaps in audit coverage if resources are created in unexpected regions.

**Remediation:**

```hcl
resource "aws_cloudtrail" "example" {
  name                          = "example-trail"
  s3_bucket_name                = aws_s3_bucket.trail.id
  is_multi_region_trail         = true
}
```

**Resource type:** `aws_cloudtrail`
