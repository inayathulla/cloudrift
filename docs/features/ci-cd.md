# CI/CD Integration

Cloudrift integrates into CI/CD pipelines to catch drift and policy violations before deployment.

## GitHub Actions

### Basic Workflow

```yaml
name: Cloudrift Scan
on:
  pull_request:
    branches: [main]

jobs:
  scan:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.24'

      - name: Install Cloudrift
        run: go install github.com/inayathulla/cloudrift@latest

      - name: Generate Terraform Plan
        run: |
          terraform init
          terraform plan -out=tfplan
          terraform show -json tfplan > plan.json

      - name: Run Cloudrift Scan
        run: cloudrift scan --service=s3 --fail-on-violation --no-emoji
        env:
          AWS_ACCESS_KEY_ID: ${{ secrets.AWS_ACCESS_KEY_ID }}
          AWS_SECRET_ACCESS_KEY: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
          AWS_REGION: us-east-1
```

### With SARIF Upload

Upload results to GitHub's Security tab:

```yaml
name: Cloudrift Security Scan
on:
  push:
    branches: [main]
  pull_request:

permissions:
  security-events: write

jobs:
  scan:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.24'

      - name: Install Cloudrift
        run: go install github.com/inayathulla/cloudrift@latest

      - name: Generate Terraform Plan
        run: |
          terraform init
          terraform plan -out=tfplan
          terraform show -json tfplan > plan.json

      - name: Run Cloudrift Scan
        run: |
          cloudrift scan --service=s3 \
            --format=sarif --output=results.sarif \
            --no-emoji
        continue-on-error: true
        env:
          AWS_ACCESS_KEY_ID: ${{ secrets.AWS_ACCESS_KEY_ID }}
          AWS_SECRET_ACCESS_KEY: ${{ secrets.AWS_SECRET_ACCESS_KEY }}

      - name: Upload SARIF
        uses: github/codeql-action/upload-sarif@v3
        with:
          sarif_file: results.sarif
```

### Framework-Specific Scan

```yaml
      - name: HIPAA Compliance Check
        run: |
          cloudrift scan --service=s3 \
            --frameworks=hipaa \
            --fail-on-violation \
            --no-emoji
```

### Docker-Based Workflow

```yaml
      - name: Run Cloudrift via Docker
        run: |
          docker run --rm \
            -e AWS_ACCESS_KEY_ID=${{ secrets.AWS_ACCESS_KEY_ID }} \
            -e AWS_SECRET_ACCESS_KEY=${{ secrets.AWS_SECRET_ACCESS_KEY }} \
            -v $(pwd):/work \
            inayathulla/cloudrift:latest scan \
            --config=/work/cloudrift-s3.yml \
            --service=s3 \
            --fail-on-violation \
            --no-emoji
```

---

## GitLab CI

```yaml
cloudrift-scan:
  stage: test
  image: golang:1.24
  before_script:
    - go install github.com/inayathulla/cloudrift@latest
  script:
    - terraform init
    - terraform plan -out=tfplan
    - terraform show -json tfplan > plan.json
    - cloudrift scan --service=s3 --fail-on-violation --no-emoji
  variables:
    AWS_ACCESS_KEY_ID: $AWS_ACCESS_KEY_ID
    AWS_SECRET_ACCESS_KEY: $AWS_SECRET_ACCESS_KEY
    AWS_REGION: us-east-1
  only:
    - merge_requests
```

---

## Pipeline Gating

Use `--fail-on-violation` to gate deployments:

| Exit Code | Meaning | Pipeline Result |
|-----------|---------|-----------------|
| `0` | No violations | :material-check: Pass |
| `1` | Scan error | :material-close: Fail |
| `2` | Violations found | :material-close: Fail |

!!! tip "Gradual adoption"
    Start without `--fail-on-violation` to see results without breaking builds. Enable it once your team has addressed existing violations.

---

## JSON Reports as Artifacts

Save scan results as CI artifacts for auditing:

```yaml
      - name: Run Scan
        run: |
          cloudrift scan --service=s3 \
            --format=json --output=cloudrift-report.json \
            --no-emoji

      - name: Upload Report
        uses: actions/upload-artifact@v4
        with:
          name: cloudrift-report
          path: cloudrift-report.json
```
