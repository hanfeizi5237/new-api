package service

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupMarketplaceServiceTestDB(t *testing.T) *gorm.DB {
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
		&model.Token{},
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
		t.Fatalf("failed to migrate marketplace service tables: %v", err)
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

func seedMarketplaceServiceUser(t *testing.T, db *gorm.DB, username string) *model.User {
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

func seedMarketplaceServiceSupply(t *testing.T, db *gorm.DB, user *model.User, supplyStatus string, verifyStatus string, quotaUnit string) (*model.SellerProfile, *model.SupplyAccount) {
	t.Helper()
	if supplyStatus == "" {
		supplyStatus = "active"
	}
	if verifyStatus == "" {
		verifyStatus = "pending"
	}
	if quotaUnit == "" {
		quotaUnit = "token"
	}
	seller := &model.SellerProfile{
		UserId:      user.Id,
		SellerCode:  fmt.Sprintf("seller-%d", user.Id),
		DisplayName: fmt.Sprintf("Seller-%d", user.Id),
		Status:      "active",
	}
	if err := db.Create(seller).Error; err != nil {
		t.Fatalf("failed to create seller: %v", err)
	}
	vendor := &model.Vendor{
		Name:   fmt.Sprintf("vendor-%d", user.Id),
		Status: 1,
	}
	if err := db.Create(vendor).Error; err != nil {
		t.Fatalf("failed to create vendor: %v", err)
	}
	supply := &model.SupplyAccount{
		SellerId:         seller.Id,
		SupplyCode:       fmt.Sprintf("supply-%d", user.Id),
		ProviderCode:     "openai",
		VendorId:         vendor.Id,
		ModelName:        "gpt-4o-mini",
		QuotaUnit:        quotaUnit,
		TotalCapacity:    100000,
		SellableCapacity: 80000,
		Status:           supplyStatus,
		VerifyStatus:     verifyStatus,
	}
	if err := db.Create(supply).Error; err != nil {
		t.Fatalf("failed to create supply: %v", err)
	}
	if err := db.Create(&model.InventorySnapshot{
		SupplyAccountId: supply.Id,
		AvailableAmount: supply.SellableCapacity,
		RiskDiscountBps: 10000,
		HealthScore:     100,
		SyncStatus:      "ok",
	}).Error; err != nil {
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
		t.Fatalf("failed to create binding: %v", err)
	}
	return channel, binding
}

func seedMarketplaceListing(t *testing.T, db *gorm.DB, seller *model.SellerProfile, supply *model.SupplyAccount, packageAmount int64, unitPriceMinor int64) (*model.Listing, *model.ListingSKU) {
	t.Helper()
	if packageAmount <= 0 {
		packageAmount = 1000
	}
	if unitPriceMinor <= 0 {
		unitPriceMinor = 199
	}
	listing := &model.Listing{
		SellerId:        seller.Id,
		SupplyAccountId: supply.Id,
		ListingCode:     fmt.Sprintf("listing-%d-%d", seller.Id, supply.Id),
		Title:           "Marketplace Listing",
		VendorId:        supply.VendorId,
		ModelName:       supply.ModelName,
		SaleMode:        "fixed_price",
		PricingUnit:     "per_token_package",
		ValidityDays:    30,
		AuditStatus:     "approved",
		Status:          "active",
	}
	if err := db.Create(listing).Error; err != nil {
		t.Fatalf("failed to create listing: %v", err)
	}
	sku := &model.ListingSKU{
		ListingId:      listing.Id,
		SkuCode:        fmt.Sprintf("sku-%d-%d", listing.Id, supply.Id),
		PackageAmount:  packageAmount,
		PackageUnit:    "token",
		UnitPriceMinor: unitPriceMinor,
		MinQuantity:    1,
		MaxQuantity:    5,
		Status:         "active",
		SortOrder:      1,
	}
	if err := db.Create(sku).Error; err != nil {
		t.Fatalf("failed to create listing sku: %v", err)
	}
	return listing, sku
}

func makeSellerSecretCiphertext(t *testing.T, key string, plaintext string, version string) string {
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

func seedSellerSecretRecord(t *testing.T, db *gorm.DB, sellerId int, supplyAccountId int, ciphertext string, fingerprint string, status string, verifyStatus string) *model.SellerSecret {
	t.Helper()
	if status == "" {
		status = "draft"
	}
	if verifyStatus == "" {
		verifyStatus = "pending"
	}
	secret := &model.SellerSecret{
		SellerId:        sellerId,
		SupplyAccountId: supplyAccountId,
		SecretType:      "api_key",
		ProviderCode:    "openai",
		Ciphertext:      ciphertext,
		CipherVersion:   "v1",
		Fingerprint:     fingerprint,
		MaskedValue:     "sk-****test",
		Status:          status,
		VerifyStatus:    verifyStatus,
	}
	if err := db.Create(secret).Error; err != nil {
		t.Fatalf("failed to create seller secret: %v", err)
	}
	return secret
}

func TestVerifySellerSecretRejectsMalformedCiphertext(t *testing.T) {
	db := setupMarketplaceServiceTestDB(t)
	user := seedMarketplaceServiceUser(t, db, "svc-secret-malformed")
	seller, supply := seedMarketplaceServiceSupply(t, db, user, "paused", "pending", "token")
	_, _ = seedMarketplaceChannelBinding(t, db, supply, "old-runtime-key")
	secret := seedSellerSecretRecord(t, db, seller.Id, supply.Id, "not-json", "fp-malformed", "draft", "pending")
	t.Setenv("SELLER_SECRET_MASTER_KEY", strings.Repeat("k", 32))

	if _, err := VerifySellerSecret(secret.Id, user.Id); err == nil {
		t.Fatalf("expected verify to fail for malformed ciphertext")
	}

	updatedSecret, err := model.GetSellerSecretByID(secret.Id)
	if err != nil {
		t.Fatalf("failed to reload secret: %v", err)
	}
	if updatedSecret.Status == "active" || updatedSecret.VerifyStatus == "success" {
		t.Fatalf("expected malformed secret to remain unusable, got %+v", updatedSecret)
	}
}

func TestVerifySellerSecretPromotesNewSecretAndRotatesOldSecret(t *testing.T) {
	db := setupMarketplaceServiceTestDB(t)
	user := seedMarketplaceServiceUser(t, db, "svc-secret-second-active")
	seller, supply := seedMarketplaceServiceSupply(t, db, user, "active", "success", "token")
	_, _ = seedMarketplaceChannelBinding(t, db, supply, "old-runtime-key")
	t.Setenv("SELLER_SECRET_MASTER_KEY", strings.Repeat("s", 32))

	activePayload := makeSellerSecretCiphertext(t, strings.Repeat("s", 32), "sk-live-old", "v1")
	_ = seedSellerSecretRecord(t, db, seller.Id, supply.Id, activePayload, "fp-old", "active", "success")
	candidatePayload := makeSellerSecretCiphertext(t, strings.Repeat("s", 32), "sk-live-new", "v1")
	candidate := seedSellerSecretRecord(t, db, seller.Id, supply.Id, candidatePayload, "fp-new", "draft", "pending")

	verified, err := VerifySellerSecret(candidate.Id, user.Id)
	if err != nil {
		t.Fatalf("expected verify to promote candidate secret, got error: %v", err)
	}
	if verified.Status != "active" || verified.VerifyStatus != "success" {
		t.Fatalf("expected candidate secret active/success, got %+v", verified)
	}

	var reloaded []model.SellerSecret
	if err := db.Where("supply_account_id = ?", supply.Id).Order("id asc").Find(&reloaded).Error; err != nil {
		t.Fatalf("failed to reload secrets: %v", err)
	}
	if len(reloaded) != 2 {
		t.Fatalf("expected two secrets after rotation verify, got %d", len(reloaded))
	}
	if reloaded[0].Status != "rotating" {
		t.Fatalf("expected old secret to become rotating, got %+v", reloaded[0])
	}
	if reloaded[1].Status != "active" {
		t.Fatalf("expected new secret to become active, got %+v", reloaded[1])
	}

	var activeCount int64
	if err := db.Model(&model.SellerSecret{}).Where("supply_account_id = ? AND status = ?", supply.Id, "active").Count(&activeCount).Error; err != nil {
		t.Fatalf("failed to count active secrets: %v", err)
	}
	if activeCount != 1 {
		t.Fatalf("expected exactly one active secret, got %d", activeCount)
	}

	var rotatingCount int64
	if err := db.Model(&model.SellerSecret{}).Where("supply_account_id = ? AND status = ?", supply.Id, "rotating").Count(&rotatingCount).Error; err != nil {
		t.Fatalf("failed to count rotating secrets: %v", err)
	}
	if rotatingCount != 1 {
		t.Fatalf("expected exactly one rotating secret, got %d", rotatingCount)
	}
}

func TestDisableAndRecoverSecretRecomputeSupplyFromAllSecrets(t *testing.T) {
	db := setupMarketplaceServiceTestDB(t)
	user := seedMarketplaceServiceUser(t, db, "svc-secret-recompute")
	seller, supply := seedMarketplaceServiceSupply(t, db, user, "active", "success", "token")
	t.Setenv("SELLER_SECRET_MASTER_KEY", strings.Repeat("r", 32))

	activePayload := makeSellerSecretCiphertext(t, strings.Repeat("r", 32), "sk-active", "v1")
	_ = seedSellerSecretRecord(t, db, seller.Id, supply.Id, activePayload, "fp-active", "active", "success")
	draftPayload := makeSellerSecretCiphertext(t, strings.Repeat("r", 32), "sk-draft", "v1")
	draft := seedSellerSecretRecord(t, db, seller.Id, supply.Id, draftPayload, "fp-draft", "draft", "pending")

	if _, err := DisableSellerSecret(draft.Id, user.Id, "disable candidate"); err != nil {
		t.Fatalf("disable secret returned error: %v", err)
	}
	updatedSupply, err := model.GetSupplyAccountByID(supply.Id)
	if err != nil {
		t.Fatalf("failed to reload supply after disable: %v", err)
	}
	if updatedSupply.Status != "active" || updatedSupply.VerifyStatus != "success" {
		t.Fatalf("expected supply to stay active/success after disabling draft secret, got %+v", updatedSupply)
	}

	if _, err := RecoverSellerSecret(draft.Id, user.Id, "recover candidate"); err != nil {
		t.Fatalf("recover secret returned error: %v", err)
	}
	updatedSupply, err = model.GetSupplyAccountByID(supply.Id)
	if err != nil {
		t.Fatalf("failed to reload supply after recover: %v", err)
	}
	if updatedSupply.Status != "active" || updatedSupply.VerifyStatus != "success" {
		t.Fatalf("expected supply to stay active/success after recovering non-primary secret, got %+v", updatedSupply)
	}
}

func TestVerifySellerSecretSyncsRuntimeChannelMirror(t *testing.T) {
	db := setupMarketplaceServiceTestDB(t)
	user := seedMarketplaceServiceUser(t, db, "svc-secret-sync")
	seller, supply := seedMarketplaceServiceSupply(t, db, user, "paused", "pending", "token")
	channel, _ := seedMarketplaceChannelBinding(t, db, supply, "old-runtime-key")
	t.Setenv("SELLER_SECRET_MASTER_KEY", strings.Repeat("m", 32))

	secretPayload := makeSellerSecretCiphertext(t, strings.Repeat("m", 32), "sk-runtime-live", "v1")
	secret := seedSellerSecretRecord(t, db, seller.Id, supply.Id, secretPayload, "fp-sync", "draft", "pending")

	verified, err := VerifySellerSecret(secret.Id, user.Id)
	if err != nil {
		t.Fatalf("verify secret returned error: %v", err)
	}
	if verified.Status != "active" || verified.VerifyStatus != "success" {
		t.Fatalf("expected verified secret active/success, got %+v", verified)
	}

	reloadedChannel, err := model.GetChannelById(channel.Id, true)
	if err != nil {
		t.Fatalf("failed to reload channel: %v", err)
	}
	if reloadedChannel.Key != "sk-runtime-live" {
		t.Fatalf("expected channel key to sync runtime plaintext, got %q", reloadedChannel.Key)
	}
	otherInfo := reloadedChannel.GetOtherInfo()
	if otherInfo["managed_by"] != "seller_secret" {
		t.Fatalf("expected channel managed_by seller_secret, got %+v", otherInfo)
	}
}

func TestVerifySellerSecretUsesProviderProbeBeforeActivation(t *testing.T) {
	db := setupMarketplaceServiceTestDB(t)
	user := seedMarketplaceServiceUser(t, db, "svc-secret-provider-probe")
	seller, supply := seedMarketplaceServiceSupply(t, db, user, "paused", "pending", "token")
	_, _ = seedMarketplaceChannelBinding(t, db, supply, "old-runtime-key")
	t.Setenv("SELLER_SECRET_MASTER_KEY", strings.Repeat("p", 32))

	secretPayload := makeSellerSecretCiphertext(t, strings.Repeat("p", 32), "sk-provider-live", "v1")
	secret := seedSellerSecretRecord(t, db, seller.Id, supply.Id, secretPayload, "fp-probe", "draft", "pending")

	previousProbe := sellerSecretLiveProbeFunc
	probeCalls := 0
	sellerSecretLiveProbeFunc = func(secret *model.SellerSecret, runtimeKey string) error {
		probeCalls++
		if runtimeKey != "sk-provider-live" {
			t.Fatalf("expected decrypted runtime key to be probed, got %q", runtimeKey)
		}
		return fmt.Errorf("provider probe failed")
	}
	t.Cleanup(func() {
		sellerSecretLiveProbeFunc = previousProbe
	})

	if _, err := VerifySellerSecret(secret.Id, user.Id); err == nil {
		t.Fatalf("expected verify to fail when provider probe fails")
	}
	if probeCalls != 1 {
		t.Fatalf("expected provider probe to run exactly once, got %d", probeCalls)
	}

	updatedSecret, err := model.GetSellerSecretByID(secret.Id)
	if err != nil {
		t.Fatalf("failed to reload secret: %v", err)
	}
	if updatedSecret.Status == "active" || updatedSecret.VerifyStatus == "success" {
		t.Fatalf("expected provider probe failure to block activation, got %+v", updatedSecret)
	}
}
