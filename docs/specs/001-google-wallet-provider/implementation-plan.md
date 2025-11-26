# Implementation Plan

## Validation Checklist

- [x] All specification file paths are correct and exist
- [x] Context priming section is complete
- [x] All implementation phases are defined
- [x] Each phase follows TDD: Prime → Test → Implement → Validate
- [x] Dependencies between phases are clear (no circular dependencies)
- [x] Parallel work is properly tagged with `[parallel: true]`
- [x] Activity hints provided for specialist selection `[activity: type]`
- [x] Every phase references relevant SDD sections
- [x] Every test references PRD acceptance criteria
- [x] Integration & E2E tests defined in final phase
- [x] Project commands match actual project setup
- [x] A developer could follow this plan independently

---

## Specification Compliance Guidelines

### How to Ensure Specification Adherence

1. **Before Each Phase**: Complete the Pre-Implementation Specification Gate
2. **During Implementation**: Reference specific SDD sections in each task
3. **After Each Task**: Run Specification Compliance checks
4. **Phase Completion**: Verify all specification requirements are met

### Deviation Protocol

If implementation cannot follow specification exactly:
1. Document the deviation and reason
2. Get approval before proceeding
3. Update SDD if the deviation is an improvement
4. Never deviate without documentation

## Metadata Reference

- `[parallel: true]` - Tasks that can run concurrently
- `[component: component-name]` - For multi-component features
- `[ref: document/section; lines: 1, 2-3]` - Links to specifications, patterns, or interfaces and (if applicable) line(s)
- `[activity: type]` - Activity hint for specialist agent selection

---

## Context Priming

*GATE: You MUST fully read all files mentioned in this section before starting any implementation.*

**Specification**:

- `docs/specs/001-google-wallet-provider/product-requirements.md` - Product Requirements
- `docs/specs/001-google-wallet-provider/solution-design.md` - Solution Design

**Supporting Documentation**:

- `docs/patterns/terraform-plugin-framework.md` - Terraform Plugin Framework Patterns
- `docs/interfaces/google-wallet-api.md` - Google Wallet API Contract
- `DESIGN.md` - Original Design Document (reference patterns)

**Key Design Decisions**:

- **ADR-001**: Use Terraform Plugin Framework (not SDK v2) for modern provider development
- **ADR-002**: Issuer soft delete strategy - API doesn't support deletion, so remove from state with warning
- **ADR-003**: Authoritative permissions pattern - full replacement model matching API behavior
- **ADR-004**: Credential flexibility - support file path OR JSON content with environment variable fallback
- **ADR-005**: String-based Issuer IDs - use `types.String` for 19-digit IDs to avoid int64 overflow
- **ADR-006**: Minimalist first release - focus on Issuers and Permissions, defer pass classes to v2

**Implementation Context**:

- Commands to run:
  - `task test` - Run unit tests
  - `task testacc` - Run acceptance tests (requires `GOOGLEWALLET_CREDENTIALS`)
  - `task build` - Build provider binary
  - `task lint` - Run golangci-lint
  - `task generate` - Generate documentation
  - `task install` - Install provider locally
- Patterns to follow: `[ref: docs/patterns/terraform-plugin-framework.md]`
- Interfaces to implement: `[ref: docs/interfaces/google-wallet-api.md]`

---

## Implementation Phases

### Phase 1: Project Scaffolding & Tooling

