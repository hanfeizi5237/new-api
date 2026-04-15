package service

import (
	"errors"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"gorm.io/gorm"
)

type CreateListingInput struct {
	Listing model.Listing
	SKUs    []model.ListingSKU
}

type MarketListingDetail struct {
	Listing   *model.Listing            `json:"listing"`
	SKUs      []model.ListingSKU        `json:"skus"`
	Inventory *model.InventorySnapshot  `json:"inventory,omitempty"`
}

func ListListings(keyword string, status string, auditStatus string, sellerId int, offset int, limit int) ([]*model.Listing, int64, error) {
	return model.GetListings(keyword, status, auditStatus, sellerId, offset, limit)
}

func ListPublicMarketListings(keyword string, offset int, limit int) ([]*model.Listing, int64, error) {
	return model.GetListings(keyword, "active", "approved", 0, offset, limit)
}

func GetPublicMarketListingDetail(id int) (*MarketListingDetail, error) {
	if id <= 0 {
		return nil, errors.New("invalid listing id")
	}
	listing, err := model.GetListingByID(id)
	if err != nil {
		return nil, err
	}
	if listing.Status != "active" || listing.AuditStatus != "approved" {
		return nil, errors.New("listing is not available")
	}
	skus, err := model.GetListingSKUsByListingID(listing.Id)
	if err != nil {
		return nil, err
	}
	filteredSKUs := make([]model.ListingSKU, 0, len(skus))
	for _, sku := range skus {
		if sku.Status == "active" {
			filteredSKUs = append(filteredSKUs, sku)
		}
	}
	snapshot, err := model.GetInventorySnapshotBySupplyAccountID(listing.SupplyAccountId)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		snapshot = nil
	}
	return &MarketListingDetail{
		Listing:   listing,
		SKUs:      filteredSKUs,
		Inventory: snapshot,
	}, nil
}

func CreateListingWithSKUs(input CreateListingInput) (*model.Listing, []model.ListingSKU, error) {
	listing := input.Listing
	if listing.SellerId <= 0 || listing.SupplyAccountId <= 0 {
		return nil, nil, errors.New("listing.seller_id and listing.supply_account_id are required")
	}
	if strings.TrimSpace(listing.ListingCode) == "" {
		return nil, nil, errors.New("listing.listing_code is required")
	}
	if strings.TrimSpace(listing.Title) == "" {
		return nil, nil, errors.New("listing.title is required")
	}
	if len(input.SKUs) == 0 {
		return nil, nil, errors.New("at least one sku is required")
	}

	seller, err := model.GetSellerByID(listing.SellerId)
	if err != nil {
		return nil, nil, err
	}
	if seller.Status != "active" {
		return nil, nil, errors.New("seller is not active")
	}
	supply, err := model.GetSupplyAccountByID(listing.SupplyAccountId)
	if err != nil {
		return nil, nil, err
	}
	if supply.SellerId != seller.Id {
		return nil, nil, errors.New("supply_account does not belong to seller")
	}
	if supply.Status != "active" || supply.VerifyStatus != "success" {
		return nil, nil, errors.New("supply_account is not ready for listing")
	}
	if supply.QuotaUnit != "token" {
		return nil, nil, errors.New("M1 only supports token-based supply units")
	}

	var activeSecretCount int64
	if err := model.DB.Model(&model.SellerSecret{}).
		Where("supply_account_id = ? AND status = ? AND verify_status = ?", supply.Id, "active", "success").
		Count(&activeSecretCount).Error; err != nil {
		return nil, nil, err
	}
	if activeSecretCount == 0 {
		return nil, nil, errors.New("supply_account has no active verified seller secret")
	}

	var activeBindings []model.SupplyChannelBinding
	if err := model.DB.
		Where("supply_account_id = ? AND status = ?", supply.Id, "active").
		Find(&activeBindings).Error; err != nil {
		return nil, nil, err
	}
	if len(activeBindings) == 0 {
		return nil, nil, errors.New("supply_account has no active channel binding")
	}
	for _, binding := range activeBindings {
		channel, err := model.GetChannelById(binding.ChannelId, true)
		if err != nil {
			return nil, nil, err
		}
		if channel.Status != common.ChannelStatusEnabled {
			return nil, nil, errors.New("supply_account has inactive channel binding")
		}
		if !containsExactModel(channel.GetModels(), supply.ModelName) {
			return nil, nil, errors.New("supply_account channel binding does not support supply model")
		}
	}

	if listing.VendorId != 0 && listing.VendorId != supply.VendorId {
		return nil, nil, errors.New("listing.vendor_id must match supply_account.vendor_id")
	}
	if listing.ModelName != "" && listing.ModelName != supply.ModelName {
		return nil, nil, errors.New("listing.model_name must match supply_account.model_name")
	}
	listing.VendorId = supply.VendorId
	listing.ModelName = supply.ModelName
	if listing.SaleMode == "" {
		listing.SaleMode = "fixed_price"
	}
	if listing.PricingUnit == "" {
		listing.PricingUnit = "per_token_package"
	}

	skus := make([]model.ListingSKU, len(input.SKUs))
	copy(skus, input.SKUs)
	for i := range skus {
		if strings.TrimSpace(skus[i].SkuCode) == "" {
			return nil, nil, errors.New("sku_code is required")
		}
		if skus[i].PackageAmount <= 0 {
			return nil, nil, errors.New("package_amount must be greater than 0")
		}
		if skus[i].UnitPriceMinor < 0 || skus[i].OriginalPriceMinor < 0 {
			return nil, nil, errors.New("price must be non-negative")
		}
		if skus[i].MinQuantity <= 0 {
			skus[i].MinQuantity = 1
		}
		if skus[i].MaxQuantity < skus[i].MinQuantity {
			return nil, nil, errors.New("max_quantity must be greater than or equal to min_quantity")
		}
		if skus[i].PackageUnit == "" {
			skus[i].PackageUnit = supply.QuotaUnit
		}
		if skus[i].PackageUnit != "token" {
			return nil, nil, errors.New("M1 only supports token-based sku units")
		}
	}

	err = model.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&listing).Error; err != nil {
			return err
		}
		for i := range skus {
			skus[i].ListingId = listing.Id
		}
		if err := tx.Create(&skus).Error; err != nil {
			return err
		}
		var snapshot model.InventorySnapshot
		err := tx.Where("supply_account_id = ?", supply.Id).First(&snapshot).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			snapshot = model.InventorySnapshot{
				SupplyAccountId: supply.Id,
				AvailableAmount: supply.SellableCapacity,
				RiskDiscountBps: 10000,
				HealthScore:     100,
				SyncStatus:      "ok",
			}
			return tx.Create(&snapshot).Error
		}
		return err
	})
	if err != nil {
		return nil, nil, err
	}
	return &listing, skus, nil
}

