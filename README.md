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

### To use Cloudrift:

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

Repeat the same process for `vuln-export` or any other Terraform-based repo.

---

## ✨ Features (coming soon)
- Detect drift between Terraform and live AWS state
- Catch unmanaged or deleted cloud resources
- Integrate into CI/CD pipelines
- Slack/email notifications
- Simple CLI and JSON output

---

## 📦 Installation

### Install via Go (Recommended for developers)
```bash
go install github.com/inayathulla/cloudrift@latest
```
Make sure your `$GOPATH/bin` is in your `PATH`. Add this to your `~/.zshrc` or `~/.bashrc` if needed:
```bash
export PATH="$HOME/go/bin:$PATH"
```
Then reload your terminal:
```bash
source ~/.zshrc
```
Now run:
```bash
cloudrift scan --config=config/cloudrift.yml
```

---
## 🤝 Contributing

### 🧪 Development Guidelines
- Use clear commit messages (e.g., feat: add EC2 drift detection)
- Keep code modular (e.g., one service = one detector)
- Follow Go formatting: go fmt ./...
- Add unit tests for new components

### 📁 Code Structure
    cmd/              ← CLI entrypoint 
    internal/
        aws/          ← AWS fetchers
        detector/     ← Drift comparison logic
        parser/       ← Terraform plan parsing
        models/       ← Shared structs

### 🧪 Testing
Before submitting a PR:
```bash
go test ./...
```
### 📬 Submitting a Pull Request
- Push your branch
- Open a pull request to main
- Briefly explain what your change does and why
- We'll review your PR and respond quickly 🙌

### 🙋‍♂️ Questions or Feedback?
Open an issue or reach out via GitHub Discussions

---

## 📝 License
Apache License 2.0
