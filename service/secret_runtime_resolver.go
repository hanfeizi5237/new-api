package service

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"gorm.io/gorm"
)

type sellerSecretCipherEnvelope struct {
	Alg        string `json:"alg"`
	Kid        string `json:"kid"`
	Nonce      string `json:"nonce"`
	Ciphertext string `json:"ciphertext"`
}

type supplySecretState struct {
	Status           string
	VerifyStatus     string
	VerifyMessage    string
	LatestVerifiedAt int64
}

func getSellerSecretMasterKey() ([]byte, error) {
	masterKey := strings.TrimSpace(os.Getenv("SELLER_SECRET_MASTER_KEY"))
	if masterKey == "" {
		return nil, errors.New("SELLER_SECRET_MASTER_KEY is not configured")
	}
	if len(masterKey) == 32 {
		return []byte(masterKey), nil
	}
	decoded, err := base64.StdEncoding.DecodeString(masterKey)
	if err != nil {
		return nil, errors.New("SELLER_SECRET_MASTER_KEY must be 32 raw bytes or base64-encoded 32 bytes")
	}
	if len(decoded) != 32 {
		return nil, errors.New("SELLER_SECRET_MASTER_KEY must resolve to 32 bytes")
	}
	return decoded, nil
}

func getSellerSecretActiveVersion() string {
	version := strings.TrimSpace(os.Getenv("SELLER_SECRET_ACTIVE_VERSION"))
	if version == "" {
		return "v1"
	}
	return version
}

func getSellerSecretFingerprintSalt() ([]byte, error) {
	salt := strings.TrimSpace(os.Getenv("SELLER_SECRET_FINGERPRINT_SALT"))
	if salt == "" {
		return nil, errors.New("SELLER_SECRET_FINGERPRINT_SALT is not configured")
	}
	return []byte(salt), nil
}

func normalizeSellerSecretPlaintext(secretType string, plaintext string) (string, error) {
	normalized := strings.TrimSpace(plaintext)
	if normalized == "" {
		return "", errors.New("seller secret plaintext is empty")
	}
	switch strings.TrimSpace(secretType) {
	case "", "api_key", "oauth_token":
		return normalized, nil
	case "service_account", "json_blob":
		var payload any
		if err := common.UnmarshalJsonStr(normalized, &payload); err != nil {
			return "", errors.New("seller secret json payload is invalid")
		}
		bytes, err := common.Marshal(payload)
		if err != nil {
			return "", err
		}
		return string(bytes), nil
	default:
		return "", fmt.Errorf("unsupported seller secret type: %s", secretType)
	}
}

func encryptSellerSecretPlaintext(plaintext string) (string, string, error) {
	keyBytes, err := getSellerSecretMasterKey()
	if err != nil {
		return "", "", err
	}
	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return "", "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", "", err
	}
	version := getSellerSecretActiveVersion()
	cipherBytes := gcm.Seal(nil, nonce, []byte(plaintext), nil)
	envelope := sellerSecretCipherEnvelope{
		Alg:        "aes-256-gcm",
		Kid:        version,
		Nonce:      base64.StdEncoding.EncodeToString(nonce),
		Ciphertext: base64.StdEncoding.EncodeToString(cipherBytes),
	}
	bytes, err := common.Marshal(envelope)
	if err != nil {
		return "", "", err
	}
	return string(bytes), version, nil
}

func buildSellerSecretFingerprint(plaintext string) (string, error) {
	salt, err := getSellerSecretFingerprintSalt()
	if err != nil {
		return "", err
	}
	mac := hmac.New(sha256.New, salt)
	if _, err := mac.Write([]byte(plaintext)); err != nil {
		return "", err
	}
	return hex.EncodeToString(mac.Sum(nil)), nil
}

func buildSellerSecretMaskedValue(secretType string, plaintext string, fingerprint string) string {
	switch strings.TrimSpace(secretType) {
	case "", "api_key":
		if len(plaintext) <= 8 {
			return plaintext[:1] + strings.Repeat("*", len(plaintext)-1)
		}
		return plaintext[:4] + strings.Repeat("*", len(plaintext)-8) + plaintext[len(plaintext)-4:]
	case "oauth_token":
		if len(plaintext) <= 6 {
			return plaintext[:1] + strings.Repeat("*", len(plaintext)-1)
		}
		return plaintext[:3] + strings.Repeat("*", len(plaintext)-6) + plaintext[len(plaintext)-3:]
	default:
		shortFingerprint := fingerprint
		if len(shortFingerprint) > 8 {
			shortFingerprint = shortFingerprint[:8]
		}
		return strings.TrimSpace(secretType) + "#" + shortFingerprint
	}
}

