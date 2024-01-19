// SPDX-FileCopyrightText: 2023 Risk.Ident GmbH <contact@riskident.com>
//
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var databaseValidators = []validator.String{
	stringvalidator.LengthBetween(1, 64),
	stringvalidator.RegexMatches(regexp.MustCompile(`^[^\/\\. "$*<>:|?\0]*$`),
		`MongoDB has restrictions on database name. We're limiting on the Windows restrictions here to be safe. See https://www.mongodb.com/docs/v6.0/reference/limits/#naming-restrictions`),
}

var optionalDatabaseValidators = []validator.String{
	stringvalidator.LengthBetween(0, 64),
	stringvalidator.RegexMatches(regexp.MustCompile(`^[^\/\\. "$*<>:|?\0]*$`),
		`MongoDB has restrictions on database name. We're limiting on the Windows restrictions here to be safe. See https://www.mongodb.com/docs/v6.0/reference/limits/#naming-restrictions`),
}
