// SPDX-FileCopyrightText: 2021 HashiCorp, Inc.
// SPDX-FileCopyrightText: 2023 Risk.Ident GmbH <contact@riskident.com>
//
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/RiskIdent/terraform-provider-mongodb-driver/internal/mongodb"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &RoleResource{}
var _ resource.ResourceWithConfigure = &RoleResource{}
var _ resource.ResourceWithImportState = &RoleResource{}

func NewRoleResource() resource.Resource {
	return &RoleResource{}
}

// RoleResource defines the resource implementation.
type RoleResource struct {
	client *mongodb.Client
}

// RoleResourceModel describes the resource data model.
type RoleResourceModel struct {
	ID         types.String             `tfsdk:"id"`
	Role       types.String             `tfsdk:"role"`
	DB         types.String             `tfsdk:"db"`
	Roles      []RoleRefResourceModel   `tfsdk:"roles"`
	Privileges []PrivilegeResourceModel `tfsdk:"privileges"`
	Timeouts   timeouts.Value           `tfsdk:"timeouts"`
}

func (u RoleResourceModel) roleAndDB() (string, string, error) {
	if !u.Role.IsNull() && !u.DB.IsNull() {
		role := u.Role.ValueString()
		db := u.DB.ValueString()
		if role != "" && db != "" {
			return role, db, nil
		}
	}
	if u.ID.IsNull() || u.ID.ValueString() == "" {
		return "", "", errors.New("missing role ID")
	}
	db, role, ok := strings.Cut(u.ID.ValueString(), ".")
	if !ok {
		return "", "", errors.New("malformed role ID, missing dot separator on db and role")
	}
	if db == "" {
		return "", "", errors.New("malformed role ID, missing db")
	}
	if role == "" {
		return "", "", errors.New("malformed role ID, missing role")
	}
	return role, db, nil
}

func (u *RoleResourceModel) applyRole(role mongodb.Role) error {
	u.ID = types.StringValue(role.ID)
	u.Role = types.StringValue(role.Role)
	u.DB = types.StringValue(role.DB)
	if u.Roles != nil {
		u.Roles = toTypesRoleRefResourceSlice(u.Roles, role.Roles)
	}
	if u.Privileges != nil {
		privs, err := toTypesPrivilegeResourceSlice(role.Privileges)
		if err != nil {
			return err
		}
		u.Privileges = privs
	}
	return nil
}

func (r *RoleResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_role"
}

func (r *RoleResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Role resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Role unique ID in MongoDB. Is composed from the `db` and `role` fields.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"role": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Rolename for this MongoDB role.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					// MongoDB restricts rolename on "at least one character and
					// cannot be larger than 7MB". Lol!
					// https://www.mongodb.com/docs/v6.0/reference/command/createRole/#rolename-limits
					stringvalidator.LengthBetween(1, 7*1000*1000),
				},
			},
			"db": schema.StringAttribute{
				Required: true,
				MarkdownDescription: "Database this MongoDB role belongs to.\n\n" +
					// Indenting here because the documentation generation doesn't do it
					"  MongoDB has some restrictions on database names. Such as:\n\n" +
					"  - Cannot contain any of the following characters (we're following Windows limits): `/\\. \"$*<>:|?`\n" +
					"  - Cannot create roles in the `local` database.\n" +
					"  - Cannot be empty.\n" +
					"  - Cannot be longer than 64 characters.\n\n" +
					"  See documentation:\n\n" +
					"  - <https://www.mongodb.com/docs/manual/reference/command/createRole/#local-database>\n" +
					"  - <https://www.mongodb.com/docs/v6.0/reference/limits/#naming-restrictions>",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: append(
					databaseValidators,
					stringvalidator.NoneOf("local"),
				),
			},
			"roles": schema.SetNestedAttribute{
				Optional:            true,
				MarkdownDescription: "Roles this role inherits privileges from.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"role": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "Role name",
							Validators: []validator.String{
								stringvalidator.LengthAtLeast(1),
							},
						},
						"db": schema.StringAttribute{
							Optional:            true,
							MarkdownDescription: "Database this role belongs to. Leave unset to target same database as role.",
							Validators:          databaseValidators,
						},
					},
				},
			},
			"privileges": schema.SetNestedAttribute{
				Optional:            true,
				MarkdownDescription: "Privileges this role has.",
				NestedObject:        privilegeResourceNestedSchema,
			},
			"timeouts": timeouts.AttributesAll(ctx),
		},
	}
}

func (r *RoleResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.client = client
}

func (r *RoleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *RoleResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createTimeout, diags := data.Timeouts.Create(ctx, DefaultTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, createTimeout)
	defer cancel()

	roleName, dbName, err := data.roleAndDB()
	if err != nil {
		resp.Diagnostics.AddError("Data Error", fmt.Sprintf("Unable to resolve rolename and database, got error: %s", err))
		return
	}

	role, err := r.client.CreateDBRole(ctx, dbName, mongodb.NewRole{
		Role:       roleName,
		Roles:      fromTypesRoleRefResourceSlice(data.Roles),
		Privileges: fromTypesPrivilegeResourceSlice(data.Privileges),
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create role, got error: %s", err))
		return
	}

	if err := data.applyRole(role); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to interpret database response, got error: %s", err))
		return
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "created role")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RoleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *RoleResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	readTimeout, diags := data.Timeouts.Read(ctx, DefaultTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, readTimeout)
	defer cancel()

	roleName, dbName, err := data.roleAndDB()
	if err != nil {
		resp.Diagnostics.AddError("Data Error", fmt.Sprintf("Unable to resolve rolename and database, got error: %s", err))
		return
	}

	role, err := r.client.GetDBRole(ctx, dbName, roleName)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read role, got error: %s", err))
		return
	}

	if err := data.applyRole(role); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to interpret database response, got error: %s", err))
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RoleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *RoleResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateTimeout, diags := data.Timeouts.Update(ctx, DefaultTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, updateTimeout)
	defer cancel()

	roleName, dbName, err := data.roleAndDB()
	if err != nil {
		resp.Diagnostics.AddError("Data Error", fmt.Sprintf("Unable to resolve rolename and database, got error: %s", err))
		return
	}

	role, err := r.client.UpdateDBRole(ctx, dbName, mongodb.UpdateRole{
		Role:       roleName,
		Roles:      fromTypesRoleRefResourceSlice(data.Roles),
		Privileges: fromTypesPrivilegeResourceSlice(data.Privileges),
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update role, got error: %s", err))
		return
	}

	if err := data.applyRole(role); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to interpret database response, got error: %s", err))
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RoleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *RoleResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleteTimeout, diags := data.Timeouts.Delete(ctx, DefaultTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, deleteTimeout)
	defer cancel()

	roleName, dbName, err := data.roleAndDB()
	if err != nil {
		resp.Diagnostics.AddError("Data Error", fmt.Sprintf("Unable to resolve rolename and database, got error: %s", err))
		return
	}

	if err := r.client.DeleteDBRole(ctx, dbName, roleName); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete role, got error: %s", err))
		return
	}
}

func (r *RoleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
