# RDS Policies

Cloudrift includes **5 built-in policies** for Amazon RDS, covering database security, encryption, backup resilience, and high availability.

## Summary

| ID | Policy Name | Severity | Frameworks |
|----|------------|----------|------------|
| [RDS-001](#rds-001) | RDS Storage Encryption Required | <span class="severity-high">HIGH</span> | HIPAA, PCI DSS, ISO 27001, GDPR, SOC 2 |
| [RDS-002](#rds-002) | RDS No Public Access | <span class="severity-critical">CRITICAL</span> | HIPAA, PCI DSS, ISO 27001, GDPR, SOC 2 |
| [RDS-003](#rds-003) | RDS Backup Retention Period | <span class="severity-medium">MEDIUM</span> | HIPAA, ISO 27001, SOC 2 |
| [RDS-004](#rds-004) | RDS Deletion Protection | <span class="severity-medium">MEDIUM</span> | ISO 27001, SOC 2 |
| [RDS-005](#rds-005) | RDS Multi-AZ Recommended | <span class="severity-low">LOW</span> | HIPAA, ISO 27001, SOC 2 |

---

## RDS-001

### RDS Storage Encryption Required

<span class="severity-high">HIGH</span>

**Frameworks:** HIPAA, PCI DSS, ISO 27001, GDPR, SOC 2

**Resource type:** `aws_db_instance`

**Description:**
RDS instance must have storage encryption enabled. Unencrypted database storage leaves data at rest vulnerable to unauthorized access if the underlying storage media is compromised. Encryption at rest is a baseline requirement for most compliance frameworks handling sensitive or regulated data.

**Remediation:**

Enable storage encryption on the RDS instance. Note that encryption can only be enabled at creation time -- existing unencrypted instances must be migrated by creating an encrypted snapshot and restoring from it.

```hcl
resource "aws_db_instance" "example" {
  identifier     = "my-database"
  engine         = "postgres"
  engine_version = "15.4"
  instance_class = "db.t3.medium"

  allocated_storage = 20
  storage_encrypted = true
  kms_key_id        = aws_kms_key.rds.arn

  db_name  = "mydb"
  username = "admin"
  password = var.db_password

  skip_final_snapshot = false
}
```

---

## RDS-002

### RDS No Public Access

<span class="severity-critical">CRITICAL</span>

**Frameworks:** HIPAA, PCI DSS, ISO 27001, GDPR, SOC 2

**Resource type:** `aws_db_instance`

**Description:**
RDS instance must not be publicly accessible. A publicly accessible database has a public IP address and can be reached from the internet, making it a direct target for brute-force attacks, SQL injection, and data exfiltration. Databases should reside in private subnets and be accessible only through application-tier resources or VPN connections.

**Remediation:**

Set `publicly_accessible` to `false` and place the RDS instance in a private subnet group.

```hcl
resource "aws_db_instance" "example" {
  identifier     = "my-database"
  engine         = "postgres"
  engine_version = "15.4"
  instance_class = "db.t3.medium"

  allocated_storage    = 20
  publicly_accessible  = false
  db_subnet_group_name = aws_db_subnet_group.private.name

  db_name  = "mydb"
  username = "admin"
  password = var.db_password

  skip_final_snapshot = false
}

resource "aws_db_subnet_group" "private" {
  name       = "private-db-subnet-group"
  subnet_ids = aws_subnet.private[*].id
}
```

---

## RDS-003

### RDS Backup Retention Period

<span class="severity-medium">MEDIUM</span>

**Frameworks:** HIPAA, ISO 27001, SOC 2

**Resource type:** `aws_db_instance`

**Description:**
RDS instance has backup retention less than 7 days. Short backup retention periods increase the risk of data loss and limit the ability to recover from accidental deletions, corruption, or security incidents. A minimum of 7 days provides adequate recovery point coverage for most workloads.

**Remediation:**

Set `backup_retention_period` to 7 or higher to ensure sufficient point-in-time recovery coverage.

```hcl
resource "aws_db_instance" "example" {
  identifier     = "my-database"
  engine         = "postgres"
  engine_version = "15.4"
  instance_class = "db.t3.medium"

  allocated_storage       = 20
  backup_retention_period = 7
  backup_window           = "03:00-04:00"

  db_name  = "mydb"
  username = "admin"
  password = var.db_password

  skip_final_snapshot = false
}
```

---

## RDS-004

### RDS Deletion Protection

<span class="severity-medium">MEDIUM</span>

**Frameworks:** ISO 27001, SOC 2

**Resource type:** `aws_db_instance`

**Description:**
RDS instance does not have deletion protection enabled. Without deletion protection, the database can be accidentally deleted through the AWS Console, CLI, or API calls, including through Terraform destroy operations. Enabling deletion protection adds a safeguard that requires the protection to be explicitly disabled before the instance can be deleted.

**Remediation:**

Enable deletion protection on the RDS instance.

```hcl
resource "aws_db_instance" "example" {
  identifier     = "my-database"
  engine         = "postgres"
  engine_version = "15.4"
  instance_class = "db.t3.medium"

  allocated_storage   = 20
  deletion_protection = true

  db_name  = "mydb"
  username = "admin"
  password = var.db_password

  skip_final_snapshot = false
}
```

---

## RDS-005

### RDS Multi-AZ Recommended

<span class="severity-low">LOW</span>

**Frameworks:** HIPAA, ISO 27001, SOC 2

**Resource type:** `aws_db_instance`

**Description:**
RDS instance is not configured for Multi-AZ deployment. Single-AZ deployments are vulnerable to availability zone outages, which can cause extended downtime. Multi-AZ provides a synchronous standby replica in a different availability zone with automatic failover, improving both availability and durability.

**Remediation:**

Enable Multi-AZ deployment for production databases to ensure high availability and automatic failover.

```hcl
resource "aws_db_instance" "example" {
  identifier     = "my-database"
  engine         = "postgres"
  engine_version = "15.4"
  instance_class = "db.t3.medium"

  allocated_storage = 20
  multi_az          = true

  db_name  = "mydb"
  username = "admin"
  password = var.db_password

  skip_final_snapshot = false
}
```
