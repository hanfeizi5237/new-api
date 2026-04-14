package controller

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/dto"
	"github.com/QuantumNous/new-api/model"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	relayconstant "github.com/QuantumNous/new-api/relay/constant"
	"github.com/QuantumNous/new-api/service"
	"github.com/QuantumNous/new-api/types"
	"github.com/gin-gonic/gin"
)

func TestGetChannelUsesMarketplaceEntitlementChannelBeforeContextFallback(t *testing.T) {
	db := setupMarketplaceControllerTestDB(t)

	buyer := seedMarketplaceUser(t, db, "relay-marketplace-buyer")
	sellerUser := seedMarketplaceUser(t, db, "relay-marketplace-seller")
	_, supply := seedMarketplaceSellerWithSupply(t, db, sellerUser.Id)
	entitlementChannel, _ := seedMarketplaceChannelBinding(t, db, supply, "market-runtime-key")

	otherChannel := &model.Channel{
		Name:   "non-marketplace-context-channel",
		Key:    "context-runtime-key",
		Status: common.ChannelStatusEnabled,
		Models: supply.ModelName,
		Group:  "default",
	}
	if err := db.Create(otherChannel).Error; err != nil {
		t.Fatalf("failed to create context fallback channel: %v", err)
	}

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest("POST", "/v1/chat/completions", nil)
	ctx.Set("channel_id", otherChannel.Id)
	ctx.Set("channel_type", otherChannel.Type)
	ctx.Set("channel_name", otherChannel.Name)
	ctx.Set("auto_ban", true)
	common.SetContextKey(ctx, constant.ContextKeyUsingGroup, "default")

	retry := 0
	channel, apiErr := getChannel(ctx, &relaycommon.RelayInfo{
		UserId:                buyer.Id,
		OriginModelName:       supply.ModelName,
		BillingSource:         service.BillingSourceMarketplaceEntitlement,
		SupplyAccountId:       supply.Id,
		EntitlementChannelIDs: []int{entitlementChannel.Id},
	}, &service.RetryParam{
		Ctx:        ctx,
		TokenGroup: "default",
		ModelName:  supply.ModelName,
		Retry:      &retry,
	})
	if apiErr != nil {
		t.Fatalf("expected marketplace channel selection success, got error: %v", apiErr)
	}
	if channel == nil {
		t.Fatalf("expected marketplace channel")
	}
	if channel.Id != entitlementChannel.Id {
		t.Fatalf("expected marketplace entitlement channel %d, got %d", entitlementChannel.Id, channel.Id)
	}
	if got := common.GetContextKeyInt(ctx, constant.ContextKeyChannelId); got != entitlementChannel.Id {
		t.Fatalf("expected context channel id switched to entitlement channel %d, got %d", entitlementChannel.Id, got)
	}
}

func TestMarketplacePaidRelayRequestUsesEntitlementChannelAndWritesUsageLedger(t *testing.T) {
	db := setupMarketplaceControllerTestDB(t)

	order, supply := seedMarketplacePayableOrder(t, db, "relay-marketplace-paid-buyer", "relay-marketplace-paid-seller")
	var buyer model.User
	if err := db.First(&buyer, order.BuyerUserId).Error; err != nil {
		t.Fatalf("failed to load buyer: %v", err)
	}
	buyer.Quota = 999999
	if err := db.Save(&buyer).Error; err != nil {
		t.Fatalf("failed to update buyer quota: %v", err)
	}

	paidOrder, err := service.CompleteMarketOrderPayment(service.CompleteMarketOrderPaymentInput{
		OrderNo:            order.OrderNo,
		PaymentMethod:      "stripe",
		PaymentTradeNo:     "pi_relay_marketplace_paid_001",
		Currency:           order.Currency,
		PayableAmountMinor: order.PayableAmountMinor,
		ProviderPayload:    `{"type":"checkout.session.completed"}`,
	})
	if err != nil {
		t.Fatalf("failed to complete marketplace payment: %v", err)
	}

	var binding model.SupplyChannelBinding
	if err := db.Where("supply_account_id = ? AND status = ?", supply.Id, "active").First(&binding).Error; err != nil {
		t.Fatalf("failed to load supply channel binding: %v", err)
	}
	entitlementChannel, err := model.GetChannelById(binding.ChannelId, true)
	if err != nil {
		t.Fatalf("failed to load entitlement channel: %v", err)
	}

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)
	ctx.Set("token_name", "marketplace-integration-token")
	ctx.Set(common.RequestIdKey, "req-marketplace-controller-integration-001")
	common.SetContextKey(ctx, constant.ContextKeyUsingGroup, "default")

	startedAt := time.Now()
	relayInfo := &relaycommon.RelayInfo{
		UserId:            buyer.Id,
		UserGroup:         "default",
		UsingGroup:        "default",
		OriginModelName:   supply.ModelName,
		RequestId:         "req-marketplace-controller-integration-001",
		TokenId:           2001,
		TokenKey:          "marketplace-integration-token-key",
		RelayMode:         relayconstant.RelayModeChatCompletions,
		StartTime:         startedAt,
		FirstResponseTime: startedAt,
		PriceData: types.PriceData{
			ModelRatio:      1,
			CompletionRatio: 1,
			CacheRatio:      1,
			UsePrice:        false,
			GroupRatioInfo: types.GroupRatioInfo{
				GroupRatio:        1,
				GroupSpecialRatio: 1,
			},
		},
		UserSetting: dto.UserSetting{
			BillingPreference: "wallet_first",
		},
	}

	session, apiErr := service.NewBillingSession(ctx, relayInfo, 400)
	if apiErr != nil {
		t.Fatalf("expected marketplace billing session after payment, got error: %v", apiErr)
	}
	relayInfo.Billing = session

	retry := 0
	selectedChannel, channelErr := getChannel(ctx, relayInfo, &service.RetryParam{
		Ctx:        ctx,
		TokenGroup: "default",
		ModelName:  supply.ModelName,
		Retry:      &retry,
	})
	if channelErr != nil {
		t.Fatalf("expected entitlement channel selection success, got error: %v", channelErr)
	}
	if selectedChannel.Id != entitlementChannel.Id {
		t.Fatalf("expected paid relay request to use entitlement channel %d, got %d", entitlementChannel.Id, selectedChannel.Id)
	}

	relayInfo.InitChannelMeta(ctx)

	service.PostTextConsumeQuota(ctx, relayInfo, &dto.Usage{
		PromptTokens:     400,
		CompletionTokens: 0,
		TotalTokens:      400,
		PromptTokensDetails: dto.InputTokenDetails{
			TextTokens: 400,
		},
	}, nil)

	var ledger model.UsageLedger
	if err := db.Where("request_id = ? AND ledger_status = ?", relayInfo.RequestId, "success").First(&ledger).Error; err != nil {
		t.Fatalf("expected marketplace usage ledger after relay consume, got error: %v", err)
	}
	if ledger.ChannelId != entitlementChannel.Id {
		t.Fatalf("expected usage ledger channel_id=%d, got %+v", entitlementChannel.Id, ledger)
	}
	if ledger.OrderId != paidOrder.Id || ledger.SupplyAccountId != supply.Id || ledger.BillingSource != service.BillingSourceMarketplaceEntitlement {
		t.Fatalf("unexpected marketplace usage ledger payload: %+v", ledger)
	}
}
