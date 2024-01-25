// SPDX-FileCopyrightText: 2023 Risk.Ident GmbH <contact@riskident.com>
//
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccRoleResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
resource "mongodb_role" "example" {
  role = "test-role"
  db   = "testdb-roleresource"
  privileges = [
    {
      resource = { db = "testdb-roleresource", collection = "" }
      actions  = ["collMod"]
    },
  ]
  roles = [
    { role = "readWrite" },
  ]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mongodb_role.example", "id", "testdb-roleresource.test-role"),
					resource.TestCheckResourceAttr("mongodb_role.example", "privileges.#", "1"),
					resource.TestCheckResourceAttr("mongodb_role.example", "privileges.0.resource.db", "testdb-roleresource"),
					resource.TestCheckResourceAttr("mongodb_role.example", "privileges.0.resource.collection", ""),
					resource.TestCheckResourceAttr("mongodb_role.example", "privileges.0.actions.#", "1"),
					resource.TestCheckResourceAttr("mongodb_role.example", "privileges.0.actions.0", "collMod"),
					resource.TestCheckResourceAttr("mongodb_role.example", "roles.#", "1"),
					resource.TestCheckResourceAttr("mongodb_role.example", "roles.0.role", "readWrite"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "mongodb_role.example",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"privileges", "roles"},
			},
			// Update and Read testing
			{
				Config: providerConfig + `
resource "mongodb_role" "example" {
  role = "test-role"
  db   = "testdb-roleresource"
  privileges = [
    {
      resource = { db = "testdb-roleresource", collection = "" }
      actions  = ["insert", "update"]
    },
  ]
  roles = [
    { role = "read" },
  ]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mongodb_role.example", "id", "testdb-roleresource.test-role"),
					resource.TestCheckResourceAttr("mongodb_role.example", "privileges.#", "1"),
					resource.TestCheckResourceAttr("mongodb_role.example", "privileges.0.resource.db", "testdb-roleresource"),
					resource.TestCheckResourceAttr("mongodb_role.example", "privileges.0.resource.collection", ""),
					resource.TestCheckResourceAttr("mongodb_role.example", "privileges.0.actions.#", "2"),
					resource.TestCheckResourceAttr("mongodb_role.example", "privileges.0.actions.0", "insert"),
					resource.TestCheckResourceAttr("mongodb_role.example", "privileges.0.actions.1", "update"),
					resource.TestCheckResourceAttr("mongodb_role.example", "roles.#", "1"),
					resource.TestCheckResourceAttr("mongodb_role.example", "roles.0.role", "read"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
