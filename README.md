<p align="center">
  <h1 align="center">Cloudrift</h1>
  <p align="center">
    <strong>Pre-apply infrastructure governance for Terraform</strong>
  </p>
  <p align="center">
    Validate infrastructure changes against policies and live AWS state — before you apply
  </p>
</p>

<p align="center">
  <a href="LICENSE"><img src="https://img.shields.io/badge/License-Apache_2.0-blue.svg" alt="License"></a>
  <a href="https://goreportcard.com/report/github.com/inayathulla/cloudrift"><img src="https://goreportcard.com/badge/github.com/inayathulla/cloudrift" alt="Go Report Card"></a>
  <a href="https://github.com/inayathulla/cloudrift/actions/workflows/tests.yml"><img src="https://github.com/inayathulla/cloudrift/actions/workflows/tests.yml/badge.svg" alt="Go Test"></a>
  <a href="https://hub.docker.com/r/inayathulla/cloudrift"><img src="https://img.shields.io/docker/pulls/inayathulla/cloudrift" alt="Docker Pulls"></a>
  <a href="https://github.com/inayathulla/cloudrift/stargazers"><img src="https://img.shields.io/github/stars/inayathulla/cloudrift?style=social" alt="GitHub stars"></a>
</p>

<p align="center">
  <a href="https://tldrsec.com/p/tldr-sec-287"><img src="https://img.shields.io/badge/Featured%20in-TLDR%20Sec%20%23287-blueviolet?logo=security&style=flat-square" alt="Featured in TLDR Sec"></a>
</p>

---

**Cloudrift** is an open-source infrastructure governance tool that validates your Terraform plans against live AWS state and security policies — catching misconfigurations **before** `terraform apply`, not after.

```
┌─────────────────────────────────────────────────────────────┐
│                  CLOUDRIFT UNIQUE POSITION                   │
│                                                              │
│   Terraform Plan  ──┐                                        │
│                     ├──▶  Policy Engine  ──▶  ALLOW/BLOCK   │
│   Live AWS State  ──┘        (OPA)                           │
│                                                              │
│   Competitors check EITHER plan OR live state                │
│   Cloudrift checks BOTH — catches drift AND policy violations│
└─────────────────────────────────────────────────────────────┘
```

![Demo](assets/s3_scanning.gif)

## Table of Contents