- [x] T1 Phase 1 - Project Foundation `[activity: scaffold-project]`

    - [x] T1.1 Prime Context
        - [x] T1.1.1 Read project structure requirements `[ref: solution-design.md; lines: 50-80]`
        - [x] T1.1.2 Review reference provider structure `[ref: terraform-plugin-framework.md; lines: 11-31]`

    - [x] T1.2 Create Project Structure `[activity: scaffold-project]`
        - [x] T1.2.1 Create `go.mod` with module path `github.com/truetickets/terraform-provider-googlewallet`
        - [x] T1.2.2 Create `main.go` entry point with provider factory
        - [x] T1.2.3 Create `internal/provider/` directory structure
        - [x] T1.2.4 Create empty placeholder files for all resources/data sources

    - [x] T1.3 Configure Tooling `[activity: configure-tooling]`
        - [x] T1.3.1 Create `Taskfile.yml` with standard provider commands
        - [x] T1.3.2 Create `GNUmakefile` for backwards compatibility
        - [x] T1.3.3 Create `.golangci.yml` with linting configuration
        - [x] T1.3.4 Create `.goreleaser.yml` for release automation
        - [x] T1.3.5 Create `.gitignore` for Go and Terraform artifacts

    - [x] T1.4 Configure Documentation Generation `[activity: configure-tooling]`
        - [x] T1.4.1 Create `templates/` directory with provider, resource, data-source templates
        - [x] T1.4.2 Create `tools/tools.go` for tfplugindocs dependency

    - [x] T1.5 Validate
        - [x] T1.5.1 Verify `go mod tidy` succeeds
        - [x] T1.5.2 Verify `task lint` passes (empty code)
        - [x] T1.5.3 Verify project compiles with `task build`

---

### Phase 2: Provider Configuration & Client

- [ ] T2 Phase 2 - Provider Core `[activity: implement-provider]`

    - [ ] T2.1 Prime Context
        - [ ] T2.1.1 Read provider schema requirements `[ref: solution-design.md; lines: 110-180]`
        - [ ] T2.1.2 Read authentication requirements `[ref: google-wallet-api.md; lines: 179-204]`
        - [ ] T2.1.3 Review provider patterns `[ref: terraform-plugin-framework.md; lines: 36-113]`

    - [ ] T2.2 Write Tests `[activity: write-tests]`
        - [ ] T2.2.1 Unit test: Provider metadata returns "googlewallet" `[ref: PRD Feature 1]`
        - [ ] T2.2.2 Unit test: Schema includes `credentials` attribute as optional, sensitive
        - [ ] T2.2.3 Unit test: Configure reads `GOOGLEWALLET_CREDENTIALS` env var fallback
        - [ ] T2.2.4 Unit test: Configure reads `GOOGLE_CREDENTIALS` env var as secondary fallback
        - [ ] T2.2.5 Unit test: Missing credentials produces attribute error
        - [ ] T2.2.6 Unit test: File path credentials are resolved
        - [ ] T2.2.7 Unit test: JSON content credentials are accepted

    - [ ] T2.3 Implement Provider `[activity: implement-code]`
        - [ ] T2.3.1 Create `internal/provider/provider.go` with `GoogleWalletProvider` struct
        - [ ] T2.3.2 Implement `Metadata()` method returning "googlewallet" type name
        - [ ] T2.3.3 Implement `Schema()` method with credentials attribute
        - [ ] T2.3.4 Implement `Configure()` method with env var fallback chain
        - [ ] T2.3.5 Implement `Resources()` returning empty slice (placeholder)
        - [ ] T2.3.6 Implement `DataSources()` returning empty slice (placeholder)

    - [ ] T2.4 Implement Client Wrapper `[activity: implement-code]`
        - [ ] T2.4.1 Create `internal/provider/client.go` with `Client` struct
        - [ ] T2.4.2 Implement `NewClient(ctx, credentials)` factory function
        - [ ] T2.4.3 Handle credentials as file path (detect with file existence check)
        - [ ] T2.4.4 Handle credentials as raw JSON content
        - [ ] T2.4.5 Create `walletobjects.Service` instance with appropriate options
        - [ ] T2.4.6 Implement wrapper methods for Issuer operations (stubs)
        - [ ] T2.4.7 Implement wrapper methods for Permissions operations (stubs)

    - [ ] T2.5 Validate
        - [ ] T2.5.1 All unit tests pass: `task test`
        - [ ] T2.5.2 Lint passes: `task lint`
        - [ ] T2.5.3 Provider compiles and loads: `task build`

---

### Phase 3: Issuer Resource & Data Sources

