package service

import (
	"errors"
	"testing"

	"github.com/QuantumNous/new-api/model"
	"gorm.io/gorm"
)

func TestCreateListingWithSKUsRejectsUnreadySupply(t *testing.T) {
	db := setupMarketplaceServiceTestDB(t)
	user := seedMarketplaceServiceUser(t, db, "svc-listing-unready")
	seller, supply := seedMarketplaceServiceSupply(t, db, user, "paused", "pending", "token")

	_, _, err := CreateListingWithSKUs(CreateListingInput{
		Listing: model.Listing{
			SellerId:        seller.Id,
			SupplyAccountId: supply.Id,
			ListingCode:     "listing-unready-001",
			Title:           "Unready Listing",
		},
		SKUs: []model.ListingSKU{
			{
				SkuCode:        "sku-unready-001",
				PackageAmount:  1000,
				PackageUnit:    "token",
				UnitPriceMinor: 199,
				MinQuantity:    1,
				MaxQuantity:    2,
			},
		},
	})
	if err == nil {
		t.Fatalf("expected listing creation to reject unready supply")
	}
}

func TestCreateListingWithSKUsRejectsNonTokenPackageUnit(t *testing.T) {
	db := setupMarketplaceServiceTestDB(t)
	user := seedMarketplaceServiceUser(t, db, "svc-listing-unit")
	seller, supply := seedMarketplaceServiceSupply(t, db, user, "active", "success", "request")
	_, _ = seedMarketplaceChannelBinding(t, db, supply, "preloaded-runtime-key")
	_ = seedSellerSecretRecord(t, db, seller.Id, supply.Id, `{"alg":"aes-256-gcm","kid":"v1","nonce":"bm9uY2U=","ciphertext":"Y2lwaGVy"}`, "fp-token-unit", "active", "success")

	_, _, err := CreateListingWithSKUs(CreateListingInput{
		Listing: model.Listing{
			SellerId:        seller.Id,
			SupplyAccountId: supply.Id,
			ListingCode:     "listing-unit-001",
			Title:           "Wrong Unit Listing",
		},
		SKUs: []model.ListingSKU{
			{
				SkuCode:        "sku-unit-001",
				PackageAmount:  1000,
				PackageUnit:    "request",
				UnitPriceMinor: 199,
				MinQuantity:    1,
				MaxQuantity:    2,
			},
		},
	})
	if err == nil {
		t.Fatalf("expected listing creation to reject non-token unit supply")
	}
}

func TestCreateListingWithSKUsRejectsSupplyMismatchMetadata(t *testing.T) {
	db := setupMarketplaceServiceTestDB(t)
	user := seedMarketplaceServiceUser(t, db, "svc-listing-mismatch")
	seller, supply := seedMarketplaceServiceSupply(t, db, user, "active", "success", "token")
	_, _ = seedMarketplaceChannelBinding(t, db, supply, "preloaded-runtime-key")
	_ = seedSellerSecretRecord(t, db, seller.Id, supply.Id, `{"alg":"aes-256-gcm","kid":"v1","nonce":"bm9uY2U=","ciphertext":"Y2lwaGVy"}`, "fp-token-match", "active", "success")

	_, _, err := CreateListingWithSKUs(CreateListingInput{
		Listing: model.Listing{
			SellerId:        seller.Id,
			SupplyAccountId: supply.Id,
			ListingCode:     "listing-mismatch-001",
			Title:           "Mismatch Listing",
			VendorId:        99,
			ModelName:       "other-model",
		},
		SKUs: []model.ListingSKU{
			{
				SkuCode:        "sku-mismatch-001",
				PackageAmount:  1000,
				PackageUnit:    "token",
				UnitPriceMinor: 199,
				MinQuantity:    1,
				MaxQuantity:    2,
			},
		},
	})
	if err == nil {
		t.Fatalf("expected listing creation to reject mismatched vendor/model")
	}
}

