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

type SellerSecretAuditActor struct {
	ActorUserID int
	ActorType   string
	RequestID   string
	IP          string
	Meta        string
}

func ListSellerSecrets(sellerId int, supplyAccountId int, status string, offset int, limit int) ([]*model.SellerSecret, int64, error) {
	return model.GetSellerSecrets(sellerId, supplyAccountId, status, offset, limit)
}

func CreateSellerSecret(input CreateSellerSecretInput, actor SellerSecretAuditActor, reason string) (*model.SellerSecret, error) {
	if input.SellerId <= 0 || input.SupplyAccountId <= 0 {
		return nil, errors.New("seller_id and supply_account_id are required")
	}
	if actor.ActorUserID <= 0 {
		return nil, errors.New("actor_user_id is required")
	}
	normalizedPlaintext, err := normalizeSellerSecretPlaintext(input.SecretType, input.Plaintext)
	if err != nil {
		return nil, err
	}

	seller, err := model.GetSellerByID(input.SellerId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("seller not found")
		}
		return nil, err
	}
	supply, err := model.GetSupplyAccountByID(input.SupplyAccountId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("supply_account not found")
		}
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
		audit := buildSellerSecretAuditRecord(secret, actor, "create", reason, "success")
		if err := tx.Create(&audit).Error; err != nil {
			return err
		}
		createdSecret = secret
		return nil
	})
	if err != nil {
		return nil, err
	}
	recordSellerSecretOperationLog(actor.ActorUserID, createdSecret, "import", "success", reason)
	syncMarketplaceInventoryAfterMutation(createdSecret.SupplyAccountId, "seller_secret_created")
	return secret, nil
}

func DisableSellerSecret(id int, actor SellerSecretAuditActor, reason string) (*model.SellerSecret, error) {
	if actor.ActorUserID <= 0 {
		return nil, errors.New("actor_user_id is required")
	}
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
		audit := buildSellerSecretAuditRecord(secret, actor, "disable", reason, "success")
		if err := tx.Create(&audit).Error; err != nil {
			return err
		}
		return recomputeSupplyAccountSecretStateTx(tx, secret.SupplyAccountId)
	})
	if err != nil {
		return nil, err
	}
	recordSellerSecretOperationLog(actor.ActorUserID, secret, "disable", "success", reason)
	syncMarketplaceInventoryAfterMutation(secret.SupplyAccountId, "seller_secret_disabled")
	return model.GetSellerSecretByID(id)
}

func RecoverSellerSecret(id int, actor SellerSecretAuditActor, reason string) (*model.SellerSecret, error) {
	if actor.ActorUserID <= 0 {
		return nil, errors.New("actor_user_id is required")
	}
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
		audit := buildSellerSecretAuditRecord(secret, actor, "recover", reason, "success")
		if err := tx.Create(&audit).Error; err != nil {
			return err
		}
		return recomputeSupplyAccountSecretStateTx(tx, secret.SupplyAccountId)
	})
	if err != nil {
		return nil, err
	}
	recordSellerSecretOperationLog(actor.ActorUserID, secret, "recover", "success", reason)
	syncMarketplaceInventoryAfterMutation(secret.SupplyAccountId, "seller_secret_recovered")
	return model.GetSellerSecretByID(id)
}

func buildSellerSecretAuditRecord(
	secret *model.SellerSecret,
	actor SellerSecretAuditActor,
	action string,
	reason string,
	result string,
) model.SellerSecretAudit {
	actorType := strings.TrimSpace(actor.ActorType)
	if actorType == "" {
		actorType = "admin"
	}
	meta := strings.TrimSpace(actor.Meta)
	if meta == "" && (strings.TrimSpace(actor.RequestID) != "" || strings.TrimSpace(actor.IP) != "") {
		meta = common.MapToJsonStr(map[string]any{
			"request_id": strings.TrimSpace(actor.RequestID),
			"ip":         strings.TrimSpace(actor.IP),
		})
	}
	return model.SellerSecretAudit{
		SellerSecretId:  secret.Id,
		SellerId:        secret.SellerId,
		SupplyAccountId: secret.SupplyAccountId,
		ActorUserId:     actor.ActorUserID,
		ActorType:       actorType,
		Action:          action,
		Reason:          reason,
		RequestId:       strings.TrimSpace(actor.RequestID),
		Ip:              strings.TrimSpace(actor.IP),
		Result:          result,
		Meta:            meta,
	}
}

func buildSellerSecretAuditRecordWithExtraMeta(
	secret *model.SellerSecret,
	actor SellerSecretAuditActor,
	action string,
	reason string,
	result string,
	extraMeta map[string]any,
) model.SellerSecretAudit {
	audit := buildSellerSecretAuditRecord(secret, actor, action, reason, result)
	audit.Meta = mergeSellerSecretAuditMeta(audit.Meta, extraMeta)
	return audit
}

func mergeSellerSecretAuditMeta(baseMeta string, extraMeta map[string]any) string {
	if len(extraMeta) == 0 {
		return strings.TrimSpace(baseMeta)
	}
	merged := map[string]any{}
	trimmedBaseMeta := strings.TrimSpace(baseMeta)
	if trimmedBaseMeta != "" {
		if err := common.UnmarshalJsonStr(trimmedBaseMeta, &merged); err != nil {
			merged = map[string]any{
				"raw_meta": trimmedBaseMeta,
			}
		}
	}
	for key, value := range extraMeta {
		merged[key] = value
	}
	if len(merged) == 0 {
		return ""
	}
	return common.MapToJsonStr(merged)
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
