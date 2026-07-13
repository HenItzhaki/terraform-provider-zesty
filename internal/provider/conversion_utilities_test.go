package provider_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zesty-co/terraform-provider-zesty/internal/models"
	"github.com/zesty-co/terraform-provider-zesty/internal/provider"
)

func TestToModel(t *testing.T) {
	tests := []struct {
		name             string
		account          *models.Account
		expectedErrorMsg string
	}{
		{
			name:             "nil roleARN",
			account:          &models.Account{AdditionalData: map[string]any{"externalID": "ext"}, AccountID: "acc", CloudProvider: "aws"},
			expectedErrorMsg: "Missing role ARN for account",
		},
		{
			name:             "non-string roleARN",
			account:          &models.Account{AdditionalData: map[string]any{"roleARN": 123, "externalID": "ext"}, AccountID: "acc", CloudProvider: "aws"},
			expectedErrorMsg: "Erroneous role ARN for account",
		},
		{
			name:             "missing externalID",
			account:          &models.Account{AdditionalData: map[string]any{"roleARN": "arn:aws"}, AccountID: "acc", CloudProvider: "aws"},
			expectedErrorMsg: "Missing external ID for account",
		},
		{
			name:             "non-string externalID",
			account:          &models.Account{AdditionalData: map[string]any{"roleARN": "arn:aws", "externalID": 42}, AccountID: "acc", CloudProvider: "aws"},
			expectedErrorMsg: "Erroneous external ID for account",
		},
		{
			name: "valid account with products",
			account: &models.Account{
				AccountID:     "acc",
				CloudProvider: "aws",
				AdditionalData: map[string]any{
					"roleARN":    "arn:aws:iam::123456789012:role/example",
					"externalID": "external-id",
					"values": map[string]any{
						"someKey":       "someVal",
						"anotherKey":    []any{"Hello", "World", 123},
						"yetAnotherOne": map[string]any{"Number": 1, "String": "It's a String", "String Slice": []string{"Surprise!"}},
					},
				},
				Products: map[models.Product]models.ProductDetails{
					models.Kompass: {
						Active: true,
					},
				},
			},
		},
		{
			name: "valid account with products but no values",
			account: &models.Account{
				AccountID:     "acc",
				CloudProvider: "aws",
				AdditionalData: map[string]any{
					"roleARN":    "arn:aws:iam::123456789012:role/example",
					"externalID": "external-id",
					"values":     map[string]any{},
				},
				Products: map[models.Product]models.ProductDetails{
					models.CM: {
						Active: true,
					},
				},
			},
		},
		{
			name: "no products, valid account",
			account: &models.Account{
				AccountID:     "acc",
				CloudProvider: "aws",
				AdditionalData: map[string]any{
					"roleARN":    "arn:aws:iam::123456789012:role/example",
					"externalID": "external-id",
				},
				Products: map[models.Product]models.ProductDetails{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model, diags := provider.ToModel(tt.account)
			if tt.expectedErrorMsg != "" {
				require.True(t, diags.HasError())
				require.Len(t, diags, 1)
				assert.Contains(t, diags[0].Summary(), tt.expectedErrorMsg)
				assert.Nil(t, model)
			} else {
				require.False(t, diags.HasError())
				require.NotNil(t, model)
				assert.Equal(t, types.StringValue(tt.account.AccountID), model.ID)
				assert.Equal(t, types.StringValue(string(tt.account.CloudProvider)), model.CloudProvider)
				assert.Equal(t, types.StringValue(tt.account.AdditionalData["roleARN"].(string)), model.RoleARN)
				assert.Equal(t, types.StringValue(tt.account.AdditionalData["externalID"].(string)), model.ExternalID)
				for product := range tt.account.Products {
					switch product {
					case models.CM:
						assert.NotNil(t, model.Products.CM)
					case models.Kompass:
						assert.NotNil(t, model.Products.Kompass)
					case models.ZestyDisk:
						assert.NotNil(t, model.Products.ZestyDisk)
					}
				}
			}
		})
	}
}

func TestToModelAssignsValuesOnlyToKompass(t *testing.T) {
	account := &models.Account{
		AccountID:     "acc",
		CloudProvider: "aws",
		AdditionalData: map[string]any{
			"roleARN":    "arn:aws:iam::123456789012:role/example",
			"externalID": "external-id",
			"values":     map[string]any{"kompass": "value"},
		},
		Products: map[models.Product]models.ProductDetails{
			models.CM:        {Active: true},
			models.Kompass:   {Active: true},
			models.ZestyDisk: {Active: true},
		},
	}

	model, diags := provider.ToModel(account)
	require.False(t, diags.HasError())
	require.NotNil(t, model.Products.CM)
	require.NotNil(t, model.Products.Kompass)
	require.NotNil(t, model.Products.ZestyDisk)

	assert.True(t, model.Products.CM.Values.IsNull())
	assert.Equal(t, types.StringValue("kompass: value\n"), model.Products.Kompass.Values)
	assert.True(t, model.Products.ZestyDisk.Values.IsNull())
}
