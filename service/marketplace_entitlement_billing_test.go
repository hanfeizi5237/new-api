package service

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/dto"
	"github.com/QuantumNous/new-api/model"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	relayconstant "github.com/QuantumNous/new-api/relay/constant"
	"github.com/QuantumNous/new-api/types"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func TestNewBillingSessionUsesMarketplaceEntitlementBeforeWallet(t *testing.T) {
	db := setupMarketplaceServiceTestDB(t)
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())

	fixture := seedMarketplaceEntitlementBillingFixture(t, db, "market-buyer-billing-priority", 2000)
	fixture.buyer.Quota = 999999
	if err := db.Save(fixture.buyer).Error; err != nil {
		t.Fatalf("failed to seed buyer quota: %v", err)
	}

	relayInfo := &relaycommon.RelayInfo{
		UserId:          fixture.buyer.Id,
		OriginModelName: fixture.supply.ModelName,
		RequestId:       "req-marketplace-priority-001",
		TokenId:         1001,
		TokenKey:        "market-token-priority",
		RelayMode:       relayconstant.RelayModeChatCompletions,
		UserSetting: dto.UserSetting{
			BillingPreference: "wallet_first",
		},
	}

	session, apiErr := NewBillingSession(ctx, relayInfo, 500)
	if apiErr != nil {
		t.Fatalf("expected marketplace entitlement billing session, got error: %v", apiErr)
	}
	if session == nil {
		t.Fatalf("expected billing session")
	}
	if session.funding.Source() != BillingSourceMarketplaceEntitlement {
		t.Fatalf("expected marketplace entitlement funding source, got %s", session.funding.Source())
	}
	if relayInfo.BillingSource != BillingSourceMarketplaceEntitlement {
		t.Fatalf("expected relayInfo billing source to be marketplace entitlement, got %+v", relayInfo)
	}
	if relayInfo.EntitlementLotId <= 0 || relayInfo.SupplyAccountId != fixture.supply.Id || relayInfo.OrderId != fixture.order.Id {
		t.Fatalf("expected relay info to carry entitlement metadata, got %+v", relayInfo)
	}

	lot, err := getEntitlementLotByID(fixture.lot.Id)
	if err != nil {
		t.Fatalf("failed to reload entitlement lot: %v", err)
	}
	if lot.FrozenAmount != 500 {
		t.Fatalf("expected entitlement lot frozen amount 500, got %+v", lot)
	}
	entitlement, err := getBuyerEntitlementByID(fixture.entitlement.Id)
	if err != nil {
		t.Fatalf("failed to reload buyer entitlement: %v", err)
	}
	if entitlement.TotalFrozen != 500 {
		t.Fatalf("expected buyer entitlement total frozen 500, got %+v", entitlement)
	}
}

func TestNewBillingSessionReturnsInsufficientMarketplaceEntitlementWithoutWalletFallback(t *testing.T) {
	db := setupMarketplaceServiceTestDB(t)
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())

	fixture := seedMarketplaceEntitlementBillingFixture(t, db, "market-buyer-billing-insufficient", 1000)
	fixture.buyer.Quota = 999999
	if err := db.Save(fixture.buyer).Error; err != nil {
		t.Fatalf("failed to seed buyer quota: %v", err)
	}

	relayInfo := &relaycommon.RelayInfo{
		UserId:          fixture.buyer.Id,
		OriginModelName: fixture.supply.ModelName,
		RequestId:       "req-marketplace-insufficient-001",
		TokenId:         1002,
		TokenKey:        "market-token-insufficient",
		RelayMode:       relayconstant.RelayModeChatCompletions,
		UserSetting: dto.UserSetting{
			BillingPreference: "wallet_first",
		},
	}

	session, apiErr := NewBillingSession(ctx, relayInfo, 1500)
	if apiErr == nil {
		t.Fatalf("expected insufficient marketplace entitlement error, got session=%+v", session)
	}
	if apiErr.GetErrorCode() != types.ErrorCodeInsufficientMarketplaceEntitlement {
		t.Fatalf("expected insufficient marketplace entitlement error code, got %s (%v)", apiErr.GetErrorCode(), apiErr)
	}
	if session != nil {
		t.Fatalf("expected no billing session on insufficient marketplace entitlement, got %+v", session)
	}

	entitlement, err := getBuyerEntitlementByID(fixture.entitlement.Id)
	if err != nil {
		t.Fatalf("failed to reload buyer entitlement: %v", err)
	}
	if entitlement.TotalFrozen != 0 {
		t.Fatalf("expected insufficient entitlement not to freeze any quota, got %+v", entitlement)
	}
}

