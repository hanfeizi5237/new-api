package controller

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting"
	"github.com/QuantumNous/new-api/setting/ratio_setting"
	"github.com/gin-gonic/gin"
)

func TestGetPricingUsesDisplayGroupRatio(t *testing.T) {
	db := setupModelListControllerTestDB(t)

	originalGroupRatio := ratio_setting.GroupRatio2JSONString()
	originalDisplayGroupRatio := ratio_setting.DisplayGroupRatio2JSONString()
	originalUserUsableGroups := setting.UserUsableGroups2JSONString()
	t.Cleanup(func() {
		_ = ratio_setting.UpdateGroupRatioByJSONString(originalGroupRatio)
		_ = ratio_setting.UpdateDisplayGroupRatioByJSONString(originalDisplayGroupRatio)
		_ = setting.UpdateUserUsableGroupsByJSONString(originalUserUsableGroups)
		model.InvalidatePricingCache()
	})

	if err := ratio_setting.UpdateGroupRatioByJSONString(`{"default":1.5}`); err != nil {
		t.Fatalf("failed to set group ratio: %v", err)
	}
	if err := ratio_setting.UpdateDisplayGroupRatioByJSONString(`{"default":1.0}`); err != nil {
		t.Fatalf("failed to set display group ratio: %v", err)
	}
	if err := setting.UpdateUserUsableGroupsByJSONString(`{"default":"Default group"}`); err != nil {
		t.Fatalf("failed to set user usable groups: %v", err)
	}

	if err := db.Create(&model.User{
		Id:       2001,
		Username: "pricing-user",
		Password: "password",
		Group:    "default",
		Status:   common.UserStatusEnabled,
	}).Error; err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	if err := db.Create(&model.Channel{
		Id:     1,
		Name:   "pricing-channel",
		Group:  "default",
		Models: "zz-display-ratio-model",
		Type:   constant.ChannelTypeOpenAI,
		Status: common.ChannelStatusEnabled,
	}).Error; err != nil {
		t.Fatalf("failed to create channel: %v", err)
	}

	if err := db.Create(&model.Ability{
		Group:     "default",
		Model:     "zz-display-ratio-model",
		ChannelId: 1,
		Enabled:   true,
	}).Error; err != nil {
		t.Fatalf("failed to create ability: %v", err)
	}

	model.InvalidatePricingCache()

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/pricing", nil)
	ctx.Set("id", 2001)

	GetPricing(ctx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}

	var payload struct {
		Success    bool               `json:"success"`
		GroupRatio map[string]float64 `json:"group_ratio"`
	}
	if err := common.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !payload.Success {
		t.Fatal("expected success response")
	}
	if got := payload.GroupRatio["default"]; got != 1.0 {
		t.Fatalf("expected display ratio 1.0, got %v", got)
	}
}
