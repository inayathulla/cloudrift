# Scan Command

The `scan` command is Cloudrift's primary command. It compares Terraform plan JSON against live AWS infrastructure and evaluates security policies.

## Usage

```bash
cloudrift scan [flags]
```

## Flags

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--config` | `-c` | string | `cloudrift.yml` | Path to configuration file |
| `--service` | `-s` | string | `s3` | AWS service to scan (`s3`, `ec2`) |
| `--format` | `-f` | string | `console` | Output format (`console`, `json`, `sarif`) |
| `--output` | `-o` | string | stdout | Write output to file instead of stdout |
| `--policy-dir` | `-p` | string | — | Directory containing custom OPA policies |
| `--frameworks` | — | string | all | Comma-separated compliance frameworks (`hipaa,soc2,gdpr,pci_dss,iso_27001`) |
| `--fail-on-violation` | — | bool | `false` | Exit with non-zero code if policy violations found |
| `--skip-policies` | — | bool | `false` | Skip policy evaluation (drift detection only) |
| `--no-emoji` | — | bool | `false` | Use ASCII characters instead of emojis |

---

## Examples

### Basic Scans

```bash
# Scan S3 buckets (default)
cloudrift scan --service=s3

# Scan EC2 instances
cloudrift scan --service=ec2

# Use a custom config
cloudrift scan --config=/path/to/cloudrift.yml --service=s3
```

### Output Formats

```bash
# JSON output to stdout
cloudrift scan --service=s3 --format=json

# SARIF output to file
cloudrift scan --service=s3 --format=sarif --output=results.sarif

# JSON output to file
cloudrift scan --service=s3 --format=json --output=report.json
```

### Framework Filtering

```bash
# HIPAA-only compliance
cloudrift scan --service=s3 --frameworks=hipaa

# Multiple frameworks
cloudrift scan --service=s3 --frameworks=hipaa,gdpr

# SOC 2 + PCI DSS with JSON output
cloudrift scan --service=s3 --format=json --frameworks=soc2,pci_dss
```

### CI/CD Usage

```bash
# Fail pipeline on violations
cloudrift scan --service=s3 --fail-on-violation

# SARIF for GitHub Security tab
cloudrift scan --service=s3 --format=sarif --output=results.sarif --fail-on-violation

# ASCII output for CI logs (no emojis)
cloudrift scan --service=s3 --no-emoji --fail-on-violation
```

### Custom Policies

```bash
# Use custom policies alongside built-ins
cloudrift scan --service=s3 --policy-dir=./my-policies

# Skip all policies (drift detection only)
cloudrift scan --service=s3 --skip-policies
```

---

## Exit Codes

| Code | Meaning |
|------|---------|
| `0` | Scan completed successfully, no violations (or `--fail-on-violation` not set) |
| `1` | Error (invalid config, AWS credentials, plan file, etc.) |
| `2` | Policy violations found (requires `--fail-on-violation`) |

---

## Scan Pipeline

The scan command executes in 8 sequential steps:

1. **Load config** — Read `cloudrift.yml` via Viper
2. **Initialize AWS** — Load AWS SDK v2 config with profile and region
3. **Validate credentials** — Verify AWS credentials are valid
4. **Fetch identity** — Call STS `GetCallerIdentity` to display account info
5. **Load plan** — Parse Terraform plan JSON for the selected service
6. **Fetch live state** — Query AWS APIs for current resource state
7. **Detect drift** — Compare planned vs live attributes
8. **Evaluate policies** — Run OPA policies against resources and output results

Each step displays a progress spinner and elapsed time.

---

## Framework Validation

When `--frameworks` is set, Cloudrift validates the specified framework names against its known list. Unknown names cause an error with the list of valid options:

```
❌ Unknown framework: hipa
  Available frameworks: gdpr, hipaa, iso_27001, pci_dss, soc2
```