func decryptSellerSecretRuntimeValue(secret *model.SellerSecret) (string, error) {
	if secret == nil {
		return "", errors.New("seller secret is required")
	}

	var envelope sellerSecretCipherEnvelope
	if err := common.UnmarshalJsonStr(secret.Ciphertext, &envelope); err != nil {
		return "", errors.New("seller secret ciphertext is not valid JSON")
	}
	if envelope.Alg != "aes-256-gcm" {
		return "", fmt.Errorf("unsupported seller secret alg: %s", envelope.Alg)
	}
	if strings.TrimSpace(envelope.Kid) == "" || envelope.Kid != secret.CipherVersion {
		return "", errors.New("seller secret kid does not match cipher_version")
	}
	if strings.TrimSpace(envelope.Nonce) == "" || strings.TrimSpace(envelope.Ciphertext) == "" {
		return "", errors.New("seller secret nonce/ciphertext is empty")
	}

	keyBytes, err := getSellerSecretMasterKey()
	if err != nil {
		return "", err
	}
	nonceBytes, err := base64.StdEncoding.DecodeString(envelope.Nonce)
	if err != nil {
		return "", errors.New("seller secret nonce is not valid base64")
	}
	cipherBytes, err := base64.StdEncoding.DecodeString(envelope.Ciphertext)
	if err != nil {
		return "", errors.New("seller secret ciphertext body is not valid base64")
	}

	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	if len(nonceBytes) != gcm.NonceSize() {
		return "", errors.New("seller secret nonce size is invalid")
	}
	plaintext, err := gcm.Open(nil, nonceBytes, cipherBytes, nil)
	if err != nil {
		return "", errors.New("seller secret decrypt failed")
	}
	runtimeKey := strings.TrimSpace(string(plaintext))
	if runtimeKey == "" {
		return "", errors.New("seller secret plaintext is empty after decrypt")
	}
	return runtimeKey, nil
}

func probeSellerSecretBindingsTx(tx *gorm.DB, secret *model.SellerSecret) error {
	bindings, err := getActiveSupplyBindingsTx(tx, secret.SupplyAccountId)
	if err != nil {
		return err
	}
	if len(bindings) == 0 {
		return errors.New("no active supply channel bindings available for verification")
	}
	supply, err := model.GetSupplyAccountByID(secret.SupplyAccountId)
	if err != nil {
		return err
	}
	for _, binding := range bindings {
		channel, err := model.GetChannelById(binding.ChannelId, true)
		if err != nil {
			return err
		}
		if channel.Status != common.ChannelStatusEnabled {
			return fmt.Errorf("bound channel %d is not enabled", channel.Id)
		}
		if !containsExactModel(channel.GetModels(), supply.ModelName) {
			return fmt.Errorf("bound channel %d does not support model %s", channel.Id, supply.ModelName)
		}
	}
	return nil
}

func getActiveSupplyBindingsTx(tx *gorm.DB, supplyAccountId int) ([]model.SupplyChannelBinding, error) {
	var bindings []model.SupplyChannelBinding
	if err := tx.Where("supply_account_id = ? AND status = ?", supplyAccountId, "active").
		Order("priority desc, id asc").
		Find(&bindings).Error; err != nil {
		return nil, err
	}
	return bindings, nil
}

func listSecretActivationBlockersTx(tx *gorm.DB, supplyAccountId int, excludeSecretId int) ([]model.SellerSecret, error) {
	var blockers []model.SellerSecret
	if err := tx.Where("supply_account_id = ? AND id <> ? AND status IN ?", supplyAccountId, excludeSecretId, []string{"active", "rotating"}).
		Order("id asc").
		Find(&blockers).Error; err != nil {
		return nil, err
	}
	return blockers, nil
}

func syncSellerSecretRuntimeMirrorTx(tx *gorm.DB, secret *model.SellerSecret, runtimeKey string) ([]int, error) {
	bindings, err := getActiveSupplyBindingsTx(tx, secret.SupplyAccountId)
	if err != nil {
		return nil, err
	}
	if len(bindings) == 0 {
		return nil, errors.New("no active supply channel bindings available for runtime mirror sync")
	}

	channelIds := make([]int, 0, len(bindings))
	for _, binding := range bindings {
		var channel model.Channel
		if err := tx.First(&channel, "id = ?", binding.ChannelId).Error; err != nil {
			return nil, err
		}
		channel.Key = runtimeKey
		otherInfo := channel.GetOtherInfo()
		otherInfo["managed_by"] = "seller_secret"
		otherInfo["supply_account_id"] = secret.SupplyAccountId
		otherInfo["seller_secret_id"] = secret.Id
		channel.SetOtherInfo(otherInfo)
		if err := tx.Save(&channel).Error; err != nil {
			return nil, err
		}
		channelIds = append(channelIds, channel.Id)
	}
	return channelIds, nil
}

