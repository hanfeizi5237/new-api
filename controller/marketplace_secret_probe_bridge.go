package controller

import (
	"errors"
	"fmt"

	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"
)

func init() {
	service.SetSellerSecretLiveProbeFunc(marketplaceSellerSecretLiveProbe)
}

func marketplaceSellerSecretLiveProbe(secret *model.SellerSecret, runtimeKey string) error {
	if secret == nil {
		return errors.New("seller secret is required")
	}
	if runtimeKey == "" {
		return errors.New("seller secret runtime key is empty")
	}
	seller, err := model.GetSellerByID(secret.SellerId)
	if err != nil {
		return err
	}
	supply, err := model.GetSupplyAccountByID(secret.SupplyAccountId)
	if err != nil {
		return err
	}
	var bindings []model.SupplyChannelBinding
	if err := model.DB.Where("supply_account_id = ? AND status = ?", secret.SupplyAccountId, "active").
		Order("priority desc, id asc").
		Find(&bindings).Error; err != nil {
		return err
	}
	if len(bindings) == 0 {
		return errors.New("no active supply channel bindings available for provider probe")
	}
	for _, binding := range bindings {
		channel, err := model.GetChannelById(binding.ChannelId, true)
		if err != nil {
			return err
		}
		channelCopy := *channel
		channelCopy.Key = runtimeKey
		otherInfo := channelCopy.GetOtherInfo()
		otherInfo["managed_by"] = "seller_secret_probe"
		otherInfo["supply_account_id"] = secret.SupplyAccountId
		otherInfo["seller_secret_id"] = secret.Id
		channelCopy.SetOtherInfo(otherInfo)
		result := testChannel(&channelCopy, supply.ModelName, "", false)
		if result.localErr != nil {
			return fmt.Errorf("channel %d provider probe failed: %w", channel.Id, result.localErr)
		}
	}
	if seller.UserId <= 0 {
		return errors.New("seller user id is invalid")
	}
	return nil
}
