package model

import (
	"strings"

	"github.com/QuantumNous/new-api/common"
	"gorm.io/gorm"
)

type Listing struct {
	Id              int    `json:"id"`
	SellerId        int    `json:"seller_id" gorm:"not null;index:idx_listings_seller_id"`
	SupplyAccountId int    `json:"supply_account_id" gorm:"not null;index:idx_listings_supply_account_id"`
	ListingCode     string `json:"listing_code" gorm:"type:varchar(64);not null;uniqueIndex:ux_listings_listing_code"`
	Title           string `json:"title" gorm:"type:varchar(255);not null"`
	VendorId        int    `json:"vendor_id" gorm:"not null;default:0;index:idx_listings_vendor_id"`
	ModelName       string `json:"model_name" gorm:"type:varchar(128);not null;index:idx_listings_model_name"`
	SaleMode        string `json:"sale_mode" gorm:"type:varchar(32);not null;default:'fixed_price'"`
	PricingUnit     string `json:"pricing_unit" gorm:"type:varchar(32);not null;default:'per_token_package'"`
	GuaranteeLevel  string `json:"guarantee_level" gorm:"type:varchar(32);not null;default:'basic'"`
	ValidityDays    int    `json:"validity_days" gorm:"not null;default:0"`
	Description     string `json:"description" gorm:"type:text"`
	AuditStatus     string `json:"audit_status" gorm:"type:varchar(32);not null;default:'draft';index:idx_listings_audit_status"`
	Status          string `json:"status" gorm:"type:varchar(32);not null;default:'paused';index:idx_listings_status"`
	AuditRemark     string `json:"audit_remark" gorm:"type:text"`
	PublishedAt     int64  `json:"published_at" gorm:"type:bigint;not null;default:0"`
	CreatedAt       int64  `json:"created_at" gorm:"bigint"`
	UpdatedAt       int64  `json:"updated_at" gorm:"bigint"`
}

func (l *Listing) BeforeCreate(tx *gorm.DB) error {
	setMarketplaceTimestampsOnCreate(&l.CreatedAt, &l.UpdatedAt)
	return nil
}

func (l *Listing) BeforeUpdate(tx *gorm.DB) error {
	setMarketplaceTimestampOnUpdate(&l.UpdatedAt)
	return nil
}

type ListingSKU struct {
	Id                 int    `json:"id"`
	ListingId          int    `json:"listing_id" gorm:"not null;index:idx_listing_skus_listing_id"`
	SkuCode            string `json:"sku_code" gorm:"type:varchar(64);not null;uniqueIndex:ux_listing_skus_sku_code"`
	PackageAmount      int64  `json:"package_amount" gorm:"type:bigint;not null;default:0"`
	PackageUnit        string `json:"package_unit" gorm:"type:varchar(32);not null;default:'token'"`
	UnitPriceMinor     int64  `json:"unit_price_minor" gorm:"type:bigint;not null;default:0"`
	OriginalPriceMinor int64  `json:"original_price_minor" gorm:"type:bigint;not null;default:0"`
	MinQuantity        int    `json:"min_quantity" gorm:"not null;default:1"`
	MaxQuantity        int    `json:"max_quantity" gorm:"not null;default:1"`
	Status             string `json:"status" gorm:"type:varchar(32);not null;default:'active';index:idx_listing_skus_status"`
	SortOrder          int    `json:"sort_order" gorm:"not null;default:0"`
	CreatedAt          int64  `json:"created_at" gorm:"bigint"`
	UpdatedAt          int64  `json:"updated_at" gorm:"bigint"`
}

func (l *ListingSKU) BeforeCreate(tx *gorm.DB) error {
	setMarketplaceTimestampsOnCreate(&l.CreatedAt, &l.UpdatedAt)
	return nil
}

func (l *ListingSKU) BeforeUpdate(tx *gorm.DB) error {
	setMarketplaceTimestampOnUpdate(&l.UpdatedAt)
	return nil
}

