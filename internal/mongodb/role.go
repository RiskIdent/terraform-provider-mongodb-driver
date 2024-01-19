// SPDX-FileCopyrightText: 2023 Risk.Ident GmbH <contact@riskident.com>
//
// SPDX-License-Identifier: MPL-2.0

package mongodb

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
)

type Role struct {
	ID                  string      `bson:"_id"`
	Role                string      `bson:"role"`
	DB                  string      `bson:"db"`
	Privileges          []Privilege `bson:"privileges"`
	IsBuiltin           bool        `bson:"isBuiltin"`
	Roles               []RoleDBRef `bson:"roles"`
	InheritedRoles      []RoleDBRef `bson:"inheritedRoles"`
	InheritedPrivileges []Privilege `bson:"inheritedPrivileges"`
}

type Privilege struct {
	Resource ResourceWrapper `bson:"resource"`
	Actions  []string        `bson:"actions"`
}

type rolesInfoCommand struct {
	RolesInfo                      any  `bson:"rolesInfo"`
	ShowAuthenticationRestrictions bool `bson:"showAuthenticationRestrictions,omitempty"`
	ShowPrivileges                 bool `bson:"showPrivileges,omitempty"`
	ShowBuiltinRoles               bool `bson:"showBuiltinRoles,omitempty"`
}

func (c *Client) ListDBRoles(ctx context.Context, dbName string) ([]Role, error) {
	if err := c.connect(ctx); err != nil {
		return nil, err
	}
	query := rolesInfoCommand{
		RolesInfo:        1, // list roles in collection
		ShowPrivileges:   true,
		ShowBuiltinRoles: true,
	}
	return c.runRolesInfo(ctx, dbName, query)
}

func (c *Client) GetDBRole(ctx context.Context, dbName, roleName string) (Role, error) {
	if err := c.connect(ctx); err != nil {
		return Role{}, err
	}
	return c.runRolesInfoSingle(ctx, dbName, roleName)
}

func (c *Client) runRolesInfoSingle(ctx context.Context, dbName, roleName string) (Role, error) {
	query := rolesInfoCommand{
		RolesInfo:        roleName,
		ShowPrivileges:   true,
		ShowBuiltinRoles: true,
	}
	roles, err := c.runRolesInfo(ctx, dbName, query)
	if err != nil {
		return Role{}, err
	}
	if len(roles) == 0 {
		return Role{}, ErrNotFound
	}
	return roles[0], nil
}

func (c *Client) runRolesInfo(ctx context.Context, dbName string, query rolesInfoCommand) ([]Role, error) {
	db := c.client.Database(dbName)

	result := db.RunCommand(ctx, query)
	if err := result.Err(); err != nil {
		return nil, err
	}
	var response struct {
		CommandResponse `bson:",inline"`
		Roles           []Role `bson:"roles"`
	}
	if err := result.Decode(&response); err != nil {
		return nil, err
	}
	if err := validateResponse(response.CommandResponse); err != nil {
		return nil, err
	}
	return response.Roles, nil
}

type NewRole struct {
	Role       string      `bson:"createRole"`
	Privileges []Privilege `bson:"privileges"`
	Roles      []RoleDBRef `bson:"roles"`
}

func (c *Client) CreateDBRole(ctx context.Context, dbName string, newRole NewRole) (Role, error) {
	if err := c.connect(ctx); err != nil {
		return Role{}, err
	}
	if newRole.Privileges == nil {
		newRole.Privileges = []Privilege{}
	}
	if newRole.Roles == nil {
		newRole.Roles = []RoleDBRef{}
	}
	if err := c.runCreateRole(ctx, dbName, newRole); err != nil {
		return Role{}, err
	}
	role, err := c.runRolesInfoSingle(ctx, dbName, newRole.Role)
	if err != nil {
		return Role{}, fmt.Errorf("get created role: %w", err)
	}
	return role, nil
}

func (c *Client) runCreateRole(ctx context.Context, dbName string, newRole NewRole) error {
	db := c.client.Database(dbName)
	result := db.RunCommand(ctx, newRole)
	if err := result.Err(); err != nil {
		return err
	}
	var response CommandResponse
	if err := result.Decode(&response); err != nil {
		return err
	}
	if err := validateResponse(response); err != nil {
		return err
	}
	return nil
}

type UpdateRole struct {
	Role       string      `bson:"updateRole"`
	Privileges []Privilege `bson:"privileges,omitempty"`
	Roles      []RoleRef   `bson:"roles,omitempty"`
}

func (c *Client) UpdateDBRole(ctx context.Context, dbName string, update UpdateRole) (Role, error) {
	if err := c.connect(ctx); err != nil {
		return Role{}, err
	}
	if err := c.runUpdateRole(ctx, dbName, update); err != nil {
		return Role{}, err
	}
	role, err := c.runRolesInfoSingle(ctx, dbName, update.Role)
	if err != nil {
		return Role{}, fmt.Errorf("get updated role: %w", err)
	}
	return role, nil
}

func (c *Client) runUpdateRole(ctx context.Context, dbName string, update UpdateRole) error {
	db := c.client.Database(dbName)

	result := db.RunCommand(ctx, update)
	if err := result.Err(); err != nil {
		return err
	}
	var response CommandResponse
	if err := result.Decode(&response); err != nil {
		return err
	}
	if err := validateResponse(response); err != nil {
		return err
	}
	return nil
}

func (c *Client) DeleteDBRole(ctx context.Context, dbName, roleName string) error {
	if err := c.connect(ctx); err != nil {
		return err
	}
	return c.runDropRole(ctx, dbName, roleName)
}

func (c *Client) runDropRole(ctx context.Context, dbName, roleName string) error {
	db := c.client.Database(dbName)

	result := db.RunCommand(ctx, bson.D{
		{Key: "dropRole", Value: roleName},
	})
	if err := result.Err(); err != nil {
		return err
	}
	var response CommandResponse
	if err := result.Decode(&response); err != nil {
		return err
	}
	return validateResponse(response)
}
