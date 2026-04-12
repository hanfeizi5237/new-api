package service

import (
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
)

func TestCreateSellerWithSupplyRejectsInvalidChannelBinding(t *testing.T) {
	db := setupMarketplaceServiceTestDB(t)
	user := seedMarketplaceServiceUser(t, db, "svc-seller-binding")

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
