<p align="center">
  <h1 align="center">Cloudrift</h1>
  <p align="center">
    <strong>Detect drift. Defend cloud.</strong>
  </p>
  <p align="center">
    Pre-apply drift detection for Terraform and AWS
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

**Cloudrift** is an open-source drift detection tool that compares your Terraform plan against live AWS infrastructure — catching misconfigurations **before** `terraform apply`, not after.

![Demo](assets/s3_scanning.gif)

## Table of Contents

- [Why Cloudrift?](#why-cloudrift)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Configuration](#configuration)
- [How It Works](#how-it-works)
- [Use Cases](#use-cases)
- [Contributing](#contributing)
- [License](#license)

## Why Cloudrift?

Unlike post-apply tools like `driftctl` or Terraform's built-in drift detection, Cloudrift operates on your **Terraform plan** — giving you a safety net before changes reach production.

| Feature | Cloudrift | driftctl | Terraform Drift Detect |
|---------|-----------|----------|------------------------|
| **Source of Truth** | Terraform Plan + Live AWS API | Terraform State + Live AWS API | Terraform State vs Live State |
| **Timing** | Pre-apply | Post-apply | Post-apply |
| **Output Format** | JSON attribute-level diffs | CLI output | Human-readable diff |
| **CI/CD Integration** | Native JSON for automation | Requires parsing | Not designed for automation |

## Installation

### Via Go

```bash
go install github.com/inayathulla/cloudrift@latest
```

Ensure `$GOPATH/bin` is in your PATH:

```bash
export PATH="$HOME/go/bin:$PATH"
```

### Via Docker

```bash
docker pull inayathulla/cloudrift
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

Create `config/cloudrift.yml`:

```yaml
aws_profile: default
region: us-east-1
plan_path: ./plan.json
```

### 3. Run Cloudrift

**With Go:**

```bash
cloudrift scan --config=config/cloudrift.yml --service=s3
```

**With Docker:**

```bash
docker run --rm \
  -v $(pwd):/app \
  -v ~/.aws:/root/.aws:ro \
  inayathulla/cloudrift \
  cloudrift scan --config=/app/config/cloudrift.yml --service=s3
```

## Configuration

| Field | Description | Required |
|-------|-------------|----------|
| `aws_profile` | AWS credentials profile name | Yes |
| `region` | AWS region to scan | Yes |
| `plan_path` | Path to Terraform plan JSON | Yes |

### Example plan.json structure

```json
{
  "resource_changes": [
    {
      "address": "aws_s3_bucket.example",
      "type": "aws_s3_bucket",
      "change": {
        "actions": ["create"],
        "after": {
          "bucket": "my-bucket",
          "acl": "private",
          "tags": { "env": "prod" },
          "versioning": { "enabled": true }
        }
      }
    }
  ]
}
```

## How It Works

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│  Terraform Plan │────▶│    Cloudrift    │◀────│   AWS Live API  │
│     (JSON)      │     │   Comparator    │     │     (S3, etc)   │
└─────────────────┘     └────────┬────────┘     └─────────────────┘
                                 │
                                 ▼
                        ┌─────────────────┐
                        │  Drift Report   │
                        │  (CLI / JSON)   │
                        └─────────────────┘
```

1. **Parse** — Reads your Terraform plan JSON and extracts planned resource configurations
2. **Fetch** — Queries AWS APIs in parallel to get current live state
3. **Compare** — Detects attribute-level differences between plan and reality
4. **Report** — Outputs colorized CLI results or structured JSON for automation

### Supported Resources

| Resource | Status |
|----------|--------|
| S3 Buckets | Available |
| EC2 Instances | Planned |
| IAM Roles | Planned |
| Security Groups | Planned |

## Use Cases

### CI/CD Pipeline Integration

Run Cloudrift in your pipeline to catch drift before deployment:

```yaml
# GitHub Actions example
- name: Check for drift
  run: |
    cloudrift scan --config=config/cloudrift.yml --service=s3
```

### Scheduled Compliance Scans

Detect configuration drift on a schedule to maintain compliance:

```bash
# Cron job example
0 9 * * * cloudrift scan --config=/path/to/cloudrift.yml --service=s3 >> /var/log/drift.log
```

### Pre-Deploy Safety Checks

Validate that your intended changes won't conflict with manual changes made in AWS:

```bash
terraform plan -out=tfplan.binary
terraform show -json tfplan.binary > plan.json
cloudrift scan --config=config/cloudrift.yml --service=s3
terraform apply tfplan.binary  # Only if no unexpected drift
```

## Project Structure

```
cloudrift/
├── cmd/                    # CLI commands
├── internal/
│   ├── aws/                # AWS API integrations
│   ├── detector/           # Drift detection logic
│   ├── models/             # Data structures
│   └── parser/             # Terraform plan parser
├── tests/                  # Unit tests
├── config/                 # Example configuration
└── examples/               # Sample Terraform projects
```

## Contributing

Contributions are welcome! Please read our contributing guidelines before submitting a PR.

### Development

```bash
# Clone the repository
git clone https://github.com/inayathulla/cloudrift.git
cd cloudrift

# Build
go build -o cloudrift main.go

# Run tests
go test ./...

# Format code
go fmt ./...
```

### Guidelines

- Use clear commit messages (e.g., `feat: add EC2 drift detection`)
- Keep code modular — one service per detector
- Add unit tests for new components
- Run `go fmt` before committing

## Connect

- **Issues & Features:** [GitHub Issues](https://github.com/inayathulla/cloudrift/issues)
- **Email:** [inayathulla2020@gmail.com](mailto:inayathulla2020@gmail.com)
- **LinkedIn:** [Inayathulla Khan Lavani](https://www.linkedin.com/in/inayathullakhan)

---

## License

[Apache License 2.0](LICENSE)

---

<p align="center">
  <sub>Built with care for the DevOps and Security community</sub>
</p>
