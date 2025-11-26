# List only active issuers (default behavior - excludes archived and testing)
data "googlewallet_issuers" "active" {}

# List all issuers including archived ones
data "googlewallet_issuers" "with_archived" {
  include_archived = true
}

# List all issuers including test issuers (useful for test cleanup)
data "googlewallet_issuers" "with_testing" {
  include_testing = true
}

output "active_issuer_count" {
  value = length(data.googlewallet_issuers.active.issuers)
}

output "issuer_names" {
  value = [for issuer in data.googlewallet_issuers.active.issuers : issuer.name]
}
