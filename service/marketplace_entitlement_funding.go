package service

import (
	"errors"
	"fmt"
	"sort"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	relayconstant "github.com/QuantumNous/new-api/relay/constant"
	"github.com/QuantumNous/new-api/types"
	"gorm.io/gorm"
)

var (
	errNoMarketplaceEntitlement           = errors.New("no marketplace entitlement matched")
	errInsufficientMarketplaceEntitlement = errors.New("matching marketplace entitlement is insufficient")
)

type MarketplaceEntitlementFunding struct {
	requestId        string
	userId           int
	tokenId          int
	modelName        string
	relayMode        int
	relayInfo        *relaycommon.RelayInfo
	preConsumed      int
	entitlementId    int
	entitlementLotId int
	sellerId         int
	supplyAccountId  int
	listingId        int
	orderId          int
	orderItemId      int
	channelIDs       []int
}

func (m *MarketplaceEntitlementFunding) Source() string { return BillingSourceMarketplaceEntitlement }

func (m *MarketplaceEntitlementFunding) PreConsume(amount int) error {
	if amount <= 0 {
		return errNoMarketplaceEntitlement
	}
	return model.DB.Transaction(func(tx *gorm.DB) error {
		now := common.GetTimestamp()
		// Lots are ordered by expiry/priority so relay spends the earliest valid entitlement first.
		lots, err := listCandidateMarketplaceLotsTx(tx, m.userId, m.modelName, now)
		if err != nil {
			return err
		}
		matchingLotExists := false
		for _, lot := range lots {
			channelIDs, err := resolveMarketplaceLotChannelIDsTx(tx, lot.SupplyAccountId, m.modelName)
			if err != nil {
				return err
			}
			if len(channelIDs) == 0 {
				continue
			}
			matchingLotExists = true
			remainingAmount := lot.GrantedAmount - lot.UsedAmount - lot.RefundedAmount - lot.FrozenAmount
			if remainingAmount < int64(amount) {
				continue
			}

			var entitlement model.BuyerEntitlement
			if err := lockForUpdate(tx).First(&entitlement, lot.BuyerEntitlementId).Error; err != nil {
				return err
			}
			if err := tx.Model(&model.EntitlementLot{}).
				Where("id = ?", lot.Id).
				Updates(map[string]interface{}{
					"frozen_amount": lot.FrozenAmount + int64(amount),
					"updated_at":    common.GetTimestamp(),
				}).Error; err != nil {
				return err
			}
			if err := tx.Model(&model.BuyerEntitlement{}).
				Where("id = ?", entitlement.Id).
				Updates(map[string]interface{}{
					"total_frozen": entitlement.TotalFrozen + int64(amount),
					"updated_at":   common.GetTimestamp(),
				}).Error; err != nil {
				return err
			}

			m.preConsumed = amount
			m.entitlementId = entitlement.Id
			m.entitlementLotId = lot.Id
			m.sellerId = lot.SellerId
			m.supplyAccountId = lot.SupplyAccountId
			m.listingId = lot.ListingId
			m.orderId = lot.OrderId
			m.orderItemId = lot.OrderItemId
			m.channelIDs = channelIDs
			if m.relayInfo != nil {
				m.relayInfo.BillingSource = BillingSourceMarketplaceEntitlement
				m.relayInfo.EntitlementLotId = lot.Id
				m.relayInfo.SellerId = lot.SellerId
				m.relayInfo.SupplyAccountId = lot.SupplyAccountId
				m.relayInfo.ListingId = lot.ListingId
				m.relayInfo.OrderId = lot.OrderId
				m.relayInfo.OrderItemId = lot.OrderItemId
				m.relayInfo.EntitlementChannelIDs = append([]int(nil), channelIDs...)
				m.relayInfo.FinalPreConsumedQuota = amount
			}
			return nil
		}
		if matchingLotExists {
			return errInsufficientMarketplaceEntitlement
		}
		return errNoMarketplaceEntitlement
	})
}

