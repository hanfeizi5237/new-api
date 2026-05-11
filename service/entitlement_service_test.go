package service

import (
	"errors"
	"testing"

	"github.com/QuantumNous/new-api/model"
	"gorm.io/gorm"
)

func TestEntitlementListBuyerEntitlementsSupportsModelFilter(t *testing.T) {
	db := setupMarketplaceServiceTestDB(t)
	buyer := seedMarketplaceServiceUser(t, db, "market-buyer-ent")
	sellerUser := seedMarketplaceServiceUser(t, db, "market-seller-ent")
	seller, supply := seedMarketplaceServiceSupply(t, db, sellerUser, "active", "success", "token")
	_, _ = seedMarketplaceChannelBinding(t, db, supply, "runtime-key")
	_ = seedSellerSecretRecord(t, db, seller.Id, supply.Id, `{"alg":"aes-256-gcm","kid":"v1","nonce":"bm9uY2U=","ciphertext":"Y2lwaGVy"}`, "fp-ent", "active", "success")
	listing, sku := seedMarketplaceListing(t, db, seller, supply, 1000, 199)

	order, _, err := CreateMarketOrder(CreateMarketOrderInput{
		BuyerUserID:    buyer.Id,
		ListingID:      listing.Id,
		SkuID:          sku.Id,
		Quantity:       1,
		IdempotencyKey: "order-ent-001",
	})
	if err != nil {
		t.Fatalf("create market order returned error: %v", err)
	}
	if _, err := CompleteMarketOrderPayment(CompleteMarketOrderPaymentInput{
		OrderNo:            order.OrderNo,
		PaymentMethod:      "stripe",
		PaymentTradeNo:     "pi_market_ent_001",
		Currency:           order.Currency,
		PayableAmountMinor: order.PayableAmountMinor,
		ProviderPayload:    `{"type":"checkout.session.completed"}`,
	}); err != nil {
		t.Fatalf("complete market order payment returned error: %v", err)
	}

	entitlements, total, err := ListBuyerEntitlements(buyer.Id, supply.ModelName, 0, 20)
	if err != nil {
		t.Fatalf("list buyer entitlements returned error: %v", err)
	}
	if total != 1 || len(entitlements) != 1 {
		t.Fatalf("expected one filtered entitlement, got total=%d len=%d", total, len(entitlements))
	}

	empty, filteredTotal, err := ListBuyerEntitlements(buyer.Id, "not-a-match", 0, 20)
	if err != nil {
		t.Fatalf("list buyer entitlements with miss filter returned error: %v", err)
	}
	if filteredTotal != 0 || len(empty) != 0 {
		t.Fatalf("expected no entitlements for unmatched filter, got total=%d len=%d", filteredTotal, len(empty))
	}
}

func TestEntitlementListBuyerEntitlementsIsScopedToBuyer(t *testing.T) {
	db := setupMarketplaceServiceTestDB(t)
	buyerA := seedMarketplaceServiceUser(t, db, "market-buyer-ent-scope-a")
	buyerB := seedMarketplaceServiceUser(t, db, "market-buyer-ent-scope-b")
	sellerUser := seedMarketplaceServiceUser(t, db, "market-seller-ent-scope")
	_, supply := seedMarketplaceServiceSupply(t, db, sellerUser, "active", "success", "token")

	entitlementA := &model.BuyerEntitlement{
		BuyerUserId:  buyerA.Id,
		VendorId:     supply.VendorId,
		ModelName:    supply.ModelName,
		TotalGranted: 2000,
		Status:       "active",
	}
	if err := db.Create(entitlementA).Error; err != nil {
		t.Fatalf("failed to create buyer A entitlement: %v", err)
	}

	entitlementB := &model.BuyerEntitlement{
		BuyerUserId:  buyerB.Id,
		VendorId:     supply.VendorId,
		ModelName:    supply.ModelName,
		TotalGranted: 4000,
		Status:       "active",
	}
	if err := db.Create(entitlementB).Error; err != nil {
		t.Fatalf("failed to create buyer B entitlement: %v", err)
	}

	entitlements, total, err := ListBuyerEntitlements(buyerA.Id, supply.ModelName, 0, 20)
	if err != nil {
		t.Fatalf("list buyer entitlements returned error: %v", err)
	}
	if total != 1 || len(entitlements) != 1 {
		t.Fatalf("expected exactly one entitlement for buyer A, got total=%d len=%d", total, len(entitlements))
	}
	if entitlements[0].Id != entitlementA.Id || entitlements[0].BuyerUserId != buyerA.Id {
		t.Fatalf("expected buyer A entitlement only, got %+v", entitlements[0])
	}
	if entitlements[0].Id == entitlementB.Id {
		t.Fatalf("buyer A entitlement query leaked buyer B entitlement id=%d", entitlementB.Id)
	}
}

