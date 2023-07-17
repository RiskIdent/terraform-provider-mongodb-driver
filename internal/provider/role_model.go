package provider

import (
	"github.2rioffice.com/platform/terraform-provider-mongodb-driver/internal/mongodb"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var RoleModelSchema = schema.NestedAttributeObject{
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
}

type RoleModel struct {
	Role types.String `tfsdk:"role"`
	DB   types.String `tfsdk:"db"`
}

func ConvertRoleSlice(slice []mongodb.RoleRef) []RoleModel {
	result := make([]RoleModel, len(slice))
	for i, role := range slice {
		result[i] = ConvertRole(role)
	}
	return result
}

func ConvertRole(role mongodb.RoleRef) RoleModel {
	return RoleModel{
		Role: types.StringValue(role.Role),
		DB:   types.StringValue(role.DB),
	}
}
