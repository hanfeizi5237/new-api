package model

import (
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestMarketOrderModelDefaultsMatchPaymentLifecycle(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&MarketOrder{}, &MarketOrderItem{}))

	order := &MarketOrder{
		OrderNo:        "MKT-DEFAULT-001",
		BuyerUserId:    1,
		IdempotencyKey: "default-idem-001",
	}
	require.NoError(t, db.Create(order).Error)
	require.Equal(t, "pending_payment", order.OrderStatus)
	require.Equal(t, "unpaid", order.PaymentStatus)
	require.Equal(t, "pending", order.EntitlementStatus)

	item := &MarketOrderItem{
		OrderId:         order.Id,
		ListingId:       1,
		SkuId:           1,
		SellerId:        1,
		SupplyAccountId: 1,
		ModelName:       "gpt-4o-mini",
	}
	require.NoError(t, db.Create(item).Error)
	require.Equal(t, "pending_payment", item.Status)
}
