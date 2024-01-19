// SPDX-FileCopyrightText: 2023 Risk.Ident GmbH <contact@riskident.com>
//
// SPDX-License-Identifier: MPL-2.0

package mongodb

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var (
	ErrNotFound = errors.New("not found")
	ErrNotOK    = errors.New("not ok")

	AppName = "terraform-provider-mongodb-driver"
)

type Client struct {
	uri         string
	credentials Credentials
	connectOnce sync.Once
	client      *mongo.Client
	connectErr  error
}

type Credentials struct {
	Username string
	Password string
}

func New(uri string, cred Credentials) *Client {
	return &Client{
		uri:         uri,
		credentials: cred,
	}
}

func (c *Client) connect(ctx context.Context) error {
	c.connectOnce.Do(func() {
		opt := options.Client().ApplyURI(c.uri).SetDirect(true)
		opt.AppName = &AppName
		if c.credentials.Username != "" {
			opt.SetAuth(options.Credential{
				Username:    c.credentials.Username,
				Password:    c.credentials.Password,
				PasswordSet: true,
			})
		}

		client, connectErr := mongo.Connect(ctx, opt)
		if connectErr != nil {
			c.connectErr = fmt.Errorf("connect: %w", connectErr)
			return
		}
		if err := client.Ping(ctx, readpref.PrimaryPreferred()); err != nil {
			c.connectErr = fmt.Errorf("ping: %w", err)
			return
		}
		c.client = client
	})
	return c.connectErr
}

type CommandResponse struct {
	OK            int       `bson:"ok"`
	OperationTime time.Time `bson:"operationTime"`
}

type RoleRef interface {
	isRoleRef()
}

// RoleDBRef is a [RoleRef] that points to a role in a specific database.
type RoleDBRef struct {
	Role string `bson:"role"`
	DB   string `bson:"db"`
}

func (RoleDBRef) isRoleRef() {}

// RoleSameDBRef is a [RoleRef] that points to a role in the same database.
type RoleSameDBRef string

func (RoleSameDBRef) isRoleRef() {}

// Mechanism is the mechanism for user authentication.
//
// [https://www.mongodb.com/docs/manual/reference/parameters/#mongodb-parameter-param.authenticationMechanisms]
type Mechanism string

var Mechanisms = []Mechanism{
	MechanismSCRAMSHA1,
	MechanismSCRAMSHA256,
	MechanismMONGODBX509,
	MechanismPLAIN,
	MechanismGSSAPI,
}

const (
	// MechanismSCRAMSHA1 is the SCRAM mechanism for creating SCRAM user credentials.
	//
	// RFC 5802 standard Salted Challenge Response Authentication Mechanism using the SHA-1 hash function.
	//
	// - Uses the SHA-1 hashing function.
	//
	// [https://www.mongodb.com/docs/manual/core/security-scram/#std-label-authentication-scram-sha-1]
	MechanismSCRAMSHA1 Mechanism = "SCRAM-SHA-1"

	// MechanismSCRAMSHA256 is the SCRAM mechanism for creating SCRAM user credentials.
	//
	// RFC 7677 standard Salted Challenge Response Authentication Mechanism using the SHA-256 hash function.
	//
	// - Uses the SHA-256 hashing function.
	// - Requires featureCompatibilityVersion set to 4.0.
	// - Requires digestPassword to be true.
	//
	// [https://www.mongodb.com/docs/manual/core/security-scram/#std-label-authentication-scram-sha-256]
	MechanismSCRAMSHA256 Mechanism = "SCRAM-SHA-256"

	// MechanismMONGODBX509 is the mechanism for MongoDB TLS/SSL certificate authentication.
	//
	// [https://www.mongodb.com/docs/manual/core/security-x.509/#std-label-security-auth-x509]
	MechanismMONGODBX509 Mechanism = "MONGODB-X509"

	// MechanismPLAIN is the mechanism for external authentication using LDAP.
	// You can also use PLAIN for authenticating in-database users.
	// PLAIN transmits passwords in plain text.
	// This mechanism is available only in MongoDB Enterprise.
	//
	// [https://www.mongodb.com/docs/manual/core/authentication/#std-label-security-auth-ldap]
	MechanismPLAIN Mechanism = "PLAIN"

	// MechanismGSSAPI is the mechanism for external authentication using Kerberos.
	// This mechanism is available only in MongoDB Enterprise.
	//
	// [https://www.mongodb.com/docs/manual/core/authentication/#std-label-security-auth-kerberos]
	MechanismGSSAPI Mechanism = "GSSAPI"
)

func validateResponse(response CommandResponse) error {
	if response.OK != 1 {
		return fmt.Errorf("%w: ok=%d", ErrNotOK, response.OK)
	}
	return nil
}
