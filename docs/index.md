# Cloudrift

**Pre-apply Terraform Drift Detection & Compliance CLI**

Cloudrift compares your Terraform plan JSON against live AWS infrastructure to detect configuration drift **before** `terraform apply`. It evaluates 49 built-in OPA security policies and scores compliance across 5 industry frameworks — all from a single CLI command.

---

## Key Features

<div class="grid cards" markdown>

-   :material-magnify-scan:{ .lg .middle } **Drift Detection**

    ---

    Compare live AWS resources (S3, EC2, IAM) against Terraform plan files. See attribute-level diffs with colorized console output.

-   :material-shield-check:{ .lg .middle } **49 Security Policies**

    ---

    OPA-powered policy engine covering S3, EC2, RDS, IAM, Security Groups, CloudTrail, KMS, Lambda, ELB, EBS, VPC, and Secrets Manager.

-   :material-clipboard-check:{ .lg .middle } **5 Compliance Frameworks**

    ---

    HIPAA, GDPR, ISO 27001, PCI DSS, and SOC 2 compliance scoring with per-framework breakdowns.

-   :material-filter:{ .lg .middle } **Framework Filtering**

    ---

    Focus on the frameworks that matter with `--frameworks=hipaa,soc2`. Only relevant policies are evaluated and scored.

-   :material-docker:{ .lg .middle } **Docker & CI/CD**

    ---

    Run as a Docker container. Integrate into GitHub Actions or GitLab CI with `--fail-on-violation` and SARIF output.

-   :material-code-json:{ .lg .middle } **3 Output Formats**

    ---

    Console (colorized), JSON (machine-readable), and SARIF (GitHub Security tab integration).

</div>

---

## Quick Start

=== "Go Install"

    ```bash
    go install github.com/inayathulla/cloudrift@latest
    ```

=== "Docker"

    ```bash
    docker pull inayathulla/cloudrift:latest
    docker run -v ~/.aws:/root/.aws:ro \
      -v $(pwd):/work \
      inayathulla/cloudrift:latest scan \
      --config=/work/cloudrift.yml --service=s3
    ```

=== "Build from Source"

    ```bash
    git clone https://github.com/inayathulla/cloudrift.git
    cd cloudrift
    go build -o cloudrift main.go
    ```

### Run Your First Scan

```bash
# 1. Generate a Terraform plan
terraform plan -out=tfplan
terraform show -json tfplan > plan.json

# 2. Create a config file
cat > cloudrift.yml <<EOF
aws_profile: default
region: us-east-1
plan_path: ./plan.json
EOF

# 3. Scan for drift and policy violations
cloudrift scan --service=s3
```

[Get Started :material-arrow-right:](getting-started/installation.md){ .md-button .md-button--primary }
[View on GitHub :material-github:](https://github.com/inayathulla/cloudrift){ .md-button }

---

## Sample Output

```
🚀 Starting Cloudrift scan...
🔐 Connected as: arn:aws:iam::123456789012:root (123456789012) [us-east-1]
✔️  Evaluated 49 policies in 23ms
⚠️  Found 2 policy violations

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

---

## Why Cloudrift?

| Feature | Cloudrift | Terraform Cloud | Checkov | driftctl |
|---------|-----------|-----------------|---------|----------|
| **Pre-apply drift detection** | :material-check: | :material-close: | :material-close: | :material-close: |
| **Live AWS comparison** | :material-check: | :material-check: | :material-close: | :material-check: |
| **OPA policy engine** | :material-check: | Sentinel | :material-check: | :material-close: |
| **Compliance scoring** | :material-check: | :material-close: | :material-check: | :material-close: |
| **Framework filtering** | :material-check: | :material-close: | :material-close: | :material-close: |
| **SARIF output** | :material-check: | :material-close: | :material-check: | :material-close: |
| **Free & open source** | :material-check: | Paid | :material-check: | :material-check: |