func TestResolveMarketplaceEntitlementBillingSessionSkipsWhenNoMarketplaceEntitlementMatches(t *testing.T) {
	db := setupMarketplaceServiceTestDB(t)
	_ = seedMarketplaceServiceUser(t, db, "market-buyer-no-entitlement")

	relayInfo := &relaycommon.RelayInfo{
		UserId:          1,
		OriginModelName: "gpt-4o-mini",
		RequestId:       "req-marketplace-no-match-001",
		TokenId:         1006,
		TokenKey:        "market-token-no-match",
		RelayMode:       relayconstant.RelayModeChatCompletions,
	}

	session, apiErr, handled := resolveMarketplaceEntitlementBillingSession(relayInfo, 400)
	if handled {
		t.Fatalf("expected marketplace entitlement resolver to skip when no entitlement matches, got session=%+v err=%v", session, apiErr)
	}
	if session != nil || apiErr != nil {
		t.Fatalf("expected no session and no error when marketplace entitlement does not match, got session=%+v err=%v", session, apiErr)
	}
	if relayInfo.BillingSource != "" || relayInfo.EntitlementLotId != 0 || relayInfo.OrderId != 0 {
		t.Fatalf("expected relayInfo to remain untouched when marketplace entitlement does not match, got %+v", relayInfo)
	}
}

func TestMarketplaceBillingSessionSettleConsumesLotAndSupplyWithoutTouchingUserQuota(t *testing.T) {
	db := setupMarketplaceServiceTestDB(t)
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())

	fixture := seedMarketplaceEntitlementBillingFixture(t, db, "market-buyer-billing-settle", 2000)
	fixture.buyer.Quota = 888888
	if err := db.Save(fixture.buyer).Error; err != nil {
		t.Fatalf("failed to seed buyer quota: %v", err)
	}

	relayInfo := &relaycommon.RelayInfo{
		UserId:          fixture.buyer.Id,
		OriginModelName: fixture.supply.ModelName,
		RequestId:       "req-marketplace-settle-001",
		TokenId:         1003,
		TokenKey:        "market-token-settle",
		RelayMode:       relayconstant.RelayModeChatCompletions,
		IsStream:        true,
		ChannelMeta: &relaycommon.ChannelMeta{
			ChannelId: fixture.channel.Id,
		},
		UserSetting: dto.UserSetting{
			BillingPreference: "subscription_first",
		},
	}

	session, apiErr := NewBillingSession(ctx, relayInfo, 500)
	if apiErr != nil {
		t.Fatalf("expected marketplace entitlement billing session, got error: %v", apiErr)
	}
	if err := session.Settle(300); err != nil {
		t.Fatalf("expected marketplace entitlement settle success, got error: %v", err)
	}

	reloadedBuyer, err := model.GetUserById(fixture.buyer.Id, true)
	if err != nil {
		t.Fatalf("failed to reload buyer: %v", err)
	}
	if reloadedBuyer.Quota != 888888 {
		t.Fatalf("expected buyer wallet quota untouched, got %+v", reloadedBuyer)
	}

	lot, err := getEntitlementLotByID(fixture.lot.Id)
	if err != nil {
		t.Fatalf("failed to reload entitlement lot: %v", err)
	}
	if lot.FrozenAmount != 0 || lot.UsedAmount != 300 {
		t.Fatalf("expected lot frozen released and used quota recorded, got %+v", lot)
	}

	entitlement, err := getBuyerEntitlementByID(fixture.entitlement.Id)
	if err != nil {
		t.Fatalf("failed to reload buyer entitlement: %v", err)
	}
	if entitlement.TotalFrozen != 0 || entitlement.TotalUsed != 300 {
		t.Fatalf("expected buyer entitlement totals updated, got %+v", entitlement)
	}

	reloadedSupply, err := model.GetSupplyAccountByID(fixture.supply.Id)
	if err != nil {
		t.Fatalf("failed to reload supply account: %v", err)
	}
	if reloadedSupply.ReservedCapacity != 1700 || reloadedSupply.UsedCapacity != 300 {
		t.Fatalf("expected supply capacity to move from reserved to used, got %+v", reloadedSupply)
	}

	snapshot, err := model.GetInventorySnapshotBySupplyAccountID(fixture.supply.Id)
	if err != nil {
		t.Fatalf("failed to reload inventory snapshot: %v", err)
	}
	if snapshot.ConsumedAmount != 300 {
		t.Fatalf("expected inventory snapshot consumed amount 300, got %+v", snapshot)
	}

	var ledger model.UsageLedger
	if err := db.Where("request_id = ? AND ledger_status = ?", relayInfo.RequestId, "success").First(&ledger).Error; err != nil {
		t.Fatalf("expected marketplace success usage ledger, got error: %v", err)
	}
	if ledger.BillingSource != BillingSourceMarketplaceEntitlement || ledger.EntitlementLotId != fixture.lot.Id || ledger.ActualQuota != 300 {
		t.Fatalf("unexpected usage ledger after settle: %+v", ledger)
	}
}

