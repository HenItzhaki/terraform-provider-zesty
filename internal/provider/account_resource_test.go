package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zesty-co/terraform-provider-zesty/internal/client"
	"github.com/zesty-co/terraform-provider-zesty/internal/models"
)

func TestAccountResourceImportStateReadsExistingAccount(t *testing.T) {
	t.Parallel()

	const (
		accountID  = "123456789012"
		externalID = "3e973fb8-0b50-4668-95a2-8b933692eb5f"
		roleARN    = "arn:aws:iam::123456789012:role/zesty"
		token      = "test-token"
	)
	region := "us-east-1"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		assert.Equal(t, http.MethodGet, req.Method)
		assert.Equal(t, "/account", req.URL.Path)
		assert.Equal(t, accountID, req.URL.Query().Get("accountID"))
		assert.Equal(t, token, req.Header.Get("x-api-key"))

		w.Header().Set("Content-Type", "application/json")
		assert.NoError(t, json.NewEncoder(w).Encode(models.Account{
			AccountID:     accountID,
			Region:        &region,
			CloudProvider: models.AWS,
			Products: map[models.Product]models.ProductDetails{
				models.Kompass: {Active: true},
			},
			AdditionalData: map[string]any{
				"roleARN":    roleARN,
				"externalID": externalID,
			},
		}))
	}))
	t.Cleanup(server.Close)

	host := server.URL
	apiClient, err := client.NewClient(&host, token)
	require.NoError(t, err)
	apiClient.HTTPClient = server.Client()

	accountResource := &AccountResource{client: apiClient}
	ctx := context.Background()
	var schemaResponse resource.SchemaResponse
	accountResource.Schema(ctx, resource.SchemaRequest{}, &schemaResponse)

	importResponse := resource.ImportStateResponse{}
	importResponse.State.Schema = schemaResponse.Schema
	importResponse.Diagnostics.Append(importResponse.State.Set(ctx, accountResourceModel{
		ID:          types.StringNull(),
		LastUpdated: types.StringNull(),
		Account: accountModel{
			ID:            types.StringNull(),
			CloudProvider: types.StringNull(),
			Region:        types.StringNull(),
			RoleARN:       types.StringNull(),
			ExternalID:    types.StringNull(),
			Products:      []productModel{},
		},
	})...)
	require.False(t, importResponse.Diagnostics.HasError())

	accountResource.ImportState(ctx, resource.ImportStateRequest{ID: accountID}, &importResponse)
	require.False(t, importResponse.Diagnostics.HasError())

	var state accountResourceModel
	require.False(t, importResponse.State.Get(ctx, &state).HasError())
	assert.Equal(t, accountID, state.ID.ValueString())
	assert.Equal(t, accountID, state.Account.ID.ValueString())
	assert.Equal(t, externalID, state.Account.ExternalID.ValueString())
	assert.Equal(t, roleARN, state.Account.RoleARN.ValueString())
	require.Len(t, state.Account.Products, 1)
	assert.Equal(t, string(models.Kompass), state.Account.Products[0].Name.ValueString())
	assert.True(t, state.Account.Products[0].Active.ValueBool())
}
