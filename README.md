# cloudrift
Detect drift. Defend cloud.

[![Docker Pulls](https://img.shields.io/docker/pulls/inayathulla/cloudrift)](https://hub.docker.com/r/inayathulla/cloudrift)


🔍 **Cloudrift** is an open-source cloud drift detection tool that helps you identify when your cloud infrastructure no longer matches your Infrastructure-as-Code (IaC) — before it causes a security or compliance incident.

## 🚀 Quick Start

### Clone the repository
```bash
git clone https://github.com/inayathulla/cloudrift.git
cd cloudrift
```


### 🔁 Using Cloudrift with your own Terraform projects

Cloudrift is designed to be used by developers to detect cloud resource drift in their own Terraform-based infrastructure projects.

### ✅ Example: compliance-export or vuln-export projects

Assume you have Terraform code stored in your repositories:

```
~/projects/
├── compliance-export/
│   ├── main.tf
│   ├── variables.tf
│   └── ...
└── vuln-export/
    ├── main.tf
    └── ...
```

To use Cloudrift:
## 🚀 Quick Start

### Clone the repository
```bash
git clone https://github.com/inayathulla/cloudrift.git
cd cloudrift
```
### 🔁 Using Cloudrift with your own Terraform projects

Cloudrift is designed to be used by developers to detect cloud resource drift in their own Terraform-based infrastructure projects.

### ✅ Example: compliance-export or vuln-export projects

Assume you have Terraform code stored in your repositories:

```
~/projects/
├── compliance-export/
│   ├── main.tf
│   ├── variables.tf
│   └── ...
└── vuln-export/
    ├── main.tf
    └── ...
```

To use Cloudrift:

### 1. Navigate to your Terraform project
```bash
cd ~/projects/compliance-export
```

### 2. Generate a Terraform plan
```bash
terraform init
terraform plan -out=compliance.binary
terraform show -json compliance.binary > compliance_plan.json
```

### 3. Update Cloudrift config (cloudrift.yaml)
```yaml
aws_profile: default
region: us-east-1
plan_path: ~/projects/compliance-export/compliance_plan.json
```

### 4. Run Cloudrift to check for drift
```bash
cd ~/projects/cloudrift
go run main.go scan
```

Repeat the same process for `vuln-export` or any other Terraform-based repo.

---

## ✨ Features (coming soon)
- Detect drift between Terraform and live AWS state
- Catch unmanaged or deleted cloud resources
- Integrate into CI/CD pipelines
- Slack/email notifications
- Simple CLI and JSON output

## 📝 License
Apache License 2.0

---

## 🤝 Contributing
Contributions welcome! Please open issues and PRs to improve Cloudrift.