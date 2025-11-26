# Changelog

All notable changes to this project will be documented in this file.

The format is based on
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this
project adheres to
[Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0] - Unreleased

### Added

- **New Resource:** `googlewallet_issuer` - Manage Google Wallet Issuers
- **New Resource:** `googlewallet_permissions` - Manage issuer
  permissions
- **New Data Source:** `googlewallet_issuer` - Retrieve issuer by ID
- **New Data Source:** `googlewallet_issuers` - List all accessible
  issuers
- **New Data Source:** `googlewallet_permissions` - Retrieve issuer
  permissions

### Features

- Google Cloud service account authentication
- Environment variable support (GOOGLEWALLET_CREDENTIALS,
  GOOGLE_CREDENTIALS)
- Full CRUD operations for Issuer resources
- Authoritative permissions management
- Import support for all resources
- Comprehensive acceptance test suite

### Notes

- Initial release of the Google Wallet Terraform provider
- Issuer IDs stored as strings to avoid int64 overflow
- Issuer deletion removes from state only (API limitation)
- Permissions resource is authoritative
