package service

import (
	"fmt"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"gorm.io/gorm"
)

func VerifySellerSecret(id int, actor SellerSecretAuditActor) (*model.SellerSecret, error) {
	if actor.ActorUserID <= 0 {
		return nil, fmt.Errorf("actor_user_id is required")
	}
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
			if err := markSellerSecretVerificationFailureTx(tx, secret, actor, verifyErr.Error()); err != nil {
				return err
			}
			return recomputeSupplyAccountSecretStateTx(tx, secret.SupplyAccountId)
		}

		if err := prepareSellerSecretActivationTx(tx, secret.SupplyAccountId, secret.Id); err != nil {
			verifyErr = err
			if err := markSellerSecretVerificationFailureTx(tx, secret, actor, verifyErr.Error()); err != nil {
				return err
			}
			return recomputeSupplyAccountSecretStateTx(tx, secret.SupplyAccountId)
		}

		if err := probeSellerSecretBindingsTx(tx, secret); err != nil {
			verifyErr = err
			if err := markSellerSecretVerificationFailureTx(tx, secret, actor, verifyErr.Error()); err != nil {
				return err
			}
			return recomputeSupplyAccountSecretStateTx(tx, secret.SupplyAccountId)
		}
		if err := sellerSecretLiveProbeFunc(secret, runtimeKey); err != nil {
			verifyErr = err
			if err := markSellerSecretVerificationFailureTx(tx, secret, actor, verifyErr.Error()); err != nil {
				return err
			}
			return recomputeSupplyAccountSecretStateTx(tx, secret.SupplyAccountId)
		}

		channelIds, err := syncSellerSecretRuntimeMirrorTx(tx, secret, runtimeKey)
		if err != nil {
			verifyErr = err
			if err := markSellerSecretVerificationFailureTx(tx, secret, actor, verifyErr.Error()); err != nil {
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

		if err := recordSellerSecretAuditTx(tx, buildSellerSecretAuditRecordWithExtraMeta(
			secret,
			actor,
			"verify_success",
			"",
			"success",
			map[string]any{
				"runtime_channels": channelIds,
			},
		)); err != nil {
			return err
		}

		if err := recordSellerSecretAuditTx(tx, buildSellerSecretAuditRecordWithExtraMeta(
			secret,
			actor,
			"sync_channel",
			"",
			"success",
			map[string]any{
				"runtime_channels": channelIds,
			},
		)); err != nil {
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
		recordSellerSecretOperationLog(actor.ActorUserID, updated, "verify", "success", "runtime mirror sync completed")
		recordSellerSecretOperationLog(actor.ActorUserID, updated, "sync_channel", "success", fmt.Sprintf("channels=%v", syncedChannelIds))
		refreshChannelCache(syncedChannelIds)
		syncMarketplaceInventoryAfterMutation(updated.SupplyAccountId, "seller_secret_verified")
		return updated, nil
	}
	verifyLogResult = "failed"
	verifyLogReason = verifyErr.Error()
	recordSellerSecretOperationLog(actor.ActorUserID, updated, "verify", verifyLogResult, verifyLogReason)
	if strings.TrimSpace(updated.VerifyMessage) == "" {
		updated.VerifyMessage = verifyErr.Error()
	}
	syncMarketplaceInventoryAfterMutation(updated.SupplyAccountId, "seller_secret_verify_failed")
	return updated, verifyErr
}
