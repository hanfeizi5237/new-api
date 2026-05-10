package controller

import (
	"testing"

	"github.com/gin-gonic/gin"
)

func TestPayMarketOrderRejectsUnsupportedPaymentMethod(t *testing.T) {
	ctx, recorder := newMarketplaceContext(t, "POST", "/api/market/orders/1/pay", PayMarketOrderRequest{
		PaymentMethod: "paypal",
	}, 1)
	ctx.Params = gin.Params{{Key: "id", Value: "1"}}

	PayMarketOrder(ctx)

	response := decodeMarketplaceResponse(t, recorder)
	if response.Success {
		t.Fatalf("expected unsupported payment method to fail")
	}
	if response.Message != "invalid payment_method" {
		t.Fatalf("expected invalid payment_method error, got %q", response.Message)
	}
}

func TestCreateMarketOrderRejectsNonPositiveValues(t *testing.T) {
	ctx, recorder := newMarketplaceContext(t, "POST", "/api/market/orders", CreateMarketOrderRequest{
		ListingId:      0,
		SkuId:          1,
		Quantity:       1,
		IdempotencyKey: "invalid-order",
	}, 1)

	CreateMarketOrder(ctx)

	response := decodeMarketplaceResponse(t, recorder)
	if response.Success {
		t.Fatalf("expected invalid market order request to fail")
	}
	if response.Message != "listing_id, sku_id, and quantity must be positive" {
		t.Fatalf("expected positive value validation error, got %q", response.Message)
	}
}
