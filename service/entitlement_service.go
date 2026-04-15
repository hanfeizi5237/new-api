package service

import (
	"errors"
	"fmt"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"gorm.io/gorm"
)

func ListBuyerEntitlements(buyerUserId int, modelName string, offset int, limit int) ([]*model.BuyerEntitlement, int64, error) {
	if buyerUserId <= 0 {
		return nil, 0, errors.New("invalid buyer_user_id")
	}
	return model.GetBuyerEntitlements(buyerUserId, modelName, offset, limit)
}

func ListEntitlementsAdmin(buyerUserId int, modelName string, status string, offset int, limit int) ([]*model.BuyerEntitlement, int64, error) {
	return model.GetEntitlements(buyerUserId, modelName, status, offset, limit)
}

func grantEntitlementsForOrderTx(tx *gorm.DB, order *model.MarketOrder, items []model.MarketOrderItem) error {
	if tx == nil || order == nil {
		return errors.New("invalid order grant context")
	}
	now := common.GetTimestamp()
	for _, item := range items {
		grantAmount := item.PackageAmount * int64(item.Quantity)
		sourceEventKey := fmt.Sprintf("%s:%d:grant", order.OrderNo, item.Id)

		var existingLot model.EntitlementLot
		if err := tx.Where("source_event_key = ?", sourceEventKey).First(&existingLot).Error; err == nil {
			if err := tx.Model(&model.MarketOrderItem{}).
				Where("id = ?", item.Id).
				Updates(map[string]interface{}{
					"granted_amount": existingLot.GrantedAmount,
					"status":         "granted",
					"updated_at":     now,
				}).Error; err != nil {
				return err
			}
			continue
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		var entitlement model.BuyerEntitlement
		err := lockForUpdate(tx).
			Where("buyer_user_id = ? AND vendor_id = ? AND model_name = ?", order.BuyerUserId, item.VendorId, item.ModelName).
			First(&entitlement).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			entitlement = model.BuyerEntitlement{
				BuyerUserId: order.BuyerUserId,
				VendorId:    item.VendorId,
				ModelName:   item.ModelName,
				Status:      "active",
			}
			if err := tx.Create(&entitlement).Error; err != nil {
				return err
			}
		} else if err != nil {
			return err
		}

		entitlement.TotalGranted += grantAmount
		if entitlement.Status == "" {
			entitlement.Status = "active"
		}
		if err := tx.Save(&entitlement).Error; err != nil {
			return err
		}

		expireAt := int64(0)
		if item.ValidityDays > 0 {
			expireAt = order.PaidAt + int64(item.ValidityDays)*24*60*60
		}
		lot := model.EntitlementLot{
			BuyerEntitlementId: entitlement.Id,
			BuyerUserId:        order.BuyerUserId,
			OrderId:            order.Id,
			OrderItemId:        item.Id,
			SellerId:           item.SellerId,
			ListingId:          item.ListingId,
			SupplyAccountId:    item.SupplyAccountId,
			GrantedAmount:      grantAmount,
			ExpireAt:           expireAt,
			PrioritySeq:        order.PaidAt,
			SourceEventKey:     sourceEventKey,
			Status:             "active",
		}
		if err := tx.Create(&lot).Error; err != nil {
			return err
		}
		if err := tx.Model(&model.MarketOrderItem{}).
			Where("id = ?", item.Id).
			Updates(map[string]interface{}{
				"granted_amount": grantAmount,
				"status":         "granted",
				"updated_at":     now,
			}).Error; err != nil {
			return err
		}
	}
	return nil
}