func UpdateListingStatus(id int, status string, auditStatus string, auditRemark string) error {
	if id <= 0 {
		return errors.New("invalid listing id")
	}
	if strings.TrimSpace(status) == "" && strings.TrimSpace(auditStatus) == "" && strings.TrimSpace(auditRemark) == "" {
		return errors.New("at least one of status, audit_status or audit_remark is required")
	}
	listing, err := model.GetListingByID(id)
	if err != nil {
		return err
	}
	nextAuditStatus, err := resolveNextListingAuditStatus(listing.AuditStatus, auditStatus)
	if err != nil {
		return err
	}
	nextStatus, err := resolveNextListingStatus(listing.Status, nextAuditStatus, status)
	if err != nil {
		return err
	}
	return model.UpdateListingStatus(id, nextStatus, nextAuditStatus, auditRemark)
}

func resolveNextListingAuditStatus(current string, requested string) (string, error) {
	requested = strings.TrimSpace(requested)
	if requested == "" {
		return current, nil
	}
	switch requested {
	case "draft", "pending_review", "approved", "rejected":
	default:
		return "", errors.New("invalid listing audit_status")
	}
	switch current {
	case "", "draft":
		if requested == "pending_review" || requested == "draft" {
			return requested, nil
		}
	case "pending_review":
		if requested == "approved" || requested == "rejected" || requested == "pending_review" {
			return requested, nil
		}
	case "approved":
		if requested == "approved" {
			return requested, nil
		}
	case "rejected":
		if requested == "rejected" || requested == "pending_review" {
			return requested, nil
		}
	}
	return "", errors.New("invalid listing audit_status transition")
}

func resolveNextListingStatus(current string, auditStatus string, requested string) (string, error) {
	requested = strings.TrimSpace(requested)
	if requested == "" {
		return current, nil
	}
	switch requested {
	case "paused", "active", "sold_out", "archived":
	default:
		return "", errors.New("invalid listing status")
	}
	if current == "archived" && requested != "archived" {
		return "", errors.New("archived listing cannot transition to another status")
	}
	if requested == "active" || requested == "sold_out" {
		if auditStatus != "approved" {
			return "", errors.New("listing must be audit approved before becoming sellable")
		}
	}
	return requested, nil
}
