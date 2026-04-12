package model

import "gorm.io/gorm"

type BuyerEntitlement struct {
	Id            int    `json:"id"`
	BuyerUserId   int    `json:"buyer_user_id" gorm:"not null;uniqueIndex:ux_buyer_entitlements_owner_model,priority:1"`
	VendorId      int    `json:"vendor_id" gorm:"not null;default:0;uniqueIndex:ux_buyer_entitlements_owner_model,priority:2"`
	ModelName     string `json:"model_name" gorm:"type:varchar(128);not null;uniqueIndex:ux_buyer_entitlements_owner_model,priority:3"`
	TotalGranted  int64  `json:"total_granted" gorm:"type:bigint;not null;default:0"`
	TotalUsed     int64  `json:"total_used" gorm:"type:bigint;not null;default:0"`
	TotalRefunded int64  `json:"total_refunded" gorm:"type:bigint;not null;default:0"`
	TotalFrozen   int64  `json:"total_frozen" gorm:"type:bigint;not null;default:0"`
	Status        string `json:"status" gorm:"type:varchar(32);not null;default:'active';index:idx_buyer_entitlements_status"`
	CreatedAt     int64  `json:"created_at" gorm:"bigint"`
	UpdatedAt     int64  `json:"updated_at" gorm:"bigint"`
}

func (b *BuyerEntitlement) BeforeCreate(tx *gorm.DB) error {
	setMarketplaceTimestampsOnCreate(&b.CreatedAt, &b.UpdatedAt)
	return nil
}

func (b *BuyerEntitlement) BeforeUpdate(tx *gorm.DB) error {
	setMarketplaceTimestampOnUpdate(&b.UpdatedAt)
	return nil
}

type EntitlementLot struct {
	Id                 int    `json:"id"`
	BuyerEntitlementId int    `json:"buyer_entitlement_id" gorm:"not null;index:idx_entitlement_lots_buyer_entitlement_id"`
	BuyerUserId        int    `json:"buyer_user_id" gorm:"not null;index:idx_entitlement_lots_buyer_user_id"`
	OrderId            int    `json:"order_id" gorm:"not null;index:idx_entitlement_lots_order_id"`
	OrderItemId        int    `json:"order_item_id" gorm:"not null;index:idx_entitlement_lots_order_item_id"`
	SellerId           int    `json:"seller_id" gorm:"not null;index:idx_entitlement_lots_seller_id"`
	ListingId          int    `json:"listing_id" gorm:"not null;index:idx_entitlement_lots_listing_id"`
	SupplyAccountId    int    `json:"supply_account_id" gorm:"not null;index:idx_entitlement_lots_supply_account_id"`
	GrantedAmount      int64  `json:"granted_amount" gorm:"type:bigint;not null;default:0"`
	UsedAmount         int64  `json:"used_amount" gorm:"type:bigint;not null;default:0"`
	RefundedAmount     int64  `json:"refunded_amount" gorm:"type:bigint;not null;default:0"`
	FrozenAmount       int64  `json:"frozen_amount" gorm:"type:bigint;not null;default:0"`
	ExpireAt           int64  `json:"expire_at" gorm:"type:bigint;not null;default:0;index:idx_entitlement_lots_expire_at"`
	PrioritySeq        int64  `json:"priority_seq" gorm:"type:bigint;not null;default:0;index:idx_entitlement_lots_priority_seq"`
	SourceEventKey     string `json:"source_event_key" gorm:"type:varchar(128);not null;uniqueIndex:ux_entitlement_lots_source_event_key"`
	Status             string `json:"status" gorm:"type:varchar(32);not null;default:'active';index:idx_entitlement_lots_status"`
	CreatedAt          int64  `json:"created_at" gorm:"bigint"`
	UpdatedAt          int64  `json:"updated_at" gorm:"bigint"`
}

func (e *EntitlementLot) BeforeCreate(tx *gorm.DB) error {
	setMarketplaceTimestampsOnCreate(&e.CreatedAt, &e.UpdatedAt)
	return nil
}

func (e *EntitlementLot) BeforeUpdate(tx *gorm.DB) error {
	setMarketplaceTimestampOnUpdate(&e.UpdatedAt)
	return nil
}

func GetBuyerEntitlements(buyerUserId int, modelName string, offset int, limit int) ([]*BuyerEntitlement, int64, error) {
	db := DB.Model(&BuyerEntitlement{}).Where("buyer_user_id = ?", buyerUserId)
	if modelName != "" {
		db = db.Where("model_name = ?", modelName)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var entitlements []*BuyerEntitlement
	if err := db.Order("id desc").Offset(offset).Limit(limit).Find(&entitlements).Error; err != nil {
		return nil, 0, err
	}
	return entitlements, total, nil
}
