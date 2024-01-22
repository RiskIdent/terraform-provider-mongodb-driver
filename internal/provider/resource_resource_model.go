// SPDX-FileCopyrightText: 2023 Risk.Ident GmbH <contact@riskident.com>
//
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"

	"github.com/RiskIdent/terraform-provider-mongodb-driver/internal/mongodb"
	"github.com/hashicorp/terraform-plugin-framework-validators/boolvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var resourceResourceAttributesSchema = map[string]schema.Attribute{
	"cluster": schema.BoolAttribute{
		Optional:            true,
		MarkdownDescription: "Set to true to target the MongoDB cluster as the resource.",
		Validators: []validator.Bool{
			boolvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("db")),
			boolvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("collection")),
		},
	},
	"any_resource": schema.BoolAttribute{
		Optional: true,
		MarkdownDescription: "Set to true to target every resource in the system. " +
			"Intended for internal use. **Do not** use this resource, " +
			"other than in exceptional circumstances.",
		Validators: []validator.Bool{
			boolvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("db")),
			boolvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("collection")),
		},
	},
	"db": schema.StringAttribute{
		Optional: true,
		MarkdownDescription: "Specify which database to target. Must be paired with the `collection` attribute. " +
			"If both the `db` and `collections` are empty strings (`\"\"`), " +
			"the resource is all collections, excluding the system collections, in all the databases. " +
			"If only the `db` attribute is an empty string (`\"\"`), " +
			"the resource is all collections with the specified `collection` name across all databases." +
			"If only the `collection` attribute is an empty string (`\"\"`), " +
			"the resource is the specified database, excluding the system collections.",
		Validators: append(optionalDatabaseValidators, []validator.String{
			stringvalidator.AlsoRequires(path.MatchRelative().AtParent().AtName("collection")),
		}...),
	},
	"collection": schema.StringAttribute{
		Optional:            true,
		MarkdownDescription: "Specify which collection to target. Must be paired with the `db` attribute.",
		Validators: []validator.String{
			stringvalidator.AlsoRequires(path.MatchRelative().AtParent().AtName("db")),
		},
	},
}

type ResourceResourceModel struct {
	Cluster     types.Bool   `tfsdk:"cluster"`
	AnyResource types.Bool   `tfsdk:"any_resource"`
	DB          types.String `tfsdk:"db"`
	Collection  types.String `tfsdk:"collection"`
}

func (r ResourceResourceModel) toResource() mongodb.Resource {
	if !r.Cluster.IsNull() && r.Cluster.ValueBool() {
		return mongodb.ResourceCluster{Cluster: true}
	}
	if !r.AnyResource.IsNull() && r.AnyResource.ValueBool() {
		return mongodb.ResourceAny{AnyResource: true}
	}
	return mongodb.ResourceCollection{
		DB:         r.DB.ValueString(),
		Collection: r.Collection.ValueString(),
	}
}

func toTypesResourceResource(resource mongodb.Resource) (ResourceResourceModel, error) {
	switch resource := resource.(type) {
	case mongodb.ResourceCluster:
		return ResourceResourceModel{Cluster: types.BoolValue(resource.Cluster)}, nil
	case mongodb.ResourceAny:
		return ResourceResourceModel{AnyResource: types.BoolValue(resource.AnyResource)}, nil
	case mongodb.ResourceCollection:
		return ResourceResourceModel{
			DB:         types.StringValue(resource.DB),
			Collection: types.StringValue(resource.Collection),
		}, nil
	default:
		return ResourceResourceModel{}, fmt.Errorf("unsupported resource type: %T", resource)
	}
}
