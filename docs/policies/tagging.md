# Tagging Policies

4 policies covering resource tagging governance.

!!! note
    TAG-001 is a **deny** rule (violation). TAG-002, TAG-003, and TAG-004 are **warn** rules (warnings).

**Applicable resource types:** `aws_s3_bucket`, `aws_instance`, `aws_security_group`, `aws_db_instance`, `aws_rds_cluster`, `aws_lambda_function`, `aws_ecs_cluster`, `aws_eks_cluster`, `aws_lb`, `aws_vpc`, `aws_ebs_volume`, `aws_kms_key`, `aws_cloudtrail`, `aws_cloudwatch_log_group`, `aws_secretsmanager_secret`

| ID | Name | Severity | Frameworks |
|----|------|----------|------------|
| [TAG-001](#tag-001) | Environment Tag Required | <span class="severity-medium">MEDIUM</span> | SOC 2 |
| [TAG-002](#tag-002) | Owner Tag Recommended | <span class="severity-low">LOW</span> | -- |
| [TAG-003](#tag-003) | Project Tag Recommended | <span class="severity-low">LOW</span> | -- |
| [TAG-004](#tag-004) | Name Tag Recommended | <span class="severity-low">LOW</span> | -- |

---

## TAG-001

**Environment Tag Required** | <span class="severity-medium">MEDIUM</span>

**Action:** deny

**Frameworks:** SOC 2

Resource is missing required 'Environment' tag. The Environment tag is essential for distinguishing between development, staging, and production resources, enabling proper access controls and change management processes.

**Remediation:**

```hcl
resource "aws_instance" "example" {
  ami           = "ami-0123456789abcdef0"
  instance_type = "t3.micro"

  tags = {
    Environment = "dev"  # Valid values: dev, staging, production
  }
}
```

**Applicable resource types:** `aws_s3_bucket`, `aws_instance`, `aws_security_group`, `aws_db_instance`, `aws_rds_cluster`, `aws_lambda_function`, `aws_ecs_cluster`, `aws_eks_cluster`, `aws_lb`, `aws_vpc`, `aws_ebs_volume`, `aws_kms_key`, `aws_cloudtrail`, `aws_cloudwatch_log_group`, `aws_secretsmanager_secret`

---

## TAG-002

**Owner Tag Recommended** | <span class="severity-low">LOW</span>

**Action:** warn

**Frameworks:** --

Resource is missing 'Owner' tag for accountability. The Owner tag identifies the team or individual responsible for a resource, enabling faster incident response and clearer cost attribution.

**Remediation:**

```hcl
resource "aws_instance" "example" {
  ami           = "ami-0123456789abcdef0"
  instance_type = "t3.micro"

  tags = {
    Owner = "platform-team"
  }
}
```

**Applicable resource types:** `aws_s3_bucket`, `aws_instance`, `aws_security_group`, `aws_db_instance`, `aws_rds_cluster`, `aws_lambda_function`, `aws_ecs_cluster`, `aws_eks_cluster`, `aws_lb`, `aws_vpc`, `aws_ebs_volume`, `aws_kms_key`, `aws_cloudtrail`, `aws_cloudwatch_log_group`, `aws_secretsmanager_secret`

---

## TAG-003

**Project Tag Recommended** | <span class="severity-low">LOW</span>

**Action:** warn

**Frameworks:** --

Resource is missing 'Project' tag for cost allocation. The Project tag enables grouping resources by project in AWS Cost Explorer and billing reports, providing visibility into per-project cloud spend.

**Remediation:**

```hcl
resource "aws_instance" "example" {
  ami           = "ami-0123456789abcdef0"
  instance_type = "t3.micro"

  tags = {
    Project = "cloudrift"
  }
}
```

**Applicable resource types:** `aws_s3_bucket`, `aws_instance`, `aws_security_group`, `aws_db_instance`, `aws_rds_cluster`, `aws_lambda_function`, `aws_ecs_cluster`, `aws_eks_cluster`, `aws_lb`, `aws_vpc`, `aws_ebs_volume`, `aws_kms_key`, `aws_cloudtrail`, `aws_cloudwatch_log_group`, `aws_secretsmanager_secret`

---

## TAG-004

**Name Tag Recommended** | <span class="severity-low">LOW</span>

**Action:** warn

**Frameworks:** --

Resource is missing 'Name' tag. The Name tag provides a human-readable identifier displayed in the AWS console, making it significantly easier to locate and manage resources across accounts and regions.

**Remediation:**

```hcl
resource "aws_instance" "example" {
  ami           = "ami-0123456789abcdef0"
  instance_type = "t3.micro"

  tags = {
    Name = "web-server-01"
  }
}
```

**Applicable resource types:** `aws_s3_bucket`, `aws_instance`, `aws_security_group`, `aws_db_instance`, `aws_rds_cluster`, `aws_lambda_function`, `aws_ecs_cluster`, `aws_eks_cluster`, `aws_lb`, `aws_vpc`, `aws_ebs_volume`, `aws_kms_key`, `aws_cloudtrail`, `aws_cloudwatch_log_group`, `aws_secretsmanager_secret`
