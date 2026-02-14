# Contributing

Thank you for your interest in contributing to Cloudrift! This guide covers the workflow for submitting changes.

## Getting Started

1. **Fork** the repository on GitHub
2. **Clone** your fork locally:

    ```bash
    git clone https://github.com/YOUR_USERNAME/cloudrift.git
    cd cloudrift
    ```

3. **Create a branch** for your feature or fix:

    ```bash
    git checkout -b feature/my-feature
    ```

4. **Build and test** to verify everything works:

    ```bash
    go build -o cloudrift main.go
    go test ./...
    ```

---

## Development Workflow

### Code Changes

1. Make your changes in the appropriate package
2. Format your code: `go fmt ./...`
3. Run tests: `go test ./...`
4. Build: `go build -o cloudrift main.go`

### Adding Features

- **New AWS service** — See [Adding Services](development/adding-services.md)
- **New OPA policy** — See [Adding Policies](development/adding-policies.md)
- **New output format** — Implement the `Formatter` interface in `internal/output/`

---

## Code Standards

### Go Conventions

- Follow [Effective Go](https://go.dev/doc/effective_go) guidelines
- Use `go fmt` for formatting (enforced in CI)
- Keep packages focused and small
- Use descriptive variable names
- Add comments for exported functions

### Testing

- Write tests for all new functionality
- Place tests in `tests/internal/` mirroring the package structure
- Use `testify` for assertions
- Aim for table-driven tests where applicable

### Policy Conventions

- Follow the existing `.rego` file patterns
- Include all metadata fields (`policy_id`, `policy_name`, `msg`, `severity`, `remediation`, `category`, `frameworks`)
- Place in the correct category directory (`security/`, `tagging/`, `cost/`)

---

## Commit Messages

Use clear, concise commit messages:

```
Add RDS drift detection support

Implements drift detection for aws_db_instance resources including
storage encryption, public access, and backup retention checks.
```

- Use imperative mood ("Add" not "Added")
- First line: summary under 72 characters
- Optional body: explain _why_, not _what_

---

## Pull Request Process

1. Ensure all tests pass: `go test ./...`
2. Ensure code is formatted: `go fmt ./...`
3. Push your branch: `git push origin feature/my-feature`
4. Open a pull request against `main`
5. Describe what changed and why
6. Link any related issues

### PR Checklist

- [ ] Tests pass (`go test ./...`)
- [ ] Code is formatted (`go fmt ./...`)
- [ ] New features have tests
- [ ] New policies include all metadata fields
- [ ] Documentation updated if needed

---

## Reporting Issues

Found a bug or have a feature request? [Open an issue](https://github.com/inayathulla/cloudrift/issues) on GitHub.

Include:

- **Expected behavior** vs **actual behavior**
- **Steps to reproduce**
- **Cloudrift version** (`cloudrift --help`)
- **Go version** (`go version`)
- **OS** and architecture

---

## License

By contributing, you agree that your contributions will be licensed under the [Apache License 2.0](https://github.com/inayathulla/cloudrift/blob/main/LICENSE).
