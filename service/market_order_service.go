package service

import (
	"errors"
	"fmt"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"gorm.io/gorm"
)

const (
	marketOrderStatusPendingPayment = "pending_payment"
	marketOrderStatusPaid           = "paid"
	marketOrderStatusClosed         = "closed"

	marketPaymentStatusUnpaid = "unpaid"
	marketPaymentStatusPaid   = "paid"
	marketPaymentStatusFailed = "failed"

	marketEntitlementStatusPending = "pending"
	marketEntitlementStatusCreated = "created"
	marketEntitlementStatusFailed  = "failed"
)

type CreateMarketOrderInput struct {
	BuyerUserID    int
	ListingID      int
	SkuID          int
	Quantity       int
	IdempotencyKey string
	Currency       string
}

type MarketOrderDetail struct {
	Order *model.MarketOrder      `json:"order"`
	Items []model.MarketOrderItem `json:"items"`
}

func CreateMarketOrder(input CreateMarketOrderInput) (*model.MarketOrder, []model.MarketOrderItem, error) {
	if input.BuyerUserID <= 0 {
		return nil, nil, errors.New("buyer_user_id is required")
	}
	if input.ListingID <= 0 || input.SkuID <= 0 {
		return nil, nil, errors.New("listing_id and sku_id are required")
	}
	if input.Quantity <= 0 {
		return nil, nil, errors.New("quantity must be greater than 0")
	}
	idempotencyKey := strings.TrimSpace(input.IdempotencyKey)
	if idempotencyKey == "" {
		return nil, nil, errors.New("idempotency_key is required")
	}

	// Reuse the original pending/paid order for the same buyer so refresh/retry does not freeze inventory twice.
	existingOrder, err := model.GetMarketOrderByIdempotencyKey(idempotencyKey)
	if err == nil {
		if existingOrder.BuyerUserId != input.BuyerUserID {
			return nil, nil, errors.New("idempotency_key already belongs to another buyer")
		}
		items, itemErr := model.GetMarketOrderItemsByOrderID(existingOrder.Id)
		return existingOrder, items, itemErr
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil, err
	}

	if _, err := model.GetUserById(input.BuyerUserID, false); err != nil {
		return nil, nil, err
	}
	listing, err := model.GetListingByID(input.ListingID)
	if err != nil {
		return nil, nil, err
	}
	if listing.Status != "active" || listing.AuditStatus != "approved" {
		return nil, nil, errors.New("listing is not available for purchase")
	}
	sku, err := model.GetListingSKUByID(input.SkuID)
	if err != nil {
		return nil, nil, err
	}
	if sku.ListingId != listing.Id {
		return nil, nil, errors.New("sku does not belong to listing")
	}
	if sku.Status != "active" {
		return nil, nil, errors.New("sku is not active")
	}
	if sku.PackageUnit != "token" {
		return nil, nil, errors.New("M1 only supports token package units")
	}
	if input.Quantity < sku.MinQuantity || input.Quantity > sku.MaxQuantity {
		return nil, nil, errors.New("quantity is out of allowed range")
	}
	seller, err := model.GetSellerByID(listing.SellerId)
	if err != nil {
		return nil, nil, err
	}
	if seller.Status != "active" {
		return nil, nil, errors.New("seller is not active")
	}
	supply, err := model.GetSupplyAccountByID(listing.SupplyAccountId)
	if err != nil {
		return nil, nil, err
	}
	if supply.Status != "active" || supply.VerifyStatus != "success" {
		return nil, nil, errors.New("supply_account is not ready")
	}

	var activeSecretCount int64
	if err := model.DB.Model(&model.SellerSecret{}).
		Where("supply_account_id = ? AND status = ? AND verify_status = ?", supply.Id, "active", "success").
		Count(&activeSecretCount).Error; err != nil {
		return nil, nil, err
	}
	if activeSecretCount == 0 {
		return nil, nil, errors.New("supply_account has no active verified seller secret")
	}
	var activeBindingCount int64
	if err := model.DB.Model(&model.SupplyChannelBinding{}).
		Where("supply_account_id = ? AND status = ?", supply.Id, "active").
		Count(&activeBindingCount).Error; err != nil {
		return nil, nil, err
	}
	if activeBindingCount == 0 {
		return nil, nil, errors.New("supply_account has no active channel binding")
	}

	now := common.GetTimestamp()
	currency := strings.ToUpper(strings.TrimSpace(input.Currency))
	if currency == "" {
		currency = "CNY"
	}
	freezeAmount := sku.PackageAmount * int64(input.Quantity)
	lineAmount := sku.UnitPriceMinor * int64(input.Quantity)
	order := model.MarketOrder{
		OrderNo:            fmt.Sprintf("MKT-%d-%s", input.BuyerUserID, strings.ToUpper(common.GetRandomString(8))),
		BuyerUserId:        input.BuyerUserID,
		Currency:           currency,
		TotalAmountMinor:   lineAmount,
		PayableAmountMinor: lineAmount,
		OrderStatus:        marketOrderStatusPendingPayment,
		PaymentStatus:      marketPaymentStatusUnpaid,
		EntitlementStatus:  marketEntitlementStatusPending,
		IdempotencyKey:     idempotencyKey,
		ExpireAt:           now + 15*60,
	}
	orderItem := model.MarketOrderItem{
		ListingId:       listing.Id,
		SkuId:           sku.Id,
		SellerId:        seller.Id,
		SupplyAccountId: supply.Id,
		VendorId:        listing.VendorId,
		ModelName:       listing.ModelName,
		Quantity:        input.Quantity,
		PackageAmount:   sku.PackageAmount,
		PackageUnit:     sku.PackageUnit,
		ValidityDays:    listing.ValidityDays,
		UnitPriceMinor:  sku.UnitPriceMinor,
		LineAmountMinor: lineAmount,
		Status:          marketOrderStatusPendingPayment,
	}

	if err := model.DB.Transaction(func(tx *gorm.DB) error {
		var lockedSupply model.SupplyAccount
		if err := lockForUpdate(tx).First(&lockedSupply, supply.Id).Error; err != nil {
			return err
		}
		var snapshot model.InventorySnapshot
		if err := lockForUpdate(tx).
			Where("supply_account_id = ?", lockedSupply.Id).
			First(&snapshot).Error; err != nil {
			return err
		}
		if snapshot.RiskDiscountBps <= 0 {
			snapshot.RiskDiscountBps = 10000
		}
		if snapshot.AvailableAmount <= 0 {
			snapshot.AvailableAmount = recomputeInventoryAvailableAmount(&lockedSupply, &snapshot)
		}
		// Keep the inventory freeze atomic while staying cross-database compatible.
		result := tx.Model(&model.InventorySnapshot{}).
			Where("id = ? AND available_amount >= ?", snapshot.Id, freezeAmount).
			Updates(map[string]interface{}{
				"frozen_amount":    gorm.Expr("frozen_amount + ?", freezeAmount),
				"available_amount": gorm.Expr("available_amount - ?", freezeAmount),
				"updated_at":       common.GetTimestamp(),
			})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return errors.New("insufficient inventory")
		}
		// Re-read snapshot to keep in-memory state consistent for downstream logic.
		if err := tx.Where("id = ?", snapshot.Id).First(&snapshot).Error; err != nil {
			return err
		}
		if err := tx.Create(&order).Error; err != nil {
			return err
		}
		orderItem.OrderId = order.Id
		return tx.Create(&orderItem).Error
	}); err != nil {
		return nil, nil, err
	}
	// Recompute listing visibility asynchronously after the write transaction commits.
	syncMarketplaceInventoryAfterMutation(supply.Id, "order_created")

	return &order, []model.MarketOrderItem{orderItem}, nil
}