- [ ] T3 Phase 3 - Issuer Management

    - [ ] T3.1 Issuer Resource `[component: issuer-resource]`

        - [ ] T3.1.1 Prime Context
            - [ ] T3.1.1.1 Read Issuer resource schema `[ref: solution-design.md; lines: 190-280]`
            - [ ] T3.1.1.2 Read Issuer API contract `[ref: google-wallet-api.md; lines: 13-68]`
            - [ ] T3.1.1.3 Review resource patterns `[ref: terraform-plugin-framework.md; lines: 117-323]`

        - [ ] T3.1.2 Write Tests `[activity: write-tests]`
            - [ ] T3.1.2.1 Unit test: Schema has required `name` and `contact_info` attributes
            - [ ] T3.1.2.2 Unit test: Schema has computed `issuer_id` with UseStateForUnknown
            - [ ] T3.1.2.3 Unit test: Schema has optional `homepage_url`
            - [ ] T3.1.2.4 Unit test: contact_info nested block validates email format
            - [ ] T3.1.2.5 Acceptance test: Create Issuer with minimal config `[ref: PRD Feature 2]`
            - [ ] T3.1.2.6 Acceptance test: Update Issuer name succeeds
            - [ ] T3.1.2.7 Acceptance test: Import existing Issuer by ID
            - [ ] T3.1.2.8 Acceptance test: Delete removes from state (verify warning logged)

        - [ ] T3.1.3 Implement Resource `[activity: implement-code]`
            - [ ] T3.1.3.1 Create `internal/provider/issuer_resource.go`
            - [ ] T3.1.3.2 Define `IssuerResourceModel` struct with tfsdk tags
            - [ ] T3.1.3.3 Define `ContactInfoModel` struct for nested block
            - [ ] T3.1.3.4 Implement `Schema()` with all attributes and validators
            - [ ] T3.1.3.5 Implement `Create()` - call API, store computed issuer_id
            - [ ] T3.1.3.6 Implement `Read()` - fetch from API, handle not found
            - [ ] T3.1.3.7 Implement `Update()` - use PATCH for partial updates
            - [ ] T3.1.3.8 Implement `Delete()` - soft delete with warning log
            - [ ] T3.1.3.9 Implement `ImportState()` - passthrough ID

        - [ ] T3.1.4 Implement Client Methods `[activity: implement-code]`
            - [ ] T3.1.4.1 Implement `client.CreateIssuer()` wrapping API insert
            - [ ] T3.1.4.2 Implement `client.GetIssuer()` wrapping API get
            - [ ] T3.1.4.3 Implement `client.PatchIssuer()` wrapping API patch
            - [ ] T3.1.4.4 Implement `isNotFoundError()` helper for 404 detection

        - [ ] T3.1.5 Validate
            - [ ] T3.1.5.1 Unit tests pass
            - [ ] T3.1.5.2 Acceptance tests pass (with real API)
            - [ ] T3.1.5.3 Lint passes

    - [ ] T3.2 Issuer Data Source `[parallel: true]` `[component: issuer-datasource]`

        - [ ] T3.2.1 Prime Context
            - [ ] T3.2.1.1 Read data source patterns `[ref: terraform-plugin-framework.md; lines: 327-354]`
            - [ ] T3.2.1.2 Read PRD requirements `[ref: PRD Feature 4]`

        - [ ] T3.2.2 Write Tests `[activity: write-tests]`
            - [ ] T3.2.2.1 Unit test: Schema requires `issuer_id` input
            - [ ] T3.2.2.2 Unit test: Schema outputs name, homepage_url, contact_info
            - [ ] T3.2.2.3 Acceptance test: Read existing Issuer by ID
            - [ ] T3.2.2.4 Acceptance test: Error on non-existent Issuer

        - [ ] T3.2.3 Implement Data Source `[activity: implement-code]`
            - [ ] T3.2.3.1 Create `internal/provider/issuer_data_source.go`
            - [ ] T3.2.3.2 Define `IssuerDataSourceModel` struct
            - [ ] T3.2.3.3 Implement `Schema()` with required issuer_id, computed outputs
            - [ ] T3.2.3.4 Implement `Read()` - fetch from API and populate state

        - [ ] T3.2.4 Validate
            - [ ] T3.2.4.1 Unit and acceptance tests pass
            - [ ] T3.2.4.2 Lint passes

    - [ ] T3.3 Issuers List Data Source `[parallel: true]` `[component: issuers-datasource]`

        - [ ] T3.3.1 Prime Context
            - [ ] T3.3.1.1 Read list data source patterns `[ref: terraform-plugin-framework.md; lines: 356-379]`
            - [ ] T3.3.1.2 Read PRD requirements `[ref: PRD Feature 5]`

        - [ ] T3.3.2 Write Tests `[activity: write-tests]`
            - [ ] T3.3.2.1 Unit test: Schema has computed `issuers` list attribute
            - [ ] T3.3.2.2 Unit test: Each issuer has issuer_id, name, homepage_url
            - [ ] T3.3.2.3 Acceptance test: List returns accessible Issuers
            - [ ] T3.3.2.4 Acceptance test: Empty list when no Issuers accessible

        - [ ] T3.3.3 Implement Data Source `[activity: implement-code]`
            - [ ] T3.3.3.1 Create `internal/provider/issuers_data_source.go`
            - [ ] T3.3.3.2 Define `IssuersDataSourceModel` with list of `IssuerModel`
            - [ ] T3.3.3.3 Implement `Schema()` with computed issuers list
            - [ ] T3.3.3.4 Implement `Read()` - list from API and populate state

        - [ ] T3.3.4 Implement Client Method `[activity: implement-code]`
            - [ ] T3.3.4.1 Implement `client.ListIssuers()` wrapping API list

        - [ ] T3.3.5 Validate
            - [ ] T3.3.5.1 Unit and acceptance tests pass
            - [ ] T3.3.5.2 Lint passes