func TestPostTextConsumeQuotaMarketplaceBillingDoesNotIncreaseUserUsedQuota(t *testing.T) {
	db := setupMarketplaceServiceTestDB(t)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)
	ctx.Set("token_name", "marketplace-token")

	fixture := seedMarketplaceEntitlementBillingFixture(t, db, "market-buyer-post-text", 2000)
	fixture.buyer.Quota = 666666
	fixture.buyer.UsedQuota = 4321
	fixture.buyer.RequestCount = 0
	if err := db.Save(fixture.buyer).Error; err != nil {
		t.Fatalf("failed to seed buyer quota usage: %v", err)
	}

	startedAt := time.Now()
	relayInfo := &relaycommon.RelayInfo{
		UserId:            fixture.buyer.Id,
		OriginModelName:   fixture.supply.ModelName,
		RequestId:         "req-marketplace-post-text-001",
		TokenId:           1004,
		TokenKey:          "market-token-post-text",
		RelayMode:         relayconstant.RelayModeChatCompletions,
		StartTime:         startedAt,
		FirstResponseTime: startedAt,
		ChannelMeta: &relaycommon.ChannelMeta{
			ChannelId: fixture.channel.Id,
		},
		PriceData: types.PriceData{
			ModelRatio:      1,
			CompletionRatio: 1,
			CacheRatio:      1,
			UsePrice:        false,
			GroupRatioInfo:  types.GroupRatioInfo{GroupRatio: 1, GroupSpecialRatio: 1},
		},
		UserSetting: dto.UserSetting{
			BillingPreference: "wallet_first",
		},
	}

	session, apiErr := NewBillingSession(ctx, relayInfo, 300)
	if apiErr != nil {
		t.Fatalf("expected marketplace entitlement billing session, got error: %v", apiErr)
	}
	relayInfo.Billing = session

	PostTextConsumeQuota(ctx, relayInfo, &dto.Usage{
		PromptTokens:     300,
		CompletionTokens: 0,
		TotalTokens:      300,
		PromptTokensDetails: dto.InputTokenDetails{
			TextTokens: 300,
		},
	}, nil)

	reloadedBuyer, err := model.GetUserById(fixture.buyer.Id, true)
	if err != nil {
		t.Fatalf("failed to reload buyer: %v", err)
	}
	if reloadedBuyer.UsedQuota != 4321 {
		t.Fatalf("expected marketplace post consume not to change wallet used_quota, got %+v", reloadedBuyer)
	}
	if reloadedBuyer.RequestCount != 1 {
		t.Fatalf("expected marketplace post consume to increment request count only, got %+v", reloadedBuyer)
	}

	lot, err := getEntitlementLotByID(fixture.lot.Id)
	if err != nil {
		t.Fatalf("failed to reload entitlement lot: %v", err)
	}
	if lot.UsedAmount != 300 || lot.FrozenAmount != 0 {
		t.Fatalf("expected entitlement lot to be settled from post text consume, got %+v", lot)
	}
}

func TestMarketplaceBillingSessionRefundReleasesFrozenLotAndWritesFailedUsageLedger(t *testing.T) {
	db := setupMarketplaceServiceTestDB(t)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)
	ctx.Set("token_name", "marketplace-refund-token")

	fixture := seedMarketplaceEntitlementBillingFixture(t, db, "market-buyer-billing-refund", 2000)

	relayInfo := &relaycommon.RelayInfo{
		UserId:          fixture.buyer.Id,
		OriginModelName: fixture.supply.ModelName,
		RequestId:       "req-marketplace-refund-001",
		TokenId:         1005,
		TokenKey:        "market-token-refund",
		RelayMode:       relayconstant.RelayModeChatCompletions,
		ChannelMeta: &relaycommon.ChannelMeta{
			ChannelId: fixture.channel.Id,
		},
		UserSetting: dto.UserSetting{
			BillingPreference: "wallet_first",
		},
	}

	session, apiErr := NewBillingSession(ctx, relayInfo, 400)
	if apiErr != nil {
		t.Fatalf("expected marketplace entitlement billing session for refund flow, got error: %v", apiErr)
	}
	if !session.NeedsRefund() {
		t.Fatalf("expected marketplace billing session to require refund before relay failure")
	}

	session.Refund(ctx)

	var (
		lot         *model.EntitlementLot
		entitlement *model.BuyerEntitlement
		ledger      model.UsageLedger
	)
	deadline := time.Now().Add(3 * time.Second)
	for {
		var err error
		lot, err = getEntitlementLotByID(fixture.lot.Id)
		if err != nil {
			t.Fatalf("failed to reload entitlement lot during refund wait: %v", err)
		}
		entitlement, err = getBuyerEntitlementByID(fixture.entitlement.Id)
		if err != nil {
			t.Fatalf("failed to reload entitlement during refund wait: %v", err)
		}
		ledgerErr := db.Where("request_id = ? AND ledger_status = ?", relayInfo.RequestId, "failed").First(&ledger).Error
		if lot.FrozenAmount == 0 && entitlement.TotalFrozen == 0 && ledgerErr == nil {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("refund did not finish in time, lot=%+v entitlement=%+v ledgerErr=%v", lot, entitlement, ledgerErr)
		}
		time.Sleep(20 * time.Millisecond)
	}

	if ledger.BillingSource != BillingSourceMarketplaceEntitlement || ledger.EntitlementLotId != fixture.lot.Id {
		t.Fatalf("unexpected refund usage ledger payload: %+v", ledger)
	}
	if ledger.ActualQuota != 0 || ledger.PreConsumedQuota != 400 {
		t.Fatalf("expected refund ledger to keep pre-consumed quota and zero actual quota, got %+v", ledger)
	}
	if ledger.OrderId != fixture.order.Id || ledger.SupplyAccountId != fixture.supply.Id || ledger.ChannelId != fixture.channel.Id {
		t.Fatalf("unexpected refund ledger routing metadata: %+v", ledger)
	}
}

