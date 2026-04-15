package controller

import (
	"net/http"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
)

type marketOrderAdminPageResponse struct {
	Items []MarketOrderView `json:"items"`
	Total int               `json:"total"`
}

type marketEntitlementAdminPageResponse struct {
	Items []model.BuyerEntitlement `json:"items"`
	Total int                      `json:"total"`
}

func TestMarketOrderAdminList(t *testing.T) {
	db := setupMarketplaceControllerTestDB(t)
	buyer := seedMarketplaceUser(t, db, "marketplace-admin-order-buyer")
	seller, supply := seedMarketplaceSellerWithSupply(t, db, buyer.Id)
	listing := &model.Listing{
		SellerId:        seller.Id,
		SupplyAccountId: supply.Id,
		ListingCode:     "listing-admin-order-001",
		Title:           "Admin Order Listing",
		VendorId:        supply.VendorId,
		ModelName:       supply.ModelName,
		SaleMode:        "fixed_price",
		PricingUnit:     "per_token_package",
		AuditStatus:     "approved",
		Status:          "active",
	}
	if err := db.Create(listing).Error; err != nil {
		t.Fatalf("failed to create listing: %v", err)
	}
	sku := &model.ListingSKU{
		ListingId:      listing.Id,
		SkuCode:        "sku-admin-order-001",
		PackageAmount:  1000,
		PackageUnit:    "token",
		UnitPriceMinor: 299,
		MinQuantity:    1,
		MaxQuantity:    5,
		Status:         "active",
	}
	if err := db.Create(sku).Error; err != nil {
		t.Fatalf("failed to create sku: %v", err)
	}
	order := &model.MarketOrder{
		OrderNo:            "MKT-ADMIN-ORDER-001",
		BuyerUserId:        buyer.Id,
		Currency:           "CNY",
		TotalAmountMinor:   299,
		PayableAmountMinor: 299,
		OrderStatus:        "paid",
		PaymentStatus:      "paid",
		EntitlementStatus:  "created",
		PaymentMethod:      "stripe",
		IdempotencyKey:     "idem-admin-order-001",
		PaidAt:             common.GetTimestamp(),
	}
	if err := db.Create(order).Error; err != nil {
		t.Fatalf("failed to create order: %v", err)
	}
	if err := db.Create(&model.MarketOrderItem{
		OrderId:         order.Id,
		ListingId:       listing.Id,
		SkuId:           sku.Id,
		SellerId:        seller.Id,
		SupplyAccountId: supply.Id,
		VendorId:        supply.VendorId,
		ModelName:       supply.ModelName,
		Quantity:        1,
		PackageAmount:   1000,
		PackageUnit:     "token",
		GrantedAmount:   1000,
		UnitPriceMinor:  299,
		LineAmountMinor: 299,
		Status:          "granted",
	}).Error; err != nil {
		t.Fatalf("failed to create order item: %v", err)
	}

	listCtx, listRecorder := newMarketplaceContext(t, http.MethodGet, "/api/marketplace/admin/orders?keyword=MKT-ADMIN&p=1&size=10", nil, buyer.Id)
	GetMarketOrdersAdmin(listCtx)
	listResp := decodeMarketplaceResponse(t, listRecorder)
	if !listResp.Success {
		t.Fatalf("expected order admin list success, got message: %s", listResp.Message)
	}

	var page marketOrderAdminPageResponse
	if err := common.Unmarshal(listResp.Data, &page); err != nil {
		t.Fatalf("failed to decode order admin page: %v", err)
	}
	if page.Total != 1 || len(page.Items) != 1 {
		t.Fatalf("expected one order in admin page, got total=%d items=%d", page.Total, len(page.Items))
	}
	if page.Items[0].Order.OrderNo != order.OrderNo || len(page.Items[0].Items) != 1 {
		t.Fatalf("expected order details in admin page, got %+v", page.Items[0])
	}
}

func TestMarketEntitlementAdminList(t *testing.T) {
	db := setupMarketplaceControllerTestDB(t)
	buyer := seedMarketplaceUser(t, db, "marketplace-admin-entitlement-buyer")
	seller, supply := seedMarketplaceSellerWithSupply(t, db, buyer.Id)
	entitlement := &model.BuyerEntitlement{
		BuyerUserId:  buyer.Id,
		VendorId:     supply.VendorId,
		ModelName:    supply.ModelName,
		TotalGranted: 5000,
		TotalUsed:    1200,
		TotalFrozen:  300,
		Status:       "active",
	}
	if err := db.Create(entitlement).Error; err != nil {
		t.Fatalf("failed to create entitlement: %v", err)
	}
	if err := db.Create(&model.EntitlementLot{
		BuyerEntitlementId: entitlement.Id,
		BuyerUserId:        buyer.Id,
		OrderId:            1,
		OrderItemId:        1,
		SellerId:           seller.Id,
		ListingId:          1,
		SupplyAccountId:    supply.Id,
		GrantedAmount:      5000,
		SourceEventKey:     "grant-admin-entitlement-001",
		Status:             "active",
	}).Error; err != nil {
		t.Fatalf("failed to create entitlement lot: %v", err)
	}

	listCtx, listRecorder := newMarketplaceContext(t, http.MethodGet, "/api/marketplace/admin/entitlements?buyer_user_id=1&model_name=gpt-4o-mini&p=1&size=10", nil, buyer.Id)
	GetMarketEntitlementsAdmin(listCtx)
	listResp := decodeMarketplaceResponse(t, listRecorder)
	if !listResp.Success {
		t.Fatalf("expected entitlement admin list success, got message: %s", listResp.Message)
	}

	var page marketEntitlementAdminPageResponse
	if err := common.Unmarshal(listResp.Data, &page); err != nil {
		t.Fatalf("failed to decode entitlement admin page: %v", err)
	}
	if page.Total != 1 || len(page.Items) != 1 {
		t.Fatalf("expected one entitlement in admin page, got total=%d items=%d", page.Total, len(page.Items))
	}
	if page.Items[0].BuyerUserId != buyer.Id || page.Items[0].ModelName != supply.ModelName {
		t.Fatalf("expected entitlement filters to match seeded entitlement, got %+v", page.Items[0])
	}
}
