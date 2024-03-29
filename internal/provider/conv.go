// SPDX-FileCopyrightText: 2023 Risk.Ident GmbH <contact@riskident.com>
//
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func toTypesStringSlice[E ~string](slice []E) []types.String {
	result := make([]types.String, len(slice))
	for i, s := range slice {
		result[i] = types.StringValue(string(s))
	}
	return result
}

func fromTypesStringSlice[E ~string](slice []types.String) []E {
	result := make([]E, len(slice))
	for i, s := range slice {
		result[i] = E(s.ValueString())
	}
	return result
}

func toTypesStringMap(m map[string]string) map[string]types.String {
	result := make(map[string]types.String, len(m))
	for key, value := range m {
		result[key] = types.StringValue(value)
	}
	return result
}

func fromTypesStringMap(m map[string]types.String) map[string]string {
	result := make(map[string]string, len(m))
	for key, value := range m {
		result[key] = value.ValueString()
	}
	return result
}

func castToStringSlice[E ~string](slice []E) []string {
	result := make([]string, len(slice))
	for i, s := range slice {
		result[i] = string(s)
	}
	return result
}
