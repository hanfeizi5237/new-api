package router

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/middleware"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestMarketplaceSellerSecretRiskRoutesRequireRoot(t *testing.T) {
	db := setupMarketplaceSellerSecretRouteTestDB(t)
	user := &model.User{
		Username:    "marketplace-risk-root-route",
		Password:    "password123",
		DisplayName: "marketplace-risk-root-route",
		Role:        common.RoleRootUser,
		Status:      common.UserStatusEnabled,
		Group:       "default",
		AffCode:     "aff-marketplace-risk-root-route",
	}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("failed to create route test user: %v", err)
	}
	supply := &model.SupplyAccount{
		SellerId:         1,
		SupplyCode:       "route-supply-risk-root",
		ProviderCode:     "openai",
		VendorId:         1,
		ModelName:        "gpt-4o-mini",
		QuotaUnit:        "token",
		TotalCapacity:    10000,
		SellableCapacity: 8000,
		Status:           "active",
		VerifyStatus:     "success",
	}
	if err := db.Create(supply).Error; err != nil {
		t.Fatalf("failed to create route test supply account: %v", err)
	}
	seller := &model.SellerProfile{
		UserId:      user.Id,
		SellerCode:  "route-seller-risk-root",
		DisplayName: "Route Seller Root",
		Status:      "active",
	}
	if err := db.Create(seller).Error; err != nil {
		t.Fatalf("failed to create route test seller profile: %v", err)
	}
	supply.SellerId = seller.Id
	if err := db.Save(supply).Error; err != nil {
		t.Fatalf("failed to update route test supply seller id: %v", err)
	}
	secret := &model.SellerSecret{
		SellerId:        seller.Id,
		SupplyAccountId: supply.Id,
		SecretType:      "api_key",
		ProviderCode:    "openai",
		Ciphertext:      `{"alg":"aes-256-gcm","kid":"v1","nonce":"bm9uY2U=","ciphertext":"Y2lwaGVy"}`,
		CipherVersion:   "v1",
		Fingerprint:     "fp-route-root-risk",
		MaskedValue:     "sk-***route",
		Status:          "active",
		VerifyStatus:    "success",
		VerifyMessage:   "verified",
	}
	if err := db.Create(secret).Error; err != nil {
		t.Fatalf("failed to create route test seller secret: %v", err)
	}

	adminRecorder := performMarketplaceSellerSecretRouteRequest(t, user.Id, common.RoleAdminUser, secret.Id, "route root disable", true)
	if strings.Contains(adminRecorder.Body.String(), `"success":true`) {
		t.Fatalf("expected ordinary admin request to be rejected, got body=%q", adminRecorder.Body.String())
	}

	reloadedSecret, err := model.GetSellerSecretByID(secret.Id)
	if err != nil {
		t.Fatalf("failed to reload secret after denied admin request: %v", err)
	}
	if reloadedSecret.Status != "active" {
		t.Fatalf("expected denied admin route request not to mutate secret, got status=%s", reloadedSecret.Status)
	}

	rootRecorder := performMarketplaceSellerSecretRouteRequest(t, user.Id, common.RoleRootUser, secret.Id, "route root disable", true)
	if !strings.Contains(rootRecorder.Body.String(), `"success":true`) {
		t.Fatalf("expected root route request to succeed, got body=%q", rootRecorder.Body.String())
	}
}

func performMarketplaceSellerSecretRouteRequest(t *testing.T, actorUserID int, actorRole int, secretID int, reason string, secureVerified bool) *httptest.ResponseRecorder {
	t.Helper()

	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(sessions.Sessions("session", cookie.NewStore([]byte("marketplace-seller-secret-route-test-secret"))))
	engine.Use(middleware.RequestId())
	engine.Use(func(c *gin.Context) {
		session := sessions.Default(c)
		session.Set("id", actorUserID)
		session.Set("username", "marketplace-route-user")
		session.Set("role", actorRole)
		session.Set("status", common.UserStatusEnabled)
		session.Set("group", "default")
		if secureVerified {
			session.Set(middleware.SecureVerificationSessionKey, common.GetTimestamp())
			session.Set("secure_verified_method", "2fa")
		}
		if err := session.Save(); err != nil {
			t.Fatalf("failed to seed route test session: %v", err)
		}
		c.Next()
	})
	SetApiRouter(engine)

	requestBody := []byte(fmt.Sprintf(`{"reason":"%s"}`, reason))
	request := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/marketplace/admin/seller-secrets/%d/disable", secretID), bytes.NewReader(requestBody))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("New-Api-User", fmt.Sprintf("%d", actorUserID))
	request.RemoteAddr = "198.51.100.60:12345"

	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, request)
	return recorder
}

func setupMarketplaceSellerSecretRouteTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	previousDB := model.DB
	previousLogDB := model.LOG_DB
	previousUsingSQLite := common.UsingSQLite
	previousUsingMySQL := common.UsingMySQL
	previousUsingPostgreSQL := common.UsingPostgreSQL
	previousRedisEnabled := common.RedisEnabled
	previousMemoryCacheEnabled := common.MemoryCacheEnabled

	common.UsingSQLite = true
	common.UsingMySQL = false
	common.UsingPostgreSQL = false
	common.RedisEnabled = false
	common.MemoryCacheEnabled = false

	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}
	model.DB = db
	model.LOG_DB = db

	if err := db.AutoMigrate(
		&model.User{},
		&model.Log{},
		&model.SellerProfile{},
		&model.SupplyAccount{},
		&model.SellerSecret{},
		&model.SellerSecretAudit{},
	); err != nil {
		t.Fatalf("failed to migrate seller secret route test tables: %v", err)
	}

	t.Cleanup(func() {
		model.DB = previousDB
		model.LOG_DB = previousLogDB
		common.UsingSQLite = previousUsingSQLite
		common.UsingMySQL = previousUsingMySQL
		common.UsingPostgreSQL = previousUsingPostgreSQL
		common.RedisEnabled = previousRedisEnabled
		common.MemoryCacheEnabled = previousMemoryCacheEnabled

		sqlDB, err := db.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	})

	return db
}