---

### Phase 4: Permissions Resource & Data Source

- [ ] T4 Phase 4 - Permissions Management

    - [ ] T4.1 Permissions Resource `[component: permissions-resource]`

        - [ ] T4.1.1 Prime Context
            - [ ] T4.1.1.1 Read permissions schema `[ref: solution-design.md; lines: 285-380]`
            - [ ] T4.1.1.2 Read permissions API contract `[ref: google-wallet-api.md; lines: 122-176]`
            - [ ] T4.1.1.3 Review authoritative pattern `[ref: terraform-plugin-framework.md; lines: 381-432]`
            - [ ] T4.1.1.4 Read business rules for permissions `[ref: PRD; lines: 219-233]`

        - [ ] T4.1.2 Write Tests `[activity: write-tests]`
            - [ ] T4.1.2.1 Unit test: Schema requires `issuer_id`
            - [ ] T4.1.2.2 Unit test: Schema has `permissions` list with email_address, role
            - [ ] T4.1.2.3 Unit test: Role validates against OWNER/READER/WRITER only
            - [ ] T4.1.2.4 Unit test: Email format validation
            - [ ] T4.1.2.5 Unit test: Duplicate emails in list produce error
            - [ ] T4.1.2.6 Acceptance test: Create permissions for Issuer `[ref: PRD Feature 3]`
            - [ ] T4.1.2.7 Acceptance test: Update replaces all permissions (authoritative)
            - [ ] T4.1.2.8 Acceptance test: Import existing permissions by Issuer ID
            - [ ] T4.1.2.9 Acceptance test: Delete sets empty permissions list

        - [ ] T4.1.3 Implement Resource `[activity: implement-code]`
            - [ ] T4.1.3.1 Create `internal/provider/permissions_resource.go`
            - [ ] T4.1.3.2 Define `PermissionsResourceModel` struct
            - [ ] T4.1.3.3 Define `PermissionModel` struct for list items
            - [ ] T4.1.3.4 Implement `Schema()` with validators (email regex, role enum)
            - [ ] T4.1.3.5 Add custom validator for duplicate email detection
            - [ ] T4.1.3.6 Implement `Create()` - set permissions via PUT
            - [ ] T4.1.3.7 Implement `Read()` - fetch current permissions
            - [ ] T4.1.3.8 Implement `Update()` - full replacement via PUT
            - [ ] T4.1.3.9 Implement `Delete()` - set empty permissions list
            - [ ] T4.1.3.10 Implement `ImportState()` - passthrough issuer_id

        - [ ] T4.1.4 Implement Client Methods `[activity: implement-code]`
            - [ ] T4.1.4.1 Implement `client.GetPermissions()` wrapping API get
            - [ ] T4.1.4.2 Implement `client.UpdatePermissions()` wrapping API update

        - [ ] T4.1.5 Validate
            - [ ] T4.1.5.1 Unit tests pass
            - [ ] T4.1.5.2 Acceptance tests pass
            - [ ] T4.1.5.3 Lint passes

    - [ ] T4.2 Permissions Data Source `[parallel: true]` `[component: permissions-datasource]`

        - [ ] T4.2.1 Prime Context
            - [ ] T4.2.1.1 Read data source requirements `[ref: PRD Feature 6]`

        - [ ] T4.2.2 Write Tests `[activity: write-tests]`
            - [ ] T4.2.2.1 Unit test: Schema requires `issuer_id` input
            - [ ] T4.2.2.2 Unit test: Schema outputs `permissions` list
            - [ ] T4.2.2.3 Acceptance test: Read permissions by Issuer ID
            - [ ] T4.2.2.4 Acceptance test: Error on non-existent Issuer

        - [ ] T4.2.3 Implement Data Source `[activity: implement-code]`
            - [ ] T4.2.3.1 Create `internal/provider/permissions_data_source.go`
            - [ ] T4.2.3.2 Define `PermissionsDataSourceModel` struct
            - [ ] T4.2.3.3 Implement `Schema()` and `Read()`

        - [ ] T4.2.4 Validate
            - [ ] T4.2.4.1 Unit and acceptance tests pass
            - [ ] T4.2.4.2 Lint passes

