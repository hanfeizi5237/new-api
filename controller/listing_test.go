package controller

import (
	"net/http"
	"strconv"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/middleware"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-gonic/gin"
)

type listingPageResponse struct {
	Items []ListingAdminView `json:"items"`
	Total int                `json:"total"`
}

func TestListingAdminCreateListAndUpdateStatus(t *testing.T) {
	db := setupMarketplaceControllerTestDB(t)
	user := seedMarketplaceUser(t, db, "listing-admin-user")
	seller, supply := seedMarketplaceSellerWithSupply(t, db, user.Id)
	supply.VerifyStatus = "success"
	if err := db.Save(supply).Error; err != nil {
		t.Fatalf("failed to update supply verify status: %v", err)
	}
	_, _ = seedMarketplaceChannelBinding(t, db, supply, "listing-runtime-key")
	if err := db.Create(&model.SellerSecret{
		SellerId:        seller.Id,
		SupplyAccountId: supply.Id,
		SecretType:      "api_key",
		ProviderCode:    "openai",
		Ciphertext:      `{"alg":"aes-256-gcm","kid":"v1","nonce":"bm9uY2U=","ciphertext":"Y2lwaGVy"}`,
		CipherVersion:   "v1",
		Fingerprint:     "fp-listing-active",
		MaskedValue:     "sk-***listing",
		Status:          "active",
		VerifyStatus:    "success",
		VerifyMessage:   "seeded active secret",
	}).Error; err != nil {
		t.Fatalf("failed to seed active secret: %v", err)
	}

	createBody := CreateListingAdminRequest{
		Listing: model.Listing{
			SellerId:        seller.Id,
			SupplyAccountId: supply.Id,
			ListingCode:     "listing-admin-001",
			Title:           "GPT Mini Package",
			SaleMode:        "fixed_price",
			PricingUnit:     "per_token_package",
		},
		SKUs: []model.ListingSKU{
			{
				SkuCode:        "sku-admin-001",
				PackageAmount:  1000,
				PackageUnit:    "token",
				UnitPriceMinor: 199,
				MinQuantity:    1,
				MaxQuantity:    5,
			},
		},
	}
	createCtx, createRecorder := newMarketplaceContext(t, http.MethodPost, "/api/listing/admin", createBody, user.Id)
	CreateListingAdmin(createCtx)
	createResp := decodeMarketplaceResponse(t, createRecorder)
	if !createResp.Success {
		t.Fatalf("expected create listing success, got message: %s", createResp.Message)
	}

	var created ListingAdminView
	if err := common.Unmarshal(createResp.Data, &created); err != nil {
		t.Fatalf("failed to decode created listing: %v", err)
	}
	if created.Listing.Id <= 0 || len(created.SKUs) != 1 {
		t.Fatalf("expected created listing with one sku, got %+v", created)
	}

	listCtx, listRecorder := newMarketplaceContext(t, http.MethodGet, "/api/listing/admin?keyword=GPT&p=1&size=10", nil, user.Id)
	GetListingAdmin(listCtx)
	listResp := decodeMarketplaceResponse(t, listRecorder)
	if !listResp.Success {
		t.Fatalf("expected list listing success, got message: %s", listResp.Message)
	}

	var page listingPageResponse
	if err := common.Unmarshal(listResp.Data, &page); err != nil {
		t.Fatalf("failed to decode listing page: %v", err)
	}
	if page.Total != 1 || len(page.Items) != 1 {
		t.Fatalf("expected one listing in page, got total=%d items=%d", page.Total, len(page.Items))
	}

	submitReviewBody := UpdateListingStatusRequest{
		AuditStatus: "pending_review",
		AuditRemark: "submit",
	}
	submitReviewCtx, submitReviewRecorder := newMarketplaceContext(t, http.MethodPut, "/api/listing/admin/"+strconv.Itoa(created.Listing.Id)+"/status", submitReviewBody, user.Id)
	submitReviewCtx.Params = gin.Params{{Key: "id", Value: strconv.Itoa(created.Listing.Id)}}
	UpdateListingAdminStatus(submitReviewCtx)
	submitReviewResp := decodeMarketplaceResponse(t, submitReviewRecorder)
	if !submitReviewResp.Success {
		t.Fatalf("expected listing submit review success, got message: %s", submitReviewResp.Message)
	}

	updateBody := UpdateListingStatusRequest{
		Status:      "active",
		AuditStatus: "approved",
		AuditRemark: "ready",
	}
	updateCtx, updateRecorder := newMarketplaceContext(t, http.MethodPut, "/api/listing/admin/"+strconv.Itoa(created.Listing.Id)+"/status", updateBody, user.Id)
	updateCtx.Params = gin.Params{{Key: "id", Value: strconv.Itoa(created.Listing.Id)}}
	UpdateListingAdminStatus(updateCtx)
	updateResp := decodeMarketplaceResponse(t, updateRecorder)
	if !updateResp.Success {
		t.Fatalf("expected update listing status success, got message: %s", updateResp.Message)
	}

	var updated ListingAdminView
	if err := common.Unmarshal(updateResp.Data, &updated); err != nil {
		t.Fatalf("failed to decode updated listing: %v", err)
	}
	if updated.Listing.Status != "active" || updated.Listing.AuditStatus != "approved" {
		t.Fatalf("expected active approved listing, got %+v", updated.Listing)
	}

	var audits []model.MarketplaceOperationAudit
	if err := db.Order("id asc").Find(&audits).Error; err != nil {
		t.Fatalf("failed to load listing audits: %v", err)
	}
	if len(audits) != 2 {
		t.Fatalf("expected exactly two listing audits, got %d", len(audits))
	}
	if audits[1].Action != "listing_status_update" || audits[1].TargetType != "listing" {
		t.Fatalf("expected listing audit action/target, got %+v", audits[1])
	}
	if audits[1].ActorUserId != user.Id || audits[1].ActorType != "admin" {
		t.Fatalf("expected listing audit actor info, got %+v", audits[1])
	}
	if audits[1].Reason != "ready" || audits[1].Result != "success" {
		t.Fatalf("expected listing audit reason/result, got %+v", audits[1])
	}
	beforeState := map[string]interface{}{}
	if err := common.UnmarshalJsonStr(audits[1].BeforeState, &beforeState); err != nil {
		t.Fatalf("failed to decode listing before_state: %v", err)
	}
	afterState := map[string]interface{}{}
	if err := common.UnmarshalJsonStr(audits[1].AfterState, &afterState); err != nil {
		t.Fatalf("failed to decode listing after_state: %v", err)
	}
	if beforeState["status"] != "paused" || beforeState["audit_status"] != "pending_review" {
		t.Fatalf("expected listing before_state paused/pending_review, got %+v", beforeState)
	}
	if afterState["status"] != "active" || afterState["audit_status"] != "approved" {
		t.Fatalf("expected listing after_state active/approved, got %+v", afterState)
	}
	meta := map[string]interface{}{}
	if err := common.UnmarshalJsonStr(audits[1].Meta, &meta); err != nil {
		t.Fatalf("failed to decode listing audit meta: %v", err)
	}
	if meta["seller_id"] != float64(seller.Id) || meta["supply_account_id"] != float64(supply.Id) {
		t.Fatalf("expected listing audit meta seller/supply ids, got %+v", meta)
	}
}

