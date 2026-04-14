package controller

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/Calcium-Ion/go-epay/epay"
	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"
	"github.com/QuantumNous/new-api/setting"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v81/webhook"
	"gorm.io/gorm"
)

func TestMarketEpayNotifyReturnsFailWhenCompletionFails(t *testing.T) {
	db := setupMarketplaceControllerTestDB(t)
	configureMarketplaceEpayTest(t)

	params := generateMarketplaceEpayParams("missing-market-order", "4.99", epay.StatusTradeSuccess)
	ctx, recorder := newMarketplaceEpayNotifyContext(params)
	MarketEpayNotify(ctx)

	if recorder.Body.String() != "fail" {
		t.Fatalf("expected epay notify to return fail when local completion fails, got %q", recorder.Body.String())
	}

	var orderCount int64
	if err := db.Model(&model.MarketOrder{}).Where("order_no = ?", "missing-market-order").Count(&orderCount).Error; err != nil {
		t.Fatalf("failed to count missing order: %v", err)
	}
	if orderCount != 0 {
		t.Fatalf("expected no market order to be created for missing callback order, got %d", orderCount)
	}
}

func TestMarketEpayNotifyRejectsAmountMismatch(t *testing.T) {
	db := setupMarketplaceControllerTestDB(t)
	configureMarketplaceEpayTest(t)

	order, supply := seedMarketplacePayableOrder(t, db, "buyer-amount-mismatch", "seller-amount-mismatch")
	params := generateMarketplaceEpayParams(order.OrderNo, "0.01", epay.StatusTradeSuccess)
	ctx, recorder := newMarketplaceEpayNotifyContext(params)
	MarketEpayNotify(ctx)

	if recorder.Body.String() != "fail" {
		t.Fatalf("expected epay notify to return fail on amount mismatch, got %q", recorder.Body.String())
	}

	reloadedOrder, err := model.GetMarketOrderByID(order.Id)
	if err != nil {
		t.Fatalf("failed to reload order: %v", err)
	}
	if reloadedOrder.PaymentStatus == "paid" || reloadedOrder.OrderStatus == "paid" {
		t.Fatalf("expected mismatched callback not to complete order, got %+v", reloadedOrder)
	}

	snapshot, err := model.GetInventorySnapshotBySupplyAccountID(supply.Id)
	if err != nil {
		t.Fatalf("failed to reload inventory snapshot: %v", err)
	}
	if snapshot.FrozenAmount == 0 {
		t.Fatalf("expected frozen inventory to remain reserved after processing failure, got %+v", snapshot)
	}
}

func TestMarketEpayNotifyClosesOrderOnFailedTradeStatus(t *testing.T) {
	db := setupMarketplaceControllerTestDB(t)
	configureMarketplaceEpayTest(t)

	order, supply := seedMarketplacePayableOrder(t, db, "buyer-failed-status", "seller-failed-status")
	params := generateMarketplaceEpayParams(order.OrderNo, "4.99", "TRADE_CLOSED")
	ctx, recorder := newMarketplaceEpayNotifyContext(params)
	MarketEpayNotify(ctx)

	if recorder.Body.String() != "success" {
		t.Fatalf("expected epay notify to acknowledge handled failed trade, got %q", recorder.Body.String())
	}

	reloadedOrder, err := model.GetMarketOrderByID(order.Id)
	if err != nil {
		t.Fatalf("failed to reload order: %v", err)
	}
	if reloadedOrder.OrderStatus != "closed" || reloadedOrder.PaymentStatus != "failed" {
		t.Fatalf("expected failed trade to close order and mark payment failed, got %+v", reloadedOrder)
	}

	snapshot, err := model.GetInventorySnapshotBySupplyAccountID(supply.Id)
	if err != nil {
		t.Fatalf("failed to reload inventory snapshot: %v", err)
	}
	if snapshot.FrozenAmount != 0 || snapshot.AvailableAmount != supply.SellableCapacity {
		t.Fatalf("expected failed trade to release frozen inventory, got %+v", snapshot)
	}
}

func TestMarketStripeWebhookRejectsInvalidAmountTotal(t *testing.T) {
	db := setupMarketplaceControllerTestDB(t)

	previousSecret := setting.StripeWebhookSecret
	setting.StripeWebhookSecret = "whsec_market_test"
	t.Cleanup(func() {
		setting.StripeWebhookSecret = previousSecret
	})

	order, supply := seedMarketplacePayableOrder(t, db, "buyer-stripe-invalid-amount", "seller-stripe-invalid-amount")
	payload, err := common.Marshal(map[string]any{
		"id":   "evt_market_stripe_invalid_amount",
		"type": "checkout.session.completed",
		"data": map[string]any{
			"object": map[string]any{
				"id":                  "cs_test_market_invalid_amount",
				"object":              "checkout.session",
				"client_reference_id": order.OrderNo,
				"payment_intent":      "pi_market_invalid_amount",
				"currency":            "cny",
				"amount_total":        "invalid-amount",
			},
		},
	})
	if err != nil {
		t.Fatalf("failed to marshal stripe payload: %v", err)
	}
	signedPayload := webhook.GenerateTestSignedPayload(&webhook.UnsignedPayload{
		Payload: payload,
		Secret:  setting.StripeWebhookSecret,
	})

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/api/market/payment/stripe/webhook", bytes.NewReader(payload))
	ctx.Request.Header.Set("Stripe-Signature", signedPayload.Header)

	MarketStripeWebhook(ctx)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected bad request for invalid stripe amount_total, got status=%d body=%q", recorder.Code, recorder.Body.String())
	}

	reloadedOrder, err := model.GetMarketOrderByID(order.Id)
	if err != nil {
		t.Fatalf("failed to reload order: %v", err)
	}
	if reloadedOrder.PaymentStatus == "paid" || reloadedOrder.OrderStatus == "paid" {
		t.Fatalf("expected invalid stripe callback not to complete order, got %+v", reloadedOrder)
	}

	snapshot, err := model.GetInventorySnapshotBySupplyAccountID(supply.Id)
	if err != nil {
		t.Fatalf("failed to reload inventory snapshot: %v", err)
	}
	if snapshot.FrozenAmount == 0 {
		t.Fatalf("expected invalid stripe callback to leave frozen inventory intact, got %+v", snapshot)
	}
}

