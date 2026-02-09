<p align="center">
  <img src="assets/cloudrift-logo.png" alt="Cloudrift Logo" width="120">
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
│                  CLOUDRIFT UNIQUE POSITION                  │
│                                                             │
│   Terraform Plan  ──┐                                       │
│                     ├──▶  Policy Engine  ──▶  ALLOW/BLOCK  │
│   Live AWS State  ──┘        (OPA)                          │
│                                                             │
│   Competitors check EITHER plan OR live state               │
│   Cloudrift checks BOTH — catches drift AND policy violations│
└─────────────────────────────────────────────────────────────┘
```

## Table of Contents

- [Why Cloudrift?](#why-cloudrift)
- [Features](#features)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Usage](#usage)
- [Output Formats](#output-formats)
- [Policy Engine](#policy-engine)
- [CI/CD Integration](#cicd-integration)
- [Desktop Dashboard](#desktop-dashboard)
- [Configuration](#configuration)
- [Project Structure](#project-structure)
- [Roadmap](#roadmap)
- [Contributing](#contributing)
- [Related Projects](#related-projects)
- [License](#license)

## Why Cloudrift?

| Feature | Cloudrift | Terraform Cloud | Checkov | driftctl |
|---------|-----------|-----------------|---------|----------|
| **Pre-apply validation** | ✅ | ❌ | ✅ | ❌ |
| **Live state comparison** | ✅ | ❌ | ❌ | ✅ |
| **Policy engine (OPA)** | ✅ | Sentinel ($$$) | ✅ | ❌ |
| **SARIF output** | ✅ | ❌ | ✅ | ❌ |
| **Open source** | ✅ | ❌ | ✅ | ✅ |
| **Self-hosted** | ✅ | ❌ | ✅ | ✅ |

**Key differentiator:** Cloudrift compares your Terraform plan against **live AWS state** — catching drift that would be silently overwritten by `terraform apply`.

## Features

- **Drift Detection** — Compare Terraform plans against live AWS infrastructure
- **Policy Engine** — 7 built-in OPA security policies + custom policy support
- **Multiple Output Formats** — Console, JSON, SARIF for CI/CD integration
- **Multi-Service Support** — S3 buckets and EC2 instances
- **CI/CD Ready** — GitHub Actions, GitLab CI, Jenkins integration
- **GitHub Security Integration** — SARIF output for Security tab

## Installation

### Via Go

```bash
go install github.com/inayathulla/cloudrift@latest
```

### Via Docker

```bash
docker pull inayathulla/cloudrift
```

### Build from Source

```bash
git clone https://github.com/inayathulla/cloudrift.git
cd cloudrift
go build -o cloudrift .
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
# Scan S3 buckets
cloudrift scan --service=s3

# Scan EC2 instances
cloudrift scan --service=ec2

# Output as JSON
cloudrift scan --service=s3 --format=json

# Fail CI/CD on policy violations
cloudrift scan --service=s3 --fail-on-violation
```

## Usage

```bash
cloudrift scan [flags]
```

### Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--config` | `-c` | `cloudrift.yml` | Path to configuration file |
| `--service` | `-s` | `s3` | AWS service to scan (s3, ec2) |
| `--format` | `-f` | `console` | Output format (console, json, sarif) |
| `--output` | `-o` | stdout | Write output to file |
| `--policy-dir` | `-p` | - | Directory with custom OPA policies |
| `--fail-on-violation` | - | `false` | Exit non-zero on violations |
| `--skip-policies` | - | `false` | Skip policy evaluation |
| `--no-emoji` | - | `false` | Use ASCII instead of emojis |

### Supported Resources

| Resource | Service | Attributes Checked |
|----------|---------|-------------------|
| S3 Buckets | `--service=s3` | ACL, tags, versioning, encryption, logging, public access block, lifecycle rules |
| EC2 Instances | `--service=ec2` | Instance type, AMI, subnet, security groups, tags, EBS optimization, monitoring |

For detailed usage instructions, see [docs/USAGE.md](docs/USAGE.md).

## Output Formats

