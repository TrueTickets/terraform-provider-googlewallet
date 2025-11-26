# List all accessible Google Wallet Issuers
data "googlewallet_issuers" "all" {}

output "all_issuers" {
  value = data.googlewallet_issuers.all.issuers
}
