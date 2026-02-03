# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Cloudrift is a Go CLI tool that detects configuration drift between Terraform Infrastructure-as-Code and live AWS infrastructure. It compares Terraform plan JSON against actual AWS API state, catching misconfigurations **before** terraform apply (pre-apply), unlike tools like driftctl which operate post-apply.

## Build & Development Commands

```bash
# Build locally
go build -o cloudrift main.go

# Install globally
go install github.com/inayathulla/cloudrift@latest

# Run tests
go test ./...

# Run a single test file
go test ./tests/internal/detector/...

# Format code
go fmt ./...

# Run the tool
cloudrift scan --config=config/cloudrift.yml --service=s3
```

## Architecture

The data flow follows a pipeline pattern:

1. **CLI Entry** (`cmd/scan.go`) - Parses config via Viper, validates AWS credentials
2. **Terraform Parsing** (`internal/parser/`) - Loads JSON plan, extracts `aws_s3_bucket` resources from `resource_changes[].change.after`
3. **AWS Fetching** (`internal/aws/s3.go`) - Fetches live bucket state via AWS SDK v2 using parallel goroutines (`errgroup.WithContext`)
4. **Drift Detection** (`internal/detector/s3.go`) - Compares planned vs live buckets, builds `DriftResult` structs
5. **Output** (`internal/detector/s3_printer.go`) - Colorized console output with attribute-level diffs

### Key Patterns

- **Parallel AWS API calls**: `internal/aws/s3.go` uses `errgroup` to fetch 7 different bucket attributes (ACL, tags, versioning, encryption, logging, public access block, lifecycle) concurrently per bucket
- **Service-based modularity**: Each AWS service (currently S3) has its own parser, detector, and printer in corresponding packages
- **Graceful error handling**: AWS API calls ignore expected "not found" errors (e.g., `NoSuchTagSet`, `NoSuchLifecycleConfiguration`)

### Data Models

Core struct in `internal/models/s3.go`:
- `S3Bucket` - Contains name, ACL, tags, versioning, encryption algorithm, logging config, `PublicAccessBlockConfig`, and `[]LifecycleRuleSummary`

### Configuration

Config file (`cloudrift.yml`):
```yaml
aws_profile: default
region: us-east-1
plan_path: ./plan.json
```

## Testing

Tests live in `tests/internal/` mirroring the `internal/` package structure. Uses testify for assertions. The main test file `tests/internal/detector/s3_test.go` covers all drift detection scenarios (ACL, tags, versioning, encryption, logging, public access block, lifecycle rules).
