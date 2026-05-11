package controller

import (
	"net/http"
	"strconv"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-gonic/gin"
)

type marketOrderBuyerPageResponse struct {
	Items []MarketOrderView `json:"items"`
	Total int               `json:"total"`
}

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

func TestGetMarketOrdersOnlyReturnsBuyerOwnedOrders(t *testing.T) {
	db := setupMarketplaceControllerTestDB(t)
	buyerAOrder, _ := seedMarketplacePayableOrder(t, db, "market-order-buyer-a", "market-order-seller-a")
	buyerBOrder, _ := seedMarketplacePayableOrder(t, db, "market-order-buyer-b", "market-order-seller-b")

	var buyerA model.User
	if err := db.First(&buyerA, buyerAOrder.BuyerUserId).Error; err != nil {
		t.Fatalf("failed to load buyer A: %v", err)
	}

	ctx, recorder := newMarketplaceContext(t, http.MethodGet, "/api/market/orders?p=1&size=10", nil, buyerA.Id)
	GetMarketOrders(ctx)

	response := decodeMarketplaceResponse(t, recorder)
	if !response.Success {
		t.Fatalf("expected buyer order list success, got message: %s", response.Message)
	}

	var page marketOrderBuyerPageResponse
	if err := common.Unmarshal(response.Data, &page); err != nil {
		t.Fatalf("failed to decode buyer order page: %v", err)
	}
	if page.Total != 1 || len(page.Items) != 1 {
		t.Fatalf("expected exactly one buyer-owned order, got total=%d len=%d", page.Total, len(page.Items))
	}
	if page.Items[0].Order.Id != buyerAOrder.Id || page.Items[0].Order.BuyerUserId != buyerA.Id {
		t.Fatalf("expected buyer A order only, got %+v", page.Items[0].Order)
	}
	if page.Items[0].Order.Id == buyerBOrder.Id {
		t.Fatalf("buyer A order list leaked buyer B order id=%d", buyerBOrder.Id)
	}
	if len(page.Items[0].Items) != 1 {
		t.Fatalf("expected buyer order items to be included, got %+v", page.Items[0].Items)
	}
}

func TestGetMarketOrderRejectsForeignBuyerAccess(t *testing.T) {
	db := setupMarketplaceControllerTestDB(t)
	buyerAOrder, _ := seedMarketplacePayableOrder(t, db, "market-order-detail-buyer-a", "market-order-detail-seller-a")
	buyerBOrder, _ := seedMarketplacePayableOrder(t, db, "market-order-detail-buyer-b", "market-order-detail-seller-b")

	var buyerA model.User
	if err := db.First(&buyerA, buyerAOrder.BuyerUserId).Error; err != nil {
		t.Fatalf("failed to load buyer A: %v", err)
	}

	ctx, recorder := newMarketplaceContext(t, http.MethodGet, "/api/market/orders/"+strconv.Itoa(buyerBOrder.Id), nil, buyerA.Id)
	ctx.Params = gin.Params{{Key: "id", Value: strconv.Itoa(buyerBOrder.Id)}}

	GetMarketOrder(ctx)

	response := decodeMarketplaceResponse(t, recorder)
	if response.Success {
		t.Fatalf("expected foreign buyer order detail access to fail")
	}
	if response.Message != "order does not belong to buyer" {
		t.Fatalf("expected ownership rejection, got %q", response.Message)
	}
}
