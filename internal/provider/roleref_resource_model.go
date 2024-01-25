// SPDX-FileCopyrightText: 2023 Risk.Ident GmbH <contact@riskident.com>
//
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"github.com/RiskIdent/terraform-provider-mongodb-driver/internal/mongodb"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type RoleRefResourceModel struct {
	Role types.String `tfsdk:"role"`
	DB   types.String `tfsdk:"db"`
}

func (r RoleRefResourceModel) toRoleRef() mongodb.RoleRef {
	if r.DB.IsNull() {
		return mongodb.RoleSameDBRef(r.Role.ValueString())
	}
	return mongodb.RoleDBRef{
		Role: r.Role.ValueString(),
		DB:   r.DB.ValueString(),
	}
}

func fromTypesRoleRefResourceSlice(roles []RoleRefResourceModel) []mongodb.RoleRef {
	result := make([]mongodb.RoleRef, len(roles))
	for i, role := range roles {
		result[i] = role.toRoleRef()
	}
	return result
}

func toTypesRoleRefResourceSlice(oldRoles []RoleRefResourceModel, roles []mongodb.RoleDBRef) []RoleRefResourceModel {
	result := make([]RoleRefResourceModel, len(roles))
	for i, role := range roles {
		result[i] = toTypesRoleRefResource(oldRoles[i], role)
	}
	return result
}

func toTypesRoleRefResource(oldRole RoleRefResourceModel, role mongodb.RoleDBRef) RoleRefResourceModel {
	newRole := RoleRefResourceModel{
		Role: types.StringValue(role.Role),
		DB:   types.StringValue(role.DB),
	}
	if oldRole.DB.IsNull() {
		newRole.DB = types.StringNull()
	}
	return newRole
}
