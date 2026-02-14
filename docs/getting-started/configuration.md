# Configuration

Cloudrift uses a YAML configuration file to define AWS credentials, region, and scan parameters.

## Config File Format

```yaml
aws_profile: default
region: us-east-1
plan_path: ./plan.json
```

## Fields

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `aws_profile` | string | yes | `default` | AWS credentials profile name from `~/.aws/credentials` |
| `region` | string | yes | `us-east-1` | AWS region to scan |
| `plan_path` | string | yes | — | Path to Terraform plan JSON file |

---

## Service-Specific Configs

Each AWS service needs its own Terraform plan file. Create separate configs per service:

=== "S3 (cloudrift.yml)"

    ```yaml
    aws_profile: default
    region: us-east-1
    plan_path: ./plan.json
    ```

=== "EC2 (cloudrift-ec2.yml)"

    ```yaml
    aws_profile: default
    region: us-east-1
    plan_path: ./ec2-plan.json
    ```

Use the `--config` flag to select the config:

```bash
cloudrift scan --config=cloudrift-ec2.yml --service=ec2
```

---

## Generating the Plan File

Cloudrift requires a Terraform plan in JSON format. Generate it with:

```bash
# 1. Create the binary plan
terraform plan -out=tfplan

# 2. Convert to JSON
terraform show -json tfplan > plan.json
```

!!! tip "Plan file scope"
    The plan file should contain all the resources you want to scan. Cloudrift extracts resources matching the selected `--service` type (e.g., `aws_s3_bucket` for S3, `aws_instance` for EC2).

### Plan File Structure

Cloudrift reads resources from the `resource_changes[].change.after` path in the plan JSON. Each resource change must contain the resource type, address, and planned attribute values.

---

## Environment Variables

AWS credentials can also be configured via environment variables:

```bash
export AWS_PROFILE=production
export AWS_REGION=eu-west-1
```

These are picked up by the AWS SDK automatically and override the config file values.

---

## Config File Locations

By default, Cloudrift looks for `cloudrift.yml` in the current working directory. Override with:

```bash
cloudrift scan --config=/path/to/config/cloudrift.yml --service=s3
```