func ListBuyerMarketOrders(buyerUserId int, offset int, limit int) ([]*model.MarketOrder, int64, error) {
	if buyerUserId <= 0 {
		return nil, 0, errors.New("invalid buyer_user_id")
	}
	return model.GetBuyerMarketOrders(buyerUserId, offset, limit)
}

func ListMarketOrdersAdmin(keyword string, buyerUserId int, orderStatus string, paymentStatus string, entitlementStatus string, offset int, limit int) ([]*model.MarketOrder, int64, error) {
	return model.GetMarketOrders(keyword, buyerUserId, orderStatus, paymentStatus, entitlementStatus, offset, limit)
}

func GetBuyerMarketOrderDetail(orderId int, buyerUserId int) (*MarketOrderDetail, error) {
	if orderId <= 0 || buyerUserId <= 0 {
		return nil, errors.New("invalid order_id or buyer_user_id")
	}
	order, err := model.GetMarketOrderByID(orderId)
	if err != nil {
		return nil, err
	}
	if order.BuyerUserId != buyerUserId {
		return nil, errors.New("order does not belong to buyer")
	}
	items, err := model.GetMarketOrderItemsByOrderID(order.Id)
	if err != nil {
		return nil, err
	}
	return &MarketOrderDetail{
		Order: order,
		Items: items,
	}, nil
}

