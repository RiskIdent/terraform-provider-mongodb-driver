// SPDX-FileCopyrightText: 2023 Risk.Ident GmbH <contact@riskident.com>
//
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/RiskIdent/terraform-provider-mongodb-driver/internal/mongodb"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

var (
	mongodbUri     = getMongoURI()
	providerConfig = `
provider "mongodb" {
  uri = "` + mongodbUri + `"
}
`

	// testAccProtoV6ProviderFactories are used to instantiate a provider during
	// acceptance testing. The factory function will be invoked for every Terraform
	// CLI command executed to create a provider server to which the CLI can
	// reattach.
	testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
		"mongodb": providerserver.NewProtocol6WithError(New("test")()),
	}
)

func getMongoURI() string {
	if uri := os.Getenv("MONGODB_URI"); uri != "" {
		return uri
	}
	return "mongodb://localhost:27017"
}

func createTestUser(t *testing.T, dbName, userName string) {
	db := mongodb.New(mongodbUri, mongodb.Credentials{})
	if _, err := db.CreateDBUser(context.Background(), dbName, mongodb.NewUser{
		User:     userName,
		Password: "secret1234",
		Roles: []mongodb.RoleRef{
			mongodb.RoleSameDBRef("readWrite"),
		},
	}); err != nil && !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("create test user: %s", err)
	}
	t.Cleanup(func() {
		if db.DeleteDBUser(context.Background(), dbName, userName) != nil {
			t.Errorf("Failed to clean up temporary testing user: %s.%s", dbName, userName)
		}
	})
}
