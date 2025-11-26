// Copyright (c) TrueTickets, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccPermissionsDataSource tests the permissions data source.
// Note: The permissions resource is authoritative - it replaces ALL permissions.
// Tests must include the service account as OWNER to maintain access.
//
// Required environment variables:
//   - GOOGLEWALLET_TEST_SA_EMAIL: Service account email (will be set as OWNER)
//   - GOOGLEWALLET_TEST_SECONDARY_EMAIL: Secondary email for testing (user, group, or SA)
func TestAccPermissionsDataSource(t *testing.T) {
	rName := fmt.Sprintf("tf-test-%s", acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum))
	saEmail := getTestServiceAccountEmail()
	secondaryEmail := getTestSecondaryEmail()

	if saEmail == "" {
		t.Skip("GOOGLEWALLET_TEST_SA_EMAIL must be set for permissions acceptance tests")
	}
	if secondaryEmail == "" {
		t.Skip("GOOGLEWALLET_TEST_SECONDARY_EMAIL must be set for permissions acceptance tests")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionsDataSourceConfig(rName, saEmail, secondaryEmail),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.googlewallet_permissions.test", "issuer_id"),
					resource.TestCheckResourceAttr("data.googlewallet_permissions.test", "permissions.#", "2"),
				),
			},
		},
	})
}

func testAccPermissionsDataSourceConfig(name, saEmail, secondaryEmail string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "googlewallet_issuer" "test" {
  name = %[1]q

  contact_info = {
    name  = "Test Contact"
    email = "test@example.com"
  }
}

resource "googlewallet_permissions" "test" {
  issuer_id = googlewallet_issuer.test.id

  permissions = [
    {
      email_address = %[2]q
      role          = "OWNER"
    },
    {
      email_address = %[3]q
      role          = "WRITER"
    }
  ]
}

data "googlewallet_permissions" "test" {
  issuer_id = googlewallet_issuer.test.id

  depends_on = [googlewallet_permissions.test]
}
`, name, saEmail, secondaryEmail)
}
