# Compliance Frameworks

Cloudrift maps its 49 built-in policies to 5 industry compliance frameworks, providing automated compliance scoring. Policy counts are computed dynamically from the `.rego` files — never hardcoded.

## Frameworks

| Framework | Policies | Description |
|-----------|----------|-------------|
| **SOC 2 Type II** | 40 | Trust services criteria — security, availability, confidentiality |
| **ISO 27001** | 39 | Information security management system controls |
| **PCI DSS** | 34 | Payment card industry data security standard |
| **HIPAA** | 26 | Health data privacy and security rules |
| **GDPR** | 18 | EU data protection and privacy regulation |

### Policy Categories

| Category | Policies | Description |
|----------|----------|-------------|
| **Security** | 42 | Encryption, access control, network, IAM, audit logging |
| **Tagging** | 4 | Resource tagging for cost allocation and governance |
| **Cost** | 3 | Instance sizing and generation optimization |

---

## Scoring Calculation

Each framework's compliance score is calculated as:

```
Score = (Passing Policies / Total Mapped Policies) x 100%
```

A policy **passes** if zero violations are found for it across all scanned resources. A single violation causes the policy to fail.

### Score Thresholds

| Score | Color | Rating |
|-------|-------|--------|
| 100% | :material-check-circle:{ style="color: #2E7D32" } Green | Full compliance |
| 80-99% | :material-alert-circle:{ style="color: #F9A825" } Yellow | Needs attention |
| < 80% | :material-close-circle:{ style="color: #D32F2F" } Red | Critical gaps |

---

## Framework Filtering

Use `--frameworks` to scope evaluation to specific frameworks:

```bash
# HIPAA-only
cloudrift scan --service=s3 --frameworks=hipaa

# Multiple frameworks
cloudrift scan --service=s3 --frameworks=hipaa,gdpr

# With JSON output
cloudrift scan --service=s3 --format=json --frameworks=soc2,pci_dss
```

When `--frameworks` is set:

- Only violations from policies mapped to selected frameworks are shown
- Compliance scoring uses filtered totals
- Console output shows the active filter in the header
- JSON output includes `active_frameworks` array

### Console Output with Filter

```
🔐 Filtering by frameworks: hipaa, gdpr

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
      COMPLIANCE SUMMARY (HIPAA, GDPR)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

### JSON Output with Filter

```json
{
  "compliance": {
    "overall_percentage": 96.43,
    "total_policies": 28,
    "active_frameworks": ["hipaa", "gdpr"]
  }
}
```

---

## Framework Validation

Unknown framework names are rejected with a list of valid options:

```
❌ Unknown framework: hipa
  Available frameworks: gdpr, hipaa, iso_27001, pci_dss, soc2
```

---

## Framework Details

### HIPAA (26 policies)

Health Insurance Portability and Accountability Act — protects sensitive patient health information.

**Focus areas:** Encryption at rest, encryption in transit, access controls, audit logging, network isolation.

**Key policies:** S3-001 (encryption), EC2-002 (root volume encryption), RDS-001/RDS-002 (database security), IAM-001 (least privilege), CT-001 (audit logging).

### GDPR (18 policies)

General Data Protection Regulation — EU regulation on data privacy and protection.

**Focus areas:** Data encryption, access controls, audit trails, data protection by design.

**Key policies:** S3-001 through S3-008 (data storage security), EBS-001/EBS-002 (volume encryption), LOG-001/LOG-002 (audit logging).

### ISO 27001 (39 policies)

International standard for information security management systems.

**Focus areas:** Comprehensive security controls, risk management, access management, operations security.

**Key policies:** Broadest coverage across all AWS services — encryption, access control, network security, key management, and audit logging.

### PCI DSS (34 policies)

Payment Card Industry Data Security Standard — protects cardholder data.

**Focus areas:** Network security, strong access control, encryption, monitoring, vulnerability management.

**Key policies:** SG-001 through SG-004 (network security), IAM-001 through IAM-003 (access control), CT-001 through CT-003 (monitoring).

### SOC 2 Type II (40 policies)

Service Organization Control 2 — covers security, availability, processing integrity, confidentiality, and privacy.

**Focus areas:** The most comprehensive framework, covering all aspects of security controls.

**Key policies:** Nearly all 49 policies — the broadest compliance framework supported.
