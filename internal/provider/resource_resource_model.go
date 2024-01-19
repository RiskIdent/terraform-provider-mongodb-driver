// SPDX-FileCopyrightText: 2023 Risk.Ident GmbH <contact@riskident.com>
//
// SPDX-License-Identifier: MPL-2.0

package provider

import "github.com/RiskIdent/terraform-provider-mongodb-driver/internal/mongodb"

type ResourceResourceModel struct {
}

func (r ResourceResourceModel) toResource() mongodb.Resource {
	if r.DB.IsNull() {
		return mongodb.RoleSameDBRef(r.Role.ValueString())
	}
	return mongodb.RoleDBRef{
		Role: r.Role.ValueString(),
		DB:   r.DB.ValueString(),
	}
}
