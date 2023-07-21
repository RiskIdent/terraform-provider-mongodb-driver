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
	DB    types.String          `tfsdk:"db"`
	Users []UserDataSourceModel `tfsdk:"users"`

	// MongoDB support more dynamic filters, but due to Terraform Plugin Framework
	// not supporting dynamic types, we have to settle on a simple map.
	// See: https://github.com/hashicorp/terraform-plugin-framework/issues/147
	Filter map[string]string `tfsdk:"filter"`
}

type UserDataSourceModel struct {
	ID         types.String              `tfsdk:"id"`
	User       types.String              `tfsdk:"user"`
	DB         types.String              `tfsdk:"db"`
	CustomData map[string]types.String   `tfsdk:"custom_data"`
	Roles      []UserRoleDataSourceModel `tfsdk:"roles"`
	Mechanisms []types.String            `tfsdk:"mechanisms"`
}

func toTypesUserDataSourceSlice(users []mongodb.User) []UserDataSourceModel {
	result := make([]UserDataSourceModel, len(users))
	for i, user := range users {
		result[i] = toTypesUserDataSource(user)
	}
	return result
}

func toTypesUserDataSource(user mongodb.User) UserDataSourceModel {
	return UserDataSourceModel{
		ID:         types.StringValue(user.ID),
		User:       types.StringValue(user.User),
		DB:         types.StringValue(user.DB),
		CustomData: toTypesStringMap(user.CustomData),
		Roles:      toTypesUserRoleDataSourceSlice(user.Roles),
		Mechanisms: toTypesStringSlice(user.Mechanisms),
	}
}

type UserRoleDataSourceModel struct {
	Role types.String `tfsdk:"role"`
	DB   types.String `tfsdk:"db"`
}

func toTypesUserRoleDataSourceSlice(slice []mongodb.RoleDBRef) []UserRoleDataSourceModel {
	result := make([]UserRoleDataSourceModel, len(slice))
	for i, role := range slice {
		result[i] = toTypesUserRoleDataSource(role)
	}
	return result
}

func toTypesUserRoleDataSource(role mongodb.RoleDBRef) UserRoleDataSourceModel {
	return UserRoleDataSourceModel{
		Role: types.StringValue(role.Role),
		DB:   types.StringValue(role.DB),
	}
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
				Optional: true,
				MarkdownDescription: "Which database to list users from. If `null`, then will list users in all databases.\n\n" +
					// Indenting here because the documentation generation doesn't do it
					"  MongoDB has some restrictions on database names. Such as:\n\n" +
					"  - Cannot contain any of the following characters (we're following Windows limits): `/\\. \"$*<>:|?`\n" +
					"  - Cannot be empty.\n" +
					"  - Cannot be longer than 64 characters.\n\n" +
					"  See documentation:\n\n" +
					"  - <https://www.mongodb.com/docs/v6.0/reference/limits/#naming-restrictions>",
				Validators: databaseValidators,
			},
			"users": schema.ListNestedAttribute{
				MarkdownDescription: "List of users fetched from MongoDB",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "User unique ID in MongoDB. Is composed from the `db` and `user` fields.",
						},
						"user": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Username for this MongoDB user.",
						},
						"db": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Database this MongoDB user belongs to.",
						},
						"custom_data": schema.MapAttribute{
							Computed:            true,
							ElementType:         types.StringType,
							MarkdownDescription: "Any custom data for this user. Map of string key and values of arbitrary values.",
						},
						"roles": schema.ListNestedAttribute{
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"role": schema.StringAttribute{
										Computed:            true,
										MarkdownDescription: "Role name",
									},
									"db": schema.StringAttribute{
										Computed:            true,
										MarkdownDescription: "Database this role belongs to.",
									},
								},
							},
							MarkdownDescription: "Roles this user belongs to.",
						},
						"mechanisms": schema.ListAttribute{
							Computed:            true,
							ElementType:         types.StringType,
							MarkdownDescription: "Authentication mechanisms this user can use.",
						},
					},
				},
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

	state.Users = toTypesUserDataSourceSlice(users)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
