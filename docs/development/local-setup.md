# Local Setup

## Prerequisites

| Tool | Version | Install |
|------|---------|---------|
| Go | 1.24+ | [go.dev/dl](https://go.dev/dl/) or `brew install go` |
| AWS CLI | v2 | [aws.amazon.com/cli](https://aws.amazon.com/cli/) |
| Terraform | 1.0+ | [terraform.io](https://www.terraform.io/downloads) (for generating plan files) |
| Git | 2.0+ | `brew install git` |

## Clone and Build

```bash
git clone https://github.com/inayathulla/cloudrift.git
cd cloudrift
go build -o cloudrift main.go
```

## Install Globally

```bash
go install .
```

This places the `cloudrift` binary in `$GOPATH/bin`.

---

## Running Tests

```bash
# All tests
go test ./...

# Verbose output
go test -v ./...

# Specific package
go test ./tests/internal/detector/...
go test ./tests/internal/policy/...

# With count (bypass cache)
go test -count=1 ./...
```

---

## Code Formatting

```bash
go fmt ./...
```

---

## Project Layout

```bash
# View the project structure
ls -la internal/
```

See [Project Structure](../architecture/project-structure.md) for the full directory layout.

---

## IDE Setup

### VS Code

Recommended extensions:

- **Go** (`golang.go`) — IntelliSense, debugging, formatting
- **OPA** (`tsandall.opa`) — Rego syntax highlighting

### GoLand / IntelliJ

Go support is built-in. Enable the OPA plugin for `.rego` file support.

---

## Running a Local Scan

```bash
# 1. Ensure AWS credentials are configured
aws sts get-caller-identity

# 2. Generate a Terraform plan
cd /path/to/terraform/project
terraform plan -out=tfplan
terraform show -json tfplan > plan.json

# 3. Create a config
cat > cloudrift.yml <<EOF
aws_profile: default
region: us-east-1
plan_path: /path/to/plan.json
EOF

# 4. Run the scan
./cloudrift scan --service=s3
```