func TestListingAdminRejectsInvalidStatusTransitions(t *testing.T) {
	db := setupMarketplaceControllerTestDB(t)
	user := seedMarketplaceUser(t, db, "listing-admin-invalid")
	seller, supply := seedMarketplaceSellerWithSupply(t, db, user.Id)
	supply.VerifyStatus = "success"
	if err := db.Save(supply).Error; err != nil {
		t.Fatalf("failed to update supply verify status: %v", err)
	}
	_, _ = seedMarketplaceChannelBinding(t, db, supply, "listing-runtime-key")
	if err := db.Create(&model.SellerSecret{
		SellerId:        seller.Id,
		SupplyAccountId: supply.Id,
		SecretType:      "api_key",
		ProviderCode:    "openai",
		Ciphertext:      `{"alg":"aes-256-gcm","kid":"v1","nonce":"bm9uY2U=","ciphertext":"Y2lwaGVy"}`,
		CipherVersion:   "v1",
		Fingerprint:     "fp-listing-invalid",
		MaskedValue:     "sk-***listing",
		Status:          "active",
		VerifyStatus:    "success",
		VerifyMessage:   "seeded active secret",
	}).Error; err != nil {
		t.Fatalf("failed to seed active secret: %v", err)
	}
	createBody := CreateListingAdminRequest{
		Listing: model.Listing{
			SellerId:        seller.Id,
			SupplyAccountId: supply.Id,
			ListingCode:     "listing-admin-invalid-001",
			Title:           "GPT Mini Package",
			SaleMode:        "fixed_price",
			PricingUnit:     "per_token_package",
		},
		SKUs: []model.ListingSKU{
			{
				SkuCode:        "sku-admin-invalid-001",
				PackageAmount:  1000,
				PackageUnit:    "token",
				UnitPriceMinor: 199,
				MinQuantity:    1,
				MaxQuantity:    5,
			},
		},
	}
	createCtx, createRecorder := newMarketplaceContext(t, http.MethodPost, "/api/listing/admin", createBody, user.Id)
	CreateListingAdmin(createCtx)
	createResp := decodeMarketplaceResponse(t, createRecorder)
	if !createResp.Success {
		t.Fatalf("expected create listing success, got message: %s", createResp.Message)
	}
	var created ListingAdminView
	if err := common.Unmarshal(createResp.Data, &created); err != nil {
		t.Fatalf("failed to decode created listing: %v", err)
	}

	invalidActiveBody := UpdateListingStatusRequest{Status: "active"}
	invalidActiveCtx, invalidActiveRecorder := newMarketplaceContext(t, http.MethodPut, "/api/listing/admin/"+strconv.Itoa(created.Listing.Id)+"/status", invalidActiveBody, user.Id)
	invalidActiveCtx.Params = gin.Params{{Key: "id", Value: strconv.Itoa(created.Listing.Id)}}
	UpdateListingAdminStatus(invalidActiveCtx)
	invalidActiveResp := decodeMarketplaceResponse(t, invalidActiveRecorder)
	if invalidActiveResp.Success {
		t.Fatalf("expected draft listing activation to fail")
	}

	invalidStatusBody := UpdateListingStatusRequest{Status: "ghost_status"}
	invalidStatusCtx, invalidStatusRecorder := newMarketplaceContext(t, http.MethodPut, "/api/listing/admin/"+strconv.Itoa(created.Listing.Id)+"/status", invalidStatusBody, user.Id)
	invalidStatusCtx.Params = gin.Params{{Key: "id", Value: strconv.Itoa(created.Listing.Id)}}
	UpdateListingAdminStatus(invalidStatusCtx)
	invalidStatusResp := decodeMarketplaceResponse(t, invalidStatusRecorder)
	if invalidStatusResp.Success {
		t.Fatalf("expected invalid listing status to fail")
	}
}

