# VPC Policies

2 policies covering network security.

| ID | Name | Severity | Frameworks |
|----|------|----------|------------|
| [VPC-001](#vpc-001) | Default Security Group Restrict All | <span class="severity-high">HIGH</span> | PCI DSS, ISO 27001, SOC 2 |
| [VPC-002](#vpc-002) | Subnet No Auto-Assign Public IP | <span class="severity-medium">MEDIUM</span> | PCI DSS, ISO 27001 |

---

## VPC-001

**Default Security Group Restrict All** | <span class="severity-high">HIGH</span>

**Frameworks:** PCI DSS, ISO 27001, SOC 2

Default security group must not have any ingress or egress rules. The default security group is automatically associated with instances that are not assigned a custom security group. Leaving rules on the default security group can inadvertently expose resources to unauthorized traffic.

**Remediation:**

```hcl
resource "aws_default_security_group" "default" {
  vpc_id = aws_vpc.example.id

  # Remove all ingress and egress rules from the default security group.
  # Use custom security groups for all traffic rules instead.
}
```

**Resource type:** `aws_default_security_group`

---

## VPC-002

**Subnet No Auto-Assign Public IP** | <span class="severity-medium">MEDIUM</span>

**Frameworks:** PCI DSS, ISO 27001

Subnet automatically assigns public IPs to launched instances. Auto-assigning public IPs increases the attack surface by making instances directly reachable from the internet. Use NAT gateways for outbound internet access from private subnets instead.

**Remediation:**

```hcl
resource "aws_subnet" "example" {
  vpc_id                  = aws_vpc.example.id
  cidr_block              = "10.0.1.0/24"
  availability_zone       = "us-east-1a"
  map_public_ip_on_launch = false
}

# Use a NAT gateway for outbound internet access
resource "aws_nat_gateway" "example" {
  allocation_id = aws_eip.nat.id
  subnet_id     = aws_subnet.public.id
}
```

**Resource type:** `aws_subnet`
