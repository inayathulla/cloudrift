# ELB Policies

3 policies covering load balancer security.

| ID | Name | Severity | Frameworks |
|----|------|----------|------------|
| [ELB-001](#elb-001) | ALB Access Logging | <span class="severity-medium">MEDIUM</span> | HIPAA, PCI DSS, ISO 27001, SOC 2 |
| [ELB-002](#elb-002) | ALB HTTPS Listener Required | <span class="severity-high">HIGH</span> | HIPAA, PCI DSS, ISO 27001, GDPR, SOC 2 |
| [ELB-003](#elb-003) | ALB Deletion Protection | <span class="severity-medium">MEDIUM</span> | ISO 27001, SOC 2 |

---

## ELB-001

**ALB Access Logging** | <span class="severity-medium">MEDIUM</span>

**Frameworks:** HIPAA, PCI DSS, ISO 27001, SOC 2

Application Load Balancer should have access logging enabled. Access logs capture detailed information about requests sent to the load balancer, including client IP, latencies, and server responses, which are essential for security analysis, troubleshooting, and compliance auditing.

**Remediation:**

```hcl
resource "aws_lb" "example" {
  name               = "example-alb"
  internal           = false
  load_balancer_type = "application"
  subnets            = var.subnet_ids

  access_logs {
    bucket  = aws_s3_bucket.lb_logs.id
    prefix  = "alb-logs"
    enabled = true
  }
}
```

**Resource type:** `aws_lb`

---

## ELB-002

**ALB HTTPS Listener Required** | <span class="severity-high">HIGH</span>

**Frameworks:** HIPAA, PCI DSS, ISO 27001, GDPR, SOC 2

Load balancer listener uses protocol other than HTTPS/TLS. Using unencrypted protocols (HTTP) for load balancer listeners exposes traffic to interception and man-in-the-middle attacks. All listeners should use HTTPS or TLS to ensure data is encrypted in transit.

**Remediation:**

```hcl
resource "aws_lb_listener" "example" {
  load_balancer_arn = aws_lb.example.arn
  port              = 443
  protocol          = "HTTPS"
  ssl_policy        = "ELBSecurityPolicy-TLS13-1-2-2021-06"
  certificate_arn   = aws_acm_certificate.example.arn

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.example.arn
  }
}
```

**Resource type:** `aws_lb_listener`

---

## ELB-003

**ALB Deletion Protection** | <span class="severity-medium">MEDIUM</span>

**Frameworks:** ISO 27001, SOC 2

Application Load Balancer does not have deletion protection enabled. Deletion protection prevents accidental or unauthorized removal of a load balancer, which could cause service outages and data loss for applications relying on it.

**Remediation:**

```hcl
resource "aws_lb" "example" {
  name                       = "example-alb"
  internal                   = false
  load_balancer_type         = "application"
  subnets                    = var.subnet_ids
  enable_deletion_protection = true
}
```

**Resource type:** `aws_lb`
