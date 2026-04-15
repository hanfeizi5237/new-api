package model

import (
	"strings"

	"gorm.io/gorm"
)

type MarketOrder struct {
	Id                  int    `json:"id"`
	OrderNo             string `json:"order_no" gorm:"type:varchar(64);not null;uniqueIndex:ux_market_orders_order_no"`
	BuyerUserId         int    `json:"buyer_user_id" gorm:"not null;index:idx_market_orders_buyer_user_id"`
	Currency            string `json:"currency" gorm:"type:varchar(8);not null;default:'CNY'"`
	TotalAmountMinor    int64  `json:"total_amount_minor" gorm:"type:bigint;not null;default:0"`
	PayableAmountMinor  int64  `json:"payable_amount_minor" gorm:"type:bigint;not null;default:0"`
	DiscountAmountMinor int64  `json:"discount_amount_minor" gorm:"type:bigint;not null;default:0"`
	OrderStatus         string `json:"order_status" gorm:"type:varchar(32);not null;default:'pending_payment';index:idx_market_orders_order_status"`
	PaymentStatus       string `json:"payment_status" gorm:"type:varchar(32);not null;default:'unpaid';index:idx_market_orders_payment_status"`
	EntitlementStatus   string `json:"entitlement_status" gorm:"type:varchar(32);not null;default:'pending'"`
	DisputeStatus       string `json:"dispute_status" gorm:"type:varchar(32);not null;default:'none'"`
	PaymentMethod       string `json:"payment_method" gorm:"type:varchar(32)"`
	PaymentTradeNo      string `json:"payment_trade_no" gorm:"type:varchar(128)"`
	ProviderPayload     string `json:"provider_payload" gorm:"type:text"`
	IdempotencyKey      string `json:"idempotency_key" gorm:"type:varchar(128);not null;uniqueIndex:ux_market_orders_idempotency_key"`
	ExpireAt            int64  `json:"expire_at" gorm:"type:bigint;not null;default:0"`
	PaidAt              int64  `json:"paid_at" gorm:"type:bigint;not null;default:0"`
	ClosedAt            int64  `json:"closed_at" gorm:"type:bigint;not null;default:0"`
	CloseReason         string `json:"close_reason" gorm:"type:text"`
	CreatedAt           int64  `json:"created_at" gorm:"bigint;index:idx_market_orders_created_at"`
	UpdatedAt           int64  `json:"updated_at" gorm:"bigint"`
}

func (m *MarketOrder) BeforeCreate(tx *gorm.DB) error {
	setMarketplaceTimestampsOnCreate(&m.CreatedAt, &m.UpdatedAt)
	return nil
}

func (m *MarketOrder) BeforeUpdate(tx *gorm.DB) error {
	setMarketplaceTimestampOnUpdate(&m.UpdatedAt)
	return nil
}

type MarketOrderItem struct {
	Id              int    `json:"id"`
	OrderId         int    `json:"order_id" gorm:"not null;index:idx_market_order_items_order_id"`
	ListingId       int    `json:"listing_id" gorm:"not null;index:idx_market_order_items_listing_id"`
	SkuId           int    `json:"sku_id" gorm:"not null;index:idx_market_order_items_sku_id"`
	SellerId        int    `json:"seller_id" gorm:"not null;index:idx_market_order_items_seller_id"`
	SupplyAccountId int    `json:"supply_account_id" gorm:"not null;index:idx_market_order_items_supply_account_id"`
	VendorId        int    `json:"vendor_id" gorm:"not null;default:0;index:idx_market_order_items_vendor_id"`
	ModelName       string `json:"model_name" gorm:"type:varchar(128);not null;index:idx_market_order_items_model_name"`
	Quantity        int    `json:"quantity" gorm:"not null;default:1"`
	PackageAmount   int64  `json:"package_amount" gorm:"type:bigint;not null;default:0"`
	PackageUnit     string `json:"package_unit" gorm:"type:varchar(32);not null;default:'token'"`
	ValidityDays    int    `json:"validity_days" gorm:"not null;default:0"`
	GrantedAmount   int64  `json:"granted_amount" gorm:"type:bigint;not null;default:0"`
	UsedAmount      int64  `json:"used_amount" gorm:"type:bigint;not null;default:0"`
	RefundedAmount  int64  `json:"refunded_amount" gorm:"type:bigint;not null;default:0"`
	UnitPriceMinor  int64  `json:"unit_price_minor" gorm:"type:bigint;not null;default:0"`
	LineAmountMinor int64  `json:"line_amount_minor" gorm:"type:bigint;not null;default:0"`
	Status          string `json:"status" gorm:"type:varchar(32);not null;default:'pending_payment';index:idx_market_order_items_status"`
	CreatedAt       int64  `json:"created_at" gorm:"bigint"`
	UpdatedAt       int64  `json:"updated_at" gorm:"bigint"`
}

func (m *MarketOrderItem) BeforeCreate(tx *gorm.DB) error {
	setMarketplaceTimestampsOnCreate(&m.CreatedAt, &m.UpdatedAt)
	return nil
}

func (m *MarketOrderItem) BeforeUpdate(tx *gorm.DB) error {
	setMarketplaceTimestampOnUpdate(&m.UpdatedAt)
	return nil
}

func GetMarketOrderByID(id int) (*MarketOrder, error) {
	var order MarketOrder
	if err := DB.First(&order, id).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

func GetMarketOrderByOrderNo(orderNo string) (*MarketOrder, error) {
	var order MarketOrder
	if err := DB.Where("order_no = ?", orderNo).First(&order).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

func GetMarketOrderByIdempotencyKey(idempotencyKey string) (*MarketOrder, error) {
	var order MarketOrder
	if err := DB.Where("idempotency_key = ?", idempotencyKey).First(&order).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

func GetMarketOrderItemsByOrderID(orderId int) ([]MarketOrderItem, error) {
	var items []MarketOrderItem
	if err := DB.Where("order_id = ?", orderId).Order("id asc").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func GetBuyerMarketOrders(buyerUserId int, offset int, limit int) ([]*MarketOrder, int64, error) {
	db := DB.Model(&MarketOrder{}).Where("buyer_user_id = ?", buyerUserId)
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var orders []*MarketOrder
	if err := db.Order("id desc").Offset(offset).Limit(limit).Find(&orders).Error; err != nil {
		return nil, 0, err
	}
	return orders, total, nil
}

func GetMarketOrders(keyword string, buyerUserId int, orderStatus string, paymentStatus string, entitlementStatus string, offset int, limit int) ([]*MarketOrder, int64, error) {
	db := DB.Model(&MarketOrder{})
	if buyerUserId > 0 {
		db = db.Where("buyer_user_id = ?", buyerUserId)
	}
	if trimmed := strings.TrimSpace(orderStatus); trimmed != "" {
		db = db.Where("order_status = ?", trimmed)
	}
	if trimmed := strings.TrimSpace(paymentStatus); trimmed != "" {
		db = db.Where("payment_status = ?", trimmed)
	}
	if trimmed := strings.TrimSpace(entitlementStatus); trimmed != "" {
		db = db.Where("entitlement_status = ?", trimmed)
	}
	if trimmed := strings.TrimSpace(keyword); trimmed != "" {
		like := "%" + trimmed + "%"
		db = db.Where(
			"order_no LIKE ? OR payment_trade_no LIKE ? OR payment_method LIKE ?",
			like, like, like,
		)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var orders []*MarketOrder
	if err := db.Order("id desc").Offset(offset).Limit(limit).Find(&orders).Error; err != nil {
		return nil, 0, err
	}
	return orders, total, nil
}
