# Policy Overview

Cloudrift ships with **49 built-in OPA policies** covering security, tagging, and cost optimization across 13 AWS resource types.

## Summary

| ID | Policy Name | Severity | Category | Frameworks |
|----|------------|----------|----------|------------|
| [S3-001](s3.md#s3-001) | S3 Encryption Required | <span class="severity-high">HIGH</span> | security | HIPAA, PCI DSS, ISO 27001, GDPR, SOC 2 |
| [S3-002](s3.md#s3-002) | S3 KMS Encryption Recommended | <span class="severity-low">LOW</span> | security | HIPAA, PCI DSS, SOC 2 |
| [S3-003](s3.md#s3-003) | S3 Block Public ACLs | <span class="severity-high">HIGH</span> | security | HIPAA, GDPR, PCI DSS, ISO 27001, SOC 2 |
| [S3-004](s3.md#s3-004) | S3 Block Public Policy | <span class="severity-high">HIGH</span> | security | HIPAA, GDPR, PCI DSS, ISO 27001, SOC 2 |
| [S3-005](s3.md#s3-005) | S3 Ignore Public ACLs | <span class="severity-high">HIGH</span> | security | HIPAA, GDPR, PCI DSS, ISO 27001, SOC 2 |
| [S3-006](s3.md#s3-006) | S3 Restrict Public Buckets | <span class="severity-high">HIGH</span> | security | HIPAA, GDPR, PCI DSS, ISO 27001, SOC 2 |
| [S3-007](s3.md#s3-007) | S3 No Public Read ACL | <span class="severity-critical">CRITICAL</span> | security | HIPAA, GDPR, PCI DSS, ISO 27001, SOC 2 |
| [S3-008](s3.md#s3-008) | S3 No Public Read-Write ACL | <span class="severity-critical">CRITICAL</span> | security | HIPAA, GDPR, PCI DSS, ISO 27001, SOC 2 |
| [S3-009](s3.md#s3-009) | S3 Versioning Recommended | <span class="severity-medium">MEDIUM</span> | security | ISO 27001, SOC 2 |
| [EC2-001](ec2.md#ec2-001) | EC2 IMDSv2 Required | <span class="severity-medium">MEDIUM</span> | security | PCI DSS, ISO 27001, SOC 2 |
| [EC2-002](ec2.md#ec2-002) | EC2 Root Volume Encryption | <span class="severity-high">HIGH</span> | security | HIPAA, PCI DSS, ISO 27001, GDPR, SOC 2 |
| [EC2-003](ec2.md#ec2-003) | EC2 Public IP Warning | <span class="severity-medium">MEDIUM</span> | security | PCI DSS, ISO 27001, SOC 2 |
| [EC2-005](ec2.md#ec2-005) | EC2 Large Instance Review | <span class="severity-medium">MEDIUM</span> | cost | — |
| [SG-001](security-groups.md#sg-001) | No Unrestricted SSH Access | <span class="severity-critical">CRITICAL</span> | security | PCI DSS, ISO 27001, SOC 2 |
| [SG-002](security-groups.md#sg-002) | No Unrestricted RDP Access | <span class="severity-critical">CRITICAL</span> | security | PCI DSS, ISO 27001, SOC 2 |
| [SG-003](security-groups.md#sg-003) | No Unrestricted All Ports Access | <span class="severity-critical">CRITICAL</span> | security | PCI DSS, ISO 27001, SOC 2 |
| [SG-004](security-groups.md#sg-004) | Database Port Public Exposure | <span class="severity-high">HIGH</span> | security | HIPAA, PCI DSS, ISO 27001, SOC 2 |
| [RDS-001](rds.md#rds-001) | RDS Storage Encryption Required | <span class="severity-high">HIGH</span> | security | HIPAA, PCI DSS, ISO 27001, GDPR, SOC 2 |
| [RDS-002](rds.md#rds-002) | RDS No Public Access | <span class="severity-critical">CRITICAL</span> | security | HIPAA, PCI DSS, ISO 27001, GDPR, SOC 2 |
| [RDS-003](rds.md#rds-003) | RDS Backup Retention Period | <span class="severity-medium">MEDIUM</span> | security | HIPAA, ISO 27001, SOC 2 |
| [RDS-004](rds.md#rds-004) | RDS Deletion Protection | <span class="severity-medium">MEDIUM</span> | security | ISO 27001, SOC 2 |
| [RDS-005](rds.md#rds-005) | RDS Multi-AZ Recommended | <span class="severity-low">LOW</span> | security | HIPAA, ISO 27001, SOC 2 |
| [IAM-001](iam.md#iam-001) | No Wildcard IAM Actions | <span class="severity-critical">CRITICAL</span> | security | HIPAA, PCI DSS, ISO 27001, GDPR, SOC 2 |
| [IAM-002](iam.md#iam-002) | No Inline Policies on Users | <span class="severity-medium">MEDIUM</span> | security | PCI DSS, ISO 27001, SOC 2 |
| [IAM-003](iam.md#iam-003) | IAM Role Trust Too Broad | <span class="severity-high">HIGH</span> | security | PCI DSS, ISO 27001, SOC 2 |
| [CT-001](cloudtrail.md#ct-001) | CloudTrail KMS Encryption | <span class="severity-high">HIGH</span> | security | HIPAA, PCI DSS, ISO 27001, GDPR, SOC 2 |
| [CT-002](cloudtrail.md#ct-002) | CloudTrail Log File Validation | <span class="severity-medium">MEDIUM</span> | security | PCI DSS, ISO 27001, SOC 2 |
| [CT-003](cloudtrail.md#ct-003) | CloudTrail Multi-Region | <span class="severity-medium">MEDIUM</span> | security | HIPAA, PCI DSS, ISO 27001, SOC 2 |
| [KMS-001](kms.md#kms-001) | KMS Key Rotation Enabled | <span class="severity-high">HIGH</span> | security | HIPAA, PCI DSS, ISO 27001, SOC 2 |
| [KMS-002](kms.md#kms-002) | KMS Key Deletion Window | <span class="severity-medium">MEDIUM</span> | security | ISO 27001, SOC 2 |
| [ELB-001](elb.md#elb-001) | ALB Access Logging | <span class="severity-medium">MEDIUM</span> | security | HIPAA, PCI DSS, ISO 27001, SOC 2 |
| [ELB-002](elb.md#elb-002) | ALB HTTPS Listener Required | <span class="severity-high">HIGH</span> | security | HIPAA, PCI DSS, ISO 27001, GDPR, SOC 2 |
| [ELB-003](elb.md#elb-003) | ALB Deletion Protection | <span class="severity-medium">MEDIUM</span> | security | ISO 27001, SOC 2 |
| [EBS-001](ebs.md#ebs-001) | EBS Volume Encryption | <span class="severity-high">HIGH</span> | security | HIPAA, PCI DSS, ISO 27001, GDPR, SOC 2 |
| [EBS-002](ebs.md#ebs-002) | EBS Snapshot Encryption | <span class="severity-high">HIGH</span> | security | HIPAA, PCI DSS, ISO 27001, GDPR |
| [LAMBDA-001](lambda.md#lambda-001) | Lambda Tracing Enabled | <span class="severity-medium">MEDIUM</span> | security | SOC 2, ISO 27001 |
| [LAMBDA-002](lambda.md#lambda-002) | Lambda VPC Configuration | <span class="severity-medium">MEDIUM</span> | security | HIPAA, PCI DSS, ISO 27001 |
| [LOG-001](logging.md#log-001) | CloudWatch Log Group Encryption | <span class="severity-medium">MEDIUM</span> | security | HIPAA, PCI DSS, GDPR, SOC 2 |
| [LOG-002](logging.md#log-002) | CloudWatch Log Retention | <span class="severity-medium">MEDIUM</span> | security | HIPAA, GDPR, SOC 2, ISO 27001 |
| [VPC-001](vpc.md#vpc-001) | Default Security Group Restrict All | <span class="severity-high">HIGH</span> | security | PCI DSS, ISO 27001, SOC 2 |
| [VPC-002](vpc.md#vpc-002) | Subnet No Auto-Assign Public IP | <span class="severity-medium">MEDIUM</span> | security | PCI DSS, ISO 27001 |
| [SECRET-001](secrets.md#secret-001) | Secrets Manager KMS Encryption | <span class="severity-medium">MEDIUM</span> | security | HIPAA, PCI DSS, GDPR, SOC 2 |
| [SECRET-002](secrets.md#secret-002) | Secrets Rotation Enabled | <span class="severity-medium">MEDIUM</span> | security | PCI DSS, ISO 27001, SOC 2 |
| [TAG-001](tagging.md#tag-001) | Environment Tag Required | <span class="severity-medium">MEDIUM</span> | tagging | SOC 2 |
| [TAG-002](tagging.md#tag-002) | Owner Tag Recommended | <span class="severity-low">LOW</span> | tagging | — |
| [TAG-003](tagging.md#tag-003) | Project Tag Recommended | <span class="severity-low">LOW</span> | tagging | — |
| [TAG-004](tagging.md#tag-004) | Name Tag Recommended | <span class="severity-low">LOW</span> | tagging | — |
| [COST-002](cost.md#cost-002) | Very Large Instance Size | <span class="severity-low">LOW</span> | cost | — |
| [COST-003](cost.md#cost-003) | Previous Generation Instance | <span class="severity-low">LOW</span> | cost | — |

## Severity Distribution

| Severity | Count | Description |
|----------|-------|-------------|
| <span class="severity-critical">CRITICAL</span> | 7 | Immediate security risk, must fix |
| <span class="severity-high">HIGH</span> | 15 | Significant security concern |
| <span class="severity-medium">MEDIUM</span> | 21 | Recommended improvement |
| <span class="severity-low">LOW</span> | 6 | Best practice advisory |

## Resource Coverage

Policies are evaluated for these Terraform resource types:

`aws_s3_bucket` `aws_instance` `aws_security_group` `aws_db_instance` `aws_iam_policy` `aws_iam_role` `aws_iam_user_policy` `aws_cloudtrail` `aws_kms_key` `aws_lb` `aws_lb_listener` `aws_ebs_volume` `aws_ebs_snapshot_copy` `aws_lambda_function` `aws_cloudwatch_log_group` `aws_default_security_group` `aws_subnet` `aws_secretsmanager_secret`