func TestUpdateListingStatusCreatesTransactionalAudit(t *testing.T) {
	db := setupMarketplaceServiceTestDB(t)
	user := seedMarketplaceServiceUser(t, db, "svc-listing-audit")
	seller, supply := seedMarketplaceServiceSupply(t, db, user, "active", "success", "token")
	listing, _ := seedMarketplaceListing(t, db, seller, supply, 1000, 199)
	if err := db.Model(&model.Listing{}).Where("id = ?", listing.Id).Updates(map[string]interface{}{
		"status":       "paused",
		"audit_status": "pending_review",
		"audit_remark": "submit",
	}).Error; err != nil {
		t.Fatalf("failed to set listing pre-state: %v", err)
	}

	err := UpdateListingStatus(listing.Id, "active", "approved", "ready", MarketplaceAuditActor{
		ActorUserID: user.Id,
		ActorType:   "admin",
		RequestID:   "req-listing-audit-001",
		IP:          "127.0.0.1",
	})
	if err != nil {
		t.Fatalf("expected listing status update success, got %v", err)
	}

	updated, err := model.GetListingByID(listing.Id)
	if err != nil {
		t.Fatalf("failed to reload listing: %v", err)
	}
	if updated.Status != "active" || updated.AuditStatus != "approved" || updated.AuditRemark != "ready" {
		t.Fatalf("expected updated listing state, got %+v", updated)
	}

	var audits []model.MarketplaceOperationAudit
	if err := db.Order("id asc").Find(&audits).Error; err != nil {
		t.Fatalf("failed to load listing audits: %v", err)
	}
	if len(audits) != 1 {
		t.Fatalf("expected exactly one listing audit, got %d", len(audits))
	}
	if audits[0].Action != "listing_status_update" || audits[0].TargetId != listing.Id {
		t.Fatalf("expected listing audit action/target, got %+v", audits[0])
	}
}

func TestUpdateListingStatusAuditSeesUpdatedStateWithinSameTransaction(t *testing.T) {
	db := setupMarketplaceServiceTestDB(t)
	user := seedMarketplaceServiceUser(t, db, "svc-listing-audit-tx")
	seller, supply := seedMarketplaceServiceSupply(t, db, user, "active", "success", "token")
	listing, _ := seedMarketplaceListing(t, db, seller, supply, 1000, 199)
	if err := db.Model(&model.Listing{}).Where("id = ?", listing.Id).Updates(map[string]interface{}{
		"status":       "paused",
		"audit_status": "pending_review",
		"audit_remark": "submit",
	}).Error; err != nil {
		t.Fatalf("failed to set listing pre-state: %v", err)
	}

	previousWriter := marketplaceOperationAuditWriter
	marketplaceOperationAuditWriter = func(tx *gorm.DB, audit *model.MarketplaceOperationAudit) error {
		var reloaded model.Listing
		if err := tx.First(&reloaded, listing.Id).Error; err != nil {
			return err
		}
		if reloaded.Status != "active" || reloaded.AuditStatus != "approved" || reloaded.AuditRemark != "ready" {
			return errors.New("listing update not visible inside audit transaction")
		}
		return previousWriter(tx, audit)
	}
	t.Cleanup(func() {
		marketplaceOperationAuditWriter = previousWriter
	})

	err := UpdateListingStatus(listing.Id, "active", "approved", "ready", MarketplaceAuditActor{
		ActorUserID: user.Id,
		ActorType:   "admin",
		RequestID:   "req-listing-audit-tx-001",
		IP:          "127.0.0.1",
	})
	if err != nil {
		t.Fatalf("expected listing status update success, got %v", err)
	}
}

func TestUpdateListingStatusRollsBackWhenAuditInsertFails(t *testing.T) {
	db := setupMarketplaceServiceTestDB(t)
	user := seedMarketplaceServiceUser(t, db, "svc-listing-audit-rollback")
	seller, supply := seedMarketplaceServiceSupply(t, db, user, "active", "success", "token")
	listing, _ := seedMarketplaceListing(t, db, seller, supply, 1000, 199)
	if err := db.Model(&model.Listing{}).Where("id = ?", listing.Id).Updates(map[string]interface{}{
		"status":       "paused",
		"audit_status": "pending_review",
		"audit_remark": "submit",
	}).Error; err != nil {
		t.Fatalf("failed to set listing pre-state: %v", err)
	}

	previousWriter := marketplaceOperationAuditWriter
	marketplaceOperationAuditWriter = func(tx *gorm.DB, audit *model.MarketplaceOperationAudit) error {
		return errors.New("forced listing audit failure")
	}
	t.Cleanup(func() {
		marketplaceOperationAuditWriter = previousWriter
	})

	err := UpdateListingStatus(listing.Id, "active", "approved", "ready", MarketplaceAuditActor{
		ActorUserID: user.Id,
		ActorType:   "admin",
		RequestID:   "req-listing-audit-rollback-001",
		IP:          "127.0.0.1",
	})
	if err == nil {
		t.Fatalf("expected listing status update to fail when audit insert fails")
	}

	reloaded, err := model.GetListingByID(listing.Id)
	if err != nil {
		t.Fatalf("failed to reload listing after rollback: %v", err)
	}
	if reloaded.Status != "paused" || reloaded.AuditStatus != "pending_review" || reloaded.AuditRemark != "submit" {
		t.Fatalf("expected listing state rollback to paused/pending_review/submit, got %+v", reloaded)
	}

	var count int64
	if err := db.Model(&model.MarketplaceOperationAudit{}).Count(&count).Error; err != nil {
		t.Fatalf("failed to count listing audits: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected zero listing audits after rollback, got %d", count)
	}
}
