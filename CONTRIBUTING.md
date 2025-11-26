# Contributing to terraform-provider-googlewallet

Thank you for your interest in contributing to the Google Wallet
Terraform Provider!

## Development Requirements

- Go 1.23+
- Terraform 1.0+ (or OpenTofu 1.0+)
- Google Cloud service account with Wallet Admin role
- Valid API credentials for testing

## Getting Started

1. Fork and clone the repository
2. Install dependencies:
    ```bash
    go mod download
    ```
3. Build the provider:
    ```bash
    go build -o terraform-provider-googlewallet
    ```

## Development Workflow

### 1. Make Your Changes

- Follow the existing code patterns
- Use the Terraform Plugin Framework (not SDK v2)
- Ensure thread-safety for parallel execution
- Add appropriate logging with `tflog`

### 2. Write Tests

- Add unit tests for new functionality
- Update acceptance tests if adding resources/data sources
- Ensure existing tests still pass

### 3. Run Quality Checks

```bash
# Format code
go fmt ./...

# Run linter
golangci-lint run ./...

# Run unit tests
go test ./...

# Run acceptance tests (requires API credentials)
TF_ACC=1 go test ./... -timeout 30m
```

### 4. Update Documentation

- Add/update templates in `templates/` directory
- Run `go generate ./...` to regenerate documentation
- Update examples in `examples/` directory

## Code Style

- Follow standard Go conventions
- Use meaningful variable and function names
- Add comments for exported functions and types
- Keep functions focused and testable

## Testing

### Unit Tests

- Test individual functions and methods
- Mock external API calls
- Aim for high code coverage

### Acceptance Tests

- Test actual API interactions
- Use `resource.Test` framework
- Clean up resources after tests

Set these environment variables for acceptance tests:

```bash
export GOOGLEWALLET_CREDENTIALS="$(cat service-account.json)"
export GOOGLEWALLET_TEST_SA_EMAIL="your-sa@project.iam.gserviceaccount.com"
export GOOGLEWALLET_TEST_SECONDARY_EMAIL="another-user@example.com"
export TF_ACC=1
```

## Adding a New Resource

1. **Define Types** in `internal/provider/models.go`
2. **Implement Resource** in `internal/provider/<resource>_resource.go`
3. **Add Tests** in `internal/provider/<resource>_resource_test.go`
4. **Register Resource** in `provider.go`
5. **Document** in `templates/resources/<resource>.md.tmpl`

## Commit Messages

Follow conventional commit format:

- `feat:` New features
- `fix:` Bug fixes
- `docs:` Documentation changes
- `test:` Test additions/changes
- `refactor:` Code refactoring
- `chore:` Maintenance tasks

## Pull Request Process

1. Create a feature branch from `main`
2. Make your changes following the guidelines above
3. Ensure all tests pass and code is properly formatted
4. Update the CHANGELOG.md with your changes
5. Create a pull request with a clear description
6. Address any review feedback

## Google Wallet API Notes

- **Issuer IDs**: 19-digit integers stored as strings to avoid overflow
- **Soft Delete**: Issuers cannot be deleted via API
- **Permissions**: Authoritative resource - replaces all permissions
- **Role Case**: API returns lowercase, provider uses uppercase

## Questions?

Open an issue for discussion. Thank you for contributing!
