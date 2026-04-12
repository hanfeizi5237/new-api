package controller

import (
	"net/http"
	"strconv"
	"testing"

	"github.com/QuantumNous/new-api/common"
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
