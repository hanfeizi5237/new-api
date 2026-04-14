package service

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/logger"
	"github.com/QuantumNous/new-api/model"
	"github.com/bytedance/gopkg/util/gopool"
	"gorm.io/gorm"
)

const (
	marketplaceInventorySyncInterval  = 5 * time.Minute
	marketplaceInventorySyncBatchSize = 200
)

var (
	marketplaceInventorySyncOnce    sync.Once
	marketplaceInventorySyncRunning atomic.Bool
)

func StartMarketplaceInventorySyncTask() {
	marketplaceInventorySyncOnce.Do(func() {
		if !common.IsMasterNode {
			return
		}
		if !common.GetEnvOrDefaultBool("MARKETPLACE_INVENTORY_SYNC_TASK_ENABLED", true) {
			common.SysLog("marketplace inventory sync task disabled by MARKETPLACE_INVENTORY_SYNC_TASK_ENABLED")
			return
		}
		gopool.Go(func() {
			logger.LogInfo(context.Background(), fmt.Sprintf("marketplace inventory sync task started: tick=%s", marketplaceInventorySyncInterval))
			runMarketplaceInventorySyncOnce()
			ticker := time.NewTicker(marketplaceInventorySyncInterval)
			defer ticker.Stop()
			for range ticker.C {
				runMarketplaceInventorySyncOnce()
			}
		})
	})
}

func runMarketplaceInventorySyncOnce() {
	if !marketplaceInventorySyncRunning.CompareAndSwap(false, true) {
		return
	}
	defer marketplaceInventorySyncRunning.Store(false)

	var lastID int
	for {
		var supplyIDs []int
		if err := model.DB.Model(&model.SupplyAccount{}).
			Where("id > ?", lastID).
			Order("id asc").
			Limit(marketplaceInventorySyncBatchSize).
			Pluck("id", &supplyIDs).Error; err != nil {
			logger.LogWarn(context.Background(), fmt.Sprintf("marketplace inventory sync query failed: %v", err))
			return
		}
		if len(supplyIDs) == 0 {
			return
		}
		for _, supplyID := range supplyIDs {
			if err := SyncMarketplaceInventoryBySupplyAccountID(supplyID, "scheduled_full_sync"); err != nil {
				logger.LogWarn(context.Background(), fmt.Sprintf("marketplace inventory sync failed for supply %d: %v", supplyID, err))
			}
		}
		lastID = supplyIDs[len(supplyIDs)-1]
		if len(supplyIDs) < marketplaceInventorySyncBatchSize {
			return
		}
	}
}

func SyncMarketplaceInventoryBySupplyAccountID(supplyAccountID int, reason string) error {
	if supplyAccountID <= 0 {
		return nil
	}
	return model.DB.Transaction(func(tx *gorm.DB) error {
		return syncMarketplaceInventoryBySupplyAccountIDTx(tx, supplyAccountID, reason)
	})
}

func syncMarketplaceInventoryBySupplyAccountIDTx(tx *gorm.DB, supplyAccountID int, reason string) error {
	if tx == nil || supplyAccountID <= 0 {
		return nil
	}

	var supply model.SupplyAccount
	if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&supply, supplyAccountID).Error; err != nil {
		return err
	}
	var seller model.SellerProfile
	if err := tx.First(&seller, supply.SellerId).Error; err != nil {
		return err
	}

	snapshot, err := loadOrCreateInventorySnapshotTx(tx, &supply)
	if err != nil {
		return err
	}
	if snapshot.RiskDiscountBps <= 0 {
		snapshot.RiskDiscountBps = 10000
	}
	if snapshot.HealthScore <= 0 {
		snapshot.HealthScore = 100
	}

	secretReady, bindingReady, syncMessage, syncStatus, err := evaluateSupplySyncHealthTx(tx, &supply, &seller)
	if err != nil {
		return err
	}
	snapshot.SyncStatus = syncStatus
	snapshot.SyncMessage = syncMessage
	snapshot.AvailableAmount = recomputeInventoryAvailableAmount(&supply, snapshot)
	snapshot.LastSyncAt = common.GetTimestamp()

	if err := tx.Save(snapshot).Error; err != nil {
		return err
	}

	var listings []model.Listing
	if err := tx.Where("supply_account_id = ?", supply.Id).Find(&listings).Error; err != nil {
		return err
	}
	for i := range listings {
		nextStatus, changed, statusReason, err := deriveSyncedListingStatusTx(tx, &listings[i], &supply, &seller, snapshot, secretReady, bindingReady)
		if err != nil {
			return err
		}
		if !changed {
			continue
		}
		if err := tx.Model(&model.Listing{}).
			Where("id = ?", listings[i].Id).
			Updates(map[string]interface{}{
				"status":       nextStatus,
				"audit_remark": statusReason,
				"updated_at":   common.GetTimestamp(),
			}).Error; err != nil {
			return err
		}
	}

	common.SysLog(fmt.Sprintf("marketplace inventory synced: supply_account_id=%d reason=%s status=%s available=%d", supply.Id, reason, snapshot.SyncStatus, snapshot.AvailableAmount))
	return nil
}

func loadOrCreateInventorySnapshotTx(tx *gorm.DB, supply *model.SupplyAccount) (*model.InventorySnapshot, error) {
	var snapshot model.InventorySnapshot
	err := tx.Set("gorm:query_option", "FOR UPDATE").
		Where("supply_account_id = ?", supply.Id).
		First(&snapshot).Error
	if err == nil {
		return &snapshot, nil
	}
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	snapshot = model.InventorySnapshot{
		SupplyAccountId: supply.Id,
		AvailableAmount: supply.SellableCapacity,
		RiskDiscountBps: 10000,
		HealthScore:     100,
		SyncStatus:      "ok",
	}
	if err := tx.Create(&snapshot).Error; err != nil {
		return nil, err
	}
	return &snapshot, nil
}

