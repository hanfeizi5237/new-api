package model

import "gorm.io/gorm"

type UsageLedger struct {
	Id               int    `json:"id"`
	EventKey         string `json:"event_key" gorm:"type:varchar(128);not null;uniqueIndex:ux_usage_ledgers_event_key"`
	RequestId        string `json:"request_id" gorm:"type:varchar(64);not null;index:idx_usage_ledgers_request_id"`
	BillingSource    string `json:"billing_source" gorm:"type:varchar(64);not null;default:'wallet';index:idx_usage_ledgers_billing_source"`
	UserId           int    `json:"user_id" gorm:"not null;index:idx_usage_ledgers_user_id"`
	TokenId          int    `json:"token_id" gorm:"not null;default:0"`
	ChannelId        int    `json:"channel_id" gorm:"not null;default:0;index:idx_usage_ledgers_channel_id"`
	SellerId         int    `json:"seller_id" gorm:"not null;default:0;index:idx_usage_ledgers_seller_id"`
	SupplyAccountId  int    `json:"supply_account_id" gorm:"not null;default:0;index:idx_usage_ledgers_supply_account_id"`
	ListingId        int    `json:"listing_id" gorm:"not null;default:0"`
	OrderId          int    `json:"order_id" gorm:"not null;default:0;index:idx_usage_ledgers_order_id"`
	OrderItemId      int    `json:"order_item_id" gorm:"not null;default:0"`
	EntitlementLotId int    `json:"entitlement_lot_id" gorm:"not null;default:0;index:idx_usage_ledgers_entitlement_lot_id"`
	ModelName        string `json:"model_name" gorm:"type:varchar(128);not null"`
	EndpointType     string `json:"endpoint_type" gorm:"type:varchar(32);not null;default:'chat'"`
	PromptTokens     int    `json:"prompt_tokens" gorm:"not null;default:0"`
	CompletionTokens int    `json:"completion_tokens" gorm:"not null;default:0"`
	TotalTokens      int    `json:"total_tokens" gorm:"not null;default:0"`
	PreConsumedQuota int    `json:"pre_consumed_quota" gorm:"not null;default:0"`
	ActualQuota      int    `json:"actual_quota" gorm:"not null;default:0"`
	RetryIndex       int    `json:"retry_index" gorm:"not null;default:0"`
	LedgerStatus     string `json:"ledger_status" gorm:"type:varchar(32);not null;default:'success'"`
	IsStream         bool   `json:"is_stream" gorm:"not null;default:false"`
	StartedAt        int64  `json:"started_at" gorm:"type:bigint;not null;default:0;index:idx_usage_ledgers_started_at"`
	FinishedAt       int64  `json:"finished_at" gorm:"type:bigint;not null;default:0"`
	ErrorCode        string `json:"error_code" gorm:"type:varchar(64)"`
	Other            string `json:"other" gorm:"type:text"`
	CreatedAt        int64  `json:"created_at" gorm:"bigint"`
	UpdatedAt        int64  `json:"updated_at" gorm:"bigint"`
}

func (u *UsageLedger) BeforeCreate(tx *gorm.DB) error {
	setMarketplaceTimestampsOnCreate(&u.CreatedAt, &u.UpdatedAt)
	if u.StartedAt == 0 {
		u.StartedAt = u.CreatedAt
	}
	return nil
}

func (u *UsageLedger) BeforeUpdate(tx *gorm.DB) error {
	setMarketplaceTimestampOnUpdate(&u.UpdatedAt)
	return nil
}