### Console (default)

```bash
cloudrift scan --service=s3
```

```
🚀 Starting Cloudrift scan...
🔐 Connected as: arn:aws:iam::123456789012:root [us-east-1]
✔️  Evaluated 7 policies in 23ms
⚠️  Found 2 policy violations

⚠️  Drift detected!
🪣 my-bucket
  🔐 Encryption mismatch:
    • expected → "AES256"
    • actual   → ""
```

### JSON

```bash
cloudrift scan --service=s3 --format=json
```

```json
{
  "service": "S3",
  "account_id": "123456789012",
  "drift_count": 1,
  "drifts": [
    {
      "resource_name": "my-bucket",
      "diffs": {
        "encryption_algorithm": ["AES256", ""]
      }
    }
  ]
}
```

### SARIF (GitHub Security)

```bash
cloudrift scan --service=s3 --format=sarif --output=results.sarif
```

Upload to GitHub Code Scanning for Security tab integration.

## Policy Engine

Cloudrift includes 7 built-in OPA security policies:

| Policy | Severity | Description |
|--------|----------|-------------|
| S3-001 | high | S3 buckets must have encryption enabled |
| S3-003 to S3-006 | high | S3 public access block settings |
| S3-007, S3-008 | critical | No public ACLs allowed |
| TAG-001 | medium | Environment tag required |
| TAG-002 to TAG-004 | low | Owner, Project, Name tags recommended |

### Custom Policies

Create custom OPA policies:

```rego
# my-policies/custom.rego
package cloudrift.custom

deny[result] {
    input.resource.type == "aws_s3_bucket"
    not input.resource.planned.tags.CostCenter

    result := {
        "policy_id": "CUSTOM-001",
        "msg": "S3 bucket must have CostCenter tag",
        "severity": "medium"
    }
}
```

```bash
cloudrift scan --service=s3 --policy-dir=./my-policies
```

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
        run: cloudrift scan --service=s3 --format=sarif --output=results.sarif --fail-on-violation
        env:
          AWS_ACCESS_KEY_ID: ${{ secrets.AWS_ACCESS_KEY_ID }}
          AWS_SECRET_ACCESS_KEY: ${{ secrets.AWS_SECRET_ACCESS_KEY }}

      - name: Upload SARIF
        uses: github/codeql-action/upload-sarif@v2
        if: always()
        with:
          sarif_file: results.sarif
```

### GitLab CI

```yaml
drift-scan:
  image: golang:1.21
  script:
    - go install github.com/inayathulla/cloudrift@latest
    - terraform init && terraform plan -out=tfplan.binary
    - terraform show -json tfplan.binary > plan.json
    - cloudrift scan --service=s3 --format=json --fail-on-violation
