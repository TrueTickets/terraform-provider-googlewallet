# Manage a Google Wallet Issuer
resource "googlewallet_issuer" "example" {
  name         = "My Organization"
  homepage_url = "https://example.com"

  contact_info = {
    name          = "Support Team"
    email         = "support@example.com"
    phone         = "+1-555-123-4567"
    alerts_emails = ["alerts@example.com"]
  }
}
