package controller

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

type marketplaceAPIResponse struct {
	Success bool            `json:"success"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func setupMarketplaceControllerTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	gin.SetMode(gin.TestMode)
	common.UsingSQLite = true
	common.UsingMySQL = false
	common.UsingPostgreSQL = false
	common.RedisEnabled = false

	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}
	model.DB = db
	model.LOG_DB = db

	if err := db.AutoMigrate(
		&model.User{},
		&model.Vendor{},
		&model.Channel{},
		&model.Log{},
		&model.SellerProfile{},
		&model.SupplyAccount{},
		&model.SellerSecret{},
		&model.SellerSecretAudit{},
		&model.SupplyChannelBinding{},
		&model.Listing{},
		&model.ListingSKU{},
		&model.InventorySnapshot{},
		&model.MarketOrder{},
		&model.MarketOrderItem{},
		&model.BuyerEntitlement{},
		&model.EntitlementLot{},
		&model.UsageLedger{},
	); err != nil {
		t.Fatalf("failed to migrate marketplace admin tables: %v", err)
	}

	t.Cleanup(func() {
		sqlDB, err := db.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	})

	return db
}

func seedMarketplaceUser(t *testing.T, db *gorm.DB, username string) *model.User {
	t.Helper()

	user := &model.User{
		Username:    username,
		Password:    "password123",
		DisplayName: username,
		Role:        common.RoleCommonUser,
		Status:      common.UserStatusEnabled,
		Group:       "default",
		AffCode:     fmt.Sprintf("aff-%s", common.GetRandomString(8)),
	}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("failed to create user: %v", err)
	}
	return user
}

func seedMarketplaceSellerWithSupply(t *testing.T, db *gorm.DB, userId int) (*model.SellerProfile, *model.SupplyAccount) {
	t.Helper()

	seller := &model.SellerProfile{
		UserId:      userId,
		SellerCode:  fmt.Sprintf("seller-%d", userId),
		DisplayName: fmt.Sprintf("Seller-%d", userId),
		Status:      "active",
	}
	if err := db.Create(seller).Error; err != nil {
		t.Fatalf("failed to create seller: %v", err)
	}
	vendor := &model.Vendor{
		Name:   fmt.Sprintf("vendor-%d", userId),
		Status: 1,
	}
	if err := db.Create(vendor).Error; err != nil {
		t.Fatalf("failed to create vendor: %v", err)
	}

	supply := &model.SupplyAccount{
		SellerId:         seller.Id,
		SupplyCode:       fmt.Sprintf("supply-%d", userId),
		ProviderCode:     "openai",
		VendorId:         vendor.Id,
		ModelName:        "gpt-4o-mini",
		QuotaUnit:        "token",
		TotalCapacity:    100000,
		SellableCapacity: 80000,
		Status:           "active",
	}
	if err := db.Create(supply).Error; err != nil {
		t.Fatalf("failed to create supply account: %v", err)
	}

	snapshot := &model.InventorySnapshot{
		SupplyAccountId: supply.Id,
		AvailableAmount: supply.SellableCapacity,
		RiskDiscountBps: 10000,
		HealthScore:     100,
		SyncStatus:      "ok",
	}
	if err := db.Create(snapshot).Error; err != nil {
		t.Fatalf("failed to create inventory snapshot: %v", err)
	}
	return seller, supply
}

func seedMarketplaceChannelBinding(t *testing.T, db *gorm.DB, supply *model.SupplyAccount, key string) (*model.Channel, *model.SupplyChannelBinding) {
	t.Helper()

	channel := &model.Channel{
		Name:   fmt.Sprintf("channel-%d", supply.Id),
		Key:    key,
		Status: common.ChannelStatusEnabled,
		Models: supply.ModelName,
		Group:  "default",
	}
	if err := db.Create(channel).Error; err != nil {
		t.Fatalf("failed to create channel: %v", err)
	}
	binding := &model.SupplyChannelBinding{
		SupplyAccountId: supply.Id,
		ChannelId:       channel.Id,
		BindingRole:     "primary",
		Status:          "active",
	}
	if err := db.Create(binding).Error; err != nil {
		t.Fatalf("failed to create supply channel binding: %v", err)
	}
	return channel, binding
}

func makeMarketplaceCiphertext(t *testing.T, key string, plaintext string, version string) string {
	t.Helper()

	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		t.Fatalf("failed to create cipher: %v", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		t.Fatalf("failed to create gcm: %v", err)
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		t.Fatalf("failed to read nonce: %v", err)
	}
	sealed := gcm.Seal(nil, nonce, []byte(plaintext), nil)
	payload := map[string]string{
		"alg":        "aes-256-gcm",
		"kid":        version,
		"nonce":      base64.StdEncoding.EncodeToString(nonce),
		"ciphertext": base64.StdEncoding.EncodeToString(sealed),
	}
	bytes, err := common.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal ciphertext payload: %v", err)
	}
	return string(bytes)
}

func newMarketplaceContext(t *testing.T, method string, target string, body any, actorUserId int) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()

	var requestBody *bytes.Reader
	if body != nil {
		payload, err := common.Marshal(body)
		if err != nil {
			t.Fatalf("failed to marshal request body: %v", err)
		}
		requestBody = bytes.NewReader(payload)
	} else {
		requestBody = bytes.NewReader(nil)
	}

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(method, target, requestBody)
	if body != nil {
		ctx.Request.Header.Set("Content-Type", "application/json")
	}
	ctx.Set("id", actorUserId)
	return ctx, recorder
}

func decodeMarketplaceResponse(t *testing.T, recorder *httptest.ResponseRecorder) marketplaceAPIResponse {
	t.Helper()

	var response marketplaceAPIResponse
	if err := common.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to decode marketplace api response: %v", err)
	}
	return response
}
