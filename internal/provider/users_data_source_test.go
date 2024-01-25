// SPDX-FileCopyrightText: 2023 Risk.Ident GmbH <contact@riskident.com>
//
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccUsersDataSource(t *testing.T) {
	createTestUser(t, "testdb", "test-user")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: providerConfig + `data "mongodb_users" "test" {
          db = "testdb"
        }`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.mongodb_users.test", "users.#", "1"),
					resource.TestCheckResourceAttr("data.mongodb_users.test", "users.0.db", "testdb"),
					resource.TestCheckResourceAttr("data.mongodb_users.test", "users.0.user", "test-user"),
					resource.TestCheckResourceAttr("data.mongodb_users.test", "users.0.id", "testdb.test-user"),
					resource.TestCheckResourceAttr("data.mongodb_users.test", "users.0.roles.#", "1"),
					resource.TestCheckResourceAttr("data.mongodb_users.test", "users.0.roles.0.role", "readWrite"),
					resource.TestCheckResourceAttr("data.mongodb_users.test", "users.0.roles.0.db", "testdb"),
				),
			},
		},
	})
}