func evaluateSupplySyncHealthTx(tx *gorm.DB, supply *model.SupplyAccount, seller *model.SellerProfile) (bool, bool, string, string, error) {
	if seller == nil || supply == nil {
		return false, false, "invalid marketplace supply context", "error", nil
	}
	if seller.Status != "active" {
		return false, false, "seller is not active", "error", nil
	}
	if supply.Status != "active" {
		return false, false, "supply_account is not active", "error", nil
	}
	if supply.VerifyStatus != "success" {
		return false, false, "supply_account verify_status is not success", "error", nil
	}

	var activeSecretCount int64
	if err := tx.Model(&model.SellerSecret{}).
		Where("supply_account_id = ? AND status = ? AND verify_status = ?", supply.Id, "active", "success").
		Count(&activeSecretCount).Error; err != nil {
		return false, false, "", "", err
	}
	if activeSecretCount == 0 {
		return false, false, "no active verified seller secret", "error", nil
	}

	bindingReady, err := hasReadyMarketplaceChannelBindingTx(tx, supply)
	if err != nil {
		return true, false, "", "", err
	}
	if !bindingReady {
		return true, false, "no active ready supply channel binding", "error", nil
	}
	return true, true, "", "ok", nil
}

func deriveSyncedListingStatusTx(tx *gorm.DB, listing *model.Listing, supply *model.SupplyAccount, seller *model.SellerProfile, snapshot *model.InventorySnapshot, secretReady bool, bindingReady bool) (string, bool, string, error) {
	if listing == nil {
		return "", false, "", nil
	}
	if listing.AuditStatus != "approved" || listing.Status == "archived" {
		return listing.Status, false, "", nil
	}

	pauseReason := deriveMarketplaceListingPauseReason(listing, supply, seller, snapshot, secretReady, bindingReady)
	if pauseReason != "" {
		if listing.Status == "paused" {
			return listing.Status, false, "", nil
		}
		return "paused", true, pauseReason, nil
	}

	soldOut, err := listingShouldBeSoldOutTx(tx, listing, snapshot)
	if err != nil {
		return "", false, "", err
	}
	if soldOut {
		if listing.Status == "sold_out" {
			return listing.Status, false, "", nil
		}
		return "sold_out", true, "inventory exhausted or below minimum purchase threshold", nil
	}

	// M1 does not auto-reactivate listings after healthy inventory returns.
	return listing.Status, false, "", nil
}

func deriveMarketplaceListingPauseReason(listing *model.Listing, supply *model.SupplyAccount, seller *model.SellerProfile, snapshot *model.InventorySnapshot, secretReady bool, bindingReady bool) string {
	if listing == nil || supply == nil || seller == nil || snapshot == nil {
		return "invalid marketplace listing sync context"
	}
	if seller.Status != "active" {
		return "seller is not active"
	}
	if supply.Status != "active" {
		return "supply_account is not active"
	}
	if supply.VerifyStatus != "success" {
		return "supply_account verify_status is not success"
	}
	if !secretReady {
		return "seller secret is not active"
	}
	if !bindingReady {
		return "supply channel binding is not ready"
	}
	if snapshot.SyncStatus == "error" {
		if snapshot.SyncMessage != "" {
			return snapshot.SyncMessage
		}
		return "inventory sync status is error"
	}
	if snapshot.HealthScore > 0 && snapshot.HealthScore < 60 {
		return "inventory health score is below threshold"
	}
	return ""
}

func listingShouldBeSoldOutTx(tx *gorm.DB, listing *model.Listing, snapshot *model.InventorySnapshot) (bool, error) {
	if listing == nil || snapshot == nil {
		return false, nil
	}
	if snapshot.AvailableAmount <= 0 {
		return true, nil
	}
	skus, err := model.GetListingSKUsByListingID(listing.Id)
	if err != nil {
		return false, err
	}
	for _, sku := range skus {
		if sku.Status != "active" {
			continue
		}
		requiredAmount := sku.PackageAmount * int64(maxInt(sku.MinQuantity, 1))
		if requiredAmount <= snapshot.AvailableAmount {
			return false, nil
		}
	}
	return true, nil
}

func hasReadyMarketplaceChannelBindingTx(tx *gorm.DB, supply *model.SupplyAccount) (bool, error) {
	if tx == nil || supply == nil {
		return false, nil
	}
	var bindings []model.SupplyChannelBinding
	if err := tx.Where("supply_account_id = ? AND status = ?", supply.Id, "active").
		Order("priority desc, id asc").
		Find(&bindings).Error; err != nil {
		return false, err
	}
	for _, binding := range bindings {
		channel, err := model.GetChannelById(binding.ChannelId, true)
		if err != nil {
			return false, err
		}
		if channel.Status != common.ChannelStatusEnabled {
			continue
		}
		if !containsExactModel(channel.GetModels(), supply.ModelName) {
			continue
		}
		return true, nil
	}
	return false, nil
}

func syncMarketplaceInventoryAfterMutation(supplyAccountID int, reason string) {
	if supplyAccountID <= 0 {
		return
	}
	if err := SyncMarketplaceInventoryBySupplyAccountID(supplyAccountID, reason); err != nil {
		common.SysLog(fmt.Sprintf("marketplace inventory sync failed: supply_account_id=%d reason=%s err=%v", supplyAccountID, reason, err))
	}
}

func maxInt(value int, fallback int) int {
	if value <= 0 {
		return fallback
	}
	return value
}
