package service

import (
	"fmt"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"gorm.io/gorm"
)

func VerifySellerSecret(id int, actorUserId int) (*model.SellerSecret, error) {
	secret, err := model.GetSellerSecretByID(id)
	if err != nil {
		return nil, err
	}
	runtimeKey, verifyErr := decryptSellerSecretRuntimeValue(secret)

	var syncedChannelIds []int
	var verifyLogResult = "success"
	var verifyLogReason string
	err = model.DB.Transaction(func(tx *gorm.DB) error {
		if verifyErr != nil {
			if err := markSellerSecretVerificationFailureTx(tx, secret, actorUserId, verifyErr.Error()); err != nil {
				return err
			}
			return recomputeSupplyAccountSecretStateTx(tx, secret.SupplyAccountId)
		}

		blockers, err := listSecretActivationBlockersTx(tx, secret.SupplyAccountId, secret.Id)
		if err != nil {
			return err
		}
		if len(blockers) > 0 {
			verifyErr = fmt.Errorf("another active or rotating seller secret already exists on supply %d", secret.SupplyAccountId)
			if err := markSellerSecretVerificationFailureTx(tx, secret, actorUserId, verifyErr.Error()); err != nil {
				return err
			}
			return recomputeSupplyAccountSecretStateTx(tx, secret.SupplyAccountId)
		}

		if err := probeSellerSecretBindingsTx(tx, secret); err != nil {
			verifyErr = err
			if err := markSellerSecretVerificationFailureTx(tx, secret, actorUserId, verifyErr.Error()); err != nil {
				return err
			}
			return recomputeSupplyAccountSecretStateTx(tx, secret.SupplyAccountId)
		}

		channelIds, err := syncSellerSecretRuntimeMirrorTx(tx, secret, runtimeKey)
		if err != nil {
			verifyErr = err
			if err := markSellerSecretVerificationFailureTx(tx, secret, actorUserId, verifyErr.Error()); err != nil {
				return err
			}
			return recomputeSupplyAccountSecretStateTx(tx, secret.SupplyAccountId)
		}
		syncedChannelIds = channelIds

		now := common.GetTimestamp()
		verifyMessage := "verified and mirrored to runtime channels"
		if err := tx.Model(&model.SellerSecret{}).Where("id = ?", id).Updates(map[string]interface{}{
			"status":           "active",
			"verify_status":    "success",
			"verify_message":   verifyMessage,
			"disabled_reason":  "",
			"last_verified_at": now,
			"updated_at":       now,
		}).Error; err != nil {
			return err
		}

		if err := recordSellerSecretAuditTx(tx, model.SellerSecretAudit{
			SellerSecretId:  secret.Id,
			SellerId:        secret.SellerId,
			SupplyAccountId: secret.SupplyAccountId,
			ActorUserId:     actorUserId,
			ActorType:       "admin",
			Action:          "verify_success",
			Result:          "success",
			Meta:            fmt.Sprintf("runtime_channels=%v", channelIds),
		}); err != nil {
			return err
		}

		if err := recordSellerSecretAuditTx(tx, model.SellerSecretAudit{
			SellerSecretId:  secret.Id,
			SellerId:        secret.SellerId,
			SupplyAccountId: secret.SupplyAccountId,
			ActorUserId:     actorUserId,
			ActorType:       "admin",
			Action:          "sync_channel",
			Result:          "success",
			Meta:            fmt.Sprintf("runtime_channels=%v", channelIds),
		}); err != nil {
			return err
		}

		return recomputeSupplyAccountSecretStateTx(tx, secret.SupplyAccountId)
	})
	if err != nil {
		return nil, err
	}

	updated, err := model.GetSellerSecretByID(id)
	if err != nil {
		return nil, err
	}
	if verifyErr == nil {
		recordSellerSecretOperationLog(actorUserId, updated, "verify", "success", "runtime mirror sync completed")
		recordSellerSecretOperationLog(actorUserId, updated, "sync_channel", "success", fmt.Sprintf("channels=%v", syncedChannelIds))
		refreshChannelCache(syncedChannelIds)
		return updated, nil
	}
	verifyLogResult = "failed"
	verifyLogReason = verifyErr.Error()
	recordSellerSecretOperationLog(actorUserId, updated, "verify", verifyLogResult, verifyLogReason)
	if strings.TrimSpace(updated.VerifyMessage) == "" {
		updated.VerifyMessage = verifyErr.Error()
	}
	return updated, verifyErr
}
