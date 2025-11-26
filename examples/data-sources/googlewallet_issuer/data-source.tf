# Read a Google Wallet Issuer by ID
data "googlewallet_issuer" "example" {
  issuer_id = "1234567890123456789"
}

output "issuer_name" {
  value = data.googlewallet_issuer.example.name
}
