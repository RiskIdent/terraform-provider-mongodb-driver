package provider

import (
	"context"
	"fmt"

	"github.2rioffice.com/platform/terraform-provider-mongodb-driver/internal/mongodb"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewUsersDataSource() datasource.DataSource {
	return &UsersDataSource{}
}

type UsersDataSource struct {
	client *mongodb.Client
}

type UsersDataSourceModel struct {
	DB    types.String `tfsdk:"db"`
	Users []UserModel  `tfsdk:"users"`

	// MongoDB support more dynamic filters, but due to Terraform Plugin Framework
	// not supporting dynamic types, we have to settle on a simple map.
	// See: https://github.com/hashicorp/terraform-plugin-framework/issues/147
	Filter map[string]string `tfsdk:"filter"`
}

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &UsersDataSource{}
	_ datasource.DataSourceWithConfigure = &UsersDataSource{}
)

func (d *UsersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_users"
}

func (d *UsersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `MongoDB user listing data source`,

		Attributes: map[string]schema.Attribute{
			"db": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Which database to list users from. If `null`, then will list users in all databases.",
			},
			"users": schema.ListNestedAttribute{
				MarkdownDescription: "List of users fetched from MongoDB",
				Computed:            true,
				NestedObject:        UserModelDataSourceSchema,
			},
			"filter": schema.MapAttribute{
				Optional:            true,
				MarkdownDescription: "Additional filters to apply.",
				ElementType:         types.StringType,
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *UsersDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*mongodb.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *mongodb.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	if err := client.Connect(ctx); err != nil {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Failed to connect to MongoDB. Error: %s", err),
		)
		return
	}

	d.client = client
}

func (d *UsersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state UsersDataSourceModel
	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var filter any
	if state.Filter != nil {
		filter = state.Filter
	}

	var users []mongodb.User
	var err error
	if state.DB.IsNull() {
		users, err = d.client.ListAllUsers(ctx, filter)
	} else {
		users, err = d.client.ListDBUsers(ctx, state.DB.ValueString(), filter)
	}

	if err != nil {
		resp.Diagnostics.AddError("Reading MongoDB users",
			fmt.Sprintf("Failed to get the list of users from MongoDB. Error: %s", err),
		)
	}

	state.Users = ConvertUserSlice(users)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
