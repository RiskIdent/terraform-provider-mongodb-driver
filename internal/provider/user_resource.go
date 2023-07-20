// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.2rioffice.com/platform/terraform-provider-mongodb-driver/internal/mongodb"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &UserResource{}
var _ resource.ResourceWithConfigure = &UserResource{}
var _ resource.ResourceWithImportState = &UserResource{}

func NewUserResource() resource.Resource {
	return &UserResource{}
}

// UserResource defines the resource implementation.
type UserResource struct {
	client *mongodb.Client
}

// UserResourceModel describes the resource data model.
type UserResourceModel struct {
	ID         types.String            `tfsdk:"id"`
	User       types.String            `tfsdk:"user"`
	DB         types.String            `tfsdk:"db"`
	Password   types.String            `tfsdk:"pwd"`
	CustomData map[string]types.String `tfsdk:"custom_data"`
	Roles      []UserRoleResourceModel `tfsdk:"roles"`
	Mechanisms []types.String          `tfsdk:"mechanisms"`
}

func (u UserResourceModel) userAndDB() (string, string, error) {
	if !u.User.IsNull() && !u.DB.IsNull() {
		user := u.User.ValueString()
		db := u.DB.ValueString()
		if user != "" && db != "" {
			return user, db, nil
		}
	}
	if u.ID.IsNull() || u.ID.ValueString() == "" {
		return "", "", errors.New("missing user ID")
	}
	db, user, ok := strings.Cut(u.ID.ValueString(), ".")
	if !ok {
		return "", "", errors.New("malformed user ID, missing dot separator on db and user")
	}
	if db == "" {
		return "", "", errors.New("malformed user ID, missing db")
	}
	if user == "" {
		return "", "", errors.New("malformed user ID, missing user")
	}
	return user, db, nil
}

func (u *UserResourceModel) applyUser(user mongodb.User) {
	u.ID = types.StringValue(user.ID)
	u.User = types.StringValue(user.User)
	u.DB = types.StringValue(user.DB)
	if u.CustomData != nil {
		u.CustomData = toTypesStringMap(user.CustomData)
	}
	if u.Roles != nil {
		u.Roles = toTypesUserRoleResourceSlice(u.Roles, user.Roles)
	}
	if u.Mechanisms != nil {
		u.Mechanisms = toTypesStringSlice(user.Mechanisms)
	}
}

type UserRoleResourceModel struct {
	Role types.String `tfsdk:"role"`
	DB   types.String `tfsdk:"db"`
}

func (r UserRoleResourceModel) toRoleRef() mongodb.RoleRef {
	if r.DB.IsNull() {
		return mongodb.RoleSameDBRef(r.Role.ValueString())
	}
	return mongodb.RoleDBRef{
		Role: r.Role.ValueString(),
		DB:   r.DB.ValueString(),
	}
}

func fromTypesUserRoleResourceSlice(roles []UserRoleResourceModel) []mongodb.RoleRef {
	result := make([]mongodb.RoleRef, len(roles))
	for i, role := range roles {
		result[i] = role.toRoleRef()
	}
	return result
}

func toTypesUserRoleResourceSlice(oldRoles []UserRoleResourceModel, roles []mongodb.RoleDBRef) []UserRoleResourceModel {
	result := make([]UserRoleResourceModel, len(roles))
	for i, role := range roles {
		result[i] = toTypesUserRoleResource(oldRoles[i], role)
	}
	return result
}

func toTypesUserRoleResource(oldRole UserRoleResourceModel, role mongodb.RoleDBRef) UserRoleResourceModel {
	newRole := UserRoleResourceModel{
		Role: types.StringValue(role.Role),
		DB:   types.StringValue(role.DB),
	}
	if oldRole.DB.IsNull() {
		newRole.DB = types.StringNull()
	}
	return newRole
}

func (r *UserResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (r *UserResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "User resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "User unique ID in MongoDB. Is composed from the `db` and `user` fields.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"user": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Username for this MongoDB user.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"db": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Database this MongoDB user belongs to.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"pwd": schema.StringAttribute{
				Required:            true,
				Sensitive:           true,
				MarkdownDescription: "Password of this user.",
			},
			"custom_data": schema.MapAttribute{
				Optional:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Any custom data for this user. Map of string key and values of arbitrary values.",
			},
			"roles": schema.ListNestedAttribute{
				Optional: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"role": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "Role name",
						},
						"db": schema.StringAttribute{
							Optional:            true,
							MarkdownDescription: "Database this role belongs to. Leave unset to target same database as user.",
						},
					},
				},
				MarkdownDescription: "Roles this user belongs to.",
			},
			"mechanisms": schema.ListAttribute{
				Optional:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Authentication mechanisms this user can use.",
			},
		},
	}
}

func (r *UserResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*mongodb.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *mongodb.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	if err := client.Connect(ctx); err != nil {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Failed to connect to MongoDB. Error: %s", err),
		)
		return
	}

	r.client = client
}

func (r *UserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *UserResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	userName, dbName, err := data.userAndDB()
	if err != nil {
		resp.Diagnostics.AddError("Data Error", fmt.Sprintf("Unable to resolve username and database, got error: %s", err))
		return
	}

	user, err := r.client.CreateDBUser(ctx, dbName, mongodb.NewUser{
		User:       userName,
		Password:   data.Password.ValueString(),
		CustomData: fromTypesStringMap(data.CustomData),
		Roles:      fromTypesUserRoleResourceSlice(data.Roles),
		Mechanisms: fromTypesStringSlice[mongodb.Mechanism](data.Mechanisms),
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create user, got error: %s", err))
		return
	}

	data.applyUser(user)

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "created user")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *UserResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	userName, dbName, err := data.userAndDB()
	if err != nil {
		resp.Diagnostics.AddError("Data Error", fmt.Sprintf("Unable to resolve username and database, got error: %s", err))
		return
	}

	user, err := r.client.GetDBUser(ctx, dbName, userName)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read user, got error: %s", err))
		return
	}

	data.applyUser(user)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *UserResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	userName, dbName, err := data.userAndDB()
	if err != nil {
		resp.Diagnostics.AddError("Data Error", fmt.Sprintf("Unable to resolve username and database, got error: %s", err))
		return
	}

	user, err := r.client.UpdateDBUser(ctx, dbName, mongodb.UpdateUser{
		User:       userName,
		Password:   data.Password.ValueString(),
		CustomData: fromTypesStringMap(data.CustomData),
		Roles:      fromTypesUserRoleResourceSlice(data.Roles),
		Mechanisms: fromTypesStringSlice[mongodb.Mechanism](data.Mechanisms),
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update user, got error: %s", err))
		return
	}

	data.applyUser(user)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *UserResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	userName, dbName, err := data.userAndDB()
	if err != nil {
		resp.Diagnostics.AddError("Data Error", fmt.Sprintf("Unable to resolve username and database, got error: %s", err))
		return
	}

	if err := r.client.DeleteDBUser(ctx, dbName, userName); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete user, got error: %s", err))
		return
	}
}

func (r *UserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}