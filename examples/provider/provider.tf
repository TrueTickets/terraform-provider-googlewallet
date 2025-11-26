# Configure the Google Wallet Provider
provider "googlewallet" {
  # Credentials can be provided via:
  # - This attribute (path to JSON file or JSON content)
  # - GOOGLEWALLET_CREDENTIALS environment variable
  # - GOOGLE_CREDENTIALS environment variable
  credentials = "/path/to/service-account.json"
}