func TestGrantEntitlementsForOrderRecoversFromDuplicateBuyerEntitlementCreate(t *testing.T) {
	db := setupMarketplaceServiceTestDB(t)
	buyer := seedMarketplaceServiceUser(t, db, "market-buyer-ent-duplicate")
	sellerUser := seedMarketplaceServiceUser(t, db, "market-seller-ent-duplicate")
	seller, supply := seedMarketplaceServiceSupply(t, db, sellerUser, "active", "success", "token")
	_, _ = seedMarketplaceChannelBinding(t, db, supply, "runtime-key")
	_ = seedSellerSecretRecord(t, db, seller.Id, supply.Id, `{"alg":"aes-256-gcm","kid":"v1","nonce":"bm9uY2U=","ciphertext":"Y2lwaGVy"}`, "fp-ent-duplicate", "active", "success")
	listing, sku := seedMarketplaceListing(t, db, seller, supply, 1500, 199)

	order, _, err := CreateMarketOrder(CreateMarketOrderInput{
		BuyerUserID:    buyer.Id,
		ListingID:      listing.Id,
		SkuID:          sku.Id,
		Quantity:       1,
		IdempotencyKey: "order-ent-duplicate-001",
	})
	if err != nil {
		t.Fatalf("create market order returned error: %v", err)
	}
	order.PaidAt = order.CreatedAt

	items, err := model.GetMarketOrderItemsByOrderID(order.Id)
	if err != nil {
		t.Fatalf("failed to load order items: %v", err)
	}

	previousCreate := createBuyerEntitlementTxFunc
	duplicateInjected := false
	createBuyerEntitlementTxFunc = func(tx *gorm.DB, entitlement *model.BuyerEntitlement) error {
		if duplicateInjected {
			return previousCreate(tx, entitlement)
		}
		duplicateInjected = true
		inserted := *entitlement
		if err := tx.Create(&inserted).Error; err != nil {
			return err
		}
		return errors.New("UNIQUE constraint failed: buyer_entitlements.buyer_user_id, buyer_entitlements.vendor_id, buyer_entitlements.model_name")
	}
	t.Cleanup(func() {
		createBuyerEntitlementTxFunc = previousCreate
	})

	if err := db.Transaction(func(tx *gorm.DB) error {
		return grantEntitlementsForOrderTx(tx, order, items)
	}); err != nil {
		t.Fatalf("expected duplicate entitlement create to recover, got error: %v", err)
	}

	entitlements, total, err := ListBuyerEntitlements(buyer.Id, supply.ModelName, 0, 20)
	if err != nil {
		t.Fatalf("list buyer entitlements returned error: %v", err)
	}
	if total != 1 || len(entitlements) != 1 {
		t.Fatalf("expected one entitlement after duplicate recovery, got total=%d len=%d", total, len(entitlements))
	}
	if entitlements[0].TotalGranted != int64(sku.PackageAmount) {
		t.Fatalf("expected granted total to be accumulated once, got %+v", entitlements[0])
	}

	var lotCount int64
	if err := db.Model(&model.EntitlementLot{}).Where("order_id = ?", order.Id).Count(&lotCount).Error; err != nil {
		t.Fatalf("failed to count entitlement lots after duplicate recovery: %v", err)
	}
	if lotCount != 1 {
		t.Fatalf("expected exactly one entitlement lot after duplicate recovery, got %d", lotCount)
	}
}