func refreshChannelCache(channelIds []int) {
	if len(channelIds) == 0 || !common.MemoryCacheEnabled {
		return
	}
	for _, channelId := range channelIds {
		channel, err := model.GetChannelById(channelId, true)
		if err != nil {
			continue
		}
		model.CacheUpdateChannel(channel)
	}
}

func deriveSupplySecretState(secrets []model.SellerSecret) supplySecretState {
	state := supplySecretState{
		Status:        "paused",
		VerifyStatus:  "pending",
		VerifyMessage: "no seller secret configured",
	}

	var activeSecret *model.SellerSecret
	var pendingSecret *model.SellerSecret
	var failedSecret *model.SellerSecret

	for i := range secrets {
		secret := &secrets[i]
		if secret.Status == "active" && secret.VerifyStatus == "success" {
			if activeSecret == nil || secret.LastVerifiedAt > activeSecret.LastVerifiedAt {
				activeSecret = secret
			}
			continue
		}
		if secret.VerifyStatus == "pending" || secret.Status == "draft" || secret.Status == "verifying" || secret.Status == "rotating" {
			if pendingSecret == nil || secret.UpdatedAt > pendingSecret.UpdatedAt {
				pendingSecret = secret
			}
			continue
		}
		if failedSecret == nil || secret.UpdatedAt > failedSecret.UpdatedAt {
			failedSecret = secret
		}
	}

	if activeSecret != nil {
		state.Status = "active"
		state.VerifyStatus = "success"
		state.VerifyMessage = activeSecret.VerifyMessage
		if strings.TrimSpace(state.VerifyMessage) == "" {
			state.VerifyMessage = "active seller secret ready"
		}
		state.LatestVerifiedAt = activeSecret.LastVerifiedAt
		return state
	}
	if pendingSecret != nil {
		state.Status = "paused"
		state.VerifyStatus = "pending"
		state.VerifyMessage = pendingSecret.VerifyMessage
		if strings.TrimSpace(state.VerifyMessage) == "" {
			state.VerifyMessage = "awaiting seller secret verification"
		}
		state.LatestVerifiedAt = pendingSecret.LastVerifiedAt
		return state
	}
	if failedSecret != nil {
		state.Status = "paused"
		state.VerifyStatus = "failed"
		state.VerifyMessage = failedSecret.VerifyMessage
		if strings.TrimSpace(state.VerifyMessage) == "" {
			state.VerifyMessage = failedSecret.DisabledReason
		}
		if strings.TrimSpace(state.VerifyMessage) == "" {
			state.VerifyMessage = "no usable seller secret"
		}
		state.LatestVerifiedAt = failedSecret.LastVerifiedAt
		return state
	}
	return state
}

func recomputeSupplyAccountSecretStateTx(tx *gorm.DB, supplyAccountId int) error {
	var secrets []model.SellerSecret
	if err := tx.Where("supply_account_id = ?", supplyAccountId).
		Order("updated_at desc, id desc").
		Find(&secrets).Error; err != nil {
		return err
	}
	state := deriveSupplySecretState(secrets)
	return tx.Model(&model.SupplyAccount{}).Where("id = ?", supplyAccountId).Updates(map[string]interface{}{
		"status":             state.Status,
		"verify_status":      state.VerifyStatus,
		"verify_message":     state.VerifyMessage,
		"latest_verified_at": state.LatestVerifiedAt,
		"updated_at":         common.GetTimestamp(),
	}).Error
}

func recordSellerSecretAuditTx(tx *gorm.DB, audit model.SellerSecretAudit) error {
	return tx.Create(&audit).Error
}

func markSellerSecretVerificationFailureTx(tx *gorm.DB, secret *model.SellerSecret, actorUserId int, verifyMessage string) error {
	now := common.GetTimestamp()
	newStatus := secret.Status
	switch secret.Status {
	case "", "draft", "verifying", "rotating":
		newStatus = "draft"
	case "active":
		newStatus = "disabled"
	}
	if err := tx.Model(&model.SellerSecret{}).Where("id = ?", secret.Id).Updates(map[string]interface{}{
		"status":           newStatus,
		"verify_status":    "failed",
		"verify_message":   verifyMessage,
		"disabled_reason":  verifyMessage,
		"last_verified_at": now,
		"updated_at":       now,
	}).Error; err != nil {
		return err
	}
	return recordSellerSecretAuditTx(tx, model.SellerSecretAudit{
		SellerSecretId:  secret.Id,
		SellerId:        secret.SellerId,
		SupplyAccountId: secret.SupplyAccountId,
		ActorUserId:     actorUserId,
		ActorType:       "admin",
		Action:          "verify_failed",
		Reason:          verifyMessage,
		Result:          "failed",
		Meta:            verifyMessage,
	})
}
