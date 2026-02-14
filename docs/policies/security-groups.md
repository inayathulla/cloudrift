# Security Group Policies

Cloudrift includes **4 built-in policies** for AWS Security Groups, covering network access controls to prevent unrestricted inbound access to critical ports and services.

## Summary

| ID | Policy Name | Severity | Frameworks |
|----|------------|----------|------------|
| [SG-001](#sg-001) | No Unrestricted SSH Access | <span class="severity-critical">CRITICAL</span> | PCI DSS, ISO 27001, SOC 2 |
| [SG-002](#sg-002) | No Unrestricted RDP Access | <span class="severity-critical">CRITICAL</span> | PCI DSS, ISO 27001, SOC 2 |
| [SG-003](#sg-003) | No Unrestricted All Ports Access | <span class="severity-critical">CRITICAL</span> | PCI DSS, ISO 27001, SOC 2 |
| [SG-004](#sg-004) | Database Port Public Exposure | <span class="severity-high">HIGH</span> | HIPAA, PCI DSS, ISO 27001, SOC 2 |

---

## SG-001

### No Unrestricted SSH Access

<span class="severity-critical">CRITICAL</span>

**Frameworks:** PCI DSS, ISO 27001, SOC 2

**Resource type:** `aws_security_group`

**Description:**
Security group allows SSH (port 22) from `0.0.0.0/0`. Unrestricted SSH access exposes instances to brute-force attacks and unauthorized login attempts from any IP address on the internet. SSH access should be limited to known, trusted IP ranges such as corporate networks, bastion hosts, or VPN endpoints.

**Remediation:**

Restrict SSH to specific IP ranges; use a bastion host or VPN for remote access.

```hcl
resource "aws_security_group" "example" {
  name        = "restricted-ssh"
  description = "Allow SSH from trusted IPs only"
  vpc_id      = aws_vpc.main.id

  ingress {
    description = "SSH from corporate network"
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["10.0.0.0/8"]  # Replace with your trusted IP range
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}
```

---

## SG-002

### No Unrestricted RDP Access

<span class="severity-critical">CRITICAL</span>

**Frameworks:** PCI DSS, ISO 27001, SOC 2

**Resource type:** `aws_security_group`

**Description:**
Security group allows RDP (port 3389) from `0.0.0.0/0`. Unrestricted RDP access exposes Windows instances to brute-force attacks, credential stuffing, and exploitation of RDP vulnerabilities from any IP on the internet. RDP should be restricted to specific trusted IP ranges.

**Remediation:**

Restrict RDP to specific IPs; use a bastion host or VPN for remote access.

```hcl
resource "aws_security_group" "example" {
  name        = "restricted-rdp"
  description = "Allow RDP from trusted IPs only"
  vpc_id      = aws_vpc.main.id

  ingress {
    description = "RDP from corporate network"
    from_port   = 3389
    to_port     = 3389
    protocol    = "tcp"
    cidr_blocks = ["10.0.0.0/8"]  # Replace with your trusted IP range
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}
```

---

## SG-003

### No Unrestricted All Ports Access

<span class="severity-critical">CRITICAL</span>

**Frameworks:** PCI DSS, ISO 27001, SOC 2

**Resource type:** `aws_security_group`

**Description:**
Security group allows all ports (0-65535) from `0.0.0.0/0`. Opening all ports to the internet is the broadest possible exposure and makes every running service on the instance publicly reachable. This violates the principle of least privilege and creates a wide attack surface.

**Remediation:**

Define specific ports needed and restrict source CIDR ranges to only trusted networks.

```hcl
resource "aws_security_group" "example" {
  name        = "restricted-access"
  description = "Allow only required ports from trusted sources"
  vpc_id      = aws_vpc.main.id

  ingress {
    description = "HTTPS from internet"
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    description = "Application port from internal network"
    from_port   = 8080
    to_port     = 8080
    protocol    = "tcp"
    cidr_blocks = ["10.0.0.0/8"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}
```

---

## SG-004

### Database Port Public Exposure

<span class="severity-high">HIGH</span>

**Frameworks:** HIPAA, PCI DSS, ISO 27001, SOC 2

**Resource type:** `aws_security_group`

**Description:**
Security group exposes a database port to `0.0.0.0/0`. Affected ports include MySQL (3306), PostgreSQL (5432), MSSQL (1433), Oracle (1521), MongoDB (27017), Redis (6379), and Memcached (11211). Publicly exposing database ports allows attackers to directly target database services with brute-force attacks, SQL injection, or exploitation of known vulnerabilities.

**Remediation:**

Place databases in private subnets and restrict security group ingress to application-tier security groups or VPN CIDR ranges. Never expose database ports to the public internet.

```hcl
resource "aws_security_group" "database" {
  name        = "database-sg"
  description = "Allow database access from application tier only"
  vpc_id      = aws_vpc.main.id

  ingress {
    description     = "PostgreSQL from application tier"
    from_port       = 5432
    to_port         = 5432
    protocol        = "tcp"
    security_groups = [aws_security_group.application.id]
  }

  ingress {
    description     = "MySQL from application tier"
    from_port       = 3306
    to_port         = 3306
    protocol        = "tcp"
    security_groups = [aws_security_group.application.id]
  }

  ingress {
    description     = "Redis from application tier"
    from_port       = 6379
    to_port         = 6379
    protocol        = "tcp"
    security_groups = [aws_security_group.application.id]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}
```