func CloseExpiredMarketOrders(now int64) (int, error) {
	if now <= 0 {
		now = common.GetTimestamp()
	}
	var orderIDs []int
	if err := model.DB.Model(&model.MarketOrder{}).
		Where("order_status = ? AND payment_status = ? AND expire_at > 0 AND expire_at <= ?", marketOrderStatusPendingPayment, marketPaymentStatusUnpaid, now).
		Pluck("id", &orderIDs).Error; err != nil {
		return 0, err
	}

	closedCount := 0
	for _, orderID := range orderIDs {
		closed, err := closeMarketOrder(orderID, now, "payment_timeout")
		if err != nil {
			return closedCount, err
		}
		if closed {
			closedCount++
		}
	}
	return closedCount, nil
}

func closeMarketOrder(orderID int, now int64, reason string) (bool, error) {
	closed := false
	err := model.DB.Transaction(func(tx *gorm.DB) error {
		var order model.MarketOrder
		if err := lockForUpdate(tx).First(&order, orderID).Error; err != nil {
			return err
		}
		if order.OrderStatus != marketOrderStatusPendingPayment || order.PaymentStatus != marketPaymentStatusUnpaid {
			return nil
		}
		items, err := loadMarketOrderItemsTx(tx, order.Id)
		if err != nil {
			return err
		}
		for _, item := range items {
			releaseAmount := item.PackageAmount * int64(item.Quantity)
			if err := releaseInventoryFreezeTx(tx, item.SupplyAccountId, releaseAmount); err != nil {
				return err
			}
			if err := tx.Model(&model.MarketOrderItem{}).
				Where("id = ?", item.Id).
				Updates(map[string]interface{}{
					"status":     marketOrderStatusClosed,
					"updated_at": now,
				}).Error; err != nil {
				return err
			}
		}
		if err := tx.Model(&model.MarketOrder{}).
			Where("id = ?", order.Id).
			Updates(map[string]interface{}{
				"order_status": marketOrderStatusClosed,
				"closed_at":    now,
				"close_reason": reason,
				"updated_at":   now,
			}).Error; err != nil {
			return err
		}
		closed = true
		return nil
	})
	if err == nil && closed {
		order, getErr := model.GetMarketOrderByID(orderID)
		if getErr == nil {
			syncMarketplaceInventoryAfterMutation(orderItemsSupplyAccountID(orderID), reason)
		} else {
			_ = order
		}
	}
	return closed, err
}

func orderItemsSupplyAccountID(orderID int) int {
	items, err := model.GetMarketOrderItemsByOrderID(orderID)
	if err != nil || len(items) == 0 {
		return 0
	}
	return items[0].SupplyAccountId
}

func loadMarketOrderItemsTx(tx *gorm.DB, orderID int) ([]model.MarketOrderItem, error) {
	var items []model.MarketOrderItem
	if err := tx.Where("order_id = ?", orderID).Order("id asc").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func releaseInventoryFreezeTx(tx *gorm.DB, supplyAccountID int, releaseAmount int64) error {
	var supply model.SupplyAccount
	if err := lockForUpdate(tx).First(&supply, supplyAccountID).Error; err != nil {
		return err
	}
	var snapshot model.InventorySnapshot
	if err := lockForUpdate(tx).
		Where("supply_account_id = ?", supplyAccountID).
		First(&snapshot).Error; err != nil {
		return err
	}
	if snapshot.RiskDiscountBps <= 0 {
		snapshot.RiskDiscountBps = 10000
	}
	snapshot.FrozenAmount -= releaseAmount
	if snapshot.FrozenAmount < 0 {
		snapshot.FrozenAmount = 0
	}
	snapshot.AvailableAmount = recomputeInventoryAvailableAmount(&supply, &snapshot)
	return tx.Save(&snapshot).Error
}
