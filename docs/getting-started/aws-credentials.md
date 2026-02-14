# AWS Credentials

Cloudrift requires read-only access to AWS resources for drift detection.

## Required IAM Permissions

### Minimum Policy

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "s3:ListAllMyBuckets",
        "s3:GetBucketLocation",
        "s3:GetBucketVersioning",
        "s3:GetBucketEncryption",
        "s3:GetBucketLogging",
        "s3:GetBucketTagging",
        "s3:GetBucketPublicAccessBlock",
        "s3:GetLifecycleConfiguration",
        "s3:GetBucketAcl",
        "ec2:DescribeInstances",
        "ec2:DescribeTags",
        "sts:GetCallerIdentity"
      ],
      "Resource": "*"
    }
  ]
}
```

### Per-Service Breakdown

=== "S3"

    | Permission | Purpose |
    |-----------|---------|
    | `s3:ListAllMyBuckets` | Enumerate buckets |
    | `s3:GetBucketLocation` | Determine bucket region |
    | `s3:GetBucketVersioning` | Check versioning status |
    | `s3:GetBucketEncryption` | Check encryption config |
    | `s3:GetBucketLogging` | Check access logging |
    | `s3:GetBucketTagging` | Read resource tags |
    | `s3:GetBucketPublicAccessBlock` | Check public access settings |
    | `s3:GetLifecycleConfiguration` | Read lifecycle rules |
    | `s3:GetBucketAcl` | Read bucket ACL |

=== "EC2"

    | Permission | Purpose |
    |-----------|---------|
    | `ec2:DescribeInstances` | List and describe EC2 instances |
    | `ec2:DescribeTags` | Read instance tags |

=== "Common"

    | Permission | Purpose |
    |-----------|---------|
    | `sts:GetCallerIdentity` | Validate credentials and display account info |

---

## AWS Profile Configuration

Configure profiles in `~/.aws/credentials`:

```ini
[default]
aws_access_key_id = AKIAIOSFODNN7EXAMPLE
aws_secret_access_key = wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY

[production]
aws_access_key_id = AKIAI44QH8DHBEXAMPLE
aws_secret_access_key = je7MtGbClwBF/2Zp9Utk/h3yCo8nvbEXAMPLEKEY
```

Reference the profile in your config:

```yaml
aws_profile: production
region: us-east-1
plan_path: ./plan.json
```

---

## Environment Variables

AWS SDK v2 supports standard environment variables:

| Variable | Description |
|----------|-------------|
| `AWS_PROFILE` | Profile name from credentials file |
| `AWS_REGION` | AWS region |
| `AWS_ACCESS_KEY_ID` | Access key (overrides profile) |
| `AWS_SECRET_ACCESS_KEY` | Secret key (overrides profile) |
| `AWS_SESSION_TOKEN` | Session token for temporary credentials |

---

## IAM Roles (Assumed Roles)

For cross-account scanning, configure a role in `~/.aws/config`:

```ini
[profile cross-account]
role_arn = arn:aws:iam::ACCOUNT_ID:role/CloudriftReadOnly
source_profile = default
region = us-east-1
```

Then reference it:

```yaml
aws_profile: cross-account
```

!!! warning "Permissions"
    Cloudrift only needs **read-only** access. Never grant write or delete permissions to the scanning role.
