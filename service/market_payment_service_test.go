package service

import (
	"testing"

	"github.com/QuantumNous/new-api/model"
)

func TestMarketPaymentCompletionCreatesEntitlementOnce(t *testing.T) {
	db := setupMarketplaceServiceTestDB(t)
	buyer := seedMarketplaceServiceUser(t, db, "market-buyer-pay")
	sellerUser := seedMarketplaceServiceUser(t, db, "market-seller-pay")
	seller, supply := seedMarketplaceServiceSupply(t, db, sellerUser, "active", "success", "token")
	_, _ = seedMarketplaceChannelBinding(t, db, supply, "runtime-key")
	_ = seedSellerSecretRecord(t, db, seller.Id, supply.Id, `{"alg":"aes-256-gcm","kid":"v1","nonce":"bm9uY2U=","ciphertext":"Y2lwaGVy"}`, "fp-pay", "active", "success")
	listing, sku := seedMarketplaceListing(t, db, seller, supply, 2000, 499)

	order, _, err := CreateMarketOrder(CreateMarketOrderInput{
		BuyerUserID:    buyer.Id,
		ListingID:      listing.Id,
		SkuID:          sku.Id,
		Quantity:       1,
		IdempotencyKey: "order-pay-001",
	})
	if err != nil {
		t.Fatalf("create market order returned error: %v", err)
	}

	intent, err := PrepareMarketOrderPayment(PrepareMarketOrderPaymentInput{
		OrderID:       order.Id,
		BuyerUserID:   buyer.Id,
		PaymentMethod: "stripe",
	})
	if err != nil {
		t.Fatalf("prepare market order payment returned error: %v", err)
	}
	if intent.OrderNo != order.OrderNo || intent.PaymentMethod != "stripe" {
		t.Fatalf("unexpected market payment intent: %+v", intent)
	}

	paidOrder, err := CompleteMarketOrderPayment(CompleteMarketOrderPaymentInput{
		OrderNo:            order.OrderNo,
		PaymentMethod:      "stripe",
		PaymentTradeNo:     "pi_market_001",
		Currency:           order.Currency,
		PayableAmountMinor: order.PayableAmountMinor,
		ProviderPayload:    `{"type":"checkout.session.completed"}`,
	})
	if err != nil {
		t.Fatalf("complete market order payment returned error: %v", err)
	}
	if paidOrder.OrderStatus != "paid" || paidOrder.PaymentStatus != "paid" || paidOrder.EntitlementStatus != "created" {
		t.Fatalf("unexpected paid order state: %+v", paidOrder)
	}

	entitlements, total, err := ListBuyerEntitlements(buyer.Id, "", 0, 20)
	if err != nil {
		t.Fatalf("list buyer entitlements returned error: %v", err)
	}
	if total != 1 || len(entitlements) != 1 {
		t.Fatalf("expected one entitlement, got total=%d len=%d", total, len(entitlements))
	}
	if entitlements[0].TotalGranted != 2000 {
		t.Fatalf("expected entitlement total granted 2000, got %+v", entitlements[0])
	}

	var lotCount int64
	if err := db.Model(&model.EntitlementLot{}).Where("order_id = ?", order.Id).Count(&lotCount).Error; err != nil {
		t.Fatalf("failed to count entitlement lots: %v", err)
	}
	if lotCount != 1 {
		t.Fatalf("expected 1 entitlement lot, got %d", lotCount)
	}

	var orderItem model.MarketOrderItem
	if err := db.Where("order_id = ?", order.Id).First(&orderItem).Error; err != nil {
		t.Fatalf("failed to reload order item: %v", err)
	}
	if orderItem.GrantedAmount != 2000 || orderItem.Status != "granted" {
		t.Fatalf("unexpected order item after payment: %+v", orderItem)
	}

	snapshot, err := model.GetInventorySnapshotBySupplyAccountID(supply.Id)
	if err != nil {
		t.Fatalf("failed to reload inventory snapshot: %v", err)
	}
	if snapshot.FrozenAmount != 0 || snapshot.SoldAmount != 2000 || snapshot.AvailableAmount != supply.SellableCapacity-2000 {
		t.Fatalf("unexpected inventory snapshot after payment: %+v", snapshot)
	}

	reloadedSupply, err := model.GetSupplyAccountByID(supply.Id)
	if err != nil {
		t.Fatalf("failed to reload supply account: %v", err)
	}
	if reloadedSupply.ReservedCapacity != 2000 {
		t.Fatalf("expected reserved capacity 2000, got %+v", reloadedSupply)
	}

	if _, err := CompleteMarketOrderPayment(CompleteMarketOrderPaymentInput{
		OrderNo:            order.OrderNo,
		PaymentMethod:      "stripe",
		PaymentTradeNo:     "pi_market_001",
		Currency:           order.Currency,
		PayableAmountMinor: order.PayableAmountMinor,
		ProviderPayload:    `{"type":"checkout.session.completed"}`,
	}); err != nil {
		t.Fatalf("repeat completion should be idempotent, got error: %v", err)
	}

	entitlements, total, err = ListBuyerEntitlements(buyer.Id, "", 0, 20)
	if err != nil {
		t.Fatalf("list buyer entitlements after repeat returned error: %v", err)
	}
	if total != 1 || len(entitlements) != 1 || entitlements[0].TotalGranted != 2000 {
		t.Fatalf("repeat completion should not duplicate entitlement, got total=%d entitlements=%+v", total, entitlements)
	}

	if err := db.Model(&model.EntitlementLot{}).Where("order_id = ?", order.Id).Count(&lotCount).Error; err != nil {
		t.Fatalf("failed to count entitlement lots after repeat: %v", err)
	}
	if lotCount != 1 {
		t.Fatalf("repeat completion should not duplicate lots, got %d", lotCount)
	}
}
