package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPreserveImportedExternalID(t *testing.T) {
	t.Parallel()

	const existingExternalID = "3e973fb8-0b50-4668-95a2-8b933692eb5f"
	assert.NotContains(t, preserveImportedExternalID().Description(context.Background()), "account UUID")

	tests := map[string]struct {
		stateValue types.String
		planValue  types.String
		wantError  bool
		wantTitle  string
	}{
		"allows create": {
			stateValue: types.StringNull(),
			planValue:  types.StringValue(existingExternalID),
		},
		"allows unchanged existing UUID": {
			stateValue: types.StringValue(existingExternalID),
			planValue:  types.StringValue(existingExternalID),
		},
		"allows destroy": {
			stateValue: types.StringValue(existingExternalID),
			planValue:  types.StringNull(),
		},
		"rejects unknown UUID after import": {
			stateValue: types.StringValue(existingExternalID),
			planValue:  types.StringUnknown(),
			wantError:  true,
			wantTitle:  "Invalid External ID",
		},
		"rejects changed UUID": {
			stateValue: types.StringValue(existingExternalID),
			planValue:  types.StringValue("aabbccdd-eeff-0011-2233-445566778899"),
			wantError:  true,
			wantTitle:  "Invalid External ID",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			request := planmodifier.StringRequest{
				Path:       path.Root("account").AtName("external_id"),
				StateValue: test.stateValue,
				PlanValue:  test.planValue,
			}
			response := planmodifier.StringResponse{}

			preserveImportedExternalID().PlanModifyString(context.Background(), request, &response)

			if !test.wantError {
				require.False(t, response.Diagnostics.HasError())
				return
			}

			require.True(t, response.Diagnostics.HasError())
			require.Len(t, response.Diagnostics, 1)
			assert.Equal(t, test.wantTitle, response.Diagnostics[0].Summary())
			assert.Equal(t, importedExternalIDError(existingExternalID), response.Diagnostics[0].Detail())
			assert.Contains(t, response.Diagnostics[0].Detail(), existingExternalID)
			assert.NotContains(t, response.Diagnostics[0].Detail(), "account UUID")
		})
	}
}
