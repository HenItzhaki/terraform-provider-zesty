package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *AccountResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.State.Raw.IsNull() || req.Plan.Raw.IsNull() {
		return
	}

	var state accountResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var plan accountResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !kompassValuesCanUseState(state, plan) {
		return
	}

	resp.Diagnostics.Append(resp.Plan.SetAttribute(
		ctx,
		path.Root("account").AtName("products").AtName("kompass").AtName("values"),
		state.Account.Products.Kompass.Values,
	)...)
}

func kompassValuesCanUseState(state, plan accountResourceModel) bool {
	stateKompass := state.Account.Products.Kompass
	planKompass := plan.Account.Products.Kompass
	if stateKompass == nil || planKompass == nil {
		return false
	}
	if stateKompass.Values.IsNull() || stateKompass.Values.IsUnknown() || !planKompass.Values.IsUnknown() {
		return false
	}
	if stateKompass.Active.IsNull() || stateKompass.Active.IsUnknown() || !stateKompass.Active.ValueBool() {
		return false
	}

	return kompassInputsUnchanged(state, plan)
}

func kompassValuesWerePreserved(state, plan accountResourceModel) bool {
	stateKompass := state.Account.Products.Kompass
	planKompass := plan.Account.Products.Kompass
	if stateKompass == nil || planKompass == nil || planKompass.Values.IsNull() || planKompass.Values.IsUnknown() {
		return false
	}

	return kompassInputsUnchanged(state, plan) && stateKompass.Values.Equal(planKompass.Values)
}

func kompassValuesChangedUnexpectedly(state, plan accountResourceModel, result accountModel) bool {
	return kompassValuesWerePreserved(state, plan) &&
		!sameProductValues(plan.Account.Products.Kompass, result.Products.Kompass)
}

func kompassInputsUnchanged(state, plan accountResourceModel) bool {
	stateKompass := state.Account.Products.Kompass
	planKompass := plan.Account.Products.Kompass
	if stateKompass == nil || planKompass == nil {
		return false
	}

	return state.ID.Equal(plan.ID) &&
		state.Account.ID.Equal(plan.Account.ID) &&
		state.Account.CloudProvider.Equal(plan.Account.CloudProvider) &&
		state.Account.Region.Equal(plan.Account.Region) &&
		state.Account.RoleARN.Equal(plan.Account.RoleARN) &&
		state.Account.ExternalID.Equal(plan.Account.ExternalID) &&
		stateKompass.Active.Equal(planKompass.Active) &&
		sameCur(state.Account.Cur, plan.Account.Cur) &&
		sameAthena(state.Account.Athena, plan.Account.Athena)
}

func sameCur(left, right *curModel) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}

	return left.S3Bucket.Equal(right.S3Bucket) &&
		left.ExportName.Equal(right.ExportName) &&
		left.Type.Equal(right.Type)
}

func sameAthena(left, right *athenaModel) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}

	return left.AthenaDB.Equal(right.AthenaDB) &&
		left.AthenaS3Bucket.Equal(right.AthenaS3Bucket) &&
		left.AthenaProjectID.Equal(right.AthenaProjectID) &&
		left.AthenaRegion.Equal(right.AthenaRegion) &&
		left.AthenaTable.Equal(right.AthenaTable) &&
		left.AthenaWorkgroup.Equal(right.AthenaWorkgroup) &&
		left.AthenaCatalog.Equal(right.AthenaCatalog)
}

func sameProductValues(left, right *productModel) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}

	return left.Values.Equal(right.Values)
}
