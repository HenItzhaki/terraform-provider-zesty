package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/zesty-co/terraform-provider-zesty/internal/client"
)

type AccountsDataSource struct {
	client *client.Client
}

var (
	_ datasource.DataSource              = &AccountsDataSource{}
	_ datasource.DataSourceWithConfigure = &AccountsDataSource{}
)

func NewAccountsDataSource() datasource.DataSource {
	return &AccountsDataSource{}
}

func (d *AccountsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_accounts"
}

type accountsDataSourceModel struct {
	Accounts []accountModel `tfsdk:"accounts"`
}

type accountModel struct {
	ID            types.String  `tfsdk:"id"`
	CloudProvider types.String  `tfsdk:"cloud_provider"`
	Region        types.String  `tfsdk:"region"`
	RoleARN       types.String  `tfsdk:"role_arn"`
	ExternalID    types.String  `tfsdk:"external_id"`
	Products      productsModel `tfsdk:"products"`
	Cur           *curModel     `tfsdk:"cur"`
	Athena        *athenaModel  `tfsdk:"athena"`
}

type productModel struct {
	Active types.Bool   `tfsdk:"active"`
	Values types.String `tfsdk:"values"`
}

type productsModel struct {
	CM        *productModel `tfsdk:"cm"`
	Kompass   *productModel `tfsdk:"kompass"`
	ZestyDisk *productModel `tfsdk:"zesty_disk"`
}

type curModel struct {
	S3Bucket   types.String `tfsdk:"s3_bucket"`
	ExportName types.String `tfsdk:"cur_export_name"`
	Type       types.String `tfsdk:"cur_type"`
}

type athenaModel struct {
	AthenaDB        types.String `tfsdk:"athena_db"`
	AthenaS3Bucket  types.String `tfsdk:"athena_s3_bucket"`
	AthenaProjectID types.String `tfsdk:"athena_project_id"`
	AthenaRegion    types.String `tfsdk:"athena_region"`
	AthenaTable     types.String `tfsdk:"athena_table"`
	AthenaWorkgroup types.String `tfsdk:"athena_workgroup"`
	AthenaCatalog   types.String `tfsdk:"athena_catalog"`
}

// Schema defines the schema for the data source.
func (d *AccountsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches the list of accounts.",
		Attributes: map[string]schema.Attribute{
			"accounts": schema.ListNestedAttribute{
				Description: "List of accounts.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Account ID",
							Computed:    true,
						},
						"cloud_provider": schema.StringAttribute{
							Description: "Name of cloud provider (e.g. AWS, GCP, Azure)",
							Computed:    true,
						},
						"role_arn": schema.StringAttribute{
							Description: "Role ARN generated on the cloud provider",
							Computed:    true,
						},
						"external_id": schema.StringAttribute{
							Description: "External ID (UUID)",
							Computed:    true,
						},
						"region": schema.StringAttribute{
							Optional: true,
							Computed: false,
						},
						"products": schema.SingleNestedAttribute{
							Description: "Products activated on the account.",
							Computed:    true,
							Attributes: map[string]schema.Attribute{
								"cm":         accountsDataSourceProductSchema(),
								"kompass":    accountsDataSourceProductSchema(),
								"zesty_disk": accountsDataSourceProductSchema(),
							},
						},
						"cur": schema.SingleNestedAttribute{
							Description: "Cur export data for the account",
							Optional:    true,
							Attributes: map[string]schema.Attribute{
								"s3_bucket": schema.StringAttribute{
									Description: "S3 bucket name for the cur export",
									Required:    true,
								},
								"cur_export_name": schema.StringAttribute{
									Description: "The cur export file name",
									Required:    true,
								},
								"cur_type": schema.StringAttribute{
									Optional: true,
									Computed: true,
								},
							},
						},
						"athena": schema.SingleNestedAttribute{
							Description: "Athena resources data for the account",
							Optional:    true,
							Attributes: map[string]schema.Attribute{
								"athena_db": schema.StringAttribute{
									Description: "The athena db associated with the cur report",
									Required:    true,
								},
								"athena_s3_bucket": schema.StringAttribute{
									Description: "The s3 bucket for athena's results",
									Required:    true,
								},
								"athena_project_id": schema.StringAttribute{
									Description: "Athen's project id",
									Required:    true,
								},
								"athena_region": schema.StringAttribute{
									Description: "The athena instance's region",
									Required:    true,
								},
								"athena_table": schema.StringAttribute{
									Description: "The athena DB table.",
									Required:    true,
								},
								"athena_workgroup": schema.StringAttribute{
									Description: "The athena workgroup",
									Required:    true,
								},
								"athena_catalog": schema.StringAttribute{
									Description: "The athena catalog",
									Required:    true,
								},
							},
						},
					},
				},
			},
		},
	}
}

func accountsDataSourceProductSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Computed: true,
		Attributes: map[string]schema.Attribute{
			"active": schema.BoolAttribute{
				Description: "Status of product",
				Computed:    true,
			},
			"values": schema.StringAttribute{
				Description: "Key-value pairs of product-specific values",
				Computed:    true,
			},
		},
	}
}

func (d *AccountsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state accountsDataSourceModel

	accounts, err := d.client.GetAccounts()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Zesty Onboarded Accounts",
			err.Error(),
		)
		return
	}

	tflog.Info(ctx, "Received accounts", map[string]any{"count": len(*accounts)})

	for _, account := range *accounts {
		roleARN, exists := account.AdditionalData["roleARN"]
		if !exists {
			resp.Diagnostics.AddError(
				"Missing role ARN for account",
				account.AccountID,
			)
			return
		}
		roleARNString, ok := roleARN.(string)
		if !ok {
			resp.Diagnostics.AddError(
				"Erroneous role ARN for account",
				account.AccountID,
			)
			return
		}

		externalID, exists := account.AdditionalData["externalID"]
		if !exists {
			resp.Diagnostics.AddError(
				"Missing external ID for account",
				account.AccountID,
			)
			return
		}
		externalIDString, ok := externalID.(string)
		if !ok {
			resp.Diagnostics.AddError(
				"Erroneous external ID for account",
				account.AccountID,
			)
			return
		}
		accountState := accountModel{
			ID:            types.StringValue(account.AccountID),
			CloudProvider: types.StringValue(string(account.CloudProvider)),
			RoleARN:       types.StringValue(roleARNString),
			ExternalID:    types.StringValue(externalIDString),
		}

		for name, details := range account.Products {
			if !setProductModel(&accountState.Products, name, productModel{
				Active: types.BoolValue(details.Active),
				Values: types.StringNull(),
			}) {
				resp.Diagnostics.AddError(
					"Unsupported product from API",
					fmt.Sprintf("Product %q is not supported by the Terraform schema.", name),
				)
				return
			}
		}

		tflog.Info(ctx, "Adding account to state", map[string]any{"account": accountState})

		state.Accounts = append(state.Accounts, accountState)
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (d *AccountsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected: *client.Client, got: %T.\nPlease report this issue to Zesty Support.", req.ProviderData),
		)

		return
	}

	d.client = client
}
