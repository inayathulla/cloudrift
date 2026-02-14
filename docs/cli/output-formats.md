# Output Formats

Cloudrift supports three output formats: Console, JSON, and SARIF.

## Console (Default)

The default format produces colorized terminal output with emojis, drift details, policy violations, and compliance scoring.

```bash
cloudrift scan --service=s3
```

### Sample Output

```
🚀 Starting Cloudrift scan...
✔️  AWS config loaded in 45ms
✔️  Credentials valided in 120ms
🔐 Connected as: arn:aws:iam::123456789012:root (123456789012) [us-east-1] in 89ms
📄 Plan loaded from json in 5ms
✔️  Live S3 state fetched in 234ms
✔️  Drift detection completed
✔️  Evaluated 49 policies in 23ms
⚠️  Found 2 policy violations
✔️  Scan completed in 516ms!

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
              POLICY EVALUATION
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

❌ VIOLATIONS (2)

  [high] S3-001
  📍 Resource: aws_s3_bucket.data
  💬 S3 bucket must have server-side encryption enabled
  🔧 Add server_side_encryption_configuration block

  [medium] S3-009
  📍 Resource: aws_s3_bucket.data
  💬 S3 bucket does not have versioning enabled
  🔧 Enable versioning in aws_s3_bucket_versioning resource

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
            COMPLIANCE SUMMARY
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  Overall: 95.9% (47/49 policies passing)

  Categories:
    cost         100.0% (3/3)
    security     95.2% (40/42)
    tagging      100.0% (4/4)

  Frameworks:
    gdpr         94.4% (17/18)
    hipaa        96.2% (25/26)
    iso_27001    97.4% (38/39)
    pci_dss      97.1% (33/34)
    soc2         97.5% (39/40)
```

!!! tip "ASCII mode"
    Use `--no-emoji` for CI/CD environments that don't render Unicode emojis.

---

## JSON

Machine-readable JSON output with full drift details, policy results, and compliance scoring.

```bash
cloudrift scan --service=s3 --format=json
```

### Sample Output

```json
{
  "service": "s3",
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
        "versioning_enabled": ["true", "false"]
      },
      "severity": "warning"
    }
  ],
  "policy_result": {
    "violations": [
      {
        "policy_id": "S3-001",
        "policy_name": "S3 Encryption Required",
        "message": "S3 bucket must have server-side encryption enabled",
        "severity": "high",
        "resource_type": "aws_s3_bucket",
        "resource_address": "aws_s3_bucket.data",
        "remediation": "Add server_side_encryption_configuration block",
        "category": "security",
        "frameworks": ["hipaa", "pci_dss", "iso_27001", "gdpr", "soc2"]
      }
    ],
    "warnings": [],
    "passed": 48,
    "failed": 1,
    "compliance": {
      "overall_percentage": 97.96,
      "total_policies": 49,
      "passing_policies": 48,
      "failing_policies": 1,
      "categories": {
        "security": { "percentage": 97.62, "passed": 41, "failed": 1, "total": 42 },
        "tagging": { "percentage": 100, "passed": 4, "failed": 0, "total": 4 },
        "cost": { "percentage": 100, "passed": 3, "failed": 0, "total": 3 }
      },
      "frameworks": {
        "hipaa": { "percentage": 96.15, "passed": 25, "failed": 1, "total": 26 },
        "pci_dss": { "percentage": 97.06, "passed": 33, "failed": 1, "total": 34 },
        "iso_27001": { "percentage": 97.44, "passed": 38, "failed": 1, "total": 39 },
        "gdpr": { "percentage": 94.44, "passed": 17, "failed": 1, "total": 18 },
        "soc2": { "percentage": 97.5, "passed": 39, "failed": 1, "total": 40 }
      },
      "active_frameworks": ["hipaa", "soc2"]
    }
  },
  "scan_duration_ms": 516,
  "timestamp": "2024-02-14T10:30:00Z"
}
```

!!! note "`active_frameworks`"
    The `active_frameworks` field only appears when `--frameworks` is set. It tells downstream tools which frameworks were selected.

---

## SARIF

Static Analysis Results Interchange Format for integration with GitHub Code Scanning, GitLab SAST, and other tools.

```bash
cloudrift scan --service=s3 --format=sarif --output=results.sarif
```

SARIF output follows the [SARIF 2.1.0 specification](https://docs.oasis-open.org/sarif/sarif/v2.1.0/sarif-v2.1.0.html) and includes:

- **Rules** — Drift detection rules (DRIFT001, DRIFT002, DRIFT003)
- **Results** — Individual drift findings with severity mapping
- **Tool information** — Cloudrift version and description

### GitHub Integration

Upload SARIF results to GitHub's Security tab:

```yaml
- uses: github/codeql-action/upload-sarif@v3
  with:
    sarif_file: results.sarif
```

---

## Writing to Files

Use `--output` to write to a file instead of stdout:

```bash
# JSON report
cloudrift scan --service=s3 --format=json --output=report.json

# SARIF for CI
cloudrift scan --service=s3 --format=sarif --output=drift.sarif
```

When `--output` is used with console format, the colorized output is still written to the terminal while the structured output goes to the file.
