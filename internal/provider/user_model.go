package provider

import (
	"github.2rioffice.com/platform/terraform-provider-mongodb-driver/internal/mongodb"
	dataschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var UserModelDataSourceSchema = dataschema.NestedAttributeObject{
	Attributes: map[string]dataschema.Attribute{
		"id": dataschema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "User unique ID in MongoDB. Is composed from the `db` and `user` fields.",
		},
		"user": dataschema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "Username for this MongoDB user.",
		},
		"db": dataschema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "Database this MongoDB user belongs to.",
		},
		"custom_data": dataschema.MapAttribute{
			Computed:            true,
			ElementType:         types.StringType,
			MarkdownDescription: "Any custom data for this user. Map of string key and values of arbitrary values.",
		},
		"roles": dataschema.ListNestedAttribute{
			Computed:            true,
			NestedObject:        RoleModelSchema,
			MarkdownDescription: "Roles this user belongs to.",
		},
		"mechanisms": dataschema.ListAttribute{
			Computed:            true,
			ElementType:         types.StringType,
			MarkdownDescription: "Authentication mechanisms this user can use.",
		},
	},
}

type UserModel struct {
	ID         types.String            `tfsdk:"id"`
	User       types.String            `tfsdk:"user"`
	DB         types.String            `tfsdk:"db"`
	CustomData map[string]types.String `tfsdk:"custom_data"`
	Roles      []RoleModel             `tfsdk:"roles"`
	Mechanisms []types.String          `tfsdk:"mechanisms"`
}

func ConvertUserSlice(users []mongodb.User) []UserModel {
	result := make([]UserModel, len(users))
	for i, user := range users {
		result[i] = ConvertUser(user)
	}
	return result
}

func ConvertUser(user mongodb.User) UserModel {
	return UserModel{
		ID:         types.StringValue(user.ID),
		User:       types.StringValue(user.User),
		DB:         types.StringValue(user.DB),
		CustomData: ConvertStringMap(user.CustomData),
		Roles:      ConvertRoleSlice(user.Roles),
		Mechanisms: ConvertStringSlice(user.Mechanisms),
	}
}

func ConvertStringSlice(slice []string) []types.String {
	result := make([]types.String, len(slice))
	for i, s := range slice {
		result[i] = types.StringValue(s)
	}
	return result
}

func ConvertStringMap(m map[string]string) map[string]types.String {
	result := make(map[string]types.String, len(m))
	for key, value := range m {
		result[key] = types.StringValue(value)
	}
	return result
}
