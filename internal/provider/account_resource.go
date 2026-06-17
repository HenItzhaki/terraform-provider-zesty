package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/zesty-co/terraform-provider-zesty/internal/client"
	"github.com/zesty-co/terraform-provider-zesty/internal/models"
)

type AccountResource struct {
	client *client.Client
}

var (
	_ resource.Resource                = &AccountResource{}
	_ resource.ResourceWithConfigure   = &AccountResource{}
	_ resource.ResourceWithImportState = &AccountResource{}
)

func NewAccountResource() resource.Resource {
	return &AccountResource{}
}

func (r *AccountResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_account"
}

type accountResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Account     accountModel `tfsdk:"account"`
	LastUpdated types.String `tfsdk:"last_updated"`
}

// Schema defines the schema for the resource.
func (r *AccountResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an account.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Account ID",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"last_updated": schema.StringAttribute{
				Description: "Timestamp of the last Terraform update of the account.",
				Computed:    true,
			},
			"account": schema.SingleNestedAttribute{
				Required: true,
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						Description: "Account ID",
						Required:    true,
					},
					"cloud_provider": schema.StringAttribute{
						Description: "Name of cloud provider (e.g. AWS, GCP, Azure)",
						Required:    true,
					},
					"role_arn": schema.StringAttribute{
						Description: "Role ARN generated on the cloud provider",
						Required:    true,
					},
					"external_id": schema.StringAttribute{
						Description: "External ID (UUID)",
						Required:    true,
					},
					"region": schema.StringAttribute{
						Description: "Region of the cloud provider",
						Optional:    true,
						Default:     stringdefault.StaticString("us-east-1"),
						Computed:    true,
					},
					"products": schema.ListNestedAttribute{
						Description: "List of products activated on the account",
						Required:    true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"name": schema.StringAttribute{
									Description: "Name of product (e.g. Kompass)",
									Required:    true,
								},
								"active": schema.BoolAttribute{
									Description: "Status of product",
									Required:    true,
								},
								"values": schema.StringAttribute{
									Description: "Key-value pairs of product-specific values",
									Computed:    true,
								},
							},
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
								Default:  stringdefault.StaticString("cur_v2"),
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
	}
}

func (r *AccountResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.client = client
}

