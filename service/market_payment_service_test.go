package service

import (
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
)

func TestPrepareMarketOrderPaymentCreatesProviderIntentAndPersistsReference(t *testing.T) {
	db := setupMarketplaceServiceTestDB(t)
	buyer := seedMarketplaceServiceUser(t, db, "market-buyer-intent")
	sellerUser := seedMarketplaceServiceUser(t, db, "market-seller-intent")
	seller, supply := seedMarketplaceServiceSupply(t, db, sellerUser, "active", "success", "token")
	_, _ = seedMarketplaceChannelBinding(t, db, supply, "runtime-key")
	_ = seedSellerSecretRecord(t, db, seller.Id, supply.Id, `{"alg":"aes-256-gcm","kid":"v1","nonce":"bm9uY2U=","ciphertext":"Y2lwaGVy"}`, "fp-intent", "active", "success")
	listing, sku := seedMarketplaceListing(t, db, seller, supply, 2000, 499)

	order, _, err := CreateMarketOrder(CreateMarketOrderInput{
		BuyerUserID:    buyer.Id,
		ListingID:      listing.Id,
		SkuID:          sku.Id,
		Quantity:       1,
		IdempotencyKey: "order-pay-intent-001",
	})
	if err != nil {
		t.Fatalf("create market order returned error: %v", err)
	}

	previousStripeCreator := marketStripePaymentIntentCreator
	marketStripePaymentIntentCreator = func(input marketProviderPaymentInitInput) (*marketProviderPaymentInitResult, error) {
		if input.Order == nil || input.Order.OrderNo != order.OrderNo {
			t.Fatalf("unexpected provider order input: %+v", input.Order)
		}
		return &marketProviderPaymentInitResult{
			ProviderReference: "cs_test_market_001",
			RedirectURL:       "https://checkout.stripe.test/session/cs_test_market_001",
			ProviderPayload:   `{"client_reference_id":"` + order.OrderNo + `"}`,
		}, nil
	}
	t.Cleanup(func() {
		marketStripePaymentIntentCreator = previousStripeCreator
	})

	intent, err := PrepareMarketOrderPayment(PrepareMarketOrderPaymentInput{
		OrderID:       order.Id,
		BuyerUserID:   buyer.Id,
		PaymentMethod: "stripe",
	})
	if err != nil {
		t.Fatalf("prepare market order payment returned error: %v", err)
	}
	if intent.ProviderReference != "cs_test_market_001" || intent.RedirectURL == "" {
		t.Fatalf("expected provider intent reference and redirect url, got %+v", intent)
	}

	reloadedOrder, err := model.GetMarketOrderByID(order.Id)
	if err != nil {
		t.Fatalf("failed to reload order: %v", err)
	}
	if reloadedOrder.PaymentMethod != "stripe" {
		t.Fatalf("expected payment method persisted, got %+v", reloadedOrder)
	}
	if !strings.Contains(reloadedOrder.ProviderPayload, order.OrderNo) {
		t.Fatalf("expected provider payload to persist platform reference, got %q", reloadedOrder.ProviderPayload)
	}
}

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

	previousStripeCreator := marketStripePaymentIntentCreator
	marketStripePaymentIntentCreator = func(input marketProviderPaymentInitInput) (*marketProviderPaymentInitResult, error) {
		return &marketProviderPaymentInitResult{
			ProviderReference: "cs_test_market_pay_001",
			RedirectURL:       "https://checkout.stripe.test/session/cs_test_market_pay_001",
			ProviderPayload:   `{"client_reference_id":"` + order.OrderNo + `"}`,
		}, nil
	}
	t.Cleanup(func() {
		marketStripePaymentIntentCreator = previousStripeCreator
	})

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

