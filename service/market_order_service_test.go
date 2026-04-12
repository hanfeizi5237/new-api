package service

import (
	"testing"

	"github.com/QuantumNous/new-api/model"
)

func TestMarketOrderCreateFreezesInventoryAndIsIdempotent(t *testing.T) {
	db := setupMarketplaceServiceTestDB(t)
	buyer := seedMarketplaceServiceUser(t, db, "market-buyer-create")
	sellerUser := seedMarketplaceServiceUser(t, db, "market-seller-create")
	seller, supply := seedMarketplaceServiceSupply(t, db, sellerUser, "active", "success", "token")
	_, _ = seedMarketplaceChannelBinding(t, db, supply, "runtime-key")
	_ = seedSellerSecretRecord(t, db, seller.Id, supply.Id, `{"alg":"aes-256-gcm","kid":"v1","nonce":"bm9uY2U=","ciphertext":"Y2lwaGVy"}`, "fp-order", "active", "success")
	listing, sku := seedMarketplaceListing(t, db, seller, supply, 1200, 299)

	order, items, err := CreateMarketOrder(CreateMarketOrderInput{
		BuyerUserID:    buyer.Id,
		ListingID:      listing.Id,
		SkuID:          sku.Id,
		Quantity:       2,
		IdempotencyKey: "order-idem-001",
	})
	if err != nil {
		t.Fatalf("create market order returned error: %v", err)
	}
	if order.OrderStatus != "pending_payment" || order.PaymentStatus != "unpaid" || order.EntitlementStatus != "pending" {
		t.Fatalf("unexpected order status after create: %+v", order)
	}
	if order.PayableAmountMinor != 598 {
		t.Fatalf("expected payable amount 598, got %d", order.PayableAmountMinor)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 order item, got %d", len(items))
	}
	if items[0].LineAmountMinor != 598 || items[0].PackageAmount != 1200 || items[0].Quantity != 2 {
		t.Fatalf("unexpected order item snapshot: %+v", items[0])
	}

	snapshot, err := model.GetInventorySnapshotBySupplyAccountID(supply.Id)
	if err != nil {
		t.Fatalf("failed to reload inventory snapshot: %v", err)
	}
	if snapshot.FrozenAmount != 2400 {
		t.Fatalf("expected frozen amount 2400, got %d", snapshot.FrozenAmount)
	}
	if snapshot.AvailableAmount != supply.SellableCapacity-2400 {
		t.Fatalf("expected available amount %d, got %d", supply.SellableCapacity-2400, snapshot.AvailableAmount)
	}

	secondOrder, secondItems, err := CreateMarketOrder(CreateMarketOrderInput{
		BuyerUserID:    buyer.Id,
		ListingID:      listing.Id,
		SkuID:          sku.Id,
		Quantity:       2,
		IdempotencyKey: "order-idem-001",
	})
	if err != nil {
		t.Fatalf("second create market order returned error: %v", err)
	}
	if secondOrder.Id != order.Id || len(secondItems) != 1 || secondItems[0].Id != items[0].Id {
		t.Fatalf("expected idempotent create to return existing order, got order=%+v items=%+v", secondOrder, secondItems)
	}

	snapshot, err = model.GetInventorySnapshotBySupplyAccountID(supply.Id)
	if err != nil {
		t.Fatalf("failed to reload inventory snapshot after second create: %v", err)
	}
	if snapshot.FrozenAmount != 2400 {
		t.Fatalf("expected frozen amount unchanged at 2400, got %d", snapshot.FrozenAmount)
	}
}

func TestCloseExpiredMarketOrdersReleasesFrozenInventory(t *testing.T) {
	db := setupMarketplaceServiceTestDB(t)
	buyer := seedMarketplaceServiceUser(t, db, "market-buyer-expire")
	sellerUser := seedMarketplaceServiceUser(t, db, "market-seller-expire")
	seller, supply := seedMarketplaceServiceSupply(t, db, sellerUser, "active", "success", "token")
	_, _ = seedMarketplaceChannelBinding(t, db, supply, "runtime-key")
	_ = seedSellerSecretRecord(t, db, seller.Id, supply.Id, `{"alg":"aes-256-gcm","kid":"v1","nonce":"bm9uY2U=","ciphertext":"Y2lwaGVy"}`, "fp-expire", "active", "success")
	listing, sku := seedMarketplaceListing(t, db, seller, supply, 1500, 399)

	order, _, err := CreateMarketOrder(CreateMarketOrderInput{
		BuyerUserID:    buyer.Id,
		ListingID:      listing.Id,
		SkuID:          sku.Id,
		Quantity:       1,
		IdempotencyKey: "order-expire-001",
	})
	if err != nil {
		t.Fatalf("create market order returned error: %v", err)
	}
	if err := db.Model(&model.MarketOrder{}).Where("id = ?", order.Id).Update("expire_at", order.CreatedAt-1).Error; err != nil {
		t.Fatalf("failed to backdate order expiry: %v", err)
	}

	closedCount, err := CloseExpiredMarketOrders(order.CreatedAt)
	if err != nil {
		t.Fatalf("close expired market orders returned error: %v", err)
	}
	if closedCount != 1 {
		t.Fatalf("expected 1 closed order, got %d", closedCount)
	}

	reloadedOrder, err := model.GetMarketOrderByID(order.Id)
	if err != nil {
		t.Fatalf("failed to reload order: %v", err)
	}
	if reloadedOrder.OrderStatus != "closed" || reloadedOrder.PaymentStatus != "unpaid" {
		t.Fatalf("unexpected order after expire close: %+v", reloadedOrder)
	}
	snapshot, err := model.GetInventorySnapshotBySupplyAccountID(supply.Id)
	if err != nil {
		t.Fatalf("failed to reload inventory snapshot: %v", err)
	}
	if snapshot.FrozenAmount != 0 || snapshot.AvailableAmount != supply.SellableCapacity {
		t.Fatalf("expected inventory release after close, got %+v", snapshot)
	}
}
