package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestKompassValuesCanUseState(t *testing.T) {
	tests := []struct {
		name     string
		mutate   func(state, plan *accountResourceModel)
		expected bool
	}{
		{
			name: "cm only change",
			mutate: func(_, plan *accountResourceModel) {
				plan.Account.Products.CM = &productModel{Active: types.BoolValue(true), Values: types.StringUnknown()}
			},
			expected: true,
		},
		{
			name: "zesty disk only change",
			mutate: func(_, plan *accountResourceModel) {
				plan.Account.Products.ZestyDisk = &productModel{Active: types.BoolValue(true), Values: types.StringUnknown()}
			},
			expected: true,
		},
		{
			name: "kompass active change",
			mutate: func(_, plan *accountResourceModel) {
				plan.Account.Products.Kompass.Active = types.BoolValue(false)
			},
			expected: false,
		},
		{
			name: "role ARN change",
			mutate: func(_, plan *accountResourceModel) {
				plan.Account.RoleARN = types.StringValue("arn:aws:iam::123456789012:role/new")
			},
			expected: false,
		},
		{
			name: "external ID change",
			mutate: func(_, plan *accountResourceModel) {
				plan.Account.ExternalID = types.StringValue("new-external-id")
			},
			expected: false,
		},
		{
			name: "region change",
			mutate: func(_, plan *accountResourceModel) {
				plan.Account.Region = types.StringValue("eu-west-1")
			},
			expected: false,
		},
		{
			name: "CUR change",
			mutate: func(_, plan *accountResourceModel) {
				plan.Account.Cur = testCurModel()
			},
			expected: false,
		},
		{
			name: "Athena change",
			mutate: func(_, plan *accountResourceModel) {
				plan.Account.Athena = testAthenaModel()
			},
			expected: false,
		},
		{
			name: "state values unknown",
			mutate: func(state, _ *accountResourceModel) {
				state.Account.Products.Kompass.Values = types.StringUnknown()
			},
			expected: false,
		},
		{
			name: "plan values already known",
			mutate: func(_, plan *accountResourceModel) {
				plan.Account.Products.Kompass.Values = types.StringValue("planned-values")
			},
			expected: false,
		},
		{
			name: "kompass missing from plan",
			mutate: func(_, plan *accountResourceModel) {
				plan.Account.Products.Kompass = nil
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := testAccountResourceModel(types.StringValue("kompass-values"))
			plan := testAccountResourceModel(types.StringUnknown())
			tt.mutate(&state, &plan)

			assert.Equal(t, tt.expected, kompassValuesCanUseState(state, plan))
		})
	}
}

func TestKompassValuesWerePreserved(t *testing.T) {
	state := testAccountResourceModel(types.StringValue("kompass-values"))
	plan := testAccountResourceModel(types.StringValue("kompass-values"))
	plan.Account.Products.CM = &productModel{Active: types.BoolValue(true), Values: types.StringUnknown()}

	assert.True(t, kompassValuesWerePreserved(state, plan))

	plan.Account.Products.Kompass.Values = types.StringValue("different-values")
	assert.False(t, kompassValuesWerePreserved(state, plan))
}

func TestKompassValuesChangedUnexpectedly(t *testing.T) {
	state := testAccountResourceModel(types.StringValue("kompass-values"))
	plan := testAccountResourceModel(types.StringValue("kompass-values"))
	plan.Account.Products.CM = &productModel{Active: types.BoolValue(true), Values: types.StringUnknown()}
	result := testAccountResourceModel(types.StringValue("rotated-values")).Account

	assert.True(t, kompassValuesChangedUnexpectedly(state, plan, result))

	result.Products.Kompass.Values = types.StringValue("kompass-values")
	assert.False(t, kompassValuesChangedUnexpectedly(state, plan, result))
}

func testAccountResourceModel(kompassValues types.String) accountResourceModel {
	return accountResourceModel{
		ID: types.StringValue("123456789012"),
		Account: accountModel{
			ID:            types.StringValue("123456789012"),
			CloudProvider: types.StringValue("AWS"),
			Region:        types.StringValue("us-east-1"),
			RoleARN:       types.StringValue("arn:aws:iam::123456789012:role/example"),
			ExternalID:    types.StringValue("external-id"),
			Products: productsModel{
				Kompass: &productModel{
					Active: types.BoolValue(true),
					Values: kompassValues,
				},
			},
		},
		LastUpdated: types.StringValue("previous-update"),
	}
}

func testCurModel() *curModel {
	return &curModel{
		S3Bucket:   types.StringValue("bucket"),
		ExportName: types.StringValue("export"),
		Type:       types.StringValue("cur_v2"),
	}
}

func testAthenaModel() *athenaModel {
	return &athenaModel{
		AthenaDB:        types.StringValue("database"),
		AthenaS3Bucket:  types.StringValue("bucket"),
		AthenaProjectID: types.StringValue("project"),
		AthenaRegion:    types.StringValue("us-east-1"),
		AthenaTable:     types.StringValue("table"),
		AthenaWorkgroup: types.StringValue("workgroup"),
		AthenaCatalog:   types.StringValue("catalog"),
	}
}
