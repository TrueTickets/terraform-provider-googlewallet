# Read permissions for a Google Wallet Issuer
data "googlewallet_permissions" "example" {
  issuer_id = "1234567890123456789"
}

output "issuer_permissions" {
  value = data.googlewallet_permissions.example.permissions
}
