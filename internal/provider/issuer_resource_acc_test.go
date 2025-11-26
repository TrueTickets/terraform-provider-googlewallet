// Copyright (c) TrueTickets, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccIssuerResource tests the full lifecycle of an issuer resource.
// Note: This test creates real resources in Google Wallet and cannot clean up
// due to API limitations (issuers cannot be deleted). Test issuers are prefixed
// with "[TESTING] " so they can be filtered out in the data source.
func TestAccIssuerResource(t *testing.T) {
	rName := fmt.Sprintf("%stf-test-%s", TestingPrefix, acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		// Note: No CheckDestroy because issuers cannot be deleted
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccIssuerResourceConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("googlewallet_issuer.test", "id"),
					resource.TestCheckResourceAttr("googlewallet_issuer.test", "name", rName),
					resource.TestCheckResourceAttr("googlewallet_issuer.test", "homepage_url", "https://example.com"),
					resource.TestCheckResourceAttr("googlewallet_issuer.test", "contact_info.name", "Test Contact"),
					resource.TestCheckResourceAttr("googlewallet_issuer.test", "contact_info.email", "test@example.com"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "googlewallet_issuer.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update testing
			{
				Config: testAccIssuerResourceConfigUpdated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("googlewallet_issuer.test", "name", rName+"-updated"),
					resource.TestCheckResourceAttr("googlewallet_issuer.test", "homepage_url", "https://updated.example.com"),
				),
			},
		},
	})
}

func testAccIssuerResourceConfig(name string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "googlewallet_issuer" "test" {
  name         = %[1]q
  homepage_url = "https://example.com"

  contact_info = {
    name  = "Test Contact"
    email = "test@example.com"
    phone = "+1-555-123-4567"
  }
}
`, name)
}

func testAccIssuerResourceConfigUpdated(name string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "googlewallet_issuer" "test" {
  name         = "%[1]s-updated"
  homepage_url = "https://updated.example.com"

  contact_info = {
    name  = "Updated Contact"
    email = "updated@example.com"
    phone = "+1-555-987-6543"
  }
}
`, name)
}
