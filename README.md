# Terraform Provider for Google Wallet

A Terraform/OpenTofu provider for managing Google Wallet resources,
including Issuers and Permissions.

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >=
  1.0 (or [OpenTofu](https://opentofu.org/) >= 1.0)
- Go >= 1.23 (for building from source)
- A Google Cloud service account with the Wallet Admin role
  (`roles/wallet.admin`)

## Installation

### From Terraform Registry

```hcl
terraform {
  required_providers {
    googlewallet = {
      source  = "truetickets/googlewallet"
      version = "~> 1.0"
    }
  }
}
```

### Building from Source

```bash
git clone https://github.com/truetickets/terraform-provider-googlewallet.git
cd terraform-provider-googlewallet
go build -o terraform-provider-googlewallet
```

## Authentication

The provider requires a Google Cloud service account with the following
permissions:

- `roles/wallet.admin` - For full management of Wallet resources

Credentials can be provided in one of the following ways (in order of
priority):

1. **Provider attribute**: Set the `credentials` attribute in the
   provider configuration
2. **GOOGLEWALLET_CREDENTIALS**: Environment variable with the JSON
   service account key
3. **GOOGLE_CREDENTIALS**: Environment variable (fallback, for
   compatibility)

### Example: Using environment variable (recommended)

```bash
export GOOGLEWALLET_CREDENTIALS=$(cat service-account.json)
```

```hcl
provider "googlewallet" {}
```

### Example: Using provider attribute

```hcl
provider "googlewallet" {
  credentials = file("service-account.json")
}
```

## Resources

### googlewallet_issuer

Manages a Google Wallet Issuer. An issuer is an entity that can create
and manage Google Wallet passes.

> **Note:** Google Wallet API does not support deleting issuers. When
> this resource is destroyed, Terraform will remove it from state but
> the issuer will continue to exist in Google Wallet.

```hcl
resource "googlewallet_issuer" "example" {
  name         = "My Company"
  homepage_url = "https://example.com"

  contact_info = {
    name          = "Support Team"
    email         = "support@example.com"
    phone         = "+1-555-123-4567"
    alerts_emails = ["alerts@example.com", "ops@example.com"]
  }
}
```

#### Attributes

| Attribute                    | Type         | Required | Description                         |
| ---------------------------- | ------------ | -------- | ----------------------------------- |
| `name`                       | String       | Yes      | The account name of the issuer      |
| `homepage_url`               | String       | No       | URL for the issuer's home page      |
| `contact_info`               | Object       | No       | Contact information for the issuer  |
| `contact_info.name`          | String       | No       | Primary contact name                |
| `contact_info.email`         | String       | No       | Primary contact email address       |
| `contact_info.phone`         | String       | No       | Primary contact phone number        |
| `contact_info.alerts_emails` | List(String) | No       | Email addresses that receive alerts |

#### Read-Only Attributes

| Attribute | Type   | Description                              |
| --------- | ------ | ---------------------------------------- |
| `id`      | String | The unique identifier assigned by Google |

### googlewallet_permissions

Manages permissions for a Google Wallet Issuer. This resource is
**authoritative** - it manages all permissions for an issuer. Any
permissions not defined in this resource will be removed when applied.

> **Warning:** This resource will replace ALL permissions for the
> issuer. Make sure to include all desired permissions in your
> configuration.

```hcl
resource "googlewallet_permissions" "example" {
  issuer_id = googlewallet_issuer.example.id

  permissions = [
    {
      email_address = "admin@example.com"
      role          = "OWNER"
    },
    {
      email_address = "developer@example.com"
      role          = "WRITER"
    },
    {
      email_address = "viewer@example.com"
      role          = "READER"
    }
  ]
}
```

#### Attributes

| Attribute                     | Type         | Required | Description                                              |
| ----------------------------- | ------------ | -------- | -------------------------------------------------------- |
| `issuer_id`                   | String       | Yes      | The ID of the issuer to manage permissions for           |
| `permissions`                 | List(Object) | Yes      | The complete list of permissions                         |
| `permissions[].email_address` | String       | Yes      | The email address of the user, group, or service account |
| `permissions[].role`          | String       | Yes      | The role to grant: `OWNER`, `WRITER`, or `READER`        |

## Data Sources

### googlewallet_issuer

Fetches a single Google Wallet Issuer by ID.

```hcl
data "googlewallet_issuer" "example" {
  id = "1234567890123456789"
}

output "issuer_name" {
  value = data.googlewallet_issuer.example.name
}
```

### googlewallet_issuers

Lists all Google Wallet Issuers accessible to the authenticated service
account.

```hcl
data "googlewallet_issuers" "all" {}

output "all_issuers" {
  value = data.googlewallet_issuers.all.issuers
}
```

### googlewallet_permissions

Fetches the current permissions for a Google Wallet Issuer.

```hcl
data "googlewallet_permissions" "example" {
  issuer_id = "1234567890123456789"
}

output "owners" {
  value = [for p in data.googlewallet_permissions.example.permissions : p.email_address if p.role == "OWNER"]
}
```

## Development

### Building

```bash
go build -o terraform-provider-googlewallet
```

### Testing

```bash
# Run unit tests
go test ./...

# Run acceptance tests (requires real credentials)
TF_ACC=1 GOOGLEWALLET_CREDENTIALS=$(cat service-account.json) go test ./... -v
```

### Linting

```bash
golangci-lint run ./...
```

## License

This provider is licensed under the
[Mozilla Public License 2.0](LICENSE).

## Contributing

Contributions are welcome! Please read our contributing guidelines
before submitting a pull request.
