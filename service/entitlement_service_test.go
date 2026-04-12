package service

import "testing"

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
