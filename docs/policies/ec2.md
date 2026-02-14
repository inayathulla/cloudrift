# EC2 Policies

4 policies covering EC2 instance metadata, encryption, network exposure, and cost optimization.

| ID | Name | Severity | Frameworks |
|----|------|----------|------------|
| [EC2-001](#ec2-001) | EC2 IMDSv2 Required | <span class="severity-medium">MEDIUM</span> | PCI DSS, ISO 27001, SOC 2 |
| [EC2-002](#ec2-002) | EC2 Root Volume Encryption | <span class="severity-high">HIGH</span> | HIPAA, PCI DSS, ISO 27001, GDPR, SOC 2 |
| [EC2-003](#ec2-003) | EC2 Public IP Warning | <span class="severity-medium">MEDIUM</span> | PCI DSS, ISO 27001, SOC 2 |
| [EC2-005](#ec2-005) | EC2 Large Instance Review | <span class="severity-medium">MEDIUM</span> | -- |

---

## EC2-001

**EC2 IMDSv2 Required** | <span class="severity-medium">MEDIUM</span>

**Frameworks:** PCI DSS, ISO 27001, SOC 2

EC2 instance should require IMDSv2 (`http_tokens = required`). IMDSv1 is vulnerable to Server-Side Request Forgery (SSRF) attacks that can expose instance credentials.

**Remediation:**

```hcl
resource "aws_instance" "example" {
  ami           = "ami-0123456789abcdef0"
  instance_type = "t3.micro"

  metadata_options {
    http_endpoint = "enabled"
    http_tokens   = "required"
  }
}
```

**Resource type:** `aws_instance`

---

## EC2-002

**EC2 Root Volume Encryption** | <span class="severity-high">HIGH</span>

**Frameworks:** HIPAA, PCI DSS, ISO 27001, GDPR, SOC 2

EC2 instance must have an encrypted root volume. Unencrypted volumes risk exposing sensitive data at rest, violating multiple compliance frameworks.

**Remediation:**

```hcl
resource "aws_instance" "example" {
  ami           = "ami-0123456789abcdef0"
  instance_type = "t3.micro"

  root_block_device {
    encrypted  = true
    kms_key_id = aws_kms_key.example.arn
  }
}
```

**Resource type:** `aws_instance`

---

## EC2-003

**EC2 Public IP Warning** | <span class="severity-medium">MEDIUM</span>

**Frameworks:** PCI DSS, ISO 27001, SOC 2

EC2 instance will have a public IP assigned. Instances with public IPs are directly reachable from the internet, increasing the attack surface. Use a load balancer or NAT gateway for outbound access instead.

**Remediation:**

```hcl
resource "aws_instance" "example" {
  ami           = "ami-0123456789abcdef0"
  instance_type = "t3.micro"

  associate_public_ip_address = false
}
```

**Resource type:** `aws_instance`

---

## EC2-005

**EC2 Large Instance Review** | <span class="severity-medium">MEDIUM</span>

**Category:** cost

**Frameworks:** --

EC2 instance uses a very large or expensive instance type. Large instances significantly increase cloud spend and may be over-provisioned for the workload. Review whether right-sizing or Spot Instances can reduce costs.

**Remediation:**

```hcl
resource "aws_instance" "example" {
  # Consider downsizing from large instance types (e.g., x1e.32xlarge)
  # to a smaller instance that fits your workload requirements
  instance_type = "m5.xlarge"

  # Alternatively, use a Spot Instance for fault-tolerant workloads
  # instance_market_options {
  #   market_type = "spot"
  #   spot_options {
  #     max_price = "0.05"
  #   }
  # }
}
```

**Resource type:** `aws_instance`