func (m *MarketplaceEntitlementFunding) Settle(delta int) error {
	actualQuota := m.preConsumed + delta
	if actualQuota < 0 {
		return fmt.Errorf("invalid marketplace entitlement settle delta: %d", delta)
	}
	return model.DB.Transaction(func(tx *gorm.DB) error {
		var lot model.EntitlementLot
		if err := lockForUpdate(tx).First(&lot, m.entitlementLotId).Error; err != nil {
			return err
		}
		var entitlement model.BuyerEntitlement
		if err := lockForUpdate(tx).First(&entitlement, lot.BuyerEntitlementId).Error; err != nil {
			return err
		}
		var supply model.SupplyAccount
		if err := lockForUpdate(tx).First(&supply, m.supplyAccountId).Error; err != nil {
			return err
		}
		snapshot, err := loadOrCreateInventorySnapshotTx(tx, &supply)
		if err != nil {
			return err
		}

		frozenRelease := minInt64(lot.FrozenAmount, int64(m.preConsumed))
		additionalNeeded := actualQuota - m.preConsumed
		if additionalNeeded > 0 {
			remainingAvailable := lot.GrantedAmount - lot.UsedAmount - lot.RefundedAmount - lot.FrozenAmount
			if remainingAvailable < int64(additionalNeeded) {
				return errInsufficientMarketplaceEntitlement
			}
		}

		lot.FrozenAmount -= frozenRelease
		entitlement.TotalFrozen -= frozenRelease
		if entitlement.TotalFrozen < 0 {
			entitlement.TotalFrozen = 0
		}
		if actualQuota > 0 {
			lot.UsedAmount += int64(actualQuota)
			entitlement.TotalUsed += int64(actualQuota)
			supply.ReservedCapacity -= int64(actualQuota)
			if supply.ReservedCapacity < 0 {
				supply.ReservedCapacity = 0
			}
			supply.UsedCapacity += int64(actualQuota)
			snapshot.ConsumedAmount += int64(actualQuota)
		}
		snapshot.AvailableAmount = recomputeInventoryAvailableAmount(&supply, snapshot)
		snapshot.LastSyncAt = common.GetTimestamp()

		if err := tx.Save(&lot).Error; err != nil {
			return err
		}
		if err := tx.Save(&entitlement).Error; err != nil {
			return err
		}
		if err := tx.Save(&supply).Error; err != nil {
			return err
		}
		if err := tx.Save(snapshot).Error; err != nil {
			return err
		}
		// Success and refund flows both upsert by event key so webhook retries do not duplicate usage ledgers.
		return upsertMarketplaceUsageLedgerTx(tx, marketplaceUsageLedgerInput{
			EventKey:         m.requestId + ":success",
			RequestID:        m.requestId,
			LedgerStatus:     "success",
			BillingSource:    BillingSourceMarketplaceEntitlement,
			UserID:           m.userId,
			TokenID:          m.tokenId,
			ChannelID:        marketplaceRelayChannelID(m.relayInfo),
			SellerID:         m.sellerId,
			SupplyAccountID:  m.supplyAccountId,
			ListingID:        m.listingId,
			OrderID:          m.orderId,
			OrderItemID:      m.orderItemId,
			EntitlementLotID: m.entitlementLotId,
			ModelName:        m.modelName,
			EndpointType:     marketplaceEndpointType(m.relayMode),
			PreConsumedQuota: m.preConsumed,
			ActualQuota:      actualQuota,
			RetryIndex:       marketplaceRelayRetryIndex(m.relayInfo),
			IsStream:         marketplaceRelayIsStream(m.relayInfo),
			StartedAt:        marketplaceRelayStartedAt(m.relayInfo),
			FinishedAt:       common.GetTimestamp(),
		})
	})
}

