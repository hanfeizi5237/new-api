package service

import (
	"errors"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"gorm.io/gorm"
)

type PrepareMarketOrderPaymentInput struct {
	OrderID       int
	BuyerUserID   int
	PaymentMethod string
}

type MarketPaymentIntent struct {
	OrderID            int    `json:"order_id"`
	OrderNo            string `json:"order_no"`
	PaymentMethod      string `json:"payment_method"`
	Currency           string `json:"currency"`
	PayableAmountMinor int64  `json:"payable_amount_minor"`
	OrderStatus        string `json:"order_status"`
	PaymentStatus      string `json:"payment_status"`
}

type CompleteMarketOrderPaymentInput struct {
	OrderNo            string
	PaymentMethod      string
	PaymentTradeNo     string
	Currency           string
	PayableAmountMinor int64
	ProviderPayload    string
}

func PrepareMarketOrderPayment(input PrepareMarketOrderPaymentInput) (*MarketPaymentIntent, error) {
	if input.OrderID <= 0 || input.BuyerUserID <= 0 {
		return nil, errors.New("invalid order_id or buyer_user_id")
	}
	paymentMethod := strings.TrimSpace(input.PaymentMethod)
	if paymentMethod == "" {
		return nil, errors.New("payment_method is required")
	}
	order, err := model.GetMarketOrderByID(input.OrderID)
	if err != nil {
		return nil, err
	}
	if order.BuyerUserId != input.BuyerUserID {
		return nil, errors.New("order does not belong to buyer")
	}
	if order.OrderStatus != marketOrderStatusPendingPayment && order.OrderStatus != marketOrderStatusPaid {
		return nil, errors.New("order is not payable")
	}
	if err := model.DB.Model(&model.MarketOrder{}).
		Where("id = ?", order.Id).
		Updates(map[string]interface{}{
			"payment_method": paymentMethod,
			"updated_at":     common.GetTimestamp(),
		}).Error; err != nil {
		return nil, err
	}
	order.PaymentMethod = paymentMethod
	return &MarketPaymentIntent{
		OrderID:            order.Id,
		OrderNo:            order.OrderNo,
		PaymentMethod:      paymentMethod,
		Currency:           order.Currency,
		PayableAmountMinor: order.PayableAmountMinor,
		OrderStatus:        order.OrderStatus,
		PaymentStatus:      order.PaymentStatus,
	}, nil
}

func CompleteMarketOrderPayment(input CompleteMarketOrderPaymentInput) (*model.MarketOrder, error) {
	orderNo := strings.TrimSpace(input.OrderNo)
	if orderNo == "" {
		return nil, errors.New("order_no is required")
	}

	var completedOrder *model.MarketOrder
	err := model.DB.Transaction(func(tx *gorm.DB) error {
		var order model.MarketOrder
		if err := tx.Set("gorm:query_option", "FOR UPDATE").
			Where("order_no = ?", orderNo).
			First(&order).Error; err != nil {
			return err
		}
		if order.PaymentStatus == marketPaymentStatusPaid && order.EntitlementStatus == marketEntitlementStatusCreated {
			completedOrder = &order
			return nil
		}
		if order.OrderStatus == marketOrderStatusClosed {
			return errors.New("order is already closed")
		}
		if input.PayableAmountMinor > 0 && input.PayableAmountMinor != order.PayableAmountMinor {
			return errors.New("payment amount mismatch")
		}
		if trimmedCurrency := strings.ToUpper(strings.TrimSpace(input.Currency)); trimmedCurrency != "" && trimmedCurrency != order.Currency {
			return errors.New("payment currency mismatch")
		}

		items, err := loadMarketOrderItemsTx(tx, order.Id)
		if err != nil {
			return err
		}
		paidAt := common.GetTimestamp()
		for _, item := range items {
			grantAmount := item.PackageAmount * int64(item.Quantity)
			var supply model.SupplyAccount
			if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&supply, item.SupplyAccountId).Error; err != nil {
				return err
			}
			var snapshot model.InventorySnapshot
			if err := tx.Set("gorm:query_option", "FOR UPDATE").
				Where("supply_account_id = ?", item.SupplyAccountId).
				First(&snapshot).Error; err != nil {
				return err
			}
			if snapshot.RiskDiscountBps <= 0 {
				snapshot.RiskDiscountBps = 10000
			}
			snapshot.FrozenAmount -= grantAmount
			if snapshot.FrozenAmount < 0 {
				snapshot.FrozenAmount = 0
			}
			snapshot.SoldAmount += grantAmount
			supply.ReservedCapacity += grantAmount
			snapshot.AvailableAmount = recomputeInventoryAvailableAmount(&supply, &snapshot)
			if err := tx.Save(&supply).Error; err != nil {
				return err
			}
			if err := tx.Save(&snapshot).Error; err != nil {
				return err
			}
		}

		order.PaidAt = paidAt
		order.OrderStatus = marketOrderStatusPaid
		order.PaymentStatus = marketPaymentStatusPaid
		order.EntitlementStatus = marketEntitlementStatusCreated
		if trimmedMethod := strings.TrimSpace(input.PaymentMethod); trimmedMethod != "" {
			order.PaymentMethod = trimmedMethod
		}
		if trimmedTradeNo := strings.TrimSpace(input.PaymentTradeNo); trimmedTradeNo != "" {
			order.PaymentTradeNo = trimmedTradeNo
		}
		if trimmedPayload := strings.TrimSpace(input.ProviderPayload); trimmedPayload != "" {
			order.ProviderPayload = trimmedPayload
		}
		if err := tx.Save(&order).Error; err != nil {
			return err
		}
		if err := grantEntitlementsForOrderTx(tx, &order, items); err != nil {
			if saveErr := tx.Model(&model.MarketOrder{}).
				Where("id = ?", order.Id).
				Updates(map[string]interface{}{
					"entitlement_status": marketEntitlementStatusFailed,
					"updated_at":         common.GetTimestamp(),
				}).Error; saveErr != nil {
				return saveErr
			}
			return err
		}
		completedOrder = &order
		return nil
	})
	if err != nil {
		return nil, err
	}
	return completedOrder, nil
}
