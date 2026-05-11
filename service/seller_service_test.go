package service

import (
	"errors"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"gorm.io/gorm"
)

func TestCreateSellerWithSupplyRejectsInvalidChannelBinding(t *testing.T) {
	db := setupMarketplaceServiceTestDB(t)
	user := seedMarketplaceServiceUser(t, db, "svc-seller-binding")
	vendor := &model.Vendor{Name: "vendor-binding", Status: 1}
	if err := db.Create(vendor).Error; err != nil {
		t.Fatalf("failed to create vendor: %v", err)
	}

	if _, _, err := CreateSellerWithSupply(CreateSellerInput{
		Seller: model.SellerProfile{
			UserId:      user.Id,
			SellerCode:  "seller-binding-001",
			DisplayName: "Seller Binding",
			Status:      "active",
		},
		SupplyAccount: model.SupplyAccount{
			SupplyCode:       "supply-binding-001",
			ProviderCode:     "openai",
			VendorId:         vendor.Id,
			ModelName:        "gpt-4o-mini",
			QuotaUnit:        "token",
			TotalCapacity:    100000,
			SellableCapacity: 80000,
			Status:           "active",
		},
		Bindings: []model.SupplyChannelBinding{
			{
				ChannelId:   99999,
				BindingRole: "primary",
				Status:      "active",
			},
		},
	}); err == nil {
		t.Fatalf("expected invalid channel binding to be rejected")
	}

	channel := &model.Channel{
		Name:   "wrong-model-channel",
		Key:    "sk-test",
		Status: common.ChannelStatusEnabled,
		Models: "gpt-3.5-turbo",
		Group:  "default",
	}
	if err := db.Create(channel).Error; err != nil {
		t.Fatalf("failed to create channel: %v", err)
	}
	if _, _, err := CreateSellerWithSupply(CreateSellerInput{
		Seller: model.SellerProfile{
			UserId:      user.Id,
			SellerCode:  "seller-binding-002",
			DisplayName: "Seller Binding 2",
			Status:      "active",
		},
		SupplyAccount: model.SupplyAccount{
			SupplyCode:       "supply-binding-002",
			ProviderCode:     "openai",
			VendorId:         vendor.Id,
			ModelName:        "gpt-4o-mini",
			QuotaUnit:        "token",
			TotalCapacity:    100000,
			SellableCapacity: 80000,
			Status:           "active",
		},
		Bindings: []model.SupplyChannelBinding{
			{
				ChannelId:   channel.Id,
				BindingRole: "primary",
				Status:      "active",
			},
		},
	}); err == nil {
		t.Fatalf("expected mismatched channel model binding to be rejected")
	}
}

func TestCreateSellerWithSupplyRequiresActiveVendor(t *testing.T) {
	db := setupMarketplaceServiceTestDB(t)
	user := seedMarketplaceServiceUser(t, db, "svc-seller-vendor")

	if _, _, err := CreateSellerWithSupply(CreateSellerInput{
		Seller: model.SellerProfile{
			UserId:      user.Id,
			SellerCode:  "seller-vendor-001",
			DisplayName: "Seller Vendor",
			Status:      "active",
		},
		SupplyAccount: model.SupplyAccount{
			SupplyCode:       "supply-vendor-001",
			ProviderCode:     "openai",
			ModelName:        "gpt-4o-mini",
			QuotaUnit:        "token",
			TotalCapacity:    100000,
			SellableCapacity: 80000,
		},
	}); err == nil {
		t.Fatalf("expected missing vendor_id to be rejected")
	}

	disabledVendor := &model.Vendor{Name: "vendor-disabled"}
	if err := db.Create(disabledVendor).Error; err != nil {
		t.Fatalf("failed to create disabled vendor: %v", err)
	}
	if err := db.Model(&model.Vendor{}).Where("id = ?", disabledVendor.Id).Update("status", 0).Error; err != nil {
		t.Fatalf("failed to disable vendor: %v", err)
	}

	if _, _, err := CreateSellerWithSupply(CreateSellerInput{
		Seller: model.SellerProfile{
			UserId:      user.Id,
			SellerCode:  "seller-vendor-002",
			DisplayName: "Seller Vendor 2",
			Status:      "active",
		},
		SupplyAccount: model.SupplyAccount{
			SupplyCode:       "supply-vendor-002",
			ProviderCode:     "openai",
			VendorId:         disabledVendor.Id,
			ModelName:        "gpt-4o-mini",
			QuotaUnit:        "token",
			TotalCapacity:    100000,
			SellableCapacity: 80000,
		},
	}); err == nil {
		t.Fatalf("expected disabled vendor to be rejected")
	}
}

