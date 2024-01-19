// SPDX-FileCopyrightText: 2023 Risk.Ident GmbH <contact@riskident.com>
//
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"github.com/RiskIdent/terraform-provider-mongodb-driver/internal/mongodb"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type PrivilegeResourceModel struct {
	Resource ResourceResourceModel `tfsdk:"resource"`
	Actions  []types.String        `tfsdk:"actions"`
}

func (r PrivilegeResourceModel) toPrivilege() mongodb.Privilege {
	return mongodb.Privilege{
		Resource: mongodb.ResourceWrapper{Union: r.Resource.toResource()},
		Actions:  toTypesStringSlice(r.Actions),
	}
}