type marketplaceEntitlementBillingFixture struct {
	buyer       *model.User
	seller      *model.SellerProfile
	supply      *model.SupplyAccount
	channel     *model.Channel
	listing     *model.Listing
	sku         *model.ListingSKU
	order       *model.MarketOrder
	entitlement *model.BuyerEntitlement
	lot         *model.EntitlementLot
}

func seedMarketplaceEntitlementBillingFixture(t *testing.T, db *gorm.DB, buyerName string, packageAmount int64) marketplaceEntitlementBillingFixture {
	t.Helper()

	buyer := seedMarketplaceServiceUser(t, db, buyerName)
	sellerUser := seedMarketplaceServiceUser(t, db, buyerName+"-seller")
	seller, supply := seedMarketplaceServiceSupply(t, db, sellerUser, "active", "success", "token")
	channel, _ := seedMarketplaceChannelBinding(t, db, supply, "runtime-key")
	_ = seedSellerSecretRecord(t, db, seller.Id, supply.Id, `{"alg":"aes-256-gcm","kid":"v1","nonce":"bm9uY2U=","ciphertext":"Y2lwaGVy"}`, "fp-"+buyerName, "active", "success")
	listing, sku := seedMarketplaceListing(t, db, seller, supply, packageAmount, 299)

	order, _, err := CreateMarketOrder(CreateMarketOrderInput{
		BuyerUserID:    buyer.Id,
		ListingID:      listing.Id,
		SkuID:          sku.Id,
		Quantity:       1,
		IdempotencyKey: "order-" + buyerName,
	})
	if err != nil {
		t.Fatalf("create market order returned error: %v", err)
	}
	if _, err := CompleteMarketOrderPayment(CompleteMarketOrderPaymentInput{
		OrderNo:            order.OrderNo,
		PaymentMethod:      "stripe",
		PaymentTradeNo:     "pi_" + buyerName,
		Currency:           order.Currency,
		PayableAmountMinor: order.PayableAmountMinor,
		ProviderPayload:    `{"type":"checkout.session.completed"}`,
	}); err != nil {
		t.Fatalf("complete market order payment returned error: %v", err)
	}

	entitlements, total, err := ListBuyerEntitlements(buyer.Id, supply.ModelName, 0, 10)
	if err != nil {
		t.Fatalf("list buyer entitlements returned error: %v", err)
	}
	if total != 1 || len(entitlements) != 1 {
		t.Fatalf("expected one buyer entitlement, got total=%d len=%d", total, len(entitlements))
	}

	var lot model.EntitlementLot
	if err := db.Where("order_id = ?", order.Id).First(&lot).Error; err != nil {
		t.Fatalf("failed to load entitlement lot: %v", err)
	}

	return marketplaceEntitlementBillingFixture{
		buyer:       buyer,
		seller:      seller,
		supply:      supply,
		channel:     channel,
		listing:     listing,
		sku:         sku,
		order:       order,
		entitlement: entitlements[0],
		lot:         &lot,
	}
}

func getBuyerEntitlementByID(id int) (*model.BuyerEntitlement, error) {
	var entitlement model.BuyerEntitlement
	if err := model.DB.First(&entitlement, id).Error; err != nil {
		return nil, err
	}
	return &entitlement, nil
}

func getEntitlementLotByID(id int) (*model.EntitlementLot, error) {
	var lot model.EntitlementLot
	if err := model.DB.First(&lot, id).Error; err != nil {
		return nil, err
	}
	return &lot, nil
}