func TestFailMarketOrderPaymentClosesOrderAndReleasesFrozenInventory(t *testing.T) {
	db := setupMarketplaceServiceTestDB(t)
	buyer := seedMarketplaceServiceUser(t, db, "market-buyer-fail")
	sellerUser := seedMarketplaceServiceUser(t, db, "market-seller-fail")
	seller, supply := seedMarketplaceServiceSupply(t, db, sellerUser, "active", "success", "token")
	_, _ = seedMarketplaceChannelBinding(t, db, supply, "runtime-key")
	_ = seedSellerSecretRecord(t, db, seller.Id, supply.Id, `{"alg":"aes-256-gcm","kid":"v1","nonce":"bm9uY2U=","ciphertext":"Y2lwaGVy"}`, "fp-fail", "active", "success")
	listing, sku := seedMarketplaceListing(t, db, seller, supply, 1500, 299)

	order, _, err := CreateMarketOrder(CreateMarketOrderInput{
		BuyerUserID:    buyer.Id,
		ListingID:      listing.Id,
		SkuID:          sku.Id,
		Quantity:       1,
		IdempotencyKey: "order-pay-fail-001",
	})
	if err != nil {
		t.Fatalf("create market order returned error: %v", err)
	}

	snapshot, err := model.GetInventorySnapshotBySupplyAccountID(supply.Id)
	if err != nil {
		t.Fatalf("failed to load inventory snapshot before fail: %v", err)
	}
	if snapshot.FrozenAmount != 1500 {
		t.Fatalf("expected frozen amount 1500 before payment fail, got %+v", snapshot)
	}

	failedOrder, err := FailMarketOrderPayment(FailMarketOrderPaymentInput{
		OrderNo:            order.OrderNo,
		PaymentMethod:      "waffo",
		PaymentTradeNo:     "payment_failed_market_001",
		Currency:           order.Currency,
		PayableAmountMinor: order.PayableAmountMinor,
		FailureReason:      "provider_declined",
		ProviderPayload:    `{"orderStatus":"PAY_FAIL"}`,
	})
	if err != nil {
		t.Fatalf("fail market order payment returned error: %v", err)
	}
	if failedOrder.OrderStatus != "closed" || failedOrder.PaymentStatus != "failed" {
		t.Fatalf("expected closed/failed order, got %+v", failedOrder)
	}

	var item model.MarketOrderItem
	if err := db.Where("order_id = ?", order.Id).First(&item).Error; err != nil {
		t.Fatalf("failed to reload order item: %v", err)
	}
	if item.Status != "closed" {
		t.Fatalf("expected closed order item after payment failure, got %+v", item)
	}

	snapshot, err = model.GetInventorySnapshotBySupplyAccountID(supply.Id)
	if err != nil {
		t.Fatalf("failed to reload inventory snapshot after fail: %v", err)
	}
	if snapshot.FrozenAmount != 0 || snapshot.AvailableAmount != supply.SellableCapacity {
		t.Fatalf("expected released frozen inventory after payment failure, got %+v", snapshot)
	}
}

