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
	OrderID            int               `json:"order_id"`
	OrderNo            string            `json:"order_no"`
	PaymentMethod      string            `json:"payment_method"`
	Currency           string            `json:"currency"`
	PayableAmountMinor int64             `json:"payable_amount_minor"`
	OrderStatus        string            `json:"order_status"`
	PaymentStatus      string            `json:"payment_status"`
	ProviderReference  string            `json:"provider_reference,omitempty"`
	RedirectURL        string            `json:"redirect_url,omitempty"`
	FormURL            string            `json:"form_url,omitempty"`
	FormParams         map[string]string `json:"form_params,omitempty"`
}

type CompleteMarketOrderPaymentInput struct {
	OrderNo            string
	PaymentMethod      string
	PaymentTradeNo     string
	Currency           string
	PayableAmountMinor int64
	ProviderPayload    string
}

type FailMarketOrderPaymentInput struct {
	OrderNo            string
	PaymentMethod      string
	PaymentTradeNo     string
	Currency           string
	PayableAmountMinor int64
	FailureReason      string
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

	intent := &MarketPaymentIntent{
		OrderID:            order.Id,
		OrderNo:            order.OrderNo,
		PaymentMethod:      paymentMethod,
		Currency:           order.Currency,
		PayableAmountMinor: order.PayableAmountMinor,
		OrderStatus:        order.OrderStatus,
		PaymentStatus:      order.PaymentStatus,
	}
	if order.PaymentStatus == marketPaymentStatusPaid {
		return intent, nil
	}

	// Reuse the same provider intent while the order is still unpaid so repeated clicks do not spawn duplicate checkouts.
	if cachedIntent, ok := decodeStoredMarketPaymentIntent(order, paymentMethod); ok {
		return cachedIntent, nil
	}

	buyer, err := model.GetUserById(order.BuyerUserId, true)
	if err != nil {
		return nil, err
	}

	providerIntent, err := createMarketProviderPaymentIntent(marketProviderPaymentInitInput{
		Order:                  order,
		Buyer:                  buyer,
		RequestedPaymentMethod: paymentMethod,
	})
	if err != nil {
		return nil, err
	}
	storedPayload, err := encodeStoredMarketPaymentIntent(paymentMethod, providerIntent)
	if err != nil {
		return nil, err
	}

	if err := model.DB.Model(&model.MarketOrder{}).
		Where("id = ?", order.Id).
		Updates(map[string]interface{}{
			"payment_method":   paymentMethod,
			"provider_payload": storedPayload,
			"updated_at":       common.GetTimestamp(),
		}).Error; err != nil {
		return nil, err
	}

	intent.ProviderReference = providerIntent.ProviderReference
	intent.RedirectURL = providerIntent.RedirectURL
	intent.FormURL = providerIntent.FormURL
	intent.FormParams = providerIntent.FormParams
	return intent, nil
}

func CompleteMarketOrderPayment(input CompleteMarketOrderPaymentInput) (*model.MarketOrder, error) {
	orderNo := strings.TrimSpace(input.OrderNo)
	if orderNo == "" {
		return nil, errors.New("order_no is required")
	}

	var completedOrder *model.MarketOrder
	err := model.DB.Transaction(func(tx *gorm.DB) error {
		var order model.MarketOrder
		if err := lockForUpdate(tx).
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
		// A paid order may re-enter here when the PSP retries after entitlement grant failed.
		// In that case we only replay entitlement creation and keep inventory counters untouched.
		retryEntitlementGrantOnly := order.PaymentStatus == marketPaymentStatusPaid && order.EntitlementStatus == marketEntitlementStatusFailed

		items, err := loadMarketOrderItemsTx(tx, order.Id)
		if err != nil {
			return err
		}
		paidAt := order.PaidAt
		if paidAt <= 0 {
			paidAt = common.GetTimestamp()
		}
		if !retryEntitlementGrantOnly {
			for _, item := range items {
				grantAmount := item.PackageAmount * int64(item.Quantity)
				var supply model.SupplyAccount
				if err := lockForUpdate(tx).First(&supply, item.SupplyAccountId).Error; err != nil {
					return err
				}
				var snapshot model.InventorySnapshot
				if err := lockForUpdate(tx).
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
	for _, item := range itemsForOrder(completedOrder) {
		syncMarketplaceInventoryAfterMutation(item.SupplyAccountId, "payment_completed")
	}
	return completedOrder, nil
}

func FailMarketOrderPayment(input FailMarketOrderPaymentInput) (*model.MarketOrder, error) {
	orderNo := strings.TrimSpace(input.OrderNo)
	if orderNo == "" {
		return nil, errors.New("order_no is required")
	}
	failureReason := strings.TrimSpace(input.FailureReason)
	if failureReason == "" {
		failureReason = "payment_failed"
	}

	var failedOrder *model.MarketOrder
	err := model.DB.Transaction(func(tx *gorm.DB) error {
		var order model.MarketOrder
		if err := lockForUpdate(tx).
			Where("order_no = ?", orderNo).
			First(&order).Error; err != nil {
			return err
		}
		if order.PaymentStatus == marketPaymentStatusPaid {
			return errors.New("order is already paid")
		}
		if order.OrderStatus == marketOrderStatusClosed && order.PaymentStatus == marketPaymentStatusFailed {
			failedOrder = &order
			return nil
		}
		items, err := loadMarketOrderItemsTx(tx, order.Id)
		if err != nil {
			return err
		}
		now := common.GetTimestamp()
		for _, item := range items {
			releaseAmount := item.PackageAmount * int64(item.Quantity)
			// Failed/expired orders must give the frozen inventory back before the order is closed.
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

		updates := map[string]interface{}{
			"order_status":   marketOrderStatusClosed,
			"payment_status": marketPaymentStatusFailed,
			"closed_at":      now,
			"close_reason":   failureReason,
			"updated_at":     now,
		}
		if trimmedMethod := strings.TrimSpace(input.PaymentMethod); trimmedMethod != "" {
			updates["payment_method"] = trimmedMethod
		}
		if trimmedTradeNo := strings.TrimSpace(input.PaymentTradeNo); trimmedTradeNo != "" {
			updates["payment_trade_no"] = trimmedTradeNo
		}
		if trimmedPayload := strings.TrimSpace(input.ProviderPayload); trimmedPayload != "" {
			updates["provider_payload"] = trimmedPayload
		}
		if err := tx.Model(&model.MarketOrder{}).
			Where("id = ?", order.Id).
			Updates(updates).Error; err != nil {
			return err
		}
		order.OrderStatus = marketOrderStatusClosed
		order.PaymentStatus = marketPaymentStatusFailed
		order.CloseReason = failureReason
		order.ClosedAt = now
		if method, ok := updates["payment_method"].(string); ok {
			order.PaymentMethod = method
		}
		if tradeNo, ok := updates["payment_trade_no"].(string); ok {
			order.PaymentTradeNo = tradeNo
		}
		if payload, ok := updates["provider_payload"].(string); ok {
			order.ProviderPayload = payload
		}
		failedOrder = &order
		return nil
	})
	if err != nil {
		return nil, err
	}
	for _, item := range itemsForOrder(failedOrder) {
		syncMarketplaceInventoryAfterMutation(item.SupplyAccountId, "payment_failed")
	}
	return failedOrder, nil
}

func itemsForOrder(order *model.MarketOrder) []model.MarketOrderItem {
	if order == nil || order.Id <= 0 {
		return nil
	}
	items, err := model.GetMarketOrderItemsByOrderID(order.Id)
	if err != nil {
		return nil
	}
	return items
}
