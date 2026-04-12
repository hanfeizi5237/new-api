package controller

import (
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-gonic/gin"
)

type sellerSecretPageResponse struct {
	Items []SellerSecretView `json:"items"`
	Total int                `json:"total"`
}

func TestSellerSecretAdminCreateListAndVerifyMaskCiphertext(t *testing.T) {
	db := setupMarketplaceControllerTestDB(t)
	user := seedMarketplaceUser(t, db, "secret-admin-user")
	seller, supply := seedMarketplaceSellerWithSupply(t, db, user.Id)
	_, _ = seedMarketplaceChannelBinding(t, db, supply, "old-runtime-key")
	t.Setenv("SELLER_SECRET_MASTER_KEY", strings.Repeat("k", 32))
	t.Setenv("SELLER_SECRET_FINGERPRINT_SALT", "seller-secret-salt")

	createBody := CreateSellerSecretRequest{
		SellerId:        seller.Id,
		SupplyAccountId: supply.Id,
		SecretType:      "api_key",
		Plaintext:       "  sk-runtime-admin  ",
		Reason:          "initial import",
	}
	createCtx, createRecorder := newMarketplaceContext(t, http.MethodPost, "/api/marketplace/admin/seller-secrets", createBody, user.Id)
	CreateSellerSecretAdmin(createCtx)
	createResp := decodeMarketplaceResponse(t, createRecorder)
	if !createResp.Success {
		t.Fatalf("expected create secret success, got message: %s", createResp.Message)
	}
	if strings.Contains(createRecorder.Body.String(), "ciphertext") {
		t.Fatalf("create secret response leaked ciphertext: %s", createRecorder.Body.String())
	}

	var created SellerSecretView
	if err := common.Unmarshal(createResp.Data, &created); err != nil {
		t.Fatalf("failed to decode created secret view: %v", err)
	}
	if created.Id <= 0 || created.Status != "draft" || created.VerifyStatus != "pending" {
		t.Fatalf("expected created draft secret, got %+v", created)
	}
	if created.Fingerprint == "" || created.MaskedValue == "" {
		t.Fatalf("expected server generated fingerprint and masked value, got %+v", created)
	}

	rawSecret, err := model.GetSellerSecretByID(created.Id)
	if err != nil {
		t.Fatalf("failed to load raw seller secret: %v", err)
	}
	if rawSecret.Ciphertext == "" || rawSecret.Ciphertext == "sk-runtime-admin" {
		t.Fatalf("expected ciphertext to be server-encrypted, got %+v", rawSecret)
	}

	listCtx, listRecorder := newMarketplaceContext(t, http.MethodGet, "/api/marketplace/admin/seller-secrets?seller_id="+strconv.Itoa(seller.Id), nil, user.Id)
	GetSellerSecretsAdmin(listCtx)
	listResp := decodeMarketplaceResponse(t, listRecorder)
	if !listResp.Success {
		t.Fatalf("expected list secret success, got message: %s", listResp.Message)
	}
	if strings.Contains(listRecorder.Body.String(), "ciphertext") {
		t.Fatalf("list secret response leaked ciphertext: %s", listRecorder.Body.String())
	}

	var page sellerSecretPageResponse
	if err := common.Unmarshal(listResp.Data, &page); err != nil {
		t.Fatalf("failed to decode seller secret page: %v", err)
	}
	if page.Total != 1 || len(page.Items) != 1 {
		t.Fatalf("expected one seller secret in page, got total=%d items=%d", page.Total, len(page.Items))
	}

	verifyCtx, verifyRecorder := newMarketplaceContext(t, http.MethodPost, "/api/marketplace/admin/seller-secrets/"+strconv.Itoa(created.Id)+"/verify", nil, user.Id)
	verifyCtx.Params = gin.Params{{Key: "id", Value: strconv.Itoa(created.Id)}}
	VerifySellerSecretAdmin(verifyCtx)
	verifyResp := decodeMarketplaceResponse(t, verifyRecorder)
	if !verifyResp.Success {
		t.Fatalf("expected verify secret success, got message: %s", verifyResp.Message)
	}

	var verified SellerSecretView
	if err := common.Unmarshal(verifyResp.Data, &verified); err != nil {
		t.Fatalf("failed to decode verified secret view: %v", err)
	}
	if verified.Status != "active" || verified.VerifyStatus != "success" {
		t.Fatalf("expected active verified secret, got %+v", verified)
	}

	verifyLogCount := int64(0)
	if err := db.Model(&model.Log{}).Count(&verifyLogCount).Error; err != nil {
		t.Fatalf("failed to count operation logs: %v", err)
	}
	if verifyLogCount < 2 {
		t.Fatalf("expected operation logs for import + verify, got %d", verifyLogCount)
	}

	disableBody := SellerSecretActionRequest{Reason: "manual disable"}
	disableCtx, disableRecorder := newMarketplaceContext(t, http.MethodPost, "/api/marketplace/admin/seller-secrets/"+strconv.Itoa(created.Id)+"/disable", disableBody, user.Id)
	disableCtx.Params = gin.Params{{Key: "id", Value: strconv.Itoa(created.Id)}}
	DisableSellerSecretAdmin(disableCtx)
	disableResp := decodeMarketplaceResponse(t, disableRecorder)
	if !disableResp.Success {
		t.Fatalf("expected disable secret success, got message: %s", disableResp.Message)
	}

	recoverBody := SellerSecretActionRequest{Reason: "retry verify"}
	recoverCtx, recoverRecorder := newMarketplaceContext(t, http.MethodPost, "/api/marketplace/admin/seller-secrets/"+strconv.Itoa(created.Id)+"/recover", recoverBody, user.Id)
	recoverCtx.Params = gin.Params{{Key: "id", Value: strconv.Itoa(created.Id)}}
	RecoverSellerSecretAdmin(recoverCtx)
	recoverResp := decodeMarketplaceResponse(t, recoverRecorder)
	if !recoverResp.Success {
		t.Fatalf("expected recover secret success, got message: %s", recoverResp.Message)
	}

	if err := db.Model(&model.Log{}).Count(&verifyLogCount).Error; err != nil {
		t.Fatalf("failed to count operation logs after disable/recover: %v", err)
	}
	if verifyLogCount < 4 {
		t.Fatalf("expected operation logs for import/verify/disable/recover, got %d", verifyLogCount)
	}
}
