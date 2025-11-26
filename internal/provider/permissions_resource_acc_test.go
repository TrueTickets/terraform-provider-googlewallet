// Copyright (c) TrueTickets, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// testAccPermissionsImportStateIdFunc returns the issuer_id for import testing.
func testAccPermissionsImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("resource not found: %s", resourceName)
		}
		return rs.Primary.Attributes["issuer_id"], nil
	}
}

// getTestServiceAccountEmail returns the service account email for permissions tests.
// This must be set via GOOGLEWALLET_TEST_SA_EMAIL environment variable.
// The SA must have OWNER access to maintain issuer access during tests.
func getTestServiceAccountEmail() string {
	return os.Getenv("GOOGLEWALLET_TEST_SA_EMAIL")
}

// getTestSecondaryEmail returns a secondary email for testing multiple permissions.
// This must be set via GOOGLEWALLET_TEST_SECONDARY_EMAIL environment variable.
// Can be a user account, group, or another service account.
func getTestSecondaryEmail() string {
	return os.Getenv("GOOGLEWALLET_TEST_SECONDARY_EMAIL")
}

// TestAccPermissionsResource tests the full lifecycle of a permissions resource.
// Note: The permissions resource is authoritative - it replaces ALL permissions.
// Tests must include the service account as OWNER to maintain access.
//
// Required environment variables:
//   - GOOGLEWALLET_TEST_SA_EMAIL: Service account email (will be set as OWNER)
//   - GOOGLEWALLET_TEST_SECONDARY_EMAIL: Secondary email for testing (user, group, or SA)
func TestAccPermissionsResource(t *testing.T) {
	rName := fmt.Sprintf("%stf-test-%s", TestingPrefix, acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum))
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
			// Create and Read testing
			{
				Config: testAccPermissionsResourceConfig(rName, saEmail, secondaryEmail),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("googlewallet_permissions.test", "issuer_id"),
					resource.TestCheckResourceAttr("googlewallet_permissions.test", "permissions.#", "2"),
				),
			},
			// ImportState testing
			{
				ResourceName:                         "googlewallet_permissions.test",
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    testAccPermissionsImportStateIdFunc("googlewallet_permissions.test"),
				ImportStateVerifyIdentifierAttribute: "issuer_id",
			},
			// Update testing - change role from READER to WRITER
			{
				Config: testAccPermissionsResourceConfigUpdated(rName, saEmail, secondaryEmail),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("googlewallet_permissions.test", "permissions.#", "2"),
				),
			},
		},
	})
}

func testAccPermissionsResourceConfig(name, saEmail, secondaryEmail string) string {
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
`, name, saEmail, secondaryEmail)
}

func testAccPermissionsResourceConfigUpdated(name, saEmail, secondaryEmail string) string {
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
      # Upgraded from WRITER to OWNER
      email_address = %[3]q
      role          = "OWNER"
    }
  ]
}
`, name, saEmail, secondaryEmail)
}