func (r *AccountResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan accountResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := models.Payload{
		AccountID:     plan.Account.ID.ValueString(),
		Region:        plan.Account.Region.ValueStringPointer(),
		CloudProvider: models.CloudProvider(plan.Account.CloudProvider.ValueString()),
		RoleARN:       plan.Account.RoleARN.ValueString(),
		ExternalID:    plan.Account.ExternalID.ValueString(),
		Products:      map[models.Product]models.ProductDetails{},
	}
	for _, product := range plan.Account.Products {
		payload.Products[models.Product(product.Name.ValueString())] = models.ProductDetails{
			Active: product.Active.ValueBool(),
		}
	}

	if plan.Account.Cur != nil {
		payload.Cur = &models.CurDetails{
			S3Bucket:   plan.Account.Cur.S3Bucket.ValueString(),
			ExportName: plan.Account.Cur.ExportName.ValueString(),
			Type:       plan.Account.Cur.Type.ValueString(),
		}
	}

	if plan.Account.Athena != nil {
		payload.Athena = &models.AthenaDetails{
			AthenaDB:        plan.Account.Athena.AthenaDB.ValueString(),
			AthenaS3Bucket:  plan.Account.Athena.AthenaS3Bucket.ValueString(),
			AthenaProjectID: plan.Account.Athena.AthenaProjectID.ValueString(),
			AthenaRegion:    plan.Account.Athena.AthenaRegion.ValueString(),
			AthenaTable:     plan.Account.Athena.AthenaTable.ValueString(),
			AthenaWorkgroup: plan.Account.Athena.AthenaWorkgroup.ValueString(),
			AthenaCatalog:   plan.Account.Athena.AthenaCatalog.ValueString(),
		}
	}

	tflog.Info(ctx, "Sending create request", map[string]any{"payload": payload})
	account, err := r.client.CreateAccount(payload)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating account",
			"Could not create account, unexpected error: "+err.Error(),
		)
		return
	}

	plan.ID = types.StringValue(account.AccountID)
	model, diag := ToModel(account)
	resp.Diagnostics.Append(diag...)
	if diag != nil {
		return
	}

	plan.Account = *model
	tflog.Info(ctx, "Create result", map[string]any{"account": plan.Account})
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *AccountResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state accountResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Sending get request", map[string]any{"id": state.ID.ValueString()})
	account, err := r.client.GetAccount(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Zesty Account",
			"Could not read account ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	model, diag := ToModel(account)
	resp.Diagnostics.Append(diag...)
	if diag != nil {
		return
	}

	state.Account = *model
	tflog.Info(ctx, "Read result", map[string]any{"account": state.Account})

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *AccountResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan accountResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := models.Payload{
		AccountID:     plan.Account.ID.ValueString(),
		Region:        plan.Account.Region.ValueStringPointer(),
		CloudProvider: models.CloudProvider(plan.Account.CloudProvider.ValueString()),
		RoleARN:       plan.Account.RoleARN.ValueString(),
		ExternalID:    plan.Account.ExternalID.ValueString(),
		Products:      map[models.Product]models.ProductDetails{},
	}
	for _, product := range plan.Account.Products {
		payload.Products[models.Product(product.Name.ValueString())] = models.ProductDetails{
			Active: product.Active.ValueBool(),
		}
	}

	if plan.Account.Cur != nil {
		payload.Cur = &models.CurDetails{
			S3Bucket:   plan.Account.Cur.S3Bucket.ValueString(),
			ExportName: plan.Account.Cur.ExportName.ValueString(),
			Type:       plan.Account.Cur.Type.ValueString(),
		}
	}

	if plan.Account.Athena != nil {
		payload.Athena = &models.AthenaDetails{
			AthenaDB:        plan.Account.Athena.AthenaDB.ValueString(),
			AthenaS3Bucket:  plan.Account.Athena.AthenaS3Bucket.ValueString(),
			AthenaProjectID: plan.Account.Athena.AthenaProjectID.ValueString(),
			AthenaRegion:    plan.Account.Athena.AthenaRegion.ValueString(),
			AthenaTable:     plan.Account.Athena.AthenaTable.ValueString(),
			AthenaWorkgroup: plan.Account.Athena.AthenaWorkgroup.ValueString(),
			AthenaCatalog:   plan.Account.Athena.AthenaCatalog.ValueString(),
		}
	}

	tflog.Info(ctx, "Sending update request", map[string]any{"payload": payload})
	updatedAccount, err := r.client.UpdateAccount(payload)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Zesty Account",
			"Could not update account, unexpected error: "+err.Error(),
		)
		return
	}

	model, diag := ToModel(updatedAccount)
	resp.Diagnostics.Append(diag...)
	if diag != nil {
		return
	}

	plan.ID = types.StringValue(model.ID.ValueString())
	plan.Account = *model
	tflog.Info(ctx, "Update result", map[string]any{"account": plan.Account})
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *AccountResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state accountResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	payload := models.Payload{
		AccountID:     state.Account.ID.ValueString(),
		CloudProvider: models.CloudProvider(state.Account.CloudProvider.ValueString()),
		RoleARN:       state.Account.RoleARN.ValueString(),
		ExternalID:    state.Account.ExternalID.ValueString(),
	}

	err := r.client.DeleteAccount(payload)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting account",
			"Could not delete account, unexpected error: "+err.Error(),
		)
		return
	}
}

func (r *AccountResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id := req.ID
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)

	account, err := r.client.GetAccount(id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error importing resource",
			fmt.Sprintf("Could not read resource with ID %q: %s", id, err),
		)
		return
	}

	model, diag := ToModel(account)
	resp.Diagnostics.Append(diag...)
	if diag != nil {
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("account"), model)...)
}
