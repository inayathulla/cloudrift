# Adding a New AWS Service

This guide walks through adding support for a new AWS service to Cloudrift. Each service requires 5 new files.

!!! info "Currently supported services"
    Cloudrift ships with drift detection for **S3**, **EC2**, and **IAM**. Use this guide to add additional services.

## Checklist

- [ ] Create the data model (`internal/models/`)
- [ ] Create the plan parser (`internal/parser/`)
- [ ] Create the AWS client (`internal/aws/`)
- [ ] Create the drift detector (`internal/detector/`)
- [ ] Create the console printer (`internal/detector/`)
- [ ] Register in `cmd/scan.go`
- [ ] Add tests (`tests/internal/`)

---

## Step 1: Data Model

Create `internal/models/<service>.go`:

```go
package models

type RDSInstance struct {
    Id                string
    Name              string
    Engine            string
    EngineVersion     string
    InstanceClass     string
    StorageEncrypted  bool
    PubliclyAccessible bool
    MultiAZ           bool
    Tags              map[string]string
}
```

---

## Step 2: Plan Parser

Create `internal/parser/<service>.go`:

```go
package parser

import "github.com/inayathulla/cloudrift/internal/models"

func ParseRDSPlan(planPath string) ([]models.RDSInstance, error) {
    // Read plan JSON
    // Extract resource_changes where type == "aws_db_instance"
    // Map change.after to RDSInstance structs
    return instances, nil
}
```

Register a loader function in `internal/common/bootstrap.go`.

---

## Step 3: AWS Client

Create `internal/aws/<service>.go`:

```go
package aws

import (
    "context"
    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/service/rds"
    "github.com/inayathulla/cloudrift/internal/models"
)

func FetchRDSInstances(cfg aws.Config) ([]models.RDSInstance, error) {
    client := rds.NewFromConfig(cfg)
    // Call DescribeDBInstances with pagination
    // Map to []models.RDSInstance
    return instances, nil
}
```

!!! tip "Parallel fetching"
    Use `errgroup.WithContext` for attributes that require separate API calls, following the S3 pattern in `internal/aws/s3.go`.

---

## Step 4: Drift Detector

Create `internal/detector/<service>.go`:

```go
package detector

type RDSDriftDetector struct {
    cfg aws.Config
}

func NewRDSDriftDetector(cfg aws.Config) *RDSDriftDetector {
    return &RDSDriftDetector{cfg: cfg}
}

func (d *RDSDriftDetector) FetchLiveState() (interface{}, error) {
    return aws.FetchRDSInstances(d.cfg)
}

func (d *RDSDriftDetector) DetectDrift(plan, live interface{}) ([]DriftResult, error) {
    // Compare planned vs live attributes
    // Return DriftResult for each resource
}
```

---

## Step 5: Console Printer

Create `internal/detector/<service>_printer.go`:

```go
package detector

type RDSDriftResultPrinter struct{}

func (p RDSDriftResultPrinter) PrintDrift(results []DriftResult, plan, live interface{}) {
    // Colorized console output for RDS drift
}
```

---

## Step 6: Register in scan.go

Add a new case in the `switch service` block in `cmd/scan.go`:

```go
case "rds":
    det = detector.NewRDSDriftDetector(cfg)
    printer = detector.RDSDriftResultPrinter{}
    serviceName = "RDS"
    pr, err := common.LoadRDSPlan(planPath)
    if err != nil {
        // handle error
    }
    planResources = pr
    planCount = len(pr)
```

---

## Step 7: Add Tests

Create tests in `tests/internal/`:

- `tests/internal/detector/<service>_test.go` — Drift detection tests
- `tests/internal/models/<service>_test.go` — Model tests
- `tests/internal/parser/<service>_test.go` — Parser tests

Follow existing test patterns using `testify` assertions.
