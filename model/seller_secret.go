package model

import (
	"strings"

	"github.com/QuantumNous/new-api/common"
	"gorm.io/gorm"
)

type SellerSecret struct {
	Id              int    `json:"id"`
	SellerId        int    `json:"seller_id" gorm:"not null;index:idx_seller_secrets_seller_id"`
	SupplyAccountId int    `json:"supply_account_id" gorm:"not null;uniqueIndex:ux_seller_secrets_supply_fingerprint,priority:1;index:idx_seller_secrets_supply_account_id"`
	SecretType      string `json:"secret_type" gorm:"type:varchar(32);not null;default:'api_key'"`
	ProviderCode    string `json:"provider_code" gorm:"type:varchar(64);not null;index:idx_seller_secrets_provider_code"`
	Ciphertext      string `json:"ciphertext" gorm:"type:text;not null"`
	CipherVersion   string `json:"cipher_version" gorm:"type:varchar(32);not null;default:'v1'"`
	Fingerprint     string `json:"fingerprint" gorm:"type:varchar(128);not null;uniqueIndex:ux_seller_secrets_supply_fingerprint,priority:2"`
	MaskedValue     string `json:"masked_value" gorm:"type:varchar(255)"`
	Status          string `json:"status" gorm:"type:varchar(32);not null;default:'draft';index:idx_seller_secrets_status"`
	VerifyStatus    string `json:"verify_status" gorm:"type:varchar(32);not null;default:'pending';index:idx_seller_secrets_verify_status"`
	LastVerifiedAt  int64  `json:"last_verified_at" gorm:"type:bigint;not null;default:0"`
	LastUsedAt      int64  `json:"last_used_at" gorm:"type:bigint;not null;default:0"`
	LastRotationAt  int64  `json:"last_rotation_at" gorm:"type:bigint;not null;default:0"`
	ExpiresAt       int64  `json:"expires_at" gorm:"type:bigint;not null;default:0"`
	DisabledReason  string `json:"disabled_reason" gorm:"type:text"`
	VerifyMessage   string `json:"verify_message" gorm:"type:text"`
	Meta            string `json:"meta" gorm:"type:text"`
	CreatedAt       int64  `json:"created_at" gorm:"bigint"`
	UpdatedAt       int64  `json:"updated_at" gorm:"bigint"`
}

func (s *SellerSecret) BeforeCreate(tx *gorm.DB) error {
	setMarketplaceTimestampsOnCreate(&s.CreatedAt, &s.UpdatedAt)
	return nil
}

func (s *SellerSecret) BeforeUpdate(tx *gorm.DB) error {
	setMarketplaceTimestampOnUpdate(&s.UpdatedAt)
	return nil
}

type SellerSecretAudit struct {
	Id              int    `json:"id"`
	SellerSecretId  int    `json:"seller_secret_id" gorm:"not null;index:idx_seller_secret_audits_secret_id"`
	SellerId        int    `json:"seller_id" gorm:"not null;index:idx_seller_secret_audits_seller_id"`
	SupplyAccountId int    `json:"supply_account_id" gorm:"not null;index:idx_seller_secret_audits_supply_account_id"`
	ActorUserId     int    `json:"actor_user_id" gorm:"not null;default:0;index:idx_seller_secret_audits_actor_user_id"`
	ActorType       string `json:"actor_type" gorm:"type:varchar(32);not null;default:'system'"`
	Action          string `json:"action" gorm:"type:varchar(64);not null;index:idx_seller_secret_audits_action"`
	Reason          string `json:"reason" gorm:"type:text"`
	RequestId       string `json:"request_id" gorm:"type:varchar(64);index:idx_seller_secret_audits_request_id"`
	Ip              string `json:"ip" gorm:"type:varchar(64)"`
	Result          string `json:"result" gorm:"type:varchar(32);not null;default:'success'"`
	Meta            string `json:"meta" gorm:"type:text"`
	CreatedAt       int64  `json:"created_at" gorm:"bigint;index:idx_seller_secret_audits_created_at"`
	UpdatedAt       int64  `json:"updated_at" gorm:"bigint"`
}

func (s *SellerSecretAudit) BeforeCreate(tx *gorm.DB) error {
	setMarketplaceTimestampsOnCreate(&s.CreatedAt, &s.UpdatedAt)
	return nil
}

func (s *SellerSecretAudit) BeforeUpdate(tx *gorm.DB) error {
	setMarketplaceTimestampOnUpdate(&s.UpdatedAt)
	return nil
}

func GetSellerSecretByID(id int) (*SellerSecret, error) {
	var secret SellerSecret
	if err := DB.First(&secret, id).Error; err != nil {
		return nil, err
	}
	return &secret, nil
}

func GetSellerSecrets(sellerId int, supplyAccountId int, status string, offset int, limit int) ([]*SellerSecret, int64, error) {
	db := DB.Model(&SellerSecret{})
	if sellerId > 0 {
		db = db.Where("seller_id = ?", sellerId)
	}
	if supplyAccountId > 0 {
		db = db.Where("supply_account_id = ?", supplyAccountId)
	}
	if trimmed := strings.TrimSpace(status); trimmed != "" {
		db = db.Where("status = ?", trimmed)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var secrets []*SellerSecret
	if err := db.Order("id desc").Offset(offset).Limit(limit).Find(&secrets).Error; err != nil {
		return nil, 0, err
	}
	return secrets, total, nil
}

func CreateSellerSecret(secret *SellerSecret) error {
	return DB.Create(secret).Error
}

func UpdateSellerSecret(id int, updates map[string]interface{}) error {
	updates["updated_at"] = common.GetTimestamp()
	return DB.Model(&SellerSecret{}).Where("id = ?", id).Updates(updates).Error
}

func CreateSellerSecretAudit(audit *SellerSecretAudit) error {
	return DB.Create(audit).Error
}
