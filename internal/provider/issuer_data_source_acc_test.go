// Copyright (c) TrueTickets, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccIssuerDataSource tests the issuer data source.
// Test issuers are prefixed with "[TESTING] " so they can be filtered out.
func TestAccIssuerDataSource(t *testing.T) {
	rName := fmt.Sprintf("%stf-test-%s", TestingPrefix, acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccIssuerDataSourceConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.googlewallet_issuer.test", "id"),
					resource.TestCheckResourceAttr("data.googlewallet_issuer.test", "name", rName),
				),
			},
		},
	})
}

// TestAccIssuersDataSource tests the issuers list data source.
func TestAccIssuersDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccIssuersDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.googlewallet_issuers.all", "issuers.#"),
				),
			},
		},
	})
}

func testAccIssuerDataSourceConfig(name string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "googlewallet_issuer" "test" {
  name         = %[1]q
  homepage_url = "https://example.com"

  contact_info = {
    name  = "Test Contact"
    email = "test@example.com"
  }
}

data "googlewallet_issuer" "test" {
  id = googlewallet_issuer.test.id
}
`, name)
}

func testAccIssuersDataSourceConfig() string {
	return testAccProviderConfig() + `
data "googlewallet_issuers" "all" {}
`
}
