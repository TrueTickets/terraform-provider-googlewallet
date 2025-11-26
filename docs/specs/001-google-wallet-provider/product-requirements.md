# Product Requirements Document

## Validation Checklist

- [x] All required sections are complete
- [x] No [NEEDS CLARIFICATION] markers remain
- [x] Problem statement is specific and measurable
- [x] Problem is validated by evidence (not assumptions)
- [x] Context → Problem → Solution flow makes sense
- [x] Every persona has at least one user journey
- [x] All MoSCoW categories addressed (Must/Should/Could/Won't)
- [x] Every feature has testable acceptance criteria
- [x] Every metric has corresponding tracking events
- [x] No feature redundancy (check for duplicates)
- [x] No contradictions between sections
- [x] No technical implementation details included
- [x] A new team member could understand this PRD

---

## Product Overview

### Vision
Enable infrastructure teams to manage Google Wallet Issuers and access permissions as code, bringing the same declarative, version-controlled, auditable approach to mobile wallet infrastructure that they use for all other cloud resources.

### Problem Statement
Organizations using Google Wallet for digital passes (membership cards, event tickets, loyalty programs) must currently manage their Issuer accounts and permissions through the Google Pay & Wallet Console UI or custom scripts. This creates several pain points:

1. **No Infrastructure as Code**: Unlike AWS, GCP, and Azure resources which can be managed via Terraform/OpenTofu, Google Wallet configuration requires manual console clicks or bespoke automation
2. **Audit Trail Gaps**: Changes to Issuer settings and permissions are not tracked in version control, making compliance and change management difficult
3. **Environment Drift**: Without declarative configuration, development, staging, and production Issuer configurations can diverge silently
4. **Access Management Complexity**: Granting/revoking permissions to team members or service accounts requires manual intervention and is error-prone
5. **Integration Friction**: CI/CD pipelines cannot provision or configure Google Wallet resources alongside other infrastructure

**Evidence**: TrueTickets operates multiple Issuers across environments and has experienced permission drift, manual configuration errors, and audit challenges that would be solved by Infrastructure as Code.

### Value Proposition
terraform-provider-googlewallet enables:
- **Declarative Management**: Define Issuers and Permissions in HCL, track in Git, review via PRs
- **Environment Consistency**: Same Terraform patterns across dev/staging/prod
- **Audit Compliance**: All changes tracked, reviewed, and versioned
- **Automation Ready**: Integrate Issuer management into existing CI/CD pipelines
- **Team Productivity**: Eliminate manual console work and permission management overhead

## User Personas

### Primary Persona: Platform/DevOps Engineer
- **Demographics:** Mid-career (3-10 years experience), responsible for infrastructure automation, strong Terraform/OpenTofu expertise
- **Goals:**
  - Manage ALL infrastructure as code without exceptions
  - Automate environment provisioning end-to-end
  - Ensure consistent, auditable configurations across environments
- **Pain Points:**
  - Google Wallet is a "gap" in their IaC coverage
  - Must switch to console UI for wallet configuration
  - Cannot include wallet setup in infrastructure modules

### Secondary Personas

#### Mobile/Backend Developer
- **Demographics:** Developers integrating Google Wallet passes into applications
- **Goals:**
  - Quickly get Issuer credentials for development
  - Test pass creation without production access
- **Pain Points:**
  - Waiting for manual provisioning of development Issuers
  - Unclear what permissions they have

#### Security/Compliance Officer
- **Demographics:** Responsible for access control and audit compliance
- **Goals:**
  - Ensure least-privilege access to Issuers
  - Audit who has access to what
  - Review permission changes before they take effect
- **Pain Points:**
  - No visibility into permission changes over time
  - Manual permission reviews are tedious and error-prone

## User Journey Maps

### Primary User Journey: Infrastructure Provisioning
1. **Awareness:** Platform engineer discovers they need to provision a new Google Wallet Issuer for a new environment or product launch
2. **Consideration:** They check if there's a Terraform provider for Google Wallet (like there is for GCP, AWS, etc.)
3. **Adoption:** They find terraform-provider-googlewallet, install it, and configure credentials
4. **Usage:**
   - Write HCL to define Issuer with name, contact info
   - Write HCL to define Permissions for team members/service accounts
   - Run `terraform plan` to preview changes
   - Run `terraform apply` to create resources
   - Commit configuration to Git
5. **Retention:** Provider becomes standard part of infrastructure modules, used for all new Issuers

### Secondary User Journeys

#### Permission Management Journey
1. **Trigger:** New team member needs access to Issuer, or employee leaves and needs removal
2. **Action:** Update `googlewallet_permissions` resource in Terraform config
3. **Review:** PR review ensures correct role assignment
4. **Apply:** Terraform apply updates permissions atomically
5. **Audit:** Git history shows who made what change and when

#### Import Existing Infrastructure Journey
1. **Trigger:** Organization has existing Issuers created manually
2. **Discovery:** Use `googlewallet_issuers` data source to list all accessible Issuers
3. **Import:** Run `terraform import googlewallet_issuer.main <issuer_id>`
4. **Verify:** Run `terraform plan` to ensure state matches actual configuration
5. **Manage:** Going forward, all changes made via Terraform

## Feature Requirements

### Must Have Features

#### Feature 1: Provider Authentication
- **User Story:** As a platform engineer, I want to authenticate the provider with my GCP service account so that I can manage Google Wallet resources programmatically
- **Acceptance Criteria:**
  - [ ] Provider accepts `credentials` attribute with path to service account JSON file
  - [ ] Provider accepts `credentials` attribute with raw JSON content
  - [ ] Provider reads `GOOGLEWALLET_CREDENTIALS` environment variable as fallback
  - [ ] Provider reads `GOOGLE_CREDENTIALS` environment variable as secondary fallback
  - [ ] Clear error message when credentials are missing or invalid
  - [ ] Provider validates credentials during configuration phase

#### Feature 2: Issuer Resource Management
- **User Story:** As a platform engineer, I want to create and update Google Wallet Issuers so that I can manage Issuer configuration as code
- **Acceptance Criteria:**
  - [ ] Can create new Issuer with `name` and `contact_info`
  - [ ] `issuer_id` is computed and stored in state after creation
  - [ ] Can update Issuer `name`, `contact_info`, `homepage_url`
  - [ ] Can import existing Issuer by ID
  - [ ] Destroy removes resource from state (with warning that Issuer persists in Google)
  - [ ] Read refreshes state from Google Wallet API

#### Feature 3: Permissions Resource Management
- **User Story:** As a platform engineer, I want to manage access permissions for Issuers so that I can control who has what level of access
- **Acceptance Criteria:**
  - [ ] Can define authoritative permission list for an Issuer
  - [ ] Each permission entry specifies `email_address` and `role`
  - [ ] Valid roles: `OWNER`, `READER`, `WRITER`
  - [ ] Update replaces entire permission set (authoritative)
  - [ ] Destroy removes all permissions (empty list)
  - [ ] Can import existing permissions by Issuer ID
  - [ ] Validates email format and role values

#### Feature 4: Issuer Data Source
- **User Story:** As a platform engineer, I want to read Issuer details so that I can reference existing Issuers in my configuration
- **Acceptance Criteria:**
  - [ ] Can fetch single Issuer by `issuer_id`
  - [ ] Returns all Issuer attributes (name, contact_info, homepage_url)
  - [ ] Error if Issuer not found

#### Feature 5: Issuers List Data Source
- **User Story:** As a platform engineer, I want to list all accessible Issuers so that I can discover what Issuers exist
- **Acceptance Criteria:**
  - [ ] Returns list of all Issuers accessible by authenticated credentials
  - [ ] Each Issuer includes `issuer_id`, `name`, `homepage_url`
  - [ ] Empty list if no Issuers accessible

#### Feature 6: Permissions Data Source
- **User Story:** As a security officer, I want to query current permissions for an Issuer so that I can audit access
- **Acceptance Criteria:**
  - [ ] Can fetch permissions by `issuer_id`
  - [ ] Returns list of permission entries with `email_address` and `role`
  - [ ] Error if Issuer not found

### Should Have Features

#### Documentation and Examples
- **User Story:** As a new user, I want comprehensive documentation so that I can quickly learn how to use the provider
- **Acceptance Criteria:**
  - [ ] Provider documentation with authentication examples
  - [ ] Resource documentation with all attributes explained
  - [ ] Data source documentation with usage examples
  - [ ] Working examples in `examples/` directory
  - [ ] Import instructions for each resource

#### Validation and Error Handling
- **User Story:** As a user, I want clear validation and error messages so that I can quickly fix configuration issues
- **Acceptance Criteria:**
  - [ ] Email validation on permission entries
  - [ ] Role enum validation (only OWNER/READER/WRITER)
  - [ ] Descriptive error messages for API failures
  - [ ] Retry logic for transient API errors

### Could Have Features

#### SmartTap Merchant Data Support
- **User Story:** As an NFC pass implementer, I want to configure SmartTap merchant data so that I can enable NFC functionality
- **Acceptance Criteria:**
  - [ ] `smart_tap_merchant_data` block in Issuer resource
  - [ ] Support for `smart_tap_merchant_id` and `authentication_keys`

#### Callback Options Support
- **User Story:** As a developer, I want to configure callback URLs so that I receive notifications when passes are saved/deleted
- **Acceptance Criteria:**
  - [ ] `callback_options` block in Issuer resource
  - [ ] HTTPS URL validation

### Won't Have (This Phase)

- **Pass Class Management** - Generic/Loyalty/Event pass classes (future provider expansion)
- **Pass Object Management** - Individual pass instances (future provider expansion)
- **Message Management** - Pass notifications (future provider expansion)
- **Media Upload** - Logo/hero images (future provider expansion)
- **Terraform Cloud Integration** - Run triggers, workspace linking (standard Terraform functionality)

## Detailed Feature Specifications

### Feature: Permissions Resource (Authoritative)

**Description:** The `googlewallet_permissions` resource manages ALL permissions for a single Issuer. This is an authoritative resource - it defines the complete permission set, not incremental additions.

**User Flow:**
1. User writes HCL defining `googlewallet_permissions` with `issuer_id` and list of `permissions`
2. Terraform reads current permissions from Google Wallet API
3. Terraform shows diff between desired state and current state
4. On apply, Terraform sends complete permission list to API (replaces all existing permissions)
5. State is updated with the new permission configuration

**Business Rules:**
- Rule 1: The permission list is authoritative - any permission not in the list will be removed
- Rule 2: At least one OWNER must be retained (API may enforce this)
- Rule 3: Roles must be uppercase: `OWNER`, `READER`, `WRITER`
- Rule 4: Email addresses must be valid Google account emails (users or service accounts)
- Rule 5: Duplicate email addresses in the same permission list should error

**Edge Cases:**
- Empty permissions list → Expected: All permissions removed (may require at least one owner)
- Permission for non-existent email → Expected: API error, surfaced to user
- Invalid role value → Expected: Terraform validation error before API call
- Issuer doesn't exist → Expected: API error "Not Found"
- Permission already exists with different role → Expected: Role updated
- Service account email format → Expected: Accepted (service accounts are valid)

## Success Metrics

### Key Performance Indicators

- **Adoption:** Provider is used to manage 100% of TrueTickets Google Wallet Issuers within 30 days of release
- **Engagement:** At least one Terraform operation (plan/apply) per week per managed Issuer
- **Quality:** Zero production incidents caused by provider bugs in first 90 days
- **Business Impact:** Elimination of manual Issuer/Permission configuration tasks

### Tracking Requirements

| Event | Properties | Purpose |
|-------|------------|---------|
| Provider configured | credentials_type (file/json/env) | Understand authentication patterns |
| Resource created | resource_type, issuer_id | Track adoption of resource types |
| Resource updated | resource_type, changed_attributes | Understand update patterns |
| Resource imported | resource_type, issuer_id | Track existing infrastructure onboarding |
| API error | operation, error_code, error_message | Identify reliability issues |
| Validation error | field, validation_type | Improve validation messaging |

*Note: Tracking implemented via Terraform's built-in logging, not custom telemetry*

---

## Constraints and Assumptions

### Constraints
- **API Limitation**: Issuers cannot be deleted via API - provider must implement soft delete
- **API Limitation**: Permissions update is authoritative (full replacement only)
- **Rate Limit**: 20 requests/second per issuer account - may need retry logic
- **Authentication**: Service account must have "Developer" role in Google Pay & Wallet Console
- **Platform**: Must support darwin/linux on amd64/arm64 architectures

### Assumptions
- Users have existing GCP projects with Google Wallet API enabled
- Users have service accounts with appropriate permissions
- Users are familiar with Terraform/OpenTofu workflows
- Reference provider (terraform-provider-appleappstoreconnect) patterns are correct to follow
- `google.golang.org/api/walletobjects/v1` Go client library is stable and supported

## Risks and Mitigations

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| API behavior differs from documentation | High | Medium | Extensive acceptance testing against real API |
| No delete operation causes user confusion | Medium | High | Clear warning messages, documentation |
| Authoritative permissions accidentally removes access | High | Medium | Terraform plan shows diff; documentation warns users |
| Rate limiting causes apply failures | Medium | Low | Implement exponential backoff retry logic |
| Google changes API without notice | High | Low | Pin to API version, monitor Google announcements |
| SmartTap feature blocked by partnership requirement | Low | High | Document as restricted, gracefully handle API errors |

## Open Questions

- [x] Should we support `GOOGLE_APPLICATION_CREDENTIALS` in addition to custom env vars? → Yes, as `GOOGLE_CREDENTIALS` fallback
- [x] Is the permissions delete operation safe (empty list)? → Yes, per API documentation
- [x] Should SmartTap be in scope for initial release? → Could Have, not Must Have
- [x] Should we create a `googlewallet_issuer_permission` singular resource for non-authoritative management? → No, stick with authoritative pattern for simplicity

---

## Supporting Research

### Competitive Analysis
**No direct competitors exist.** There is no official or community Terraform/OpenTofu provider for Google Wallet API. Alternatives are:
- **Manual Console**: Google Pay & Wallet Console UI
- **Custom Scripts**: Direct API calls via gcloud or custom code
- **No Management**: Create once, never update

The terraform-provider-googlewallet fills a clear gap in the Infrastructure as Code ecosystem.

### User Research
Based on TrueTickets internal needs:
- Managing 3+ Issuers across environments (dev, staging, production)
- 10+ team members requiring various access levels
- Frequent permission changes during development cycles
- Desire to include Issuer setup in infrastructure-as-code modules

### Market Data
- **Google Wallet Growth**: Increasing adoption for loyalty, membership, and event tickets
- **IaC Adoption**: 70%+ of organizations use some form of Infrastructure as Code (2023 surveys)
- **Terraform Dominance**: Terraform/OpenTofu is the leading IaC tool for multi-cloud
- **HashiCorp Provider Ecosystem**: 3,000+ providers in Terraform Registry demonstrate demand for IaC integrations
