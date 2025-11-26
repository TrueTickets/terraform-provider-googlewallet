# List only active issuers (default behavior - excludes archived)
data "googlewallet_issuers" "active" {}

# List all issuers including archived ones
data "googlewallet_issuers" "all" {
  include_archived = true
}

output "active_issuer_count" {
  value = length(data.googlewallet_issuers.active.issuers)
}

output "issuer_names" {
  value = [for issuer in data.googlewallet_issuers.active.issuers : issuer.name]
}
