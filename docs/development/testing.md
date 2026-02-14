# Testing

Cloudrift uses Go's built-in testing framework with [testify](https://github.com/stretchr/testify) for assertions.

## Test Structure

Tests live in `tests/internal/`, mirroring the `internal/` package structure:

```
tests/
└── internal/
    ├── detector/
    │   ├── s3_test.go          # S3 drift detection scenarios
    │   ├── ec2_test.go         # EC2 drift detection scenarios
    │   └── registry_test.go    # Service registry tests
    ├── models/
    │   └── s3bucket_test.go    # Model tests
    ├── output/
    │   └── formatter_test.go   # JSON, SARIF, Console formatter tests
    ├── parser/
    │   └── parser_test.go      # Plan parser tests
    └── policy/
        ├── policy_test.go      # Policy engine evaluation tests
        └── registry_test.go    # Policy registry + framework filtering tests
```

---

## Running Tests

```bash
# All tests
go test ./...

# Verbose output
go test -v ./...

# Specific package
go test ./tests/internal/detector/...

# Specific test function
go test -v ./tests/internal/policy/... -run TestPolicyRegistry_FilterByFrameworks

# Bypass test cache
go test -count=1 ./...
```

---

## Writing Tests

### Basic Test Pattern

```go
package detector

import (
    "testing"

    "github.com/inayathulla/cloudrift/internal/detector"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestS3DriftDetector_NoDrift(t *testing.T) {
    planned := models.S3Bucket{
        Name:              "my-bucket",
        VersioningEnabled: true,
    }
    live := models.S3Bucket{
        Name:              "my-bucket",
        VersioningEnabled: true,
    }

    det := detector.NewS3DriftDetector(cfg)
    results, err := det.DetectDrift(
        []models.S3Bucket{planned},
        []models.S3Bucket{live},
    )

    require.NoError(t, err)
    assert.Empty(t, results[0].TagDiffs)
    assert.False(t, results[0].VersioningDiff)
}
```

### Table-Driven Tests

```go
func TestSeverityLevels(t *testing.T) {
    tests := []struct {
        name     string
        severity string
        expected string
    }{
        {"critical maps to error", "critical", "error"},
        {"warning maps to warning", "warning", "warning"},
        {"info maps to note", "info", "note"},
    }

    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            result := mapSeverity(tc.severity)
            assert.Equal(t, tc.expected, result)
        })
    }
}
```

---

## Key Test Files

### Drift Detection Tests

`tests/internal/detector/s3_test.go` covers:

- No drift (identical plan and live)
- ACL drift
- Tag drift (modified, extra, missing tags)
- Versioning drift
- Encryption drift
- Logging drift
- Public access block drift
- Lifecycle rule drift
- Missing bucket (in plan but not in AWS)

### Policy Registry Tests

`tests/internal/policy/registry_test.go` covers:

- Registry loads non-empty
- Total matches policy map length
- Category totals are consistent
- Expected categories exist (security, tagging, cost)
- Expected frameworks exist (5 frameworks)
- Known policies exist with correct categories
- Framework mappings are correct
- No duplicate policy IDs
- Framework filtering (single, multiple, empty)
- Filtering doesn't mutate the original
- Filtered category totals are consistent
- KnownFrameworks returns sorted list
- Idempotent loading

### Formatter Tests

`tests/internal/output/formatter_test.go` covers:

- JSON format validity and field presence
- SARIF schema compliance
- Console output for drift/no-drift
- Compliance data in JSON output
- Framework filter metadata (`active_frameworks`)
- Backward compatibility (no compliance when nil)
- Severity mapping

---

## CI Test Workflow

Tests run automatically on push and PR via GitHub Actions:

```yaml
- name: Run Tests
  run: go test -v ./... -json > test-results/report.json
```

See [CI/CD Integration](../features/ci-cd.md) for the full workflow.
