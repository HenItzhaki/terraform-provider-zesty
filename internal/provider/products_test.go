package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zesty-co/terraform-provider-zesty/internal/models"
)

func TestSetProductModelAndPayload(t *testing.T) {
	tests := []struct {
		name    string
		product models.Product
	}{
		{name: "cm", product: models.CM},
		{name: "kompass", product: models.Kompass},
		{name: "zesty_disk", product: models.ZestyDisk},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var products productsModel
			require.True(t, setProductModel(&products, tt.product, productModel{Active: types.BoolValue(true)}))
			assert.True(t, productPayload(products)[tt.product].Active)
		})
	}
}

func TestSetProductModelRejectsUnknownProduct(t *testing.T) {
	var products productsModel
	assert.False(t, setProductModel(&products, "unknown", productModel{}))
}