```

## Desktop Dashboard

**[Cloudrift UI](https://github.com/inayathulla/cloudrift-ui)** is a cross-platform dashboard that visualizes Cloudrift scan results. It runs **two ways**:

- **Web via Docker** — One command deploys a container with Flutter web app, Go API server, nginx, and Terraform. Zero dependencies.
- **Native Desktop** — Runs on macOS, Linux, or Windows, calling the CLI binary directly. No server needed.

### Quick Start (Docker)

```bash
docker build -t cloudrift-ui .
docker run -d -p 8080:80 -v ~/.aws:/root/.aws:ro --name cloudrift-ui cloudrift-ui:latest
open http://localhost:8080
```

### Quick Start (Desktop)

```bash
git clone https://github.com/inayathulla/cloudrift-ui.git
cd cloudrift-ui
flutter pub get && flutter run -d macos   # or -d linux / -d windows
```

### Key Features

| Feature | Description |
|---------|-------------|
| **Interactive Dashboard** | Clickable KPI cards, drift trend charts, severity donut, framework compliance rings |
| **Drift Visualization** | Three-column diff viewer: Attribute / Expected (Terraform) / Actual (AWS) |
| **Resource Builder** | Three modes: Terraform (auto-generate plan.json), Manual (S3/EC2 forms), Upload (drag & drop) |
| **Policy Dashboard** | 21 OPA policies with compliance mapping (HIPAA, GDPR, ISO 27001, PCI DSS), severity filters, remediation guidance |
| **Compliance Scoring** | Animated compliance rings with category breakdowns and trend tracking |
| **Scan History** | Persistent local history with trend charts and human-readable durations |
| **Go API Server** | Backend wrapping Cloudrift CLI for web mode — scan, config, health, Terraform plan generation |
| **Dark Theme** | Cybersecurity-grade dark theme with severity-coded color system |

See the [Cloudrift UI README](https://github.com/inayathulla/cloudrift-ui) for full documentation and screenshots.

## Configuration

| Field | Description | Required |
|-------|-------------|----------|
| `aws_profile` | AWS credentials profile name | Yes |
| `region` | AWS region to scan | Yes |
| `plan_path` | Path to Terraform plan JSON | Yes |

### Example Configurations

**S3 Scanning:**
```yaml
# config/cloudrift.yml
aws_profile: default
region: us-east-1
plan_path: ./examples/plan.json
```

**EC2 Scanning:**
```yaml
# config/cloudrift-ec2.yml
aws_profile: default
region: us-east-1
plan_path: ./examples/ec2-plan.json
```

## Project Structure

```
cloudrift/
├── cmd/                          # CLI commands
│   ├── root.go
│   └── scan.go
├── internal/
│   ├── aws/                      # AWS API integrations
│   │   ├── config.go             # AWS SDK configuration
│   │   ├── s3.go                 # S3 API client
│   │   └── ec2.go                # EC2 API client
│   ├── detector/                 # Drift detection logic
│   │   ├── interface.go          # Detector interface
│   │   ├── s3.go                 # S3 drift detector
│   │   ├── ec2.go                # EC2 drift detector
│   │   ├── s3_printer.go         # S3 console output
│   │   └── ec2_printer.go        # EC2 console output
│   ├── output/                   # Output formatters
│   │   ├── json.go               # JSON formatter
│   │   ├── sarif.go              # SARIF formatter
│   │   └── console.go            # Console formatter
│   ├── policy/                   # OPA policy engine
│   │   ├── engine.go             # Policy evaluation
│   │   ├── loader.go             # Policy loading
│   │   └── policies/             # Built-in policies
│   │       ├── security/
│   │       ├── tagging/
│   │       └── cost/
│   ├── models/                   # Data structures
│   └── parser/                   # Terraform plan parser
├── config/                       # Example configurations
├── examples/                     # Example Terraform plans
├── docs/                         # Documentation
│   └── USAGE.md                  # Detailed usage guide
└── tests/                        # Unit tests
```

## Roadmap

### Completed ✅
- [x] S3 drift detection
- [x] EC2 drift detection
- [x] JSON output format
- [x] SARIF output format
- [x] OPA policy engine
- [x] Built-in security policies
- [x] Custom policy support
- [x] `--fail-on-violation` flag
- [x] Desktop dashboard ([Cloudrift UI](https://github.com/inayathulla/cloudrift-ui))

### In Progress 🚧
- [ ] IAM drift detection
- [ ] Security Groups detection
- [ ] RDS drift detection

### Planned 📋
- [ ] Compliance packs (CIS, SOC2, HIPAA)
- [ ] Multi-account scanning
- [ ] Slack/PagerDuty alerts

## Contributing

Contributions are welcome!

```bash
# Clone
git clone https://github.com/inayathulla/cloudrift.git
cd cloudrift

# Build
go build -o cloudrift .

# Test
go test ./...

# Run
./cloudrift scan --service=s3 --config=config/cloudrift.yml
```

## Related Projects

| Project | Description |
|---------|-------------|
| **[Cloudrift UI](https://github.com/inayathulla/cloudrift-ui)** | Cross-platform security dashboard (Flutter) — Desktop + Web/Docker. Drift diff viewer, resource builder, 21-policy browser with HIPAA/GDPR/ISO/PCI mapping, compliance scoring, Go API server, and dark cybersecurity theme. |

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
