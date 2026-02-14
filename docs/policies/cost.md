# Cost Policies

2 policies covering instance cost optimization.

| ID | Name | Severity | Frameworks |
|----|------|----------|------------|
| [COST-002](#cost-002) | Very Large Instance Size | <span class="severity-low">LOW</span> | -- |
| [COST-003](#cost-003) | Previous Generation Instance | <span class="severity-low">LOW</span> | -- |

---

## COST-002

**Very Large Instance Size** | <span class="severity-low">LOW</span>

**Frameworks:** --

EC2 instance uses very large size (16xlarge/24xlarge) with monthly cost exceeding $5,000. Very large instances represent significant cloud spend and are often over-provisioned. Verify the workload genuinely requires this capacity, and consider auto-scaling groups to match capacity to demand.

**Remediation:**

```hcl
resource "aws_instance" "example" {
  # Review whether this instance size is necessary.
  # Consider using auto-scaling instead of a single very large instance.
  instance_type = "m5.4xlarge"  # Downsize from 16xlarge/24xlarge

  # Alternatively, use an Auto Scaling Group to scale horizontally
  # resource "aws_autoscaling_group" "example" {
  #   min_size         = 2
  #   max_size         = 10
  #   desired_capacity = 2
  #   launch_template {
  #     id      = aws_launch_template.example.id
  #     version = "$Latest"
  #   }
  # }
}
```

**Resource type:** `aws_instance`

---

## COST-003

**Previous Generation Instance** | <span class="severity-low">LOW</span>

**Frameworks:** --

EC2 instance uses previous generation family (m4, m3, c4, c3, r4, r3, i3, d2, t2). Previous generation instance types offer lower performance per dollar compared to current generation equivalents. Upgrading typically provides better performance at the same or lower cost.

**Remediation:**

```hcl
resource "aws_instance" "example" {
  # Upgrade from previous generation to current generation:
  #   m3/m4  -> m5 or m6i
  #   c3/c4  -> c5 or c6i
  #   r3/r4  -> r5 or r6i
  #   i3     -> i3en or i4i
  #   d2     -> d3 or d3en
  #   t2     -> t3 or t3a
  instance_type = "m5.xlarge"  # Upgraded from m4.xlarge
}
```

**Resource type:** `aws_instance`
