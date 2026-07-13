package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

func preserveImportedExternalID() planmodifier.String {
	return preserveImportedExternalIDModifier{}
}

type preserveImportedExternalIDModifier struct{}

func (preserveImportedExternalIDModifier) Description(context.Context) string {
	return "prevents an existing External ID from being generated or changed"
}

func (m preserveImportedExternalIDModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func importedExternalIDError(externalID string) string {
	return fmt.Sprintf("External ID must be %q for this account.", externalID)
}

func (preserveImportedExternalIDModifier) PlanModifyString(
	_ context.Context,
	req planmodifier.StringRequest,
	resp *planmodifier.StringResponse,
) {
	if req.StateValue.IsNull() || req.StateValue.IsUnknown() || req.PlanValue.IsNull() {
		return
	}

	if req.PlanValue.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid External ID",
			importedExternalIDError(req.StateValue.ValueString()),
		)
		return
	}

	if !req.PlanValue.Equal(req.StateValue) {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid External ID",
			importedExternalIDError(req.StateValue.ValueString()),
		)
	}
}