---

### Phase 5: Documentation & Examples

- [ ] T5 Phase 5 - Documentation `[activity: write-documentation]`

    - [ ] T5.1 Prime Context
        - [ ] T5.1.1 Read documentation requirements `[ref: PRD Should Have Features]`

    - [ ] T5.2 Provider Documentation
        - [ ] T5.2.1 Create `templates/index.md.tmpl` with provider overview
        - [ ] T5.2.2 Document authentication methods with examples
        - [ ] T5.2.3 Document environment variables

    - [ ] T5.3 Resource Documentation
        - [ ] T5.3.1 Create `templates/resources/issuer.md.tmpl`
        - [ ] T5.3.2 Create `templates/resources/permissions.md.tmpl`
        - [ ] T5.3.3 Include import instructions for each resource

    - [ ] T5.4 Data Source Documentation
        - [ ] T5.4.1 Create `templates/data-sources/issuer.md.tmpl`
        - [ ] T5.4.2 Create `templates/data-sources/issuers.md.tmpl`
        - [ ] T5.4.3 Create `templates/data-sources/permissions.md.tmpl`

    - [ ] T5.5 Working Examples
        - [ ] T5.5.1 Create `examples/provider/provider.tf` - basic provider config
        - [ ] T5.5.2 Create `examples/resources/googlewallet_issuer/resource.tf`
        - [ ] T5.5.3 Create `examples/resources/googlewallet_permissions/resource.tf`
        - [ ] T5.5.4 Create `examples/data-sources/googlewallet_issuer/data-source.tf`
        - [ ] T5.5.5 Create `examples/data-sources/googlewallet_issuers/data-source.tf`
        - [ ] T5.5.6 Create `examples/data-sources/googlewallet_permissions/data-source.tf`
        - [ ] T5.5.7 Create `examples/full-example/` with complete setup

    - [ ] T5.6 Validate
        - [ ] T5.6.1 Generate docs with `task generate`
        - [ ] T5.6.2 Verify all examples in examples/ are syntactically valid
        - [ ] T5.6.3 Review generated docs for completeness

---

### Phase 6: CI/CD & Release Configuration

- [ ] T6 Phase 6 - Release Automation `[activity: configure-cicd]`

    - [ ] T6.1 Prime Context
        - [ ] T6.1.1 Read CI/CD requirements `[ref: solution-design.md; lines: 450-520]`

    - [ ] T6.2 GitHub Actions Workflows
        - [ ] T6.2.1 Create `.github/workflows/test.yml` - PR validation
        - [ ] T6.2.2 Create `.github/workflows/release.yml` - tag-triggered release
        - [ ] T6.2.3 Configure matrix testing (Go versions, OS)
        - [ ] T6.2.4 Configure acceptance test job (optional, requires secrets)

    - [ ] T6.3 Release Configuration
        - [ ] T6.3.1 Finalize `.goreleaser.yml` with cross-platform builds
        - [ ] T6.3.2 Configure GPG signing for releases
        - [ ] T6.3.3 Configure changelog generation

    - [ ] T6.4 Repository Configuration
        - [ ] T6.4.1 Create `README.md` with quick start, badges, links
        - [ ] T6.4.2 Create `CHANGELOG.md` with initial release notes
        - [ ] T6.4.3 Create `LICENSE` (MPL-2.0 for Terraform providers)
        - [ ] T6.4.4 Create `CONTRIBUTING.md` with development setup

    - [ ] T6.5 Validate
        - [ ] T6.5.1 GitHub Actions workflows pass locally with `act`
        - [ ] T6.5.2 GoReleaser config validates with `goreleaser check`
        - [ ] T6.5.3 Test release build with `goreleaser release --snapshot --clean`

