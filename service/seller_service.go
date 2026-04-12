package service

import (
	"errors"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"gorm.io/gorm"
)

type CreateSellerInput struct {
	Seller        model.SellerProfile
	SupplyAccount model.SupplyAccount
	Bindings      []model.SupplyChannelBinding
}

func ListSellers(keyword string, status string, offset int, limit int) ([]*model.SellerProfile, int64, error) {
	return model.GetSellers(keyword, status, offset, limit)
}

func CreateSellerWithSupply(input CreateSellerInput) (*model.SellerProfile, *model.SupplyAccount, error) {
	if input.Seller.UserId <= 0 {
		return nil, nil, errors.New("seller.user_id is required")
	}
	if strings.TrimSpace(input.Seller.SellerCode) == "" {
		return nil, nil, errors.New("seller.seller_code is required")
	}
	if strings.TrimSpace(input.Seller.DisplayName) == "" {
		return nil, nil, errors.New("seller.display_name is required")
	}
	if strings.TrimSpace(input.SupplyAccount.SupplyCode) == "" {
		return nil, nil, errors.New("supply_account.supply_code is required")
	}
	if strings.TrimSpace(input.SupplyAccount.ProviderCode) == "" {
		return nil, nil, errors.New("supply_account.provider_code is required")
	}
	if strings.TrimSpace(input.SupplyAccount.ModelName) == "" {
		return nil, nil, errors.New("supply_account.model_name is required")
	}
	if input.SupplyAccount.TotalCapacity < 0 || input.SupplyAccount.SellableCapacity < 0 {
		return nil, nil, errors.New("supply_account capacity must be non-negative")
	}
	if input.SupplyAccount.SellableCapacity > input.SupplyAccount.TotalCapacity {
		return nil, nil, errors.New("supply_account.sellable_capacity cannot exceed total_capacity")
	}
	if _, err := model.GetUserById(input.Seller.UserId, true); err != nil {
		return nil, nil, err
	}

	seller := input.Seller
	supply := input.SupplyAccount
	if supply.QuotaUnit == "" {
		supply.QuotaUnit = "token"
	}
	for _, binding := range input.Bindings {
		if binding.ChannelId <= 0 {
			return nil, nil, errors.New("binding.channel_id is required")
		}
		channel, err := model.GetChannelById(binding.ChannelId, true)
		if err != nil {
			return nil, nil, err
		}
		if channel.Status != common.ChannelStatusEnabled {
			return nil, nil, errors.New("binding channel is not enabled")
		}
		channelModels := channel.GetModels()
		if len(channelModels) == 0 || !containsExactModel(channelModels, supply.ModelName) {
			return nil, nil, errors.New("binding channel does not support supply model")
		}
	}

	err := model.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&seller).Error; err != nil {
			return err
		}

		supply.SellerId = seller.Id
		if err := tx.Create(&supply).Error; err != nil {
			return err
		}

		snapshot := model.InventorySnapshot{
			SupplyAccountId: supply.Id,
			AvailableAmount: supply.SellableCapacity,
			RiskDiscountBps: 10000,
			HealthScore:     100,
			SyncStatus:      "ok",
		}
		if err := tx.Create(&snapshot).Error; err != nil {
			return err
		}

		if len(input.Bindings) > 0 {
			for i := range input.Bindings {
				input.Bindings[i].SupplyAccountId = supply.Id
			}
			if err := tx.Create(&input.Bindings).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	return &seller, &supply, nil
}

func containsExactModel(models []string, target string) bool {
	target = strings.TrimSpace(target)
	if target == "" {
		return false
	}
	for _, modelName := range models {
		if strings.TrimSpace(modelName) == target {
			return true
		}
	}
	return false
}

func UpdateSellerStatus(id int, status string, remark string) error {
	if id <= 0 {
		return errors.New("invalid seller id")
	}
	if strings.TrimSpace(status) == "" {
		return errors.New("status is required")
	}
	return model.UpdateSellerStatus(id, status, remark)
}
