// SPDX-FileCopyrightText: 2023 Risk.Ident GmbH <contact@riskident.com>
//
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"github.com/RiskIdent/terraform-provider-mongodb-driver/internal/mongodb"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var privilegeResourceNestedSchema = schema.NestedAttributeObject{
	Attributes: map[string]schema.Attribute{
		"resource": schema.SingleNestedAttribute{
			Required:            true,
			MarkdownDescription: "A document that specifies the resources upon which the privilege `actions` apply.",
			Attributes:          resourceResourceAttributesSchema,
		},
		"actions": schema.SetAttribute{
			Optional:            true,
			MarkdownDescription: "Database this role belongs to. Leave unset to target same database as role.",
			ElementType:         types.StringType,
		},
	},
}

type PrivilegeResourceModel struct {
	Resource ResourceResourceModel `tfsdk:"resource"`
	Actions  []types.String        `tfsdk:"actions"`
}

func (r PrivilegeResourceModel) toPrivilege() mongodb.Privilege {
	return mongodb.Privilege{
		Resource: mongodb.ResourceWrapper{Union: r.Resource.toResource()},
		Actions:  fromTypesStringSlice[string](r.Actions),
	}
}

func fromTypesPrivilegeResourceSlice(privileges []PrivilegeResourceModel) []mongodb.Privilege {
	result := make([]mongodb.Privilege, len(privileges))
	for i, priv := range privileges {
		result[i] = priv.toPrivilege()
	}
	return result
}

func toTypesPrivilegeResourceSlice(privileges []mongodb.Privilege) ([]PrivilegeResourceModel, error) {
	result := make([]PrivilegeResourceModel, len(privileges))
	for i, priv := range privileges {
		p, err := toTypesPrivilegeResource(priv)
		if err != nil {
			return nil, err
		}
		result[i] = p
	}
	return result, nil
}

func toTypesPrivilegeResource(privileges mongodb.Privilege) (PrivilegeResourceModel, error) {
	resource, err := toTypesResourceResource(privileges.Resource.Union)
	if err != nil {
		return PrivilegeResourceModel{}, err
	}
	return PrivilegeResourceModel{
		Resource: resource,
		Actions:  toTypesStringSlice(privileges.Actions),
	}, nil
}
