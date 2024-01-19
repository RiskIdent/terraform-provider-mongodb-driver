package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/RiskIdent/terraform-provider-mongodb-driver/internal/mongodb"
)

func main() {
	c := mongodb.New("mongodb://localhost:27017", mongodb.Credentials{})

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))
	ctx := context.Background()

	err := c.DeleteDBRole(ctx, "test", "my-test-role")
	if err != nil {
		slog.Warn("delete role", "error", err)
	} else {
		slog.Info("Deleted role", "role", "my-test-role")
	}

	r, err := c.CreateDBRole(ctx, "test", mongodb.NewRole{
		Role: "my-test-role",
	})
	if err != nil {
		slog.Error("create role", "error", err)
		return
	}
	printRole("Created role", r)

	roles, err := c.ListDBRoles(ctx, "test")
	if err != nil {
		slog.Error("list roles", "error", err)
		return
	}

	slog.Info("Found roles.", "count", len(roles))
	for _, role := range roles {
		printRole("List role", role)
	}
}

func printRole(msg string, role mongodb.Role) {
	slog.Info(msg,
		"role", fmt.Sprintf("%s/%s", role.DB, role.Role),
		"id", role.ID,
		"isBuiltin", role.IsBuiltin,
		"priv", len(role.Privileges),
		"inheritPriv", len(role.InheritedPrivileges),
		"inheritRoles", len(role.InheritedRoles),
	)
}
