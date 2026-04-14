package service

import (
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
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