func TestMarketPaymentRetryAfterEntitlementFailureDoesNotDoubleCountInventory(t *testing.T) {
	db := setupMarketplaceServiceTestDB(t)
	buyer := seedMarketplaceServiceUser(t, db, "market-buyer-retry")
	sellerUser := seedMarketplaceServiceUser(t, db, "market-seller-retry")
	seller, supply := seedMarketplaceServiceSupply(t, db, sellerUser, "active", "success", "token")
	_, _ = seedMarketplaceChannelBinding(t, db, supply, "runtime-key")
	_ = seedSellerSecretRecord(t, db, seller.Id, supply.Id, `{"alg":"aes-256-gcm","kid":"v1","nonce":"bm9uY2U=","ciphertext":"Y2lwaGVy"}`, "fp-retry", "active", "success")
	listing, sku := seedMarketplaceListing(t, db, seller, supply, 1800, 399)

	order, _, err := CreateMarketOrder(CreateMarketOrderInput{
		BuyerUserID:    buyer.Id,
		ListingID:      listing.Id,
		SkuID:          sku.Id,
		Quantity:       1,
		IdempotencyKey: "order-pay-retry-001",
	})
	if err != nil {
		t.Fatalf("create market order returned error: %v", err)
	}

	var orderItem model.MarketOrderItem
	if err := db.Where("order_id = ?", order.Id).First(&orderItem).Error; err != nil {
		t.Fatalf("failed to load market order item: %v", err)
	}
	grantAmount := orderItem.PackageAmount * int64(orderItem.Quantity)
	paidAt := common.GetTimestamp()

	if err := db.Model(&model.MarketOrder{}).
		Where("id = ?", order.Id).
		Updates(map[string]interface{}{
			"order_status":       marketOrderStatusPaid,
			"payment_status":     marketPaymentStatusPaid,
			"entitlement_status": marketEntitlementStatusFailed,
			"payment_method":     "stripe",
			"payment_trade_no":   "pi_market_retry_first",
			"provider_payload":   `{"type":"checkout.session.completed"}`,
			"paid_at":            paidAt,
			"updated_at":         paidAt,
		}).Error; err != nil {
		t.Fatalf("failed to seed paid+failed order state: %v", err)
	}

	if err := db.Model(&model.InventorySnapshot{}).
		Where("supply_account_id = ?", supply.Id).
		Updates(map[string]interface{}{
			"frozen_amount":    0,
			"sold_amount":      grantAmount,
			"available_amount": supply.SellableCapacity - grantAmount,
			"updated_at":       paidAt,
		}).Error; err != nil {
		t.Fatalf("failed to seed inventory snapshot after first payment: %v", err)
	}

	if err := db.Model(&model.SupplyAccount{}).
		Where("id = ?", supply.Id).
		Updates(map[string]interface{}{
			"reserved_capacity": grantAmount,
			"updated_at":        paidAt,
		}).Error; err != nil {
		t.Fatalf("failed to seed supply reserved capacity after first payment: %v", err)
	}

	paidOrder, err := CompleteMarketOrderPayment(CompleteMarketOrderPaymentInput{
		OrderNo:            order.OrderNo,
		PaymentMethod:      "stripe",
		PaymentTradeNo:     "pi_market_retry_second",
		Currency:           order.Currency,
		PayableAmountMinor: order.PayableAmountMinor,
		ProviderPayload:    `{"type":"checkout.session.completed","retry":true}`,
	})
	if err != nil {
		t.Fatalf("retry complete market order payment returned error: %v", err)
	}
	if paidOrder.OrderStatus != marketOrderStatusPaid || paidOrder.PaymentStatus != marketPaymentStatusPaid || paidOrder.EntitlementStatus != marketEntitlementStatusCreated {
		t.Fatalf("expected retry to finish entitlement grant without changing paid state, got %+v", paidOrder)
	}

	snapshot, err := model.GetInventorySnapshotBySupplyAccountID(supply.Id)
	if err != nil {
		t.Fatalf("failed to reload inventory snapshot: %v", err)
	}
	if snapshot.FrozenAmount != 0 || snapshot.SoldAmount != grantAmount || snapshot.AvailableAmount != supply.SellableCapacity-grantAmount {
		t.Fatalf("retry should not double count inventory, got %+v", snapshot)
	}

	reloadedSupply, err := model.GetSupplyAccountByID(supply.Id)
	if err != nil {
		t.Fatalf("failed to reload supply account: %v", err)
	}
	if reloadedSupply.ReservedCapacity != grantAmount {
		t.Fatalf("retry should not double count reserved capacity, got %+v", reloadedSupply)
	}

	entitlements, total, err := ListBuyerEntitlements(buyer.Id, "", 0, 20)
	if err != nil {
		t.Fatalf("list buyer entitlements returned error: %v", err)
	}
	if total != 1 || len(entitlements) != 1 || entitlements[0].TotalGranted != grantAmount {
		t.Fatalf("retry should create exactly one entitlement, got total=%d entitlements=%+v", total, entitlements)
	}

	var lotCount int64
	if err := db.Model(&model.EntitlementLot{}).Where("order_id = ?", order.Id).Count(&lotCount).Error; err != nil {
		t.Fatalf("failed to count entitlement lots: %v", err)
	}
	if lotCount != 1 {
		t.Fatalf("retry should create exactly one entitlement lot, got %d", lotCount)
	}
}
