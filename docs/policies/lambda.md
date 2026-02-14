# Lambda Policies

2 policies covering serverless security.

| ID | Name | Severity | Frameworks |
|----|------|----------|------------|
| [LAMBDA-001](#lambda-001) | Lambda Tracing Enabled | <span class="severity-medium">MEDIUM</span> | SOC 2, ISO 27001 |
| [LAMBDA-002](#lambda-002) | Lambda VPC Configuration | <span class="severity-medium">MEDIUM</span> | HIPAA, PCI DSS, ISO 27001 |

---

## LAMBDA-001

**Lambda Tracing Enabled** | <span class="severity-medium">MEDIUM</span>

**Frameworks:** SOC 2, ISO 27001

Lambda function should have X-Ray tracing enabled for observability. Active tracing helps identify performance bottlenecks, trace requests across distributed services, and provides the audit trail needed for compliance frameworks.

**Remediation:**

```hcl
resource "aws_lambda_function" "example" {
  function_name = "example-function"
  role          = aws_iam_role.lambda.arn
  handler       = "index.handler"
  runtime       = "nodejs18.x"

  tracing_config {
    mode = "Active"
  }
}
```

**Resource type:** `aws_lambda_function`

---

## LAMBDA-002

**Lambda VPC Configuration** | <span class="severity-medium">MEDIUM</span>

**Frameworks:** HIPAA, PCI DSS, ISO 27001

Lambda function is not configured to run in a VPC. Running Lambda functions inside a VPC enables access to private resources such as RDS databases and ElastiCache clusters, and allows network-level controls via security groups and NACLs required by compliance frameworks.

**Remediation:**

```hcl
resource "aws_lambda_function" "example" {
  function_name = "example-function"
  role          = aws_iam_role.lambda.arn
  handler       = "index.handler"
  runtime       = "nodejs18.x"

  vpc_config {
    subnet_ids         = [aws_subnet.private_a.id, aws_subnet.private_b.id]
    security_group_ids = [aws_security_group.lambda.id]
  }
}
```

**Resource type:** `aws_lambda_function`
