// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"os"

	"github.2rioffice.com/platform/terraform-provider-mongodb-driver/internal/mongodb"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ provider.Provider = &mongodbProvider{}
)

// New is a helper function to simplify provider server and testing implementation.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &mongodbProvider{
			version: version,
		}
	}
}

// mongodbProvider is the provider implementation.
type mongodbProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// Metadata returns the provider type name.
func (p *mongodbProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "mongodb"
	resp.Version = p.version
}

// Schema defines the provider-level schema for configuration data.
func (p *mongodbProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"uri": schema.StringAttribute{
				Optional:  false,
				Required:  true,
				Sensitive: true,
				MarkdownDescription: `Connection URI, also known as the connection string.
Tells the provider how to reach MongoDB.

The implementation is based on the Go SDK.
For a full reference of connection URI options available for the Go SDK, see this page: <https://www.mongodb.com/docs/drivers/go/current/fundamentals/connection/>`,
			},
			"username": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Allows specifying the username for the connection. Setting this will override any credentials used in the connection URI.",
			},
			"password": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "Allows specifying the password for the connection. You must also set the `username` attribute when using this attribute.",
			},
		},
	}
}

// mongodbProviderModel maps provider schema data to a Go type.
type mongodbProviderModel struct {
	URI      types.String `tfsdk:"uri"`
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
}

func (p *mongodbProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	// Retrieve provider data from configuration
	var config mongodbProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If practitioner provided a configuration value for any of the
	// attributes, it must be a known value.

	if config.URI.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("uri"),
			"Unknown MongoDB connection URI",
			"The provider cannot create the MongoDB client as there is an unknown configuration value for the MongoDB connection URI. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the MONGODB_URI environment variable.",
		)
	}

	if config.Username.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("username"),
			"Unknown MongoDB username",
			"The provider cannot create the MongoDB client as there is an unknown configuration value for the MongoDB username. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the MONGODB_USERNAME environment variable.",
		)
	}

	if config.Password.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("password"),
			"Unknown MongoDB password",
			"The provider cannot create the MongoDB client as there is an unknown configuration value for the MongoDB password. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the MONGODB_PASSWORD environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Default values to environment variables, but override
	// with Terraform configuration value if set.

	uri := os.Getenv("MONGODB_URI")
	username := os.Getenv("MONGODB_USERNAME")
	password := os.Getenv("MONGODB_PASSWORD")

	if !config.URI.IsNull() {
		uri = config.URI.ValueString()
	}

	if !config.Username.IsNull() {
		username = config.Username.ValueString()
	}

	if !config.Password.IsNull() {
		password = config.Password.ValueString()
	}

	// If any of the expected configurations are missing, return
	// errors with provider-specific guidance.

	if uri == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("uri"),
			"Missing MongoDB connection URI",
			"The provider cannot create the MongoDB client as there is a missing or empty value for the MongoDB connection URI. "+
				"Set the host value in the configuration or use the MONGODB_URI environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if password != "" && username == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("username"),
			"Missing MongoDB username",
			"The username must be specified if the password is. Cannot override only the password.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Create a new HashiCups client using the configuration values
	client, err := mongodb.New(uri, mongodb.Credentials{
		Username: username,
		Password: password,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create MongoDB Client",
			"An unexpected error occurred when creating the MongoDB client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"MongoDB Client Error: "+err.Error(),
		)
		return
	}

	// Make the HashiCups client available during DataSource and Resource
	// type Configure methods.
	resp.DataSourceData = client
	resp.ResourceData = client
}

// DataSources defines the data sources implemented in the provider.
func (p *mongodbProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewUsersDataSource,
	}
}

// Resources defines the resources implemented in the provider.
func (p *mongodbProvider) Resources(_ context.Context) []func() resource.Resource {
	return nil
}
