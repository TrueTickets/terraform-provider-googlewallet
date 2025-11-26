# Configure the Google Wallet Provider
provider "googlewallet" {
  # Credentials can be provided via:
  # - This attribute (JSON string of service account key)
  # - GOOGLEWALLET_CREDENTIALS environment variable
  # - GOOGLE_CREDENTIALS environment variable (fallback)
  credentials = file("service-account.json")
}

# Alternative: Using environment variables (recommended for security)
# export GOOGLEWALLET_CREDENTIALS=$(cat service-account.json)
# provider "googlewallet" {}
