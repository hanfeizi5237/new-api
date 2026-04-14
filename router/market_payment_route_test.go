package router

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestMarketPayRouteUsesCriticalRateLimit(t *testing.T) {
	db := setupMarketRouteTestDB(t)
	user := seedMarketRouteUser(t, db, "market-route-rate-limit")
	token := "market-route-access-token"
	user.SetAccessToken(token)
	if err := db.Save(user).Error; err != nil {
		t.Fatalf("failed to persist access token: %v", err)
	}

	previousCriticalEnabled := common.CriticalRateLimitEnable
	previousCriticalNum := common.CriticalRateLimitNum
	previousCriticalDuration := common.CriticalRateLimitDuration
	previousGlobalAPIEnabled := common.GlobalApiRateLimitEnable
	previousRedisEnabled := common.RedisEnabled
	common.CriticalRateLimitEnable = true
	common.CriticalRateLimitNum = 1
	common.CriticalRateLimitDuration = 60
	common.GlobalApiRateLimitEnable = false
	common.RedisEnabled = false
	t.Cleanup(func() {
		common.CriticalRateLimitEnable = previousCriticalEnabled
		common.CriticalRateLimitNum = previousCriticalNum
		common.CriticalRateLimitDuration = previousCriticalDuration
		common.GlobalApiRateLimitEnable = previousGlobalAPIEnabled
		common.RedisEnabled = previousRedisEnabled
	})

	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(sessions.Sessions("session", cookie.NewStore([]byte("market-route-test-secret"))))
	SetApiRouter(engine)

	requestBody := []byte(`{"payment_method":"stripe"}`)
	firstReq := httptest.NewRequest(http.MethodPost, "/api/market/orders/999/pay", bytes.NewReader(requestBody))
	firstReq.Header.Set("Content-Type", "application/json")
	firstReq.Header.Set("Authorization", token)
	firstReq.Header.Set("New-Api-User", fmt.Sprintf("%d", user.Id))
	firstRecorder := httptest.NewRecorder()
	engine.ServeHTTP(firstRecorder, firstReq)
	if firstRecorder.Code == http.StatusTooManyRequests {
		t.Fatalf("expected first pay request not to be rate limited")
	}

	secondReq := httptest.NewRequest(http.MethodPost, "/api/market/orders/999/pay", bytes.NewReader(requestBody))
	secondReq.Header.Set("Content-Type", "application/json")
	secondReq.Header.Set("Authorization", token)
	secondReq.Header.Set("New-Api-User", fmt.Sprintf("%d", user.Id))
	secondRecorder := httptest.NewRecorder()
	engine.ServeHTTP(secondRecorder, secondReq)
	if secondRecorder.Code != http.StatusTooManyRequests {
		t.Fatalf("expected second pay request to be rate limited, got status=%d body=%q", secondRecorder.Code, secondRecorder.Body.String())
	}
}

func setupMarketRouteTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	previousDB := model.DB
	previousLogDB := model.LOG_DB
	previousUsingSQLite := common.UsingSQLite
	previousUsingMySQL := common.UsingMySQL
	previousUsingPostgreSQL := common.UsingPostgreSQL

	common.UsingSQLite = true
	common.UsingMySQL = false
	common.UsingPostgreSQL = false

	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}
	model.DB = db
	model.LOG_DB = db

	if err := db.AutoMigrate(&model.User{}, &model.MarketOrder{}); err != nil {
		t.Fatalf("failed to migrate market route test tables: %v", err)
	}

	t.Cleanup(func() {
		model.DB = previousDB
		model.LOG_DB = previousLogDB
		common.UsingSQLite = previousUsingSQLite
		common.UsingMySQL = previousUsingMySQL
		common.UsingPostgreSQL = previousUsingPostgreSQL

		sqlDB, err := db.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	})

	return db
}

func seedMarketRouteUser(t *testing.T, db *gorm.DB, username string) *model.User {
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
		t.Fatalf("failed to create route test user: %v", err)
	}
	return user
}