type InventorySnapshot struct {
	Id              int    `json:"id"`
	SupplyAccountId int    `json:"supply_account_id" gorm:"not null;uniqueIndex:ux_inventory_snapshots_supply_account_id"`
	AvailableAmount int64  `json:"available_amount" gorm:"type:bigint;not null;default:0"`
	FrozenAmount    int64  `json:"frozen_amount" gorm:"type:bigint;not null;default:0"`
	SoldAmount      int64  `json:"sold_amount" gorm:"type:bigint;not null;default:0"`
	ConsumedAmount  int64  `json:"consumed_amount" gorm:"type:bigint;not null;default:0"`
	RiskDiscountBps int    `json:"risk_discount_bps" gorm:"not null;default:10000"`
	HealthScore     int    `json:"health_score" gorm:"not null;default:100"`
	SyncStatus      string `json:"sync_status" gorm:"type:varchar(32);not null;default:'ok';index:idx_inventory_snapshots_sync_status"`
	SyncMessage     string `json:"sync_message" gorm:"type:text"`
	LastSyncAt      int64  `json:"last_sync_at" gorm:"type:bigint;not null;default:0;index:idx_inventory_snapshots_last_sync_at"`
	CreatedAt       int64  `json:"created_at" gorm:"bigint"`
	UpdatedAt       int64  `json:"updated_at" gorm:"bigint"`
}

func (i *InventorySnapshot) BeforeCreate(tx *gorm.DB) error {
	setMarketplaceTimestampsOnCreate(&i.CreatedAt, &i.UpdatedAt)
	return nil
}

func (i *InventorySnapshot) BeforeUpdate(tx *gorm.DB) error {
	setMarketplaceTimestampOnUpdate(&i.UpdatedAt)
	return nil
}

func GetListingByID(id int) (*Listing, error) {
	var listing Listing
	if err := DB.First(&listing, id).Error; err != nil {
		return nil, err
	}
	return &listing, nil
}

func GetListings(keyword string, status string, auditStatus string, sellerId int, offset int, limit int) ([]*Listing, int64, error) {
	db := DB.Model(&Listing{})
	if sellerId > 0 {
		db = db.Where("seller_id = ?", sellerId)
	}
	if trimmed := strings.TrimSpace(status); trimmed != "" {
		db = db.Where("status = ?", trimmed)
	}
	if trimmed := strings.TrimSpace(auditStatus); trimmed != "" {
		db = db.Where("audit_status = ?", trimmed)
	}
	if trimmed := strings.TrimSpace(keyword); trimmed != "" {
		like := "%" + trimmed + "%"
		db = db.Where(
			"title LIKE ? OR listing_code LIKE ? OR model_name LIKE ?",
			like, like, like,
		)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var listings []*Listing
	if err := db.Order("id desc").Offset(offset).Limit(limit).Find(&listings).Error; err != nil {
		return nil, 0, err
	}
	return listings, total, nil
}

func CreateListing(listing *Listing) error {
	return DB.Create(listing).Error
}

func GetListingSKUsByListingID(listingId int) ([]ListingSKU, error) {
	var skus []ListingSKU
	if err := DB.Where("listing_id = ?", listingId).Order("sort_order asc, id asc").Find(&skus).Error; err != nil {
		return nil, err
	}
	return skus, nil
}

func GetListingSKUByID(id int) (*ListingSKU, error) {
	var sku ListingSKU
	if err := DB.First(&sku, id).Error; err != nil {
		return nil, err
	}
	return &sku, nil
}

func CreateListingSKUs(skus []ListingSKU) error {
	if len(skus) == 0 {
		return nil
	}
	return DB.Create(&skus).Error
}

func UpdateListingStatus(id int, status string, auditStatus string, auditRemark string) error {
	updates := map[string]interface{}{
		"updated_at": common.GetTimestamp(),
	}
	if status != "" {
		updates["status"] = status
	}
	if auditStatus != "" {
		updates["audit_status"] = auditStatus
	}
	if auditRemark != "" {
		updates["audit_remark"] = auditRemark
	}
	return DB.Model(&Listing{}).Where("id = ?", id).Updates(updates).Error
}

func GetInventorySnapshotBySupplyAccountID(supplyAccountId int) (*InventorySnapshot, error) {
	var snapshot InventorySnapshot
	if err := DB.Where("supply_account_id = ?", supplyAccountId).First(&snapshot).Error; err != nil {
		return nil, err
	}
	return &snapshot, nil
}

func CreateInventorySnapshot(snapshot *InventorySnapshot) error {
	return DB.Create(snapshot).Error
}