---

### Phase 7: Integration & End-to-End Validation

- [ ] T7 Integration & End-to-End Validation

    - [ ] T7.1 All unit tests passing `[activity: run-tests]`
    - [ ] T7.2 All acceptance tests passing against real Google Wallet API
    - [ ] T7.3 Integration tests for multi-resource scenarios
        - [ ] T7.3.1 Test: Create Issuer → Add Permissions → Read both
        - [ ] T7.3.2 Test: Import existing Issuer → Manage permissions
        - [ ] T7.3.3 Test: List Issuers → Select one → Read permissions
    - [ ] T7.4 End-to-end tests for complete user flows
        - [ ] T7.4.1 Test: Full provisioning flow (plan → apply → verify)
        - [ ] T7.4.2 Test: Update flow (modify → plan → apply → verify)
        - [ ] T7.4.3 Test: Import flow (import → plan shows no changes)
        - [ ] T7.4.4 Test: Destroy flow (destroy → verify soft delete)
    - [ ] T7.5 Error handling validation
        - [ ] T7.5.1 Verify clear error on invalid credentials
        - [ ] T7.5.2 Verify clear error on non-existent Issuer
        - [ ] T7.5.3 Verify clear error on invalid permission role
        - [ ] T7.5.4 Verify retry logic on transient API errors
    - [ ] T7.6 Documentation validation
        - [ ] T7.6.1 All examples in docs/ work when copied
        - [ ] T7.6.2 README quick start works for new users
        - [ ] T7.6.3 Import instructions work as documented
    - [ ] T7.7 Test coverage meets standards (>80%)
    - [ ] T7.8 Security validation
        - [ ] T7.8.1 Credentials marked sensitive in state
        - [ ] T7.8.2 No credentials logged in debug output
        - [ ] T7.8.3 File permissions appropriate on credential files
    - [ ] T7.9 Build and release verification
        - [ ] T7.9.1 Binary builds for all platforms (darwin/linux, amd64/arm64)
        - [ ] T7.9.2 Provider registers with Terraform
        - [ ] T7.9.3 Local dev override works correctly
    - [ ] T7.10 All PRD acceptance criteria verified `[ref: PRD Features 1-6]`
    - [ ] T7.11 Implementation follows SDD design `[ref: solution-design.md]`

---

## Estimated Timeline

| Phase | Duration | Dependencies |
|-------|----------|--------------|
| T1: Project Scaffolding | 2-3 hours | None |
| T2: Provider Core | 3-4 hours | T1 |
| T3: Issuer Management | 6-8 hours | T2 |
| T4: Permissions Management | 4-6 hours | T2 |
| T5: Documentation | 3-4 hours | T3, T4 |
| T6: CI/CD | 2-3 hours | T1 |
| T7: Integration & Validation | 4-6 hours | T3, T4, T5, T6 |

**Total Estimated Time**: 24-34 hours

*Note: T3 and T4 can be parallelized after T2 completion. T5 and T6 can be parallelized after T3/T4.*

---

## Risk Mitigation

| Risk | Mitigation Strategy |
|------|---------------------|
| API behavior differs from documentation | Run acceptance tests early and often against real API |
| Rate limiting during acceptance tests | Implement exponential backoff, add test delays if needed |
| SmartTap features require partnership | Defer to "Could Have", document limitations clearly |
| Credential handling edge cases | Test both file path and JSON content extensively |
| Int64 overflow for Issuer IDs | Use string type throughout, validate in tests |