func configureMarketplaceEpayTest(t *testing.T) {
	t.Helper()

	previousPayAddress := operation_setting.PayAddress
	previousEpayID := operation_setting.EpayId
	previousEpayKey := operation_setting.EpayKey
	operation_setting.PayAddress = "https://epay.test"
	operation_setting.EpayId = "market-test-partner"
	operation_setting.EpayKey = "market-test-key"
	t.Cleanup(func() {
		operation_setting.PayAddress = previousPayAddress
		operation_setting.EpayId = previousEpayID
		operation_setting.EpayKey = previousEpayKey
	})
}

func generateMarketplaceEpayParams(orderNo string, money string, tradeStatus string) map[string]string {
	return epay.GenerateParams(map[string]string{
		"pid":          operation_setting.EpayId,
		"type":         "alipay",
		"out_trade_no": orderNo,
		"trade_no":     "epay_trade_market_001",
		"name":         "Marketplace Order",
		"money":        money,
		"trade_status": tradeStatus,
	}, operation_setting.EpayKey)
}

func newMarketplaceEpayNotifyContext(params map[string]string) (*gin.Context, *httptest.ResponseRecorder) {
	query := url.Values{}
	for key, value := range params {
		query.Set(key, value)
	}
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/market/payment/epay/notify?"+query.Encode(), nil)
	return ctx, recorder
}

func seedMarketplacePayableOrder(t *testing.T, db *gorm.DB, buyerName string, sellerName string) (*model.MarketOrder, *model.SupplyAccount) {
	t.Helper()

	buyer := seedMarketplaceUser(t, db, buyerName)
	sellerUser := seedMarketplaceUser(t, db, sellerName)
	seller, supply := seedMarketplaceSellerWithSupply(t, db, sellerUser.Id)
	supply.VerifyStatus = "success"
	if err := db.Save(supply).Error; err != nil {
		t.Fatalf("failed to update supply verify status: %v", err)
	}
	_, _ = seedMarketplaceChannelBinding(t, db, supply, "runtime-key")

	if err := db.Create(&model.SellerSecret{
		SellerId:        seller.Id,
		SupplyAccountId: supply.Id,
		SecretType:      "api_key",
		ProviderCode:    "openai",
		Ciphertext:      `{"alg":"aes-256-gcm","kid":"v1","nonce":"bm9uY2U=","ciphertext":"Y2lwaGVy"}`,
		CipherVersion:   "v1",
		Fingerprint:     fmt.Sprintf("fp-%s", buyerName),
		MaskedValue:     "sk-***market",
		Status:          "active",
		VerifyStatus:    "success",
		VerifyMessage:   "seeded active secret",
	}).Error; err != nil {
		t.Fatalf("failed to seed active secret: %v", err)
	}

	listing := &model.Listing{
		SellerId:        seller.Id,
		SupplyAccountId: supply.Id,
		ListingCode:     fmt.Sprintf("listing-%s", buyerName),
		Title:           "Market Payment Fixture",
		VendorId:        supply.VendorId,
		ModelName:       supply.ModelName,
		SaleMode:        "fixed_price",
		PricingUnit:     "per_token_package",
		ValidityDays:    30,
		AuditStatus:     "approved",
		Status:          "active",
	}
	if err := db.Create(listing).Error; err != nil {
		t.Fatalf("failed to create listing: %v", err)
	}
	sku := &model.ListingSKU{
		ListingId:      listing.Id,
		SkuCode:        fmt.Sprintf("sku-%s", buyerName),
		PackageAmount:  2000,
		PackageUnit:    "token",
		UnitPriceMinor: 499,
		MinQuantity:    1,
		MaxQuantity:    5,
		Status:         "active",
	}
	if err := db.Create(sku).Error; err != nil {
		t.Fatalf("failed to create sku: %v", err)
	}

	order, _, err := service.CreateMarketOrder(service.CreateMarketOrderInput{
		BuyerUserID:    buyer.Id,
		ListingID:      listing.Id,
		SkuID:          sku.Id,
		Quantity:       1,
		IdempotencyKey: fmt.Sprintf("idem-%s", buyerName),
		Currency:       "CNY",
	})
	if err != nil {
		t.Fatalf("failed to create market order fixture: %v", err)
	}
	return order, supply
}
