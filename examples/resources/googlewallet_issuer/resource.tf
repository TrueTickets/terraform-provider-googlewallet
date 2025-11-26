# Manage a Google Wallet Issuer
resource "googlewallet_issuer" "example" {
  name = "My Organization"

  contact_info {
    name  = "Support Team"
    email = "support@example.com"
    phone = "+1-555-123-4567"
  }

  homepage_url = "https://example.com"
}
