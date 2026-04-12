package model

import (
	"strings"

	"github.com/QuantumNous/new-api/common"
	"gorm.io/gorm"
)

func setMarketplaceTimestampsOnCreate(createdAt *int64, updatedAt *int64) {
	now := common.GetTimestamp()
	if createdAt != nil && *createdAt == 0 {
		*createdAt = now
	}
	if updatedAt != nil {
		*updatedAt = now
	}
}

func setMarketplaceTimestampOnUpdate(updatedAt *int64) {
	if updatedAt == nil {
		return
	}
	*updatedAt = common.GetTimestamp()
}

type SellerProfile struct {
	Id                  int    `json:"id"`
	UserId              int    `json:"user_id" gorm:"not null;uniqueIndex:ux_seller_profiles_user_id"`
	SellerCode          string `json:"seller_code" gorm:"type:varchar(64);not null;uniqueIndex:ux_seller_profiles_seller_code"`
	SellerType          string `json:"seller_type" gorm:"type:varchar(32);not null;default:'personal'"`
	DisplayName         string `json:"display_name" gorm:"type:varchar(128);not null"`
	ContactEmail        string `json:"contact_email" gorm:"type:varchar(255)"`
	ContactPhone        string `json:"contact_phone" gorm:"type:varchar(64)"`
	CompanyName         string `json:"company_name" gorm:"type:varchar(255)"`
	LicenseNo           string `json:"license_no" gorm:"type:varchar(128)"`
	KycStatus           string `json:"kyc_status" gorm:"type:varchar(32);not null;default:'pending';index:idx_seller_profiles_kyc_status"`
	KycLevel            string `json:"kyc_level" gorm:"type:varchar(32);not null;default:'none'"`
	RiskLevel           string `json:"risk_level" gorm:"type:varchar(32);not null;default:'low';index:idx_seller_profiles_risk_level"`
	SettlementStatus    string `json:"settlement_status" gorm:"type:varchar(32);not null;default:'pending'"`
	DepositBalanceMinor int64  `json:"deposit_balance_minor" gorm:"type:bigint;not null;default:0"`
	FrozenDepositMinor  int64  `json:"frozen_deposit_minor" gorm:"type:bigint;not null;default:0"`
	Status              string `json:"status" gorm:"type:varchar(32);not null;default:'active';index:idx_seller_profiles_status"`
	Remark              string `json:"remark" gorm:"type:text"`
	Extra               string `json:"extra" gorm:"type:text"`
	CreatedAt           int64  `json:"created_at" gorm:"bigint"`
	UpdatedAt           int64  `json:"updated_at" gorm:"bigint"`
}

func (s *SellerProfile) BeforeCreate(tx *gorm.DB) error {
	setMarketplaceTimestampsOnCreate(&s.CreatedAt, &s.UpdatedAt)
	return nil
}

func (s *SellerProfile) BeforeUpdate(tx *gorm.DB) error {
	setMarketplaceTimestampOnUpdate(&s.UpdatedAt)
	return nil
}

func GetSellerByID(id int) (*SellerProfile, error) {
	var seller SellerProfile
	if err := DB.First(&seller, id).Error; err != nil {
		return nil, err
	}
	return &seller, nil
}

func GetSellers(keyword string, status string, offset int, limit int) ([]*SellerProfile, int64, error) {
	db := DB.Model(&SellerProfile{})
	if trimmed := strings.TrimSpace(status); trimmed != "" {
		db = db.Where("status = ?", trimmed)
	}
	if trimmed := strings.TrimSpace(keyword); trimmed != "" {
		like := "%" + trimmed + "%"
		db = db.Where(
			"display_name LIKE ? OR seller_code LIKE ? OR contact_email LIKE ? OR company_name LIKE ?",
			like, like, like, like,
		)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var sellers []*SellerProfile
	if err := db.Order("id desc").Offset(offset).Limit(limit).Find(&sellers).Error; err != nil {
		return nil, 0, err
	}
	return sellers, total, nil
}

func CreateSeller(seller *SellerProfile) error {
	return DB.Create(seller).Error
}

func UpdateSellerStatus(id int, status string, remark string) error {
	updates := map[string]interface{}{
		"status":     status,
		"remark":     remark,
		"updated_at": common.GetTimestamp(),
	}
	return DB.Model(&SellerProfile{}).Where("id = ?", id).Updates(updates).Error
}