func TestListingAdminRecoveryRequiresRiskScopeAndSecureVerification(t *testing.T) {
	db := setupMarketplaceControllerTestDB(t)
	user := seedMarketplaceUser(t, db, "listing-admin-recovery")
	seller, supply := seedMarketplaceSellerWithSupply(t, db, user.Id)
	supply.VerifyStatus = "success"
	if err := db.Save(supply).Error; err != nil {
		t.Fatalf("failed to update supply verify status: %v", err)
	}
	_, _ = seedMarketplaceChannelBinding(t, db, supply, "listing-recovery-runtime-key")
	if err := db.Create(&model.SellerSecret{
		SellerId:        seller.Id,
		SupplyAccountId: supply.Id,
		SecretType:      "api_key",
		ProviderCode:    "openai",
		Ciphertext:      `{"alg":"aes-256-gcm","kid":"v1","nonce":"bm9uY2U=","ciphertext":"Y2lwaGVy"}`,
		CipherVersion:   "v1",
		Fingerprint:     "fp-listing-recovery",
		MaskedValue:     "sk-***listing-recovery",
		Status:          "active",
		VerifyStatus:    "success",
		VerifyMessage:   "seeded active secret",
	}).Error; err != nil {
		t.Fatalf("failed to seed active secret: %v", err)
	}

	createBody := CreateListingAdminRequest{
		Listing: model.Listing{
			SellerId:        seller.Id,
			SupplyAccountId: supply.Id,
			ListingCode:     "listing-admin-recovery-001",
			Title:           "GPT Recovery Package",
			SaleMode:        "fixed_price",
			PricingUnit:     "per_token_package",
		},
		SKUs: []model.ListingSKU{
			{
				SkuCode:        "sku-admin-recovery-001",
				PackageAmount:  1000,
				PackageUnit:    "token",
				UnitPriceMinor: 199,
				MinQuantity:    1,
				MaxQuantity:    5,
			},
		},
	}
	createCtx, createRecorder := newMarketplaceContext(t, http.MethodPost, "/api/listing/admin", createBody, user.Id)
	CreateListingAdmin(createCtx)
	createResp := decodeMarketplaceResponse(t, createRecorder)
	if !createResp.Success {
		t.Fatalf("expected create listing success, got message: %s", createResp.Message)
	}

	var created ListingAdminView
	if err := common.Unmarshal(createResp.Data, &created); err != nil {
		t.Fatalf("failed to decode created listing: %v", err)
	}
	if err := db.Model(&model.Listing{}).Where("id = ?", created.Listing.Id).Updates(map[string]any{
		"status":       "paused",
		"audit_status": "approved",
		"audit_remark": "auto paused",
	}).Error; err != nil {
		t.Fatalf("failed to seed paused approved listing state: %v", err)
	}

	recoveryBody := UpdateListingStatusRequest{
		Status:      "active",
		AuditRemark: "manual recovery after inventory sync",
	}

	adminRecorder := performMarketplaceScopedRequestWithSession(
		t,
		http.MethodPut,
		"/api/listing/admin/"+strconv.Itoa(created.Listing.Id)+"/status",
		recoveryBody,
		user.Id,
		common.RoleAdminUser,
		"default",
		true,
		"req-listing-recovery-admin-denied",
		"198.51.100.31",
		func(engine *gin.Engine) {
			engine.PUT("/api/listing/admin/:id/status", middleware.AdminAuth(), UpdateListingAdminStatus)
		},
	)
	adminResp := decodeMarketplaceResponse(t, adminRecorder)
	if adminResp.Success {
		t.Fatalf("expected ordinary admin listing recovery to be rejected")
	}

	riskNoVerifyRecorder := performMarketplaceScopedRequestWithSession(
		t,
		http.MethodPut,
		"/api/listing/admin/"+strconv.Itoa(created.Listing.Id)+"/status",
		recoveryBody,
		user.Id,
		common.RoleAdminUser,
		"market_risk",
		false,
		"req-listing-recovery-risk-no-verify",
		"198.51.100.32",
		func(engine *gin.Engine) {
			engine.PUT("/api/listing/admin/:id/status", middleware.AdminAuth(), UpdateListingAdminStatus)
		},
	)
	riskNoVerifyResp := decodeMarketplaceResponse(t, riskNoVerifyRecorder)
	if riskNoVerifyResp.Success {
		t.Fatalf("expected risk admin listing recovery without secure verification to be rejected")
	}

	riskRecorder := performMarketplaceScopedRequestWithSession(
		t,
		http.MethodPut,
		"/api/listing/admin/"+strconv.Itoa(created.Listing.Id)+"/status",
		recoveryBody,
		user.Id,
		common.RoleAdminUser,
		"market_risk",
		true,
		"req-listing-recovery-risk-success",
		"198.51.100.33",
		func(engine *gin.Engine) {
			engine.PUT("/api/listing/admin/:id/status", middleware.AdminAuth(), UpdateListingAdminStatus)
		},
	)
	riskResp := decodeMarketplaceResponse(t, riskRecorder)
	if !riskResp.Success {
		t.Fatalf("expected risk admin listing recovery to succeed, got message: %s", riskResp.Message)
	}

	updated, err := model.GetListingByID(created.Listing.Id)
	if err != nil {
		t.Fatalf("failed to reload listing: %v", err)
	}
	if updated.Status != "active" || updated.AuditStatus != "approved" {
		t.Fatalf("expected recovered listing to be active and approved, got %+v", updated)
	}
}

func TestListingAdminRejectsInvalidSellerIDQueryAndPathID(t *testing.T) {
	listCtx, listRecorder := newMarketplaceContext(t, http.MethodGet, "/api/listing/admin?seller_id=0", nil, 1)
	GetListingAdmin(listCtx)
	listResp := decodeMarketplaceResponse(t, listRecorder)
	if listResp.Success {
		t.Fatalf("expected invalid seller_id query to fail")
	}
	if listResp.Message != "invalid seller_id" {
		t.Fatalf("expected invalid seller_id message, got %q", listResp.Message)
	}

	updateCtx, updateRecorder := newMarketplaceContext(t, http.MethodPut, "/api/listing/admin/0/status", UpdateListingStatusRequest{}, 1)
	updateCtx.Params = gin.Params{{Key: "id", Value: "0"}}
	UpdateListingAdminStatus(updateCtx)
	updateResp := decodeMarketplaceResponse(t, updateRecorder)
	if updateResp.Success {
		t.Fatalf("expected invalid listing id to fail")
	}
	if updateResp.Message != "invalid listing id" {
		t.Fatalf("expected invalid listing id message, got %q", updateResp.Message)
	}
}
