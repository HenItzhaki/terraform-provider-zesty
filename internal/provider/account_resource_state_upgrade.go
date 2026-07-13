package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/zesty-co/terraform-provider-zesty/internal/models"
)

type accountResourceModelV0 struct {
	ID          types.String   `tfsdk:"id"`
	Account     accountModelV0 `tfsdk:"account"`
	LastUpdated types.String   `tfsdk:"last_updated"`
}

type accountModelV0 struct {
	ID            types.String     `tfsdk:"id"`
	CloudProvider types.String     `tfsdk:"cloud_provider"`
	Region        types.String     `tfsdk:"region"`
	RoleARN       types.String     `tfsdk:"role_arn"`
	ExternalID    types.String     `tfsdk:"external_id"`
	Products      []productModelV0 `tfsdk:"products"`
	Cur           *curModel        `tfsdk:"cur"`
	Athena        *athenaModel     `tfsdk:"athena"`
}

type productModelV0 struct {
	Name   types.String `tfsdk:"name"`
	Active types.Bool   `tfsdk:"active"`
	Values types.String `tfsdk:"values"`
}

func (r *AccountResource) UpgradeState(_ context.Context) map[int64]resource.StateUpgrader {
	priorSchema := accountResourceSchemaV0()

	return map[int64]resource.StateUpgrader{
		0: {
			PriorSchema: &priorSchema,
			StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
				var prior accountResourceModelV0
				resp.Diagnostics.Append(req.State.Get(ctx, &prior)...)
				if resp.Diagnostics.HasError() {
					return
				}

				var products productsModel
				for _, product := range prior.Account.Products {
					if !setProductModel(&products, models.Product(product.Name.ValueString()), productModel{
						Active: product.Active,
						Values: product.Values,
					}) {
						resp.Diagnostics.AddError(
							"Unable to upgrade product state",
							fmt.Sprintf("Product %q is not supported by the Terraform schema.", product.Name.ValueString()),
						)
						return
					}
				}

				upgraded := accountResourceModel{
					ID: prior.ID,
					Account: accountModel{
						ID:            prior.Account.ID,
						CloudProvider: prior.Account.CloudProvider,
						Region:        prior.Account.Region,
						RoleARN:       prior.Account.RoleARN,
						ExternalID:    prior.Account.ExternalID,
						Products:      products,
						Cur:           prior.Account.Cur,
						Athena:        prior.Account.Athena,
					},
					LastUpdated: prior.LastUpdated,
				}

				resp.Diagnostics.Append(resp.State.Set(ctx, &upgraded)...)
			},
		},
	}
}
