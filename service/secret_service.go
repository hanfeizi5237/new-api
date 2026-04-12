package service

import (
	"errors"
	"fmt"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"gorm.io/gorm"
)

type CreateSellerSecretInput struct {
	SellerId        int
	SupplyAccountId int
	SecretType      string
	Plaintext       string
	ProviderCode    string
	ExpiresAt       int64
	Meta            string
}

func ListSellerSecrets(sellerId int, supplyAccountId int, status string, offset int, limit int) ([]*model.SellerSecret, int64, error) {
	return model.GetSellerSecrets(sellerId, supplyAccountId, status, offset, limit)
}

func CreateSellerSecret(input CreateSellerSecretInput, actorUserId int, reason string) (*model.SellerSecret, error) {
	if input.SellerId <= 0 || input.SupplyAccountId <= 0 {
		return nil, errors.New("seller_id and supply_account_id are required")
	}
	normalizedPlaintext, err := normalizeSellerSecretPlaintext(input.SecretType, input.Plaintext)
	if err != nil {
		return nil, err
	}

	seller, err := model.GetSellerByID(input.SellerId)
	if err != nil {
		return nil, err
	}
	supply, err := model.GetSupplyAccountByID(input.SupplyAccountId)
	if err != nil {
		return nil, err
	}
	if supply.SellerId != seller.Id {
		return nil, errors.New("supply_account does not belong to seller")
	}
	ciphertext, cipherVersion, err := encryptSellerSecretPlaintext(normalizedPlaintext)
	if err != nil {
		return nil, err
	}
	fingerprint, err := buildSellerSecretFingerprint(normalizedPlaintext)
	if err != nil {
		return nil, err
	}
	secretType := strings.TrimSpace(input.SecretType)
	if secretType == "" {
		secretType = "api_key"
	}
	secret := &model.SellerSecret{
		SellerId:        input.SellerId,
		SupplyAccountId: input.SupplyAccountId,
		SecretType:      secretType,
		ProviderCode:    strings.TrimSpace(input.ProviderCode),
		Ciphertext:      ciphertext,
		CipherVersion:   cipherVersion,
		Fingerprint:     fingerprint,
		MaskedValue:     buildSellerSecretMaskedValue(secretType, normalizedPlaintext, fingerprint),
		Status:          "draft",
		VerifyStatus:    "pending",
		VerifyMessage:   "awaiting verification",
		ExpiresAt:       input.ExpiresAt,
		Meta:            input.Meta,
	}
	if secret.ProviderCode == "" {
		secret.ProviderCode = supply.ProviderCode
	}

	var createdSecret *model.SellerSecret

	err = model.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(secret).Error; err != nil {
			return err
		}
		audit := model.SellerSecretAudit{
			SellerSecretId:  secret.Id,
			SellerId:        secret.SellerId,
			SupplyAccountId: secret.SupplyAccountId,
			ActorUserId:     actorUserId,
			ActorType:       "admin",
			Action:          "create",
			Reason:          reason,
			Result:          "success",
		}
		if err := tx.Create(&audit).Error; err != nil {
			return err
		}
		createdSecret = secret
		return nil
	})
	if err != nil {
		return nil, err
	}
	recordSellerSecretOperationLog(actorUserId, createdSecret, "import", "success", reason)
	return secret, nil
}

func DisableSellerSecret(id int, actorUserId int, reason string) (*model.SellerSecret, error) {
	secret, err := model.GetSellerSecretByID(id)
	if err != nil {
		return nil, err
	}
	err = model.DB.Transaction(func(tx *gorm.DB) error {
		now := common.GetTimestamp()
		if err := tx.Model(&model.SellerSecret{}).Where("id = ?", id).Updates(map[string]interface{}{
			"status":           "disabled",
			"verify_status":    "failed",
			"disabled_reason":  reason,
			"verify_message":   reason,
			"last_verified_at": now,
			"updated_at":       now,
		}).Error; err != nil {
			return err
		}
		audit := model.SellerSecretAudit{
			SellerSecretId:  secret.Id,
			SellerId:        secret.SellerId,
			SupplyAccountId: secret.SupplyAccountId,
			ActorUserId:     actorUserId,
			ActorType:       "admin",
			Action:          "disable",
			Reason:          reason,
			Result:          "success",
		}
		if err := tx.Create(&audit).Error; err != nil {
			return err
		}
		return recomputeSupplyAccountSecretStateTx(tx, secret.SupplyAccountId)
	})
	if err != nil {
		return nil, err
	}
	recordSellerSecretOperationLog(actorUserId, secret, "disable", "success", reason)
	return model.GetSellerSecretByID(id)
}

func RecoverSellerSecret(id int, actorUserId int, reason string) (*model.SellerSecret, error) {
	secret, err := model.GetSellerSecretByID(id)
	if err != nil {
		return nil, err
	}
	err = model.DB.Transaction(func(tx *gorm.DB) error {
		now := common.GetTimestamp()
		if err := tx.Model(&model.SellerSecret{}).Where("id = ?", id).Updates(map[string]interface{}{
			"status":          "draft",
			"verify_status":   "pending",
			"disabled_reason": "",
			"verify_message":  "awaiting verification",
			"updated_at":      now,
		}).Error; err != nil {
			return err
		}
		audit := model.SellerSecretAudit{
			SellerSecretId:  secret.Id,
			SellerId:        secret.SellerId,
			SupplyAccountId: secret.SupplyAccountId,
			ActorUserId:     actorUserId,
			ActorType:       "admin",
			Action:          "recover",
			Reason:          reason,
			Result:          "success",
		}
		if err := tx.Create(&audit).Error; err != nil {
			return err
		}
		return recomputeSupplyAccountSecretStateTx(tx, secret.SupplyAccountId)
	})
	if err != nil {
		return nil, err
	}
	recordSellerSecretOperationLog(actorUserId, secret, "recover", "success", reason)
	return model.GetSellerSecretByID(id)
}

func recordSellerSecretOperationLog(actorUserId int, secret *model.SellerSecret, action string, result string, reason string) {
	if actorUserId <= 0 || secret == nil {
		return
	}
	content := fmt.Sprintf(
		"seller secret %s %s: seller_secret_id=%d seller_id=%d supply_account_id=%d result=%s reason=%s",
		action,
		strings.TrimSpace(secret.SecretType),
		secret.Id,
		secret.SellerId,
		secret.SupplyAccountId,
		result,
		strings.TrimSpace(reason),
	)
	model.RecordLog(actorUserId, model.LogTypeManage, content)
}
