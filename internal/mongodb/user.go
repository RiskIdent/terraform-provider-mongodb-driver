// SPDX-FileCopyrightText: 2023 Risk.Ident GmbH <contact@riskident.com>
//
// SPDX-License-Identifier: MPL-2.0

package mongodb

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID         string            `bson:"_id"`
	UserID     primitive.Binary  `bson:"userId"`
	User       string            `bson:"user"`
	DB         string            `bson:"db"`
	CustomData map[string]string `bson:"customData"`
	Roles      []RoleDBRef       `bson:"roles"`
	Mechanisms []Mechanism       `bson:"mechanisms"`
}

func (c *Client) ListDBUsers(ctx context.Context, dbName string, filter any) ([]User, error) {
	if err := c.connect(ctx); err != nil {
		return nil, err
	}
	query := usersInfoCommand{
		UsersInfo: 1, // 1 = all in current db
		Filter:    filter,
	}
	return c.runUsersInfo(ctx, dbName, query)
}

func (c *Client) ListAllUsers(ctx context.Context, filter any) ([]User, error) {
	if err := c.connect(ctx); err != nil {
		return nil, err
	}
	query := usersInfoCommand{
		UsersInfo: bson.D{
			{Key: "forAllDBs", Value: true},
		},
		Filter: filter,
	}
	return c.runUsersInfo(ctx, "admin", query)
}

func (c *Client) GetDBUser(ctx context.Context, dbName, userName string) (User, error) {
	if err := c.connect(ctx); err != nil {
		return User{}, err
	}
	return c.runUsersInfoSingle(ctx, dbName, userName)
}

func (c *Client) runUsersInfoSingle(ctx context.Context, dbName, userName string) (User, error) {
	query := usersInfoCommand{
		UsersInfo: userName,
	}
	users, err := c.runUsersInfo(ctx, dbName, query)
	if err != nil {
		return User{}, err
	}
	if len(users) == 0 {
		return User{}, ErrNotFound
	}
	return users[0], nil
}

type usersInfoCommand struct {
	UsersInfo any `bson:"usersInfo"`
	Filter    any `bson:"filter,omitempty"`
}

func (c *Client) runUsersInfo(ctx context.Context, dbName string, query usersInfoCommand) ([]User, error) {
	db := c.client.Database(dbName)

	result := db.RunCommand(ctx, query)
	if err := result.Err(); err != nil {
		return nil, err
	}
	var response struct {
		CommandResponse `bson:",inline"`
		Users           []User `bson:"users"`
	}
	if err := result.Decode(&response); err != nil {
		return nil, err
	}
	if err := validateResponse(response.CommandResponse); err != nil {
		return nil, err
	}
	return response.Users, nil
}

type NewUser struct {
	User       string            `bson:"createUser"`
	Password   string            `bson:"pwd"`
	CustomData map[string]string `bson:"customData,omitempty"`
	Roles      []RoleRef         `bson:"roles"`
	Mechanisms []Mechanism       `bson:"mechanisms,omitempty"`
}

func (c *Client) CreateDBUser(ctx context.Context, dbName string, newUser NewUser) (User, error) {
	if err := c.connect(ctx); err != nil {
		return User{}, err
	}
	if err := c.runCreateUser(ctx, dbName, newUser); err != nil {
		return User{}, err
	}
	user, err := c.runUsersInfoSingle(ctx, dbName, newUser.User)
	if err != nil {
		return User{}, fmt.Errorf("get created user: %w", err)
	}
	return user, nil
}

func (c *Client) runCreateUser(ctx context.Context, dbName string, newUser NewUser) error {
	db := c.client.Database(dbName)

	// Empty list is never a valid value, and the MongoDB BSON encoder doesn't
	// seem to fully treat empty slices as empty, even though the docs say it should.
	if len(newUser.Mechanisms) == 0 {
		newUser.Mechanisms = nil
	}
	if newUser.Roles == nil {
		newUser.Roles = []RoleRef{}
	}

	var cmd = struct {
		NewUser        `bson:",inline"`
		DigestPassword bool `bson:"digestPassword"`
	}{
		NewUser:        newUser,
		DigestPassword: true,
	}
	result := db.RunCommand(ctx, cmd)
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

type UpdateUser struct {
	User       string            `bson:"updateUser"`
	Password   string            `bson:"pwd,omitempty"`
	CustomData map[string]string `bson:"customData,omitempty"`
	Roles      []RoleRef         `bson:"roles,omitempty"`
	Mechanisms []Mechanism       `bson:"mechanisms,omitempty"`
}

func (c *Client) UpdateDBUser(ctx context.Context, dbName string, update UpdateUser) (User, error) {
	if err := c.connect(ctx); err != nil {
		return User{}, err
	}
	if err := c.runUpdateUser(ctx, dbName, update); err != nil {
		return User{}, err
	}
	user, err := c.runUsersInfoSingle(ctx, dbName, update.User)
	if err != nil {
		return User{}, fmt.Errorf("get updated user: %w", err)
	}
	return user, nil
}

func (c *Client) runUpdateUser(ctx context.Context, dbName string, update UpdateUser) error {
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

func (c *Client) DeleteDBUser(ctx context.Context, dbName, userName string) error {
	if err := c.connect(ctx); err != nil {
		return err
	}
	return c.runDropUser(ctx, dbName, userName)
}

func (c *Client) runDropUser(ctx context.Context, dbName, userName string) error {
	db := c.client.Database(dbName)

	result := db.RunCommand(ctx, bson.D{
		{Key: "dropUser", Value: userName},
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
