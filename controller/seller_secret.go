package controller

import (
	"strconv"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"
	"github.com/gin-gonic/gin"
)

type CreateSellerSecretRequest struct {
	SellerId        int    `json:"seller_id"`
	SupplyAccountId int    `json:"supply_account_id"`
	SecretType      string `json:"secret_type"`
	Plaintext       string `json:"plaintext"`
	ProviderCode    string `json:"provider_code"`
	ExpiresAt       int64  `json:"expires_at"`
	Meta            string `json:"meta"`
	Reason          string `json:"reason"`
}

type SellerSecretActionRequest struct {
	Reason string `json:"reason"`
}

type SellerSecretView struct {
	Id              int    `json:"id"`
	SellerId        int    `json:"seller_id"`
	SupplyAccountId int    `json:"supply_account_id"`
	SecretType      string `json:"secret_type"`
	ProviderCode    string `json:"provider_code"`
	CipherVersion   string `json:"cipher_version"`
	Fingerprint     string `json:"fingerprint"`
	MaskedValue     string `json:"masked_value"`
	Status          string `json:"status"`
	VerifyStatus    string `json:"verify_status"`
	LastVerifiedAt  int64  `json:"last_verified_at"`
	LastUsedAt      int64  `json:"last_used_at"`
	LastRotationAt  int64  `json:"last_rotation_at"`
	ExpiresAt       int64  `json:"expires_at"`
	DisabledReason  string `json:"disabled_reason"`
	VerifyMessage   string `json:"verify_message"`
	Meta            string `json:"meta"`
	CreatedAt       int64  `json:"created_at"`
	UpdatedAt       int64  `json:"updated_at"`
}

func toSellerSecretView(secret *model.SellerSecret) SellerSecretView {
	return SellerSecretView{
		Id:              secret.Id,
		SellerId:        secret.SellerId,
		SupplyAccountId: secret.SupplyAccountId,
		SecretType:      secret.SecretType,
		ProviderCode:    secret.ProviderCode,
		CipherVersion:   secret.CipherVersion,
		Fingerprint:     secret.Fingerprint,
		MaskedValue:     secret.MaskedValue,
		Status:          secret.Status,
		VerifyStatus:    secret.VerifyStatus,
		LastVerifiedAt:  secret.LastVerifiedAt,
		LastUsedAt:      secret.LastUsedAt,
		LastRotationAt:  secret.LastRotationAt,
		ExpiresAt:       secret.ExpiresAt,
		DisabledReason:  secret.DisabledReason,
		VerifyMessage:   secret.VerifyMessage,
		Meta:            secret.Meta,
		CreatedAt:       secret.CreatedAt,
		UpdatedAt:       secret.UpdatedAt,
	}
}

func GetSellerSecretsAdmin(c *gin.Context) {
	pageInfo := common.GetPageQuery(c)
	sellerId, _ := strconv.Atoi(c.Query("seller_id"))
	supplyAccountId, _ := strconv.Atoi(c.Query("supply_account_id"))
	secrets, total, err := service.ListSellerSecrets(
		sellerId,
		supplyAccountId,
		c.Query("status"),
		pageInfo.GetStartIdx(),
		pageInfo.GetPageSize(),
	)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	items := make([]SellerSecretView, 0, len(secrets))
	for _, secret := range secrets {
		items = append(items, toSellerSecretView(secret))
	}
	pageInfo.SetTotal(int(total))
	pageInfo.SetItems(items)
	common.ApiSuccess(c, pageInfo)
}

func CreateSellerSecretAdmin(c *gin.Context) {
	var req CreateSellerSecretRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiError(c, err)
		return
	}
	secret, err := service.CreateSellerSecret(service.CreateSellerSecretInput{
		SellerId:        req.SellerId,
		SupplyAccountId: req.SupplyAccountId,
		SecretType:      req.SecretType,
		Plaintext:       req.Plaintext,
		ProviderCode:    req.ProviderCode,
		ExpiresAt:       req.ExpiresAt,
		Meta:            req.Meta,
	}, c.GetInt("id"), req.Reason)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, toSellerSecretView(secret))
}

func VerifySellerSecretAdmin(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		common.ApiError(c, err)
		return
	}
	secret, err := service.VerifySellerSecret(id, c.GetInt("id"))
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, toSellerSecretView(secret))
}

func DisableSellerSecretAdmin(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		common.ApiError(c, err)
		return
	}
	var req SellerSecretActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiError(c, err)
		return
	}
	secret, err := service.DisableSellerSecret(id, c.GetInt("id"), req.Reason)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, toSellerSecretView(secret))
}

func RecoverSellerSecretAdmin(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		common.ApiError(c, err)
		return
	}
	var req SellerSecretActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiError(c, err)
		return
	}
	secret, err := service.RecoverSellerSecret(id, c.GetInt("id"), req.Reason)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, toSellerSecretView(secret))
}
