package mongodb

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
	client      *mongo.Client
	connectOnce sync.Once
}

type Credentials struct {
	Username string
	Password string
}

func New(uri string, cred Credentials) (*Client, error) {
	opt := options.Client().ApplyURI(uri).SetDirect(true)
	opt.AppName = &AppName
	if cred.Username != "" {
		opt.SetAuth(options.Credential{
			Username:    cred.Username,
			Password:    cred.Password,
			PasswordSet: true,
		})
	}
	client, err := mongo.NewClient(opt)
	if err != nil {
		return nil, err
	}
	return &Client{
		client: client,
	}, nil
}

func (c *Client) Connect(ctx context.Context) error {
	var connectErr error
	c.connectOnce.Do(func() {
		ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
		defer cancel()
		if err := c.client.Connect(ctx); err != nil {
			connectErr = fmt.Errorf("connect: %w", err)
			return
		}
		if err := c.client.Ping(ctx, readpref.PrimaryPreferred()); err != nil {
			connectErr = fmt.Errorf("ping: %w", err)
			return
		}
	})
	return connectErr
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
	query := bson.D{
		{Key: "usersInfo", Value: 1}, // 1 = all in current db
	}
	if filter != nil {
		query = append(query, bson.E{Key: "filter", Value: filter})
	}
	return c.listUsers(ctx, dbName, query)
}

func (c *Client) ListAllUsers(ctx context.Context, filter any) ([]User, error) {
	query := bson.D{
		{Key: "usersInfo", Value: bson.D{
			{Key: "forAllDBs", Value: true},
		}},
	}
	if filter != nil {
		query = append(query, bson.E{Key: "filter", Value: filter})
	}
	return c.listUsers(ctx, "admin", query)
}

func (c *Client) GetDBUser(ctx context.Context, dbName, userName string) (User, error) {
	query := bson.D{
		{Key: "usersInfo", Value: userName},
	}
	users, err := c.listUsers(ctx, dbName, query)
	if err != nil {
		return User{}, err
	}
	if len(users) == 0 {
		return User{}, ErrNotFound
	}
	return users[0], nil
}

func (c *Client) listUsers(ctx context.Context, dbName string, query bson.D) ([]User, error) {
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
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
	if err := c.createDBUser(ctx, dbName, newUser); err != nil {
		return User{}, err
	}
	user, err := c.GetDBUser(ctx, dbName, newUser.User)
	if err != nil {
		return User{}, fmt.Errorf("get created user: %w", err)
	}
	return user, nil
}

func (c *Client) createDBUser(ctx context.Context, dbName string, newUser NewUser) error {
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	db := c.client.Database(dbName)

	result := db.RunCommand(ctx, newUser)
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
	if err := c.updateDBUser(ctx, dbName, update); err != nil {
		return User{}, err
	}
	user, err := c.GetDBUser(ctx, dbName, update.User)
	if err != nil {
		return User{}, fmt.Errorf("get updated user: %w", err)
	}
	return user, nil
}

func (c *Client) updateDBUser(ctx context.Context, dbName string, update UpdateUser) error {
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
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
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
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
	if err := validateResponse(response); err != nil {
		return err
	}
	return nil
}

func validateResponse(response CommandResponse) error {
	if response.OK != 1 {
		return fmt.Errorf("%w: ok=%d", ErrNotOK, response.OK)
	}
	return nil
}