func (m *MarketplaceEntitlementFunding) Refund() error {
	if m.preConsumed <= 0 {
		return nil
	}
	return model.DB.Transaction(func(tx *gorm.DB) error {
		var lot model.EntitlementLot
		if err := lockForUpdate(tx).First(&lot, m.entitlementLotId).Error; err != nil {
			return err
		}
		var entitlement model.BuyerEntitlement
		if err := lockForUpdate(tx).First(&entitlement, lot.BuyerEntitlementId).Error; err != nil {
			return err
		}
		frozenRelease := minInt64(lot.FrozenAmount, int64(m.preConsumed))
		lot.FrozenAmount -= frozenRelease
		entitlement.TotalFrozen -= frozenRelease
		if entitlement.TotalFrozen < 0 {
			entitlement.TotalFrozen = 0
		}
		if err := tx.Save(&lot).Error; err != nil {
			return err
		}
		if err := tx.Save(&entitlement).Error; err != nil {
			return err
		}
		errorCode := ""
		if m.relayInfo != nil && m.relayInfo.LastError != nil {
			errorCode = string(m.relayInfo.LastError.GetErrorCode())
		}
		// Refund only releases the pre-frozen entitlement window; no actual quota is consumed on failed relay attempts.
		return upsertMarketplaceUsageLedgerTx(tx, marketplaceUsageLedgerInput{
			EventKey:         m.requestId + ":refund_final",
			RequestID:        m.requestId,
			LedgerStatus:     "failed",
			BillingSource:    BillingSourceMarketplaceEntitlement,
			UserID:           m.userId,
			TokenID:          m.tokenId,
			ChannelID:        marketplaceRelayChannelID(m.relayInfo),
			SellerID:         m.sellerId,
			SupplyAccountID:  m.supplyAccountId,
			ListingID:        m.listingId,
			OrderID:          m.orderId,
			OrderItemID:      m.orderItemId,
			EntitlementLotID: m.entitlementLotId,
			ModelName:        m.modelName,
			EndpointType:     marketplaceEndpointType(m.relayMode),
			PreConsumedQuota: m.preConsumed,
			ActualQuota:      0,
			RetryIndex:       marketplaceRelayRetryIndex(m.relayInfo),
			LedgerErrorCode:  errorCode,
			IsStream:         marketplaceRelayIsStream(m.relayInfo),
			StartedAt:        marketplaceRelayStartedAt(m.relayInfo),
			FinishedAt:       common.GetTimestamp(),
		})
	})
}

type marketplaceUsageLedgerInput struct {
	EventKey         string
	RequestID        string
	LedgerStatus     string
	BillingSource    string
	UserID           int
	TokenID          int
	ChannelID        int
	SellerID         int
	SupplyAccountID  int
	ListingID        int
	OrderID          int
	OrderItemID      int
	EntitlementLotID int
	ModelName        string
	EndpointType     string
	PreConsumedQuota int
	ActualQuota      int
	RetryIndex       int
	LedgerErrorCode  string
	IsStream         bool
	StartedAt        int64
	FinishedAt       int64
}