- [Why Cloudrift?](#why-cloudrift)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Output Formats](#output-formats)
- [Configuration](#configuration)
- [How It Works](#how-it-works)
- [Roadmap](#roadmap)
- [Contributing](#contributing)
- [License](#license)

## Why Cloudrift?

| Feature | Cloudrift | Terraform Cloud | Checkov | driftctl |
|---------|-----------|-----------------|---------|----------|
| **Pre-apply validation** | ✅ | ❌ | ✅ | ❌ |
| **Live state comparison** | ✅ | ❌ | ❌ | ✅ |
| **Policy engine (OPA)** | 🚧 Coming | Sentinel | ✅ | ❌ |
| **SARIF output** | ✅ | ❌ | ✅ | ❌ |
| **Open source** | ✅ | ❌ | ✅ | ✅ |
| **Self-hosted** | ✅ | ❌ | ✅ | ✅ |

**Key differentiator:** Cloudrift compares your Terraform plan against **live AWS state** — catching drift that would be silently overwritten by `terraform apply`.

## Installation

### Via Go

```bash
go install github.com/inayathulla/cloudrift@latest
```

### Via Docker

```bash
docker pull inayathulla/cloudrift
```

### Via Homebrew (coming soon)

```bash
brew install cloudrift
```

## Quick Start

### 1. Generate a Terraform plan

```bash
cd your-terraform-project
terraform init
terraform plan -out=tfplan.binary
terraform show -json tfplan.binary > plan.json
```

### 2. Create a configuration file

Create `cloudrift.yml`:

```yaml
aws_profile: default
region: us-east-1
plan_path: ./plan.json
```

### 3. Run Cloudrift

```bash
# Console output (default)
cloudrift scan --service=s3

# JSON output for CI/CD
cloudrift scan --service=s3 --format=json

# SARIF output for GitHub Security tab
cloudrift scan --service=s3 --format=sarif --output=drift-report.sarif
```

## Output Formats

Cloudrift supports multiple output formats for different use cases:

### Console (default)

Colorized, human-readable output for terminal use:

```bash
cloudrift scan --service=s3
```

### JSON

Machine-readable format for CI/CD pipelines and scripting:

```bash
cloudrift scan --service=s3 --format=json
```

```json
{
  "service": "S3",
  "account_id": "123456789012",
  "total_resources": 5,
  "drift_count": 2,
  "drifts": [
    {
      "resource_name": "my-bucket",
      "resource_type": "aws_s3_bucket",
      "diffs": {
        "versioning_enabled": ["true", "false"]
      }
    }
  ]
}
```

### SARIF

[Static Analysis Results Interchange Format](https://sarifweb.azurewebsites.net/) for GitHub/GitLab Security integration:

```bash
cloudrift scan --service=s3 --format=sarif --output=drift-report.sarif
```

Upload to GitHub Code Scanning:

```yaml
# .github/workflows/drift-scan.yml
- name: Run Cloudrift
  run: cloudrift scan --service=s3 --format=sarif --output=results.sarif

- name: Upload SARIF
  uses: github/codeql-action/upload-sarif@v2
  with:
    sarif_file: results.sarif
```

## Configuration

| Field | Description | Required |
|-------|-------------|----------|
| `aws_profile` | AWS credentials profile name | Yes |
| `region` | AWS region to scan | Yes |
| `plan_path` | Path to Terraform plan JSON | Yes |

## How It Works

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│  Terraform Plan │────▶│    Cloudrift    │◀────│   AWS Live API  │
│     (JSON)      │     │     Engine      │     │   (S3, EC2...)  │
└─────────────────┘     └────────┬────────┘     └─────────────────┘
                                 │
                                 ▼
                        ┌─────────────────┐
                        │  Drift Report   │
                        │ JSON/SARIF/CLI  │
                        └─────────────────┘
```

1. **Parse** — Extracts planned resource configurations from Terraform plan JSON
2. **Fetch** — Queries AWS APIs in parallel to get current live state
3. **Compare** — Detects attribute-level differences between plan and reality
4. **Report** — Outputs results in your preferred format

### Supported Resources

| Resource | Status |
|----------|--------|
| S3 Buckets | ✅ Available |
| EC2 Instances | 🚧 Coming |
| IAM Roles | 🚧 Planned |
| Security Groups | 🚧 Planned |
| RDS Instances | 🚧 Planned |

## Roadmap

Cloudrift is evolving into a full infrastructure governance platform:

### Phase 1: Multi-Service Foundation ✅
- [x] Generic detector interface
- [x] JSON output format
- [x] SARIF output format
- [ ] EC2 drift detection
- [ ] IAM drift detection

### Phase 2: Policy Engine
- [ ] OPA (Open Policy Agent) integration
- [ ] Built-in security policies
- [ ] Custom policy support
- [ ] `--fail-on-violation` flag

### Phase 3: Compliance Packs
- [ ] CIS AWS Foundations
- [ ] SOC 2 Type II
- [ ] HIPAA
- [ ] PCI-DSS

### Phase 4: Enterprise Features
- [ ] Multi-account scanning
- [ ] Web dashboard
- [ ] Scheduled scans
- [ ] Slack/PagerDuty alerts

## CI/CD Integration

### GitHub Actions

```yaml
name: Drift Detection
on: [pull_request]

jobs:
  drift-scan:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Setup Terraform
        uses: hashicorp/setup-terraform@v3

      - name: Terraform Plan
        run: |
          terraform init
          terraform plan -out=tfplan.binary
          terraform show -json tfplan.binary > plan.json

      - name: Install Cloudrift
        run: go install github.com/inayathulla/cloudrift@latest

      - name: Run Drift Scan
        run: cloudrift scan --service=s3 --format=sarif --output=results.sarif
        env:
          AWS_ACCESS_KEY_ID: ${{ secrets.AWS_ACCESS_KEY_ID }}
          AWS_SECRET_ACCESS_KEY: ${{ secrets.AWS_SECRET_ACCESS_KEY }}

      - name: Upload SARIF
        uses: github/codeql-action/upload-sarif@v2
        with:
          sarif_file: results.sarif
```

### GitLab CI

```yaml
drift-scan:
  image: golang:1.21
  script:
    - go install github.com/inayathulla/cloudrift@latest
    - cloudrift scan --service=s3 --format=json
  artifacts:
    reports:
      sast: drift-report.json
```

## Project Structure

```
cloudrift/
├── cmd/                    # CLI commands
├── internal/
│   ├── aws/                # AWS API integrations
│   ├── detector/           # Drift detection logic
│   │   ├── interface.go    # Generic detector interface
│   │   ├── registry.go     # Service registry
│   │   └── s3.go           # S3 detector
│   ├── output/             # Output formatters
│   │   ├── json.go         # JSON formatter
│   │   ├── sarif.go        # SARIF formatter
│   │   └── console.go      # Console formatter
│   ├── models/             # Data structures
│   └── parser/             # Terraform plan parser
├── policies/               # OPA policies (coming soon)
└── tests/                  # Unit tests
```

## Contributing

Contributions are welcome! See our contributing guidelines.

```bash
# Clone
git clone https://github.com/inayathulla/cloudrift.git
cd cloudrift

# Build
go build -o cloudrift main.go

# Test
go test ./...

# Format
go fmt ./...
```

## Connect

- **Issues & Features:** [GitHub Issues](https://github.com/inayathulla/cloudrift/issues)
- **Email:** [inayathulla2020@gmail.com](mailto:inayathulla2020@gmail.com)
- **LinkedIn:** [Inayathulla Khan Lavani](https://www.linkedin.com/in/inayathullakhan)

---

## License

[Apache License 2.0](LICENSE)

---

<p align="center">
  <sub>Built for DevOps teams who believe in shift-left infrastructure governance</sub>
</p>
