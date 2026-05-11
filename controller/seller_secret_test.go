package controller

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/middleware"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"
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
	service.SetSellerSecretLiveProbeFunc(func(secret *model.SellerSecret, runtimeKey string) error {
		if runtimeKey == "" {
			return errors.New("runtime key is empty")
		}
		return nil
	})
	t.Cleanup(func() {
		service.SetSellerSecretLiveProbeFunc(nil)
	})

	createBody := CreateSellerSecretRequest{
		SellerId:        seller.Id,
		SupplyAccountId: supply.Id,
		SecretType:      "api_key",
		Plaintext:       "  sk-runtime-admin  ",
		Reason:          "initial import",
	}
	createCtx, createRecorder := newMarketplaceContext(t, http.MethodPost, "/api/marketplace/admin/seller-secrets", createBody, user.Id)
	createCtx.Set("role", common.RoleAdminUser)
	createCtx.Set(common.RequestIdKey, "req-controller-secret-create")
	createCtx.Request.RemoteAddr = "198.51.100.61:12345"
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

	var createAudit model.SellerSecretAudit
	if err := db.Where("seller_secret_id = ? AND action = ?", created.Id, "create").Order("id desc").First(&createAudit).Error; err != nil {
		t.Fatalf("failed to load create audit: %v", err)
	}
	if createAudit.RequestId != "req-controller-secret-create" || createAudit.Ip != "198.51.100.61" || createAudit.ActorType != "admin" {
		t.Fatalf("expected create audit context to be persisted, got %+v", createAudit)
	}
	if createAudit.Meta == "" || !strings.Contains(createAudit.Meta, "request_path") {
		t.Fatalf("expected create audit meta to contain request context, got %q", createAudit.Meta)
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
	verifyCtx.Set("role", common.RoleAdminUser)
	verifyCtx.Set(common.RequestIdKey, "req-controller-secret-verify")
	verifyCtx.Request.RemoteAddr = "198.51.100.62:12345"
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

	var verifyAudit model.SellerSecretAudit
	if err := db.Where("seller_secret_id = ? AND action = ?", created.Id, "verify_success").Order("id desc").First(&verifyAudit).Error; err != nil {
		t.Fatalf("failed to load verify audit: %v", err)
	}
	if verifyAudit.RequestId != "req-controller-secret-verify" || verifyAudit.Ip != "198.51.100.62" || verifyAudit.ActorType != "admin" {
		t.Fatalf("expected verify audit context to be persisted, got %+v", verifyAudit)
	}
	if verifyAudit.Meta == "" || !strings.Contains(verifyAudit.Meta, "runtime_channels") || !strings.Contains(verifyAudit.Meta, "request_path") {
		t.Fatalf("expected verify audit meta to contain runtime channel and request context, got %q", verifyAudit.Meta)
	}

	verifyLogCount := int64(0)
	if err := db.Model(&model.Log{}).Count(&verifyLogCount).Error; err != nil {
		t.Fatalf("failed to count operation logs: %v", err)
	}
	if verifyLogCount < 2 {
		t.Fatalf("expected operation logs for import + verify, got %d", verifyLogCount)
	}

	disableBody := SellerSecretActionRequest{Reason: "manual disable"}
	disableRecorder := performMarketplaceRequestWithSession(
		t,
		http.MethodPost,
		"/api/marketplace/admin/seller-secrets/"+strconv.Itoa(created.Id)+"/disable",
		disableBody,
		user.Id,
		true,
		func(engine *gin.Engine) {
			engine.POST("/api/marketplace/admin/seller-secrets/:id/disable", middleware.SecureVerificationRequired(), DisableSellerSecretAdmin)
		},
	)
	disableResp := decodeMarketplaceResponse(t, disableRecorder)
	if !disableResp.Success {
		t.Fatalf("expected disable secret success, got message: %s", disableResp.Message)
	}

	recoverBody := SellerSecretActionRequest{Reason: "retry verify"}
	recoverRecorder := performMarketplaceRequestWithSession(
		t,
		http.MethodPost,
		"/api/marketplace/admin/seller-secrets/"+strconv.Itoa(created.Id)+"/recover",
		recoverBody,
		user.Id,
		true,
		func(engine *gin.Engine) {
			engine.POST("/api/marketplace/admin/seller-secrets/:id/recover", middleware.SecureVerificationRequired(), RecoverSellerSecretAdmin)
		},
	)
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

func TestSellerSecretAdminRejectsInvalidQueryAndPathIDs(t *testing.T) {
	listCtx, listRecorder := newMarketplaceContext(t, http.MethodGet, "/api/marketplace/admin/seller-secrets?seller_id=0", nil, 1)
	GetSellerSecretsAdmin(listCtx)
	listResp := decodeMarketplaceResponse(t, listRecorder)
	if listResp.Success {
		t.Fatalf("expected invalid seller_id query to fail")
	}
	if listResp.Message != "invalid seller_id" {
		t.Fatalf("expected invalid seller_id message, got %q", listResp.Message)
	}

	verifyCtx, verifyRecorder := newMarketplaceContext(t, http.MethodPost, "/api/marketplace/admin/seller-secrets/0/verify", nil, 1)
	verifyCtx.Params = gin.Params{{Key: "id", Value: "0"}}
	VerifySellerSecretAdmin(verifyCtx)
	verifyResp := decodeMarketplaceResponse(t, verifyRecorder)
	if verifyResp.Success {
		t.Fatalf("expected invalid seller secret id to fail")
	}
	if verifyResp.Message != "invalid seller secret id" {
		t.Fatalf("expected invalid seller secret id message, got %q", verifyResp.Message)
	}
}

func TestSellerSecretRiskActionsRequireSecureVerification(t *testing.T) {
	db := setupMarketplaceControllerTestDB(t)
	user := seedMarketplaceUser(t, db, "secret-risk-admin-user")
	seller, supply := seedMarketplaceSellerWithSupply(t, db, user.Id)
	secret := &model.SellerSecret{
		SellerId:        seller.Id,
		SupplyAccountId: supply.Id,
		SecretType:      "api_key",
		ProviderCode:    "openai",
		Ciphertext:      `{"alg":"aes-256-gcm","kid":"v1","nonce":"bm9uY2U=","ciphertext":"Y2lwaGVy"}`,
		CipherVersion:   "v1",
		Fingerprint:     "fp-secret-risk",
		MaskedValue:     "sk-***risk",
		Status:          "active",
		VerifyStatus:    "success",
		VerifyMessage:   "verified",
	}
	if err := db.Create(secret).Error; err != nil {
		t.Fatalf("failed to seed seller secret: %v", err)
	}

	disableRecorder := performMarketplaceRequestWithSession(
		t,
		http.MethodPost,
		"/api/marketplace/admin/seller-secrets/"+strconv.Itoa(secret.Id)+"/disable",
		SellerSecretActionRequest{Reason: "risk disable"},
		user.Id,
		false,
		func(engine *gin.Engine) {
			engine.POST("/api/marketplace/admin/seller-secrets/:id/disable", middleware.SecureVerificationRequired(), DisableSellerSecretAdmin)
		},
	)
	disableResp := decodeMarketplaceResponse(t, disableRecorder)
	if disableResp.Success {
		t.Fatalf("expected disable without secure verification to fail")
	}
	if disableResp.Message != "需要安全验证" {
		t.Fatalf("expected secure verification failure message, got %q", disableResp.Message)
	}

	reloadedSecret, err := model.GetSellerSecretByID(secret.Id)
	if err != nil {
		t.Fatalf("failed to reload seller secret after rejected disable: %v", err)
	}
	if reloadedSecret.Status != "active" {
		t.Fatalf("expected rejected disable not to mutate secret, got status=%s", reloadedSecret.Status)
	}

	disableVerifiedRecorder := performMarketplaceRequestWithSession(
		t,
		http.MethodPost,
		"/api/marketplace/admin/seller-secrets/"+strconv.Itoa(secret.Id)+"/disable",
		SellerSecretActionRequest{Reason: "risk disable"},
		user.Id,
		true,
		func(engine *gin.Engine) {
			engine.POST("/api/marketplace/admin/seller-secrets/:id/disable", middleware.SecureVerificationRequired(), DisableSellerSecretAdmin)
		},
	)
	disableVerifiedResp := decodeMarketplaceResponse(t, disableVerifiedRecorder)
	if !disableVerifiedResp.Success {
		t.Fatalf("expected disable with secure verification to succeed, got message: %s", disableVerifiedResp.Message)
	}

	recoverRecorder := performMarketplaceRequestWithSession(
		t,
		http.MethodPost,
		"/api/marketplace/admin/seller-secrets/"+strconv.Itoa(secret.Id)+"/recover",
		SellerSecretActionRequest{Reason: "risk recover"},
		user.Id,
		false,
		func(engine *gin.Engine) {
			engine.POST("/api/marketplace/admin/seller-secrets/:id/recover", middleware.SecureVerificationRequired(), RecoverSellerSecretAdmin)
		},
	)
	recoverResp := decodeMarketplaceResponse(t, recoverRecorder)
	if recoverResp.Success {
		t.Fatalf("expected recover without secure verification to fail")
	}
	if recoverResp.Message != "需要安全验证" {
		t.Fatalf("expected secure verification failure message, got %q", recoverResp.Message)
	}

	recoverVerifiedRecorder := performMarketplaceRequestWithSession(
		t,
		http.MethodPost,
		"/api/marketplace/admin/seller-secrets/"+strconv.Itoa(secret.Id)+"/recover",
		SellerSecretActionRequest{Reason: "risk recover"},
		user.Id,
		true,
		func(engine *gin.Engine) {
			engine.POST("/api/marketplace/admin/seller-secrets/:id/recover", middleware.SecureVerificationRequired(), RecoverSellerSecretAdmin)
		},
	)
	recoverVerifiedResp := decodeMarketplaceResponse(t, recoverVerifiedRecorder)
	if !recoverVerifiedResp.Success {
		t.Fatalf("expected recover with secure verification to succeed, got message: %s", recoverVerifiedResp.Message)
	}
}

func TestSellerSecretRiskActionsRequireReason(t *testing.T) {
	db := setupMarketplaceControllerTestDB(t)
	user := seedMarketplaceUser(t, db, "secret-risk-reason-user")
	seller, supply := seedMarketplaceSellerWithSupply(t, db, user.Id)
	secret := &model.SellerSecret{
		SellerId:        seller.Id,
		SupplyAccountId: supply.Id,
		SecretType:      "api_key",
		ProviderCode:    "openai",
		Ciphertext:      `{"alg":"aes-256-gcm","kid":"v1","nonce":"bm9uY2U=","ciphertext":"Y2lwaGVy"}`,
		CipherVersion:   "v1",
		Fingerprint:     "fp-secret-risk-reason",
		MaskedValue:     "sk-***reason",
		Status:          "active",
		VerifyStatus:    "success",
		VerifyMessage:   "verified",
	}
	if err := db.Create(secret).Error; err != nil {
		t.Fatalf("failed to seed seller secret: %v", err)
	}

	disableRecorder := performMarketplaceRequestWithSession(
		t,
		http.MethodPost,
		"/api/marketplace/admin/seller-secrets/"+strconv.Itoa(secret.Id)+"/disable",
		SellerSecretActionRequest{Reason: "   "},
		user.Id,
		true,
		func(engine *gin.Engine) {
			engine.POST("/api/marketplace/admin/seller-secrets/:id/disable", middleware.SecureVerificationRequired(), DisableSellerSecretAdmin)
		},
	)
	disableResp := decodeMarketplaceResponse(t, disableRecorder)
	if disableResp.Success {
		t.Fatalf("expected disable without meaningful reason to fail")
	}
	if disableResp.Message != "reason is required" {
		t.Fatalf("expected missing reason message, got %q", disableResp.Message)
	}

	genericReasonRecorder := performMarketplaceRequestWithSession(
		t,
		http.MethodPost,
		"/api/marketplace/admin/seller-secrets/"+strconv.Itoa(secret.Id)+"/recover",
		SellerSecretActionRequest{Reason: "test"},
		user.Id,
		true,
		func(engine *gin.Engine) {
			engine.POST("/api/marketplace/admin/seller-secrets/:id/recover", middleware.SecureVerificationRequired(), RecoverSellerSecretAdmin)
		},
	)
	genericReasonResp := decodeMarketplaceResponse(t, genericReasonRecorder)
	if genericReasonResp.Success {
		t.Fatalf("expected recover with generic reason to fail")
	}
	if genericReasonResp.Message != "reason is too generic" {
		t.Fatalf("expected generic reason message, got %q", genericReasonResp.Message)
	}
}

func TestSellerSecretRiskActionsRequireRootAuth(t *testing.T) {
	db := setupMarketplaceControllerTestDB(t)
	user := seedMarketplaceUser(t, db, "secret-risk-root-user")
	seller, supply := seedMarketplaceSellerWithSupply(t, db, user.Id)
	secret := &model.SellerSecret{
		SellerId:        seller.Id,
		SupplyAccountId: supply.Id,
		SecretType:      "api_key",
		ProviderCode:    "openai",
		Ciphertext:      `{"alg":"aes-256-gcm","kid":"v1","nonce":"bm9uY2U=","ciphertext":"Y2lwaGVy"}`,
		CipherVersion:   "v1",
		Fingerprint:     "fp-secret-risk-root",
		MaskedValue:     "sk-***root",
		Status:          "active",
		VerifyStatus:    "success",
		VerifyMessage:   "verified",
	}
	if err := db.Create(secret).Error; err != nil {
		t.Fatalf("failed to seed seller secret: %v", err)
	}

	adminDisableRecorder := performMarketplaceAuthedRequestWithSession(
		t,
		http.MethodPost,
		"/api/marketplace/admin/seller-secrets/"+strconv.Itoa(secret.Id)+"/disable",
		SellerSecretActionRequest{Reason: "root-only disable"},
		user.Id,
		common.RoleAdminUser,
		true,
		"req-risk-admin-denied",
		"198.51.100.21",
		func(engine *gin.Engine) {
			engine.POST(
				"/api/marketplace/admin/seller-secrets/:id/disable",
				middleware.AdminAuth(),
				middleware.RootAuth(),
				middleware.SecureVerificationRequired(),
				DisableSellerSecretAdmin,
			)
		},
	)
	adminDisableResp := decodeMarketplaceResponse(t, adminDisableRecorder)
	if adminDisableResp.Success {
		t.Fatalf("expected ordinary admin disable to be rejected")
	}

	reloadedSecret, err := model.GetSellerSecretByID(secret.Id)
	if err != nil {
		t.Fatalf("failed to reload seller secret after denied admin disable: %v", err)
	}
	if reloadedSecret.Status != "active" {
		t.Fatalf("expected denied admin disable not to mutate secret, got status=%s", reloadedSecret.Status)
	}

	rootDisableRecorder := performMarketplaceAuthedRequestWithSession(
		t,
		http.MethodPost,
		"/api/marketplace/admin/seller-secrets/"+strconv.Itoa(secret.Id)+"/disable",
		SellerSecretActionRequest{Reason: "root-only disable"},
		user.Id,
		common.RoleRootUser,
		true,
		"req-risk-root-disable",
		"198.51.100.22",
		func(engine *gin.Engine) {
			engine.POST(
				"/api/marketplace/admin/seller-secrets/:id/disable",
				middleware.RootAuth(),
				middleware.SecureVerificationRequired(),
				DisableSellerSecretAdmin,
			)
		},
	)
	rootDisableResp := decodeMarketplaceResponse(t, rootDisableRecorder)
	if !rootDisableResp.Success {
		t.Fatalf("expected root disable to succeed, got message: %s", rootDisableResp.Message)
	}
	var disableAudit model.SellerSecretAudit
	if err := db.Where("seller_secret_id = ? AND action = ?", secret.Id, "disable").Order("id desc").First(&disableAudit).Error; err != nil {
		t.Fatalf("failed to load disable audit: %v", err)
	}
	if disableAudit.ActorType != "root" || disableAudit.RequestId != "req-risk-root-disable" || disableAudit.Ip != "198.51.100.22" {
		t.Fatalf("expected disable audit context to be persisted, got %+v", disableAudit)
	}
	if disableAudit.Meta == "" || !strings.Contains(disableAudit.Meta, "request_path") {
		t.Fatalf("expected disable audit meta to contain request context, got %q", disableAudit.Meta)
	}

	rootRecoverRecorder := performMarketplaceAuthedRequestWithSession(
		t,
		http.MethodPost,
		"/api/marketplace/admin/seller-secrets/"+strconv.Itoa(secret.Id)+"/recover",
		SellerSecretActionRequest{Reason: "root-only recover"},
		user.Id,
		common.RoleRootUser,
		true,
		"req-risk-root-recover",
		"198.51.100.23",
		func(engine *gin.Engine) {
			engine.POST(
				"/api/marketplace/admin/seller-secrets/:id/recover",
				middleware.RootAuth(),
				middleware.SecureVerificationRequired(),
				RecoverSellerSecretAdmin,
			)
		},
	)
	rootRecoverResp := decodeMarketplaceResponse(t, rootRecoverRecorder)
	if !rootRecoverResp.Success {
		t.Fatalf("expected root recover to succeed, got message: %s", rootRecoverResp.Message)
	}
	var recoverAudit model.SellerSecretAudit
	if err := db.Where("seller_secret_id = ? AND action = ?", secret.Id, "recover").Order("id desc").First(&recoverAudit).Error; err != nil {
		t.Fatalf("failed to load recover audit: %v", err)
	}
	if recoverAudit.ActorType != "root" || recoverAudit.RequestId != "req-risk-root-recover" || recoverAudit.Ip != "198.51.100.23" {
		t.Fatalf("expected recover audit context to be persisted, got %+v", recoverAudit)
	}
	if recoverAudit.Meta == "" || !strings.Contains(recoverAudit.Meta, "request_path") {
		t.Fatalf("expected recover audit meta to contain request context, got %q", recoverAudit.Meta)
	}
}

func TestSellerSecretRiskActionsRejectNonRootAtHandler(t *testing.T) {
	db := setupMarketplaceControllerTestDB(t)
	user := seedMarketplaceUser(t, db, "secret-risk-handler-guard")
	seller, supply := seedMarketplaceSellerWithSupply(t, db, user.Id)
	secret := &model.SellerSecret{
		SellerId:        seller.Id,
		SupplyAccountId: supply.Id,
		SecretType:      "api_key",
		ProviderCode:    "openai",
		Ciphertext:      `{"alg":"aes-256-gcm","kid":"v1","nonce":"bm9uY2U=","ciphertext":"Y2lwaGVy"}`,
		CipherVersion:   "v1",
		Fingerprint:     "fp-secret-handler-guard",
		MaskedValue:     "sk-***guard",
		Status:          "active",
		VerifyStatus:    "success",
		VerifyMessage:   "verified",
	}
	if err := db.Create(secret).Error; err != nil {
		t.Fatalf("failed to seed seller secret: %v", err)
	}

	adminRecorder := performMarketplaceAuthedRequestWithSession(
		t,
		http.MethodPost,
		"/api/marketplace/admin/seller-secrets/"+strconv.Itoa(secret.Id)+"/disable",
		SellerSecretActionRequest{Reason: "non-root guard"},
		user.Id,
		common.RoleAdminUser,
		true,
		"req-risk-handler-guard",
		"198.51.100.24",
		func(engine *gin.Engine) {
			engine.POST("/api/marketplace/admin/seller-secrets/:id/disable", middleware.SecureVerificationRequired(), DisableSellerSecretAdmin)
		},
	)
	adminResp := decodeMarketplaceResponse(t, adminRecorder)
	if adminResp.Success {
		t.Fatalf("expected non-root admin to be rejected by handler guard")
	}
}
