package provider

import (
	"fmt"
	"sort"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/zesty-co/terraform-provider-zesty/internal/models"
	"gopkg.in/yaml.v3"
)

func ToModel(account *models.Account) (*accountModel, diag.Diagnostics) {
	roleARN, exists := account.AdditionalData["roleARN"]
	if !exists {
		return nil, diag.Diagnostics{
			diag.NewErrorDiagnostic(
				"Missing role ARN for account",
				"account.AdditionalData.roleARN is nil or empty",
			),
		}
	}

	roleARNString, ok := roleARN.(string)
	if !ok {
		return nil, diag.Diagnostics{
			diag.NewErrorDiagnostic(
				"Erroneous role ARN for account",
				fmt.Sprintf("Expected string for role ARN but got %T", roleARN),
			),
		}
	}

	externalID, exists := account.AdditionalData["externalID"]
	if !exists {
		return nil, diag.Diagnostics{
			diag.NewErrorDiagnostic(
				"Missing external ID for account",
				"account.AdditionalData.externalID is nil or empty",
			),
		}
	}

	externalIDString, ok := externalID.(string)
	if !ok {
		return nil, diag.Diagnostics{
			diag.NewErrorDiagnostic(
				"Erroneous external ID for account",
				fmt.Sprintf("Expected string for external ID but got %T", roleARN),
			),
		}
	}

	rawValues := parseValues(account.AdditionalData)
	valuesBytes, err := yaml.Marshal(rawValues)
	if err != nil {
		return nil, diag.Diagnostics{
			diag.NewErrorDiagnostic(
				"Erroneous values from provider",
				fmt.Sprintf("Got error: %v", err),
			),
		}
	}

	model := accountModel{
		ID:            types.StringValue(account.AccountID),
		Region:        types.StringPointerValue(account.Region),
		CloudProvider: types.StringValue(string(account.CloudProvider)),
		RoleARN:       types.StringValue(roleARNString),
		ExternalID:    types.StringValue(externalIDString),
	}

	var productNames []string
	for name := range account.Products {
		productNames = append(productNames, string(name))
	}
	sort.Strings(productNames)

	model.Products = []productModel{}
	for _, name := range productNames {
		details := account.Products[models.Product(name)]
		model.Products = append(model.Products, productModel{
			Name:   types.StringValue(name),
			Active: types.BoolValue(details.Active),
			Values: types.StringValue(string(valuesBytes)),
		})
	}
	if account.Cur != nil {
		model.Cur = &curModel{
			S3Bucket:   types.StringValue(account.Cur.S3Bucket),
			ExportName: types.StringValue(account.Cur.ExportName),
			Type:       types.StringValue(account.Cur.Type),
		}
	}
	if account.Athena != nil {
		model.Athena = &athenaModel{
			AthenaDB:        types.StringValue(account.Athena.AthenaDB),
			AthenaS3Bucket:  types.StringValue(account.Athena.AthenaS3Bucket),
			AthenaProjectID: types.StringValue(account.Athena.AthenaProjectID),
			AthenaRegion:    types.StringValue(account.Athena.AthenaRegion),
			AthenaTable:     types.StringValue(account.Athena.AthenaTable),
			AthenaWorkgroup: types.StringValue(account.Athena.AthenaWorkgroup),
			AthenaCatalog:   types.StringValue(account.Athena.AthenaCatalog),
		}
	}

	return &model, nil
}

func parseValues(input map[string]any) map[string]any {
	values, ok := input["values"]
	if !ok {
		return map[string]any{}
	}
	valuesMap, ok := values.(map[string]any)
	if !ok {
		return map[string]any{}
	}

	clean := make(map[string]any)
	for k, v := range valuesMap {
		if v != nil && k != "metadata" {
			clean[k] = v
		}
	}
	return clean
}
