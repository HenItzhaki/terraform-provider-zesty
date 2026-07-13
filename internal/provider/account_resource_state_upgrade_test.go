package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccountResourceUpgradeStateConvertsProductListToMap(t *testing.T) {
	ctx := context.Background()
	priorSchema := accountResourceSchemaV0()
	priorState := tfsdk.State{Schema: priorSchema}
	prior := accountResourceModelV0{
		ID: types.StringValue("123456789012"),
		Account: accountModelV0{
			ID:            types.StringValue("123456789012"),
			CloudProvider: types.StringValue("AWS"),
			Region:        types.StringValue("us-east-1"),
			RoleARN:       types.StringValue("arn:aws:iam::123456789012:role/ZestyIamRole"),
			ExternalID:    types.StringValue("external-id"),
			Products: []productModelV0{
				{Name: types.StringValue("CM"), Active: types.BoolValue(true), Values: types.StringValue("cm-values")},
				{Name: types.StringValue("Kompass"), Active: types.BoolValue(true), Values: types.StringValue("kompass-values")},
				{Name: types.StringValue("ZestyDisk"), Active: types.BoolValue(false), Values: types.StringValue("disk-values")},
			},
		},
		LastUpdated: types.StringValue("timestamp"),
	}
	require.False(t, priorState.Set(ctx, &prior).HasError())

	currentSchema := accountResourceSchemaV1()
	resp := resource.UpgradeStateResponse{State: tfsdk.State{Schema: currentSchema}}
	upgrader := (&AccountResource{}).UpgradeState(ctx)[0]
	upgrader.StateUpgrader(ctx, resource.UpgradeStateRequest{State: &priorState}, &resp)
	require.False(t, resp.Diagnostics.HasError())

	var upgraded accountResourceModel
	require.False(t, resp.State.Get(ctx, &upgraded).HasError())
	require.NotNil(t, upgraded.Account.Products.CM)
	require.NotNil(t, upgraded.Account.Products.Kompass)
	require.NotNil(t, upgraded.Account.Products.ZestyDisk)
	assert.Equal(t, types.StringValue("cm-values"), upgraded.Account.Products.CM.Values)
	assert.Equal(t, types.StringValue("kompass-values"), upgraded.Account.Products.Kompass.Values)
	assert.Equal(t, types.BoolValue(false), upgraded.Account.Products.ZestyDisk.Active)
	assert.Equal(t, prior.LastUpdated, upgraded.LastUpdated)
}
