package provider

import "github.com/zesty-co/terraform-provider-zesty/internal/models"

func setProductModel(products *productsModel, name models.Product, product productModel) bool {
	switch name {
	case models.CM:
		products.CM = &product
	case models.Kompass:
		products.Kompass = &product
	case models.ZestyDisk:
		products.ZestyDisk = &product
	default:
		return false
	}

	return true
}

func productPayload(products productsModel) map[models.Product]models.ProductDetails {
	payload := make(map[models.Product]models.ProductDetails, 3)
	if products.CM != nil {
		payload[models.CM] = models.ProductDetails{Active: products.CM.Active.ValueBool()}
	}
	if products.Kompass != nil {
		payload[models.Kompass] = models.ProductDetails{Active: products.Kompass.Active.ValueBool()}
	}
	if products.ZestyDisk != nil {
		payload[models.ZestyDisk] = models.ProductDetails{Active: products.ZestyDisk.Active.ValueBool()}
	}

	return payload
}
