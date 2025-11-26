# Manage permissions for a Google Wallet Issuer
resource "googlewallet_permissions" "example" {
  issuer_id = googlewallet_issuer.example.issuer_id

  permissions {
    email_address = "admin@example.com"
    role          = "OWNER"
  }

  permissions {
    email_address = "developer@example.com"
    role          = "WRITER"
  }

  permissions {
    email_address = "viewer@example.com"
    role          = "READER"
  }
}
