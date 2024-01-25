// SPDX-FileCopyrightText: 2023 Risk.Ident GmbH <contact@riskident.com>
//
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccUserResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
resource "mongodb_user" "test" {
  user  = "test-user"
  db    = "testdb-userresource"
  pwd   = "secret1234"
  roles = [
    { role = "readWrite" },
  ]
  custom_data = {
    "my-custom-field" = "my-custom-value"
  }
  mechanisms = [ "SCRAM-SHA-256" ]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mongodb_user.test", "id", "testdb-userresource.test-user"),
					resource.TestCheckResourceAttr("mongodb_user.test", "custom_data.%", "1"),
					resource.TestCheckResourceAttr("mongodb_user.test", "custom_data.my-custom-field", "my-custom-value"),
					resource.TestCheckResourceAttr("mongodb_user.test", "roles.#", "1"),
					resource.TestCheckResourceAttr("mongodb_user.test", "roles.0.role", "readWrite"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "mongodb_user.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"pwd", "custom_data", "roles", "mechanisms"},
			},
			// Update and Read testing
			{
				Config: providerConfig + `
resource "mongodb_user" "test" {
  user  = "test-user"
  db    = "testdb-userresource"
  pwd   = "secret1234"
  roles = [
    { role = "read" },
  ]
  custom_data = {
    "my-custom-field" = "my-updated-custom-value"
  }
  mechanisms = [ "SCRAM-SHA-256" ]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mongodb_user.test", "id", "testdb-userresource.test-user"),
					resource.TestCheckResourceAttr("mongodb_user.test", "custom_data.%", "1"),
					resource.TestCheckResourceAttr("mongodb_user.test", "custom_data.my-custom-field", "my-updated-custom-value"),
					resource.TestCheckResourceAttr("mongodb_user.test", "roles.#", "1"),
					resource.TestCheckResourceAttr("mongodb_user.test", "roles.0.role", "read"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
