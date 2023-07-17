package mongodb

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type Client struct {
	client    *mongo.Client
	connected bool
}

type Credentials struct {
	Username string
	Password string
}

func New(uri string, cred Credentials) (*Client, error) {
	opt := options.Client().ApplyURI(uri)
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
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	if !c.connected {
		if err := c.client.Connect(ctx); err != nil {
			return fmt.Errorf("connect: %w", err)
		}
		c.connected = true
	}
	if err := c.client.Ping(ctx, readpref.PrimaryPreferred()); err != nil {
		return fmt.Errorf("ping: %w", err)
	}
	return nil
}

type CommandResponse struct {
	OK            int       `bson:"ok"`
	OperationTime time.Time `bson:"operationTime"`
}

type RoleRef struct {
	Role string `bson:"role"`
	DB   string `bson:"db"`
}

type User struct {
	ID         string            `bson:"_id"`
	UserID     primitive.Binary  `bson:"userId"`
	User       string            `bson:"user"`
	DB         string            `bson:"db"`
	CustomData map[string]string `bson:"customData"`
	Roles      []RoleRef         `bson:"roles"`
	Mechanisms []string          `bson:"mechanisms"`
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
	return response.Users, nil
}