func upsertMarketplaceUsageLedgerTx(tx *gorm.DB, input marketplaceUsageLedgerInput) error {
	var existing model.UsageLedger
	err := tx.Where("event_key = ?", input.EventKey).First(&existing).Error
	if err == nil {
		return tx.Model(&model.UsageLedger{}).
			Where("id = ?", existing.Id).
			Updates(map[string]interface{}{
				"request_id":         input.RequestID,
				"billing_source":     input.BillingSource,
				"user_id":            input.UserID,
				"token_id":           input.TokenID,
				"channel_id":         input.ChannelID,
				"seller_id":          input.SellerID,
				"supply_account_id":  input.SupplyAccountID,
				"listing_id":         input.ListingID,
				"order_id":           input.OrderID,
				"order_item_id":      input.OrderItemID,
				"entitlement_lot_id": input.EntitlementLotID,
				"model_name":         input.ModelName,
				"endpoint_type":      input.EndpointType,
				"pre_consumed_quota": input.PreConsumedQuota,
				"actual_quota":       input.ActualQuota,
				"retry_index":        input.RetryIndex,
				"ledger_status":      input.LedgerStatus,
				"is_stream":          input.IsStream,
				"started_at":         input.StartedAt,
				"finished_at":        input.FinishedAt,
				"error_code":         input.LedgerErrorCode,
				"updated_at":         common.GetTimestamp(),
			}).Error
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	ledger := model.UsageLedger{
		EventKey:         input.EventKey,
		RequestId:        input.RequestID,
		BillingSource:    input.BillingSource,
		UserId:           input.UserID,
		TokenId:          input.TokenID,
		ChannelId:        input.ChannelID,
		SellerId:         input.SellerID,
		SupplyAccountId:  input.SupplyAccountID,
		ListingId:        input.ListingID,
		OrderId:          input.OrderID,
		OrderItemId:      input.OrderItemID,
		EntitlementLotId: input.EntitlementLotID,
		ModelName:        input.ModelName,
		EndpointType:     input.EndpointType,
		PreConsumedQuota: input.PreConsumedQuota,
		ActualQuota:      input.ActualQuota,
		RetryIndex:       input.RetryIndex,
		LedgerStatus:     input.LedgerStatus,
		IsStream:         input.IsStream,
		StartedAt:        input.StartedAt,
		FinishedAt:       input.FinishedAt,
		ErrorCode:        input.LedgerErrorCode,
	}
	return tx.Create(&ledger).Error
}

func resolveMarketplaceEntitlementBillingSession(relayInfo *relaycommon.RelayInfo, preConsumedQuota int) (*BillingSession, *types.NewAPIError, bool) {
	if !supportsMarketplaceEntitlementRelay(relayInfo, preConsumedQuota) {
		return nil, nil, false
	}
	funding := &MarketplaceEntitlementFunding{
		requestId: relayInfo.RequestId,
		userId:    relayInfo.UserId,
		tokenId:   relayInfo.TokenId,
		modelName: relayInfo.OriginModelName,
		relayMode: relayInfo.RelayMode,
		relayInfo: relayInfo,
	}
	if err := funding.PreConsume(preConsumedQuota); err != nil {
		if errors.Is(err, errNoMarketplaceEntitlement) {
			return nil, nil, false
		}
		if errors.Is(err, errInsufficientMarketplaceEntitlement) {
			return nil, types.NewErrorWithStatusCode(
				fmt.Errorf("marketplace entitlement is insufficient for model %s", relayInfo.OriginModelName),
				types.ErrorCodeInsufficientMarketplaceEntitlement,
				403,
				types.ErrOptionWithSkipRetry(),
				types.ErrOptionWithNoRecordErrorLog(),
			), true
		}
		return nil, types.NewError(err, types.ErrorCodeUpdateDataError, types.ErrOptionWithSkipRetry()), true
	}
	session := &BillingSession{
		relayInfo:        relayInfo,
		funding:          funding,
		preConsumedQuota: preConsumedQuota,
	}
	session.syncRelayInfo()
	return session, nil, true
}

func supportsMarketplaceEntitlementRelay(relayInfo *relaycommon.RelayInfo, preConsumedQuota int) bool {
	if relayInfo == nil || relayInfo.UserId <= 0 || preConsumedQuota <= 0 {
		return false
	}
	switch relayInfo.RelayMode {
	case relayconstant.RelayModeChatCompletions, relayconstant.RelayModeCompletions, relayconstant.RelayModeResponses, relayconstant.RelayModeResponsesCompact:
		return relayInfo.OriginModelName != ""
	default:
		return false
	}
}

func listCandidateMarketplaceLotsTx(tx *gorm.DB, userID int, modelName string, now int64) ([]model.EntitlementLot, error) {
	var lots []model.EntitlementLot
	if err := lockForUpdate(tx).
		Where("buyer_user_id = ? AND status = ?", userID, "active").
		Where("expire_at = 0 OR expire_at > ?", now).
		Find(&lots).Error; err != nil {
		return nil, err
	}

	filtered := make([]model.EntitlementLot, 0, len(lots))
	for _, lot := range lots {
		// Every candidate must still satisfy the full marketplace responsibility chain:
		// entitlement -> listing -> supply -> active verified secret -> ready channel binding.
		entitlement, err := getBuyerEntitlementByIDTx(tx, lot.BuyerEntitlementId)
		if err != nil {
			return nil, err
		}
		if entitlement.Status != "active" || entitlement.ModelName != modelName {
			continue
		}
		listing, err := model.GetListingByID(lot.ListingId)
		if err != nil {
			return nil, err
		}
		if listing.Status != "active" || listing.AuditStatus != "approved" {
			continue
		}
		supply, err := model.GetSupplyAccountByID(lot.SupplyAccountId)
		if err != nil {
			return nil, err
		}
		if supply.Status != "active" || supply.VerifyStatus != "success" {
			continue
		}
		var activeSecretCount int64
		if err := tx.Model(&model.SellerSecret{}).
			Where("supply_account_id = ? AND status = ? AND verify_status = ?", supply.Id, "active", "success").
			Count(&activeSecretCount).Error; err != nil {
			return nil, err
		}
		if activeSecretCount == 0 {
			continue
		}
		filtered = append(filtered, lot)
	}

	sort.Slice(filtered, func(i, j int) bool {
		leftExpire := filtered[i].ExpireAt
		rightExpire := filtered[j].ExpireAt
		if leftExpire == 0 {
			leftExpire = maxInt64Value()
		}
		if rightExpire == 0 {
			rightExpire = maxInt64Value()
		}
		if leftExpire != rightExpire {
			return leftExpire < rightExpire
		}
		if filtered[i].PrioritySeq != filtered[j].PrioritySeq {
			return filtered[i].PrioritySeq < filtered[j].PrioritySeq
		}
		return filtered[i].Id < filtered[j].Id
	})
	return filtered, nil
}

func getBuyerEntitlementByIDTx(tx *gorm.DB, id int) (*model.BuyerEntitlement, error) {
	var entitlement model.BuyerEntitlement
	if err := tx.First(&entitlement, id).Error; err != nil {
		return nil, err
	}
	return &entitlement, nil
}

func resolveMarketplaceLotChannelIDsTx(tx *gorm.DB, supplyAccountID int, modelName string) ([]int, error) {
	var bindings []model.SupplyChannelBinding
	if err := tx.Where("supply_account_id = ? AND status = ?", supplyAccountID, "active").
		Order("CASE WHEN binding_role = 'primary' THEN 0 ELSE 1 END asc").
		Order("priority asc, id asc").
		Find(&bindings).Error; err != nil {
		return nil, err
	}
	channelIDs := make([]int, 0, len(bindings))
	for _, binding := range bindings {
		channel, err := model.GetChannelById(binding.ChannelId, true)
		if err != nil {
			return nil, err
		}
		if channel.Status != common.ChannelStatusEnabled {
			continue
		}
		if !containsExactModel(channel.GetModels(), modelName) {
			continue
		}
		channelIDs = append(channelIDs, channel.Id)
	}
	return channelIDs, nil
}

func GetMarketplaceEntitlementChannel(relayInfo *relaycommon.RelayInfo, retry int) (*model.Channel, error) {
	if relayInfo == nil || relayInfo.BillingSource != BillingSourceMarketplaceEntitlement {
		return nil, errNoMarketplaceEntitlement
	}
	if len(relayInfo.EntitlementChannelIDs) == 0 {
		return nil, fmt.Errorf("no marketplace entitlement channels available for supply %d", relayInfo.SupplyAccountId)
	}
	if retry < 0 {
		retry = 0
	}
	for index := retry; index < len(relayInfo.EntitlementChannelIDs); index++ {
		channel, err := model.CacheGetChannel(relayInfo.EntitlementChannelIDs[index])
		if err != nil {
			return nil, err
		}
		if channel.Status != common.ChannelStatusEnabled {
			continue
		}
		if !containsExactModel(channel.GetModels(), relayInfo.OriginModelName) {
			continue
		}
		return channel, nil
	}
	return nil, fmt.Errorf("marketplace entitlement channels exhausted for supply %d", relayInfo.SupplyAccountId)
}

func marketplaceEndpointType(relayMode int) string {
	switch relayMode {
	case relayconstant.RelayModeResponses, relayconstant.RelayModeResponsesCompact:
		return "responses"
	case relayconstant.RelayModeCompletions:
		return "completion"
	default:
		return "chat"
	}
}

func marketplaceRelayChannelID(relayInfo *relaycommon.RelayInfo) int {
	if relayInfo == nil || relayInfo.ChannelMeta == nil {
		return 0
	}
	return relayInfo.ChannelId
}

func marketplaceRelayRetryIndex(relayInfo *relaycommon.RelayInfo) int {
	if relayInfo == nil {
		return 0
	}
	return relayInfo.RetryIndex
}

func marketplaceRelayIsStream(relayInfo *relaycommon.RelayInfo) bool {
	return relayInfo != nil && relayInfo.IsStream
}

func marketplaceRelayStartedAt(relayInfo *relaycommon.RelayInfo) int64 {
	if relayInfo == nil || relayInfo.StartTime.IsZero() {
		return common.GetTimestamp()
	}
	return relayInfo.StartTime.Unix()
}

func minInt64(left int64, right int64) int64 {
	if left < right {
		return left
	}
	return right
}

func maxInt64Value() int64 {
	return int64(^uint64(0) >> 1)
}
