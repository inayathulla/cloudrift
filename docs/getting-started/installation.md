# Installation

Cloudrift can be installed via Go, Docker, or built from source.

## Go Install (Recommended)

```bash
go install github.com/inayathulla/cloudrift@latest
```

!!! note "Requires Go 1.24+"
    Cloudrift uses Go 1.24 features. Verify your Go version with `go version`.

### Verify Installation

```bash
cloudrift --help
```

---

## Docker

```bash
docker pull inayathulla/cloudrift:latest
```

### Run a Scan

```bash
docker run --rm \
  -v ~/.aws:/root/.aws:ro \
  -v $(pwd):/work \
  inayathulla/cloudrift:latest scan \
  --config=/work/cloudrift.yml \
  --service=s3
```

### Available Tags

| Tag | Description |
|-----|-------------|
| `latest` | Latest stable release |
| `v1.0.0` | Specific version |

---

## Build from Source

```bash
git clone https://github.com/inayathulla/cloudrift.git
cd cloudrift
go build -o cloudrift main.go
```

Optionally install to your Go bin path:

```bash
go install .
```

### Dockerfile

The included Dockerfile produces a minimal Alpine-based image:

```bash
docker build -t cloudrift .
```

The multi-stage build compiles a statically-linked binary (`CGO_ENABLED=0`) and copies it to an `alpine:latest` runtime image running as a non-root `cloudrift` user.

---

## Prerequisites

| Tool | Version | Required For |
|------|---------|-------------|
| Go | 1.24+ | Building from source or `go install` |
| AWS CLI | v2 | Configuring AWS credentials |
| Terraform | 1.0+ | Generating plan JSON files |
| Docker | 20+ | Running the Docker image |

## Next Steps

- [Configure Cloudrift](configuration.md) with a `cloudrift.yml` file
- [Set up AWS credentials](aws-credentials.md) with the right permissions
- [Run your first scan](../cli/scan-command.md)
