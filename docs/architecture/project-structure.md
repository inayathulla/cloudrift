# Project Structure

```
cloudrift/
├── main.go                         # Entry point
├── cmd/
│   ├── root.go                     # Base Cobra command
│   └── scan.go                     # Scan command with all flags and pipeline logic
├── internal/
│   ├── aws/                        # AWS API integrations
│   │   ├── config.go               # AWS SDK v2 configuration
│   │   ├── s3.go                   # S3 API client (parallel attribute fetching)
│   │   ├── ec2.go                  # EC2 API client (pagination support)
│   │   ├── iam.go                  # IAM API client (roles, users, policies, groups)
│   │   └── identity.go            # STS identity operations
│   ├── common/                     # Shared utilities
│   │   └── bootstrap.go           # Config loading, AWS init, credential validation
│   ├── detector/                   # Drift detection logic
│   │   ├── interface.go           # DriftResultPrinter interface
│   │   ├── registry.go            # Service detector registry
│   │   ├── s3.go                  # S3 drift detector
│   │   ├── ec2.go                 # EC2 drift detector
│   │   ├── iam.go                 # IAM drift detector
│   │   ├── s3_printer.go          # S3 console output
│   │   ├── ec2_printer.go         # EC2 console output
│   │   ├── iam_printer.go         # IAM console output
│   │   └── printer.go            # Common printer utilities
│   ├── models/                     # Data structures
│   │   ├── s3.go                  # S3Bucket, PublicAccessBlockConfig, LifecycleRuleSummary
│   │   ├── ec2.go                 # EC2Instance, BlockDevice
│   │   ├── iam.go                 # IAMRole, IAMUser, IAMPolicy, IAMGroup
│   │   └── analytics.go          # Analytics models
│   ├── output/                     # Output formatters
│   │   ├── formatter.go          # Format registry, interfaces, data types
│   │   ├── console.go            # Colorized CLI formatter
│   │   ├── json.go               # JSON formatter
│   │   └── sarif.go              # SARIF 2.1.0 formatter
│   ├── parser/                     # Terraform plan JSON parsers
│   │   ├── plan.go               # Core parsing logic
│   │   ├── s3.go                 # S3 resource parser
│   │   ├── ec2.go                # EC2 resource parser
│   │   └── iam.go                # IAM resource parser
│   └── policy/                     # OPA policy engine
│       ├── engine.go             # Policy evaluation (compile, query, parse)
│       ├── loader.go             # Embedded policy loading (//go:embed)
│       ├── registry.go           # Dynamic policy metadata extraction
│       ├── result.go             # Violation, EvaluationResult structs
│       ├── input.go              # PolicyInput structs
│       └── policies/             # 49 built-in OPA policies
│           ├── security/         # 42 security policies (16 .rego files)
│           ├── tagging/          # 4 tagging policies (1 .rego file)
│           └── cost/             # 3 cost policies (1 .rego file)
├── tests/                          # Unit test suite
│   └── internal/
│       ├── detector/             # Drift detection tests
│       ├── models/               # Model tests
│       ├── output/               # Formatter tests
│       ├── parser/               # Plan parser tests
│       └── policy/               # Policy engine + registry tests
├── config/                         # Example configurations
│   ├── cloudrift-s3.yml            # S3 scanning config
│   ├── cloudrift-ec2.yml          # EC2 scanning config
│   └── cloudrift-iam.yml          # IAM scanning config
├── docs/                           # MkDocs documentation
├── .github/workflows/             # CI/CD workflows
├── Dockerfile                      # Multi-stage Docker build
├── go.mod                          # Go module definition
└── go.sum                          # Dependency lock file
```

---

## Package Responsibilities

| Package | Responsibility |
|---------|---------------|
| `cmd` | CLI commands, flag parsing, scan pipeline orchestration, compliance scoring |
| `internal/aws` | AWS SDK v2 API calls for each service |
| `internal/common` | Shared utilities: config loading, AWS initialization, credential validation |
| `internal/detector` | Drift detection logic: compare planned vs live, build DriftResult structs |
| `internal/models` | Data structures for AWS resources (S3Bucket, EC2Instance, IAMRole, etc.) |
| `internal/output` | Output formatters (Console, JSON, SARIF) and format registry |
| `internal/parser` | Terraform plan JSON parsing, resource extraction |
| `internal/policy` | OPA policy engine: loading, compilation, evaluation, result types |
| `tests` | Unit tests mirroring the `internal/` package structure |

---

## Key Interfaces

### DriftDetector

```go
type DriftDetector interface {
    FetchLiveState() (interface{}, error)
    DetectDrift(plan interface{}, live interface{}) ([]DriftResult, error)
}
```

Implemented by `S3DriftDetector`, `EC2DriftDetector`, and `IAMDriftDetector`.

### DriftResultPrinter

```go
type DriftResultPrinter interface {
    PrintDrift(results []DriftResult, plan interface{}, live interface{})
}
```

Implemented by `S3DriftResultPrinter`, `EC2DriftResultPrinter`, and `IAMDriftResultPrinter`.

### Formatter

```go
type Formatter interface {
    Format(w io.Writer, result ScanResult) error
    Name() string
    FileExtension() string
}
```

Implemented by `JSONFormatter`, `SARIFFormatter`, and `ConsoleFormatter`.

---

## Key Data Models

### S3Bucket

```go
type S3Bucket struct {
    Id                  string
    Name                string
    Acl                 string
    Tags                map[string]string
    VersioningEnabled   bool
    EncryptionAlgorithm string
    LoggingEnabled      bool
    PublicAccessBlock   PublicAccessBlockConfig
    LifecycleRules      []LifecycleRuleSummary
}
```

### EC2Instance

```go
type EC2Instance struct {
    InstanceID         string
    InstanceType       string
    AMI                string
    SubnetID           string
    Tags               map[string]string
    EBSOptimized       bool
    Monitoring         bool
    RootBlockDevice    BlockDevice
    TerraformAddress   string
}
```

### Violation

```go
type Violation struct {
    PolicyID        string
    PolicyName      string
    Message         string
    Severity        Severity    // critical, high, medium, low, info
    ResourceType    string
    ResourceAddress string
    Remediation     string
    Category        string      // security, tagging, cost
    Frameworks      []string    // hipaa, pci_dss, iso_27001, gdpr, soc2
}
```