func TestUpdateSellerStatusCreatesTransactionalAudit(t *testing.T) {
	db := setupMarketplaceServiceTestDB(t)
	user := seedMarketplaceServiceUser(t, db, "svc-seller-audit")
	seller, _ := seedMarketplaceServiceSupply(t, db, user, "active", "success", "token")

	err := UpdateSellerStatus(seller.Id, "disabled", "risk-review", MarketplaceAuditActor{
		ActorUserID: user.Id,
		ActorType:   "admin",
		RequestID:   "req-seller-audit-001",
		IP:          "127.0.0.1",
	})
	if err != nil {
		t.Fatalf("expected seller status update success, got %v", err)
	}

	updated, err := model.GetSellerByID(seller.Id)
	if err != nil {
		t.Fatalf("failed to reload seller: %v", err)
	}
	if updated.Status != "disabled" || updated.Remark != "risk-review" {
		t.Fatalf("expected updated seller status/remark, got %+v", updated)
	}

	var audits []model.MarketplaceOperationAudit
	if err := db.Order("id asc").Find(&audits).Error; err != nil {
		t.Fatalf("failed to load seller audits: %v", err)
	}
	if len(audits) != 1 {
		t.Fatalf("expected exactly one seller audit, got %d", len(audits))
	}
	if audits[0].Action != "seller_status_update" || audits[0].TargetId != seller.Id {
		t.Fatalf("expected seller audit action/target, got %+v", audits[0])
	}
}

func TestUpdateSellerStatusAuditSeesUpdatedStateWithinSameTransaction(t *testing.T) {
	db := setupMarketplaceServiceTestDB(t)
	user := seedMarketplaceServiceUser(t, db, "svc-seller-audit-tx")
	seller, _ := seedMarketplaceServiceSupply(t, db, user, "active", "success", "token")

	previousWriter := marketplaceOperationAuditWriter
	marketplaceOperationAuditWriter = func(tx *gorm.DB, audit *model.MarketplaceOperationAudit) error {
		var reloaded model.SellerProfile
		if err := tx.First(&reloaded, seller.Id).Error; err != nil {
			return err
		}
		if reloaded.Status != "disabled" || reloaded.Remark != "risk-review" {
			return errors.New("seller update not visible inside audit transaction")
		}
		return previousWriter(tx, audit)
	}
	t.Cleanup(func() {
		marketplaceOperationAuditWriter = previousWriter
	})

	err := UpdateSellerStatus(seller.Id, "disabled", "risk-review", MarketplaceAuditActor{
		ActorUserID: user.Id,
		ActorType:   "admin",
		RequestID:   "req-seller-audit-tx-001",
		IP:          "127.0.0.1",
	})
	if err != nil {
		t.Fatalf("expected seller status update success, got %v", err)
	}
}

func TestUpdateSellerStatusRollsBackWhenAuditInsertFails(t *testing.T) {
	db := setupMarketplaceServiceTestDB(t)
	user := seedMarketplaceServiceUser(t, db, "svc-seller-audit-rollback")
	seller, _ := seedMarketplaceServiceSupply(t, db, user, "active", "success", "token")

	previousWriter := marketplaceOperationAuditWriter
	marketplaceOperationAuditWriter = func(tx *gorm.DB, audit *model.MarketplaceOperationAudit) error {
		return errors.New("forced seller audit failure")
	}
	t.Cleanup(func() {
		marketplaceOperationAuditWriter = previousWriter
	})

	err := UpdateSellerStatus(seller.Id, "disabled", "risk-review", MarketplaceAuditActor{
		ActorUserID: user.Id,
		ActorType:   "admin",
		RequestID:   "req-seller-audit-rollback-001",
		IP:          "127.0.0.1",
	})
	if err == nil {
		t.Fatalf("expected seller status update to fail when audit insert fails")
	}

	reloaded, err := model.GetSellerByID(seller.Id)
	if err != nil {
		t.Fatalf("failed to reload seller after rollback: %v", err)
	}
	if reloaded.Status != "active" || reloaded.Remark != "" {
		t.Fatalf("expected seller state rollback to active/empty remark, got %+v", reloaded)
	}

	var count int64
	if err := db.Model(&model.MarketplaceOperationAudit{}).Count(&count).Error; err != nil {
		t.Fatalf("failed to count seller audits: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected zero seller audits after rollback, got %d", count)
	}
}
