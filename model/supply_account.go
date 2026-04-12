package model

import (
	"github.com/QuantumNous/new-api/common"
	"gorm.io/gorm"
)

type SupplyAccount struct {
	Id               int    `json:"id"`
	SellerId         int    `json:"seller_id" gorm:"not null;index:idx_supply_accounts_seller_id"`
	SupplyCode       string `json:"supply_code" gorm:"type:varchar(64);not null;uniqueIndex:ux_supply_accounts_supply_code"`
	ProviderCode     string `json:"provider_code" gorm:"type:varchar(64);not null;index:idx_supply_accounts_provider_code"`
	VendorId         int    `json:"vendor_id" gorm:"not null;default:0;index:idx_supply_accounts_vendor_id"`
	ModelName        string `json:"model_name" gorm:"type:varchar(128);not null;index:idx_supply_accounts_model_name"`
	QuotaUnit        string `json:"quota_unit" gorm:"type:varchar(32);not null;default:'token'"`
	TotalCapacity    int64  `json:"total_capacity" gorm:"type:bigint;not null;default:0"`
	SellableCapacity int64  `json:"sellable_capacity" gorm:"type:bigint;not null;default:0"`
	ReservedCapacity int64  `json:"reserved_capacity" gorm:"type:bigint;not null;default:0"`
	UsedCapacity     int64  `json:"used_capacity" gorm:"type:bigint;not null;default:0"`
	ExpiresAt        int64  `json:"expires_at" gorm:"type:bigint;not null;default:0"`
	VerifyStatus     string `json:"verify_status" gorm:"type:varchar(32);not null;default:'pending'"`
	LatestVerifiedAt int64  `json:"latest_verified_at" gorm:"type:bigint;not null;default:0"`
	Status           string `json:"status" gorm:"type:varchar(32);not null;default:'active';index:idx_supply_accounts_status"`
	VerifyMessage    string `json:"verify_message" gorm:"type:text"`
	Extra            string `json:"extra" gorm:"type:text"`
	CreatedAt        int64  `json:"created_at" gorm:"bigint"`
	UpdatedAt        int64  `json:"updated_at" gorm:"bigint"`
}

func (s *SupplyAccount) BeforeCreate(tx *gorm.DB) error {
	setMarketplaceTimestampsOnCreate(&s.CreatedAt, &s.UpdatedAt)
	return nil
}

func (s *SupplyAccount) BeforeUpdate(tx *gorm.DB) error {
	setMarketplaceTimestampOnUpdate(&s.UpdatedAt)
	return nil
}

type SupplyChannelBinding struct {
	Id              int    `json:"id"`
	SupplyAccountId int    `json:"supply_account_id" gorm:"not null;uniqueIndex:ux_supply_channel_bindings_supply_channel,priority:1;index:idx_supply_channel_bindings_supply_account_id"`
	ChannelId       int    `json:"channel_id" gorm:"not null;uniqueIndex:ux_supply_channel_bindings_supply_channel,priority:2;index:idx_supply_channel_bindings_channel_id"`
	BindingRole     string `json:"binding_role" gorm:"type:varchar(32);not null;default:'primary'"`
	Priority        int64  `json:"priority" gorm:"type:bigint;not null;default:0"`
	Status          string `json:"status" gorm:"type:varchar(32);not null;default:'active';index:idx_supply_channel_bindings_status"`
	CreatedAt       int64  `json:"created_at" gorm:"bigint"`
	UpdatedAt       int64  `json:"updated_at" gorm:"bigint"`
}

func (s *SupplyChannelBinding) BeforeCreate(tx *gorm.DB) error {
	setMarketplaceTimestampsOnCreate(&s.CreatedAt, &s.UpdatedAt)
	return nil
}

func (s *SupplyChannelBinding) BeforeUpdate(tx *gorm.DB) error {
	setMarketplaceTimestampOnUpdate(&s.UpdatedAt)
	return nil
}

func GetSupplyAccountByID(id int) (*SupplyAccount, error) {
	var supply SupplyAccount
	if err := DB.First(&supply, id).Error; err != nil {
		return nil, err
	}
	return &supply, nil
}

func CreateSupplyAccount(supply *SupplyAccount) error {
	return DB.Create(supply).Error
}

func UpdateSupplyAccountVerify(id int, verifyStatus string, latestVerifiedAt int64, verifyMessage string) error {
	return DB.Model(&SupplyAccount{}).Where("id = ?", id).Updates(map[string]interface{}{
		"verify_status":      verifyStatus,
		"latest_verified_at": latestVerifiedAt,
		"verify_message":     verifyMessage,
		"updated_at":         common.GetTimestamp(),
	}).Error
}

func UpdateSupplyAccountStatus(id int, status string) error {
	return DB.Model(&SupplyAccount{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":     status,
		"updated_at": common.GetTimestamp(),
	}).Error
}

func CreateSupplyChannelBindings(bindings []SupplyChannelBinding) error {
	if len(bindings) == 0 {
		return nil
	}
	return DB.Create(&bindings).Error
}
