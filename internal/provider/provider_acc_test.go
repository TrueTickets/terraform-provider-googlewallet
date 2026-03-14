// Copyright (c) TrueTickets, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// testAccProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"googlewallet": providerserver.NewProtocol6WithError(New("test")()),
}

// testAccPreCheck validates that required environment variables are set
// before running acceptance tests.
func testAccPreCheck(t *testing.T) {
	t.Helper()

	// Check for credentials in order of precedence
	// (matching the provider's credential resolution order)
	if os.Getenv("GOOGLEWALLET_CREDENTIALS") != "" {
		return
	}
	if os.Getenv("GOOGLE_CREDENTIALS") != "" {
		return
	}
	if os.Getenv("GOOGLE_APPLICATION_CREDENTIALS") != "" {
		return
	}

	t.Skip("One of GOOGLEWALLET_CREDENTIALS, GOOGLE_CREDENTIALS, or GOOGLE_APPLICATION_CREDENTIALS must be set for acceptance tests")
}

// testAccProviderConfig returns the provider configuration for acceptance tests.
func testAccProviderConfig() string {
	return `
provider "googlewallet" {}
`
}
