package controller

import (
	"net/http"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
)

type marketEntitlementBuyerPageResponse struct {
	Items []model.BuyerEntitlement `json:"items"`
	Total int                      `json:"total"`
}

func TestGetMarketEntitlementsOnlyReturnsBuyerOwnedEntitlements(t *testing.T) {
	db := setupMarketplaceControllerTestDB(t)
	buyerA := seedMarketplaceUser(t, db, "market-entitlement-buyer-a")
	buyerB := seedMarketplaceUser(t, db, "market-entitlement-buyer-b")
	sellerUser := seedMarketplaceUser(t, db, "market-entitlement-seller")
	_, supply := seedMarketplaceSellerWithSupply(t, db, sellerUser.Id)

	entitlementA := &model.BuyerEntitlement{
		BuyerUserId:  buyerA.Id,
		VendorId:     supply.VendorId,
		ModelName:    supply.ModelName,
		TotalGranted: 3200,
		Status:       "active",
	}
	if err := db.Create(entitlementA).Error; err != nil {
		t.Fatalf("failed to create buyer A entitlement: %v", err)
	}

	entitlementB := &model.BuyerEntitlement{
		BuyerUserId:  buyerB.Id,
		VendorId:     supply.VendorId,
		ModelName:    supply.ModelName,
		TotalGranted: 6400,
		Status:       "active",
	}
	if err := db.Create(entitlementB).Error; err != nil {
		t.Fatalf("failed to create buyer B entitlement: %v", err)
	}

	ctx, recorder := newMarketplaceContext(t, http.MethodGet, "/api/market/entitlements?model_name="+supply.ModelName+"&p=1&size=10", nil, buyerA.Id)
	GetMarketEntitlements(ctx)

	response := decodeMarketplaceResponse(t, recorder)
	if !response.Success {
		t.Fatalf("expected buyer entitlement list success, got message: %s", response.Message)
	}

	var page marketEntitlementBuyerPageResponse
	if err := common.Unmarshal(response.Data, &page); err != nil {
		t.Fatalf("failed to decode buyer entitlement page: %v", err)
	}
	if page.Total != 1 || len(page.Items) != 1 {
		t.Fatalf("expected exactly one buyer-owned entitlement, got total=%d len=%d", page.Total, len(page.Items))
	}
	if page.Items[0].Id != entitlementA.Id || page.Items[0].BuyerUserId != buyerA.Id {
		t.Fatalf("expected buyer A entitlement only, got %+v", page.Items[0])
	}
	if page.Items[0].Id == entitlementB.Id {
		t.Fatalf("buyer A entitlement list leaked buyer B entitlement id=%d", entitlementB.Id)
	}
}
