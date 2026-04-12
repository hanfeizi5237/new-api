package controller

import (
	"net/http"
	"strconv"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-gonic/gin"
)

type sellerCreateResponse struct {
	Seller        model.SellerProfile `json:"seller"`
	SupplyAccount model.SupplyAccount `json:"supply_account"`
}

type sellerPageResponse struct {
	Items []model.SellerProfile `json:"items"`
	Total int                   `json:"total"`
}

func TestSellerAdminCreateListAndUpdateStatus(t *testing.T) {
	db := setupMarketplaceControllerTestDB(t)
	user := seedMarketplaceUser(t, db, "seller-admin-user")

	createBody := CreateSellerAdminRequest{
		Seller: model.SellerProfile{
			UserId:      user.Id,
			SellerCode:  "seller-admin-001",
			DisplayName: "Seller Admin",
			SellerType:  "business",
		},
		SupplyAccount: model.SupplyAccount{
			SupplyCode:       "supply-admin-001",
			ProviderCode:     "openai",
			ModelName:        "gpt-4o-mini",
			QuotaUnit:        "token",
			TotalCapacity:    120000,
			SellableCapacity: 100000,
		},
	}
	createCtx, createRecorder := newMarketplaceContext(t, http.MethodPost, "/api/seller/admin", createBody, user.Id)
	CreateSellerAdmin(createCtx)

	createResp := decodeMarketplaceResponse(t, createRecorder)
	if !createResp.Success {
		t.Fatalf("expected create seller success, got message: %s", createResp.Message)
	}

	var created sellerCreateResponse
	if err := common.Unmarshal(createResp.Data, &created); err != nil {
		t.Fatalf("failed to decode create seller response: %v", err)
	}
	if created.Seller.Id <= 0 {
		t.Fatalf("expected created seller id, got %d", created.Seller.Id)
	}
	if created.SupplyAccount.Id <= 0 || created.SupplyAccount.SellerId != created.Seller.Id {
		t.Fatalf("expected created supply account to belong to seller, got %+v", created.SupplyAccount)
	}

	listCtx, listRecorder := newMarketplaceContext(t, http.MethodGet, "/api/seller/admin?keyword=Seller&p=1&size=10", nil, user.Id)
	GetSellerAdmin(listCtx)
	listResp := decodeMarketplaceResponse(t, listRecorder)
	if !listResp.Success {
		t.Fatalf("expected list seller success, got message: %s", listResp.Message)
	}

	var page sellerPageResponse
	if err := common.Unmarshal(listResp.Data, &page); err != nil {
		t.Fatalf("failed to decode seller page: %v", err)
	}
	if page.Total != 1 || len(page.Items) != 1 {
		t.Fatalf("expected one seller in list, got total=%d items=%d", page.Total, len(page.Items))
	}

	updateBody := UpdateSellerStatusRequest{
		Status: "disabled",
		Remark: "risk-review",
	}
	updateCtx, updateRecorder := newMarketplaceContext(t, http.MethodPut, "/api/seller/admin/"+strconv.Itoa(created.Seller.Id)+"/status", updateBody, user.Id)
	updateCtx.Params = gin.Params{{Key: "id", Value: strconv.Itoa(created.Seller.Id)}}
	UpdateSellerAdminStatus(updateCtx)
	updateResp := decodeMarketplaceResponse(t, updateRecorder)
	if !updateResp.Success {
		t.Fatalf("expected update seller status success, got message: %s", updateResp.Message)
	}

	var updated model.SellerProfile
	if err := common.Unmarshal(updateResp.Data, &updated); err != nil {
		t.Fatalf("failed to decode updated seller: %v", err)
	}
	if updated.Status != "disabled" || updated.Remark != "risk-review" {
		t.Fatalf("expected disabled seller with remark, got %+v", updated)
	}
}
