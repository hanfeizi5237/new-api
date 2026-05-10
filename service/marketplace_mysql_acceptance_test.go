package service

import (
	"errors"
	"fmt"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/dto"
	"github.com/QuantumNous/new-api/model"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	relayconstant "github.com/QuantumNous/new-api/relay/constant"
	"github.com/gin-gonic/gin"
	mysqldrv "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	marketplaceMySQLAcceptanceOnce sync.Once
	marketplaceMySQLAcceptanceDB   *gorm.DB
	marketplaceMySQLAcceptanceErr  error
)

func setupMarketplaceMySQLAcceptanceDB(t *testing.T) *gorm.DB {
	t.Helper()

	host := strings.TrimSpace(os.Getenv("TEST_MYSQL_HOST"))
	if host == "" {
		t.Skip("set TEST_MYSQL_HOST/PORT/USER/PASSWORD to run MySQL marketplace acceptance tests")
	}
	port := strings.TrimSpace(os.Getenv("TEST_MYSQL_PORT"))
	if port == "" {
		port = "3306"
	}
	user := strings.TrimSpace(os.Getenv("TEST_MYSQL_USER"))
	password := os.Getenv("TEST_MYSQL_PASSWORD")
	if user == "" {
		t.Skip("set TEST_MYSQL_USER to run MySQL marketplace acceptance tests")
	}
	existingDBName := strings.TrimSpace(os.Getenv("TEST_MYSQL_DB"))
	if existingDBName != "" {
		db := setupSharedMarketplaceMySQLAcceptanceDB(t, host, port, user, password, existingDBName)
		previousDB := model.DB
		previousLogDB := model.LOG_DB
		previousUsingSQLite := common.UsingSQLite
		previousUsingMySQL := common.UsingMySQL
		previousUsingPostgreSQL := common.UsingPostgreSQL
		previousRedisEnabled := common.RedisEnabled
		previousMemoryCacheEnabled := common.MemoryCacheEnabled

		common.UsingSQLite = false
		common.UsingMySQL = true
		common.UsingPostgreSQL = false
		common.RedisEnabled = false
		common.MemoryCacheEnabled = false
		model.DB = db
		model.LOG_DB = db

		t.Cleanup(func() {
			model.DB = previousDB
			model.LOG_DB = previousLogDB
			common.UsingSQLite = previousUsingSQLite
			common.UsingMySQL = previousUsingMySQL
			common.UsingPostgreSQL = previousUsingPostgreSQL
			common.RedisEnabled = previousRedisEnabled
			common.MemoryCacheEnabled = previousMemoryCacheEnabled
		})

		return db
	}

	rootConfig := mysqldrv.NewConfig()
	rootConfig.User = user
	rootConfig.Passwd = password
	rootConfig.Net = "tcp"
	rootConfig.Addr = host + ":" + port
	rootConfig.Params = map[string]string{"charset": "utf8mb4"}
	rootConfig.ParseTime = true
	rootConfig.Loc = time.Local

	schemaName := existingDBName
	var rootDB *gorm.DB
	var err error
	if schemaName == "" {
		rootDB, err = gorm.Open(mysql.Open(rootConfig.FormatDSN()), &gorm.Config{})
		if err != nil {
			t.Fatalf("failed to open mysql server connection: %v", err)
		}

		schemaName = fmt.Sprintf("codex_marketplace_acceptance_%d", time.Now().UnixNano())
		if err := rootDB.Exec("CREATE DATABASE `" + schemaName + "` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci").Error; err != nil {
			t.Fatalf("failed to create acceptance schema %s: %v", schemaName, err)
		}
	}

	schemaConfig := mysqldrv.NewConfig()
	schemaConfig.User = user
	schemaConfig.Passwd = password
	schemaConfig.Net = "tcp"
	schemaConfig.Addr = host + ":" + port
	schemaConfig.DBName = schemaName
	schemaConfig.Params = map[string]string{"charset": "utf8mb4"}
	schemaConfig.ParseTime = true
	schemaConfig.Loc = time.Local

	previousDB := model.DB
	previousLogDB := model.LOG_DB
	previousUsingSQLite := common.UsingSQLite
	previousUsingMySQL := common.UsingMySQL
	previousUsingPostgreSQL := common.UsingPostgreSQL
	previousRedisEnabled := common.RedisEnabled
	previousMemoryCacheEnabled := common.MemoryCacheEnabled

	common.UsingSQLite = false
	common.UsingMySQL = true
	common.UsingPostgreSQL = false
	common.RedisEnabled = false
	common.MemoryCacheEnabled = false

	db, err := gorm.Open(mysql.Open(schemaConfig.FormatDSN()), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open mysql acceptance schema: %v", err)
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
		t.Fatalf("failed to migrate mysql acceptance schema: %v", err)
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
		if rootDB != nil {
			_ = rootDB.Exec("DROP DATABASE IF EXISTS `" + schemaName + "`").Error
			rootSQLDB, err := rootDB.DB()
			if err == nil {
				_ = rootSQLDB.Close()
			}
		}
	})

	return db
}

func setupSharedMarketplaceMySQLAcceptanceDB(t *testing.T, host string, port string, user string, password string, schemaName string) *gorm.DB {
	t.Helper()

	marketplaceMySQLAcceptanceOnce.Do(func() {
		schemaConfig := mysqldrv.NewConfig()
		schemaConfig.User = user
		schemaConfig.Passwd = password
		schemaConfig.Net = "tcp"
		schemaConfig.Addr = host + ":" + port
		schemaConfig.DBName = schemaName
		schemaConfig.Params = map[string]string{"charset": "utf8mb4"}
		schemaConfig.ParseTime = true
		schemaConfig.Loc = time.Local

		db, err := gorm.Open(mysql.Open(schemaConfig.FormatDSN()), &gorm.Config{})
		if err != nil {
			marketplaceMySQLAcceptanceErr = fmt.Errorf("failed to open mysql acceptance schema: %w", err)
			return
		}

		models := []any{
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
		}
		missingModels := make([]any, 0, len(models))
		for _, candidate := range models {
			if !db.Migrator().HasTable(candidate) {
				missingModels = append(missingModels, candidate)
			}
		}
		if len(missingModels) > 0 {
			if err := db.AutoMigrate(missingModels...); err != nil {
				marketplaceMySQLAcceptanceErr = fmt.Errorf("failed to migrate mysql acceptance schema: %w", err)
				return
			}
		}

		marketplaceMySQLAcceptanceDB = db
	})

	if marketplaceMySQLAcceptanceErr != nil {
		t.Fatal(marketplaceMySQLAcceptanceErr)
	}
	if marketplaceMySQLAcceptanceDB == nil {
		t.Fatal("mysql acceptance shared db was not initialized")
	}
	return marketplaceMySQLAcceptanceDB
}

func mysqlAcceptanceSuffix() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func TestMarketplaceAcceptanceMySQLPaymentFailureAndRetryDoesNotDoubleCount(t *testing.T) {
	db := setupMarketplaceMySQLAcceptanceDB(t)
	suffix := mysqlAcceptanceSuffix()
	buyer := seedMarketplaceServiceUser(t, db, "mysql-acceptance-buyer-retry-"+suffix)
	sellerUser := seedMarketplaceServiceUser(t, db, "mysql-acceptance-seller-retry-"+suffix)
	seller, supply := seedMarketplaceServiceSupply(t, db, sellerUser, "active", "success", "token")
	_, _ = seedMarketplaceChannelBinding(t, db, supply, "runtime-key")
	_ = seedSellerSecretRecord(t, db, seller.Id, supply.Id, `{"alg":"aes-256-gcm","kid":"v1","nonce":"bm9uY2U=","ciphertext":"Y2lwaGVy"}`, "fp-mysql-retry-"+suffix, "active", "success")
	listing, sku := seedMarketplaceListing(t, db, seller, supply, 2000, 499)

	order, _, err := CreateMarketOrder(CreateMarketOrderInput{
		BuyerUserID:    buyer.Id,
		ListingID:      listing.Id,
		SkuID:          sku.Id,
		Quantity:       1,
		IdempotencyKey: "mysql-acceptance-order-retry-" + suffix,
	})
	if err != nil {
		t.Fatalf("create market order returned error: %v", err)
	}

	previousGrantFunc := grantEntitlementsForOrderTxFunc
	grantEntitlementsForOrderTxFunc = func(tx *gorm.DB, order *model.MarketOrder, items []model.MarketOrderItem) error {
		return errors.New("synthetic entitlement grant failure")
	}
	t.Cleanup(func() {
		grantEntitlementsForOrderTxFunc = previousGrantFunc
	})

	paidOrder, err := CompleteMarketOrderPayment(CompleteMarketOrderPaymentInput{
		OrderNo:            order.OrderNo,
		PaymentMethod:      model.PaymentMethodStripe,
		PaymentTradeNo:     "pi_mysql_acceptance_retry_" + suffix,
		Currency:           order.Currency,
		PayableAmountMinor: order.PayableAmountMinor,
		ProviderPayload:    `{"type":"checkout.session.completed"}`,
	})
	if err == nil {
		t.Fatalf("expected synthetic entitlement grant failure to bubble up")
	}
	if paidOrder == nil || paidOrder.PaymentStatus != marketPaymentStatusPaid || paidOrder.EntitlementStatus != marketEntitlementStatusFailed {
		t.Fatalf("expected paid + entitlement_failed state after failed grant, got %+v", paidOrder)
	}

	reloadedOrder, err := model.GetMarketOrderByID(order.Id)
	if err != nil {
		t.Fatalf("failed to reload paid order after grant failure: %v", err)
	}
	if reloadedOrder.PaymentStatus != marketPaymentStatusPaid || reloadedOrder.EntitlementStatus != marketEntitlementStatusFailed {
		t.Fatalf("expected persisted paid + entitlement_failed state, got %+v", reloadedOrder)
	}

	grantEntitlementsForOrderTxFunc = previousGrantFunc

	retriedOrder, err := CompleteMarketOrderPayment(CompleteMarketOrderPaymentInput{
		OrderNo:            order.OrderNo,
		PaymentMethod:      model.PaymentMethodStripe,
		PaymentTradeNo:     "pi_mysql_acceptance_retry_" + suffix,
		Currency:           order.Currency,
		PayableAmountMinor: order.PayableAmountMinor,
		ProviderPayload:    `{"type":"checkout.session.completed"}`,
	})
	if err != nil {
		t.Fatalf("expected entitlement retry success on mysql, got error: %v", err)
	}
	if retriedOrder.EntitlementStatus != marketEntitlementStatusCreated {
		t.Fatalf("expected entitlement retry to finish successfully, got %+v", retriedOrder)
	}

	entitlements, total, err := ListBuyerEntitlements(buyer.Id, "", 0, 20)
	if err != nil {
		t.Fatalf("list buyer entitlements returned error: %v", err)
	}
	if total != 1 || len(entitlements) != 1 || entitlements[0].TotalGranted != 2000 {
		t.Fatalf("expected exactly one entitlement after retry, got total=%d entitlements=%+v", total, entitlements)
	}

	var lotCount int64
	if err := db.Model(&model.EntitlementLot{}).Where("order_id = ?", order.Id).Count(&lotCount).Error; err != nil {
		t.Fatalf("failed to count entitlement lots after retry: %v", err)
	}
	if lotCount != 1 {
		t.Fatalf("expected exactly one entitlement lot after retry, got %d", lotCount)
	}

	snapshot, err := model.GetInventorySnapshotBySupplyAccountID(supply.Id)
	if err != nil {
		t.Fatalf("failed to reload inventory snapshot after mysql retry: %v", err)
	}
	if snapshot.SoldAmount != 2000 || snapshot.FrozenAmount != 0 || snapshot.AvailableAmount != supply.SellableCapacity-2000 {
		t.Fatalf("expected inventory counters not to double count after retry, got %+v", snapshot)
	}

	reloadedSupply, err := model.GetSupplyAccountByID(supply.Id)
	if err != nil {
		t.Fatalf("failed to reload supply account after mysql retry: %v", err)
	}
	if reloadedSupply.ReservedCapacity != 2000 {
		t.Fatalf("expected reserved capacity to remain 2000 after retry, got %+v", reloadedSupply)
	}
}

func TestMarketplaceAcceptanceMySQLSettleWritesUsageLedgerAndOrderItem(t *testing.T) {
	db := setupMarketplaceMySQLAcceptanceDB(t)
	suffix := mysqlAcceptanceSuffix()
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)

	fixture := seedMarketplaceEntitlementBillingFixture(t, db, "mysql-acceptance-buyer-settle-"+suffix, 2000)
	relayInfo := &relaycommon.RelayInfo{
		UserId:          fixture.buyer.Id,
		OriginModelName: fixture.supply.ModelName,
		RequestId:       "req-mysql-acceptance-settle-" + suffix,
		TokenId:         3001,
		TokenKey:        "mysql-acceptance-token-" + suffix,
		RelayMode:       relayconstant.RelayModeChatCompletions,
		ChannelMeta: &relaycommon.ChannelMeta{
			ChannelId: fixture.channel.Id,
		},
		UserSetting: dto.UserSetting{
			BillingPreference: "wallet_first",
		},
	}

	session, apiErr := NewBillingSession(ctx, relayInfo, 300)
	if apiErr != nil {
		t.Fatalf("expected mysql marketplace billing session, got error: %v", apiErr)
	}
	marketplaceFunding, ok := session.funding.(*MarketplaceEntitlementFunding)
	if !ok {
		t.Fatalf("expected marketplace entitlement funding, got %T", session.funding)
	}
	marketplaceFunding.captureUsageSummary(textQuotaSummary{
		PromptTokens:     240,
		CompletionTokens: 60,
		TotalTokens:      300,
	})

	if err := session.Settle(300); err != nil {
		t.Fatalf("expected mysql settle success, got error: %v", err)
	}

	orderItem, err := getMarketOrderItemByID(fixture.lot.OrderItemId)
	if err != nil {
		t.Fatalf("failed to reload market order item after mysql settle: %v", err)
	}
	if orderItem.UsedAmount != 300 {
		t.Fatalf("expected mysql order item used amount 300, got %+v", orderItem)
	}

	lot, err := getEntitlementLotByID(fixture.lot.Id)
	if err != nil {
		t.Fatalf("failed to reload entitlement lot after mysql settle: %v", err)
	}
	if lot.UsedAmount != 300 {
		t.Fatalf("expected mysql entitlement lot used amount 300, got %+v", lot)
	}

	var ledger model.UsageLedger
	if err := db.Where("request_id = ? AND ledger_status = ?", relayInfo.RequestId, "success").First(&ledger).Error; err != nil {
		t.Fatalf("expected mysql success usage ledger, got error: %v", err)
	}
	if ledger.PromptTokens != 240 || ledger.CompletionTokens != 60 || ledger.TotalTokens != 300 {
		t.Fatalf("expected mysql usage ledger token details, got %+v", ledger)
	}
	if ledger.BillingSource != BillingSourceMarketplaceEntitlement || ledger.EntitlementLotId != fixture.lot.Id || ledger.OrderItemId != fixture.lot.OrderItemId {
		t.Fatalf("unexpected mysql usage ledger metadata: %+v", ledger)
	}

	retryFunding := &MarketplaceEntitlementFunding{
		requestId:        relayInfo.RequestId,
		userId:           fixture.buyer.Id,
		tokenId:          relayInfo.TokenId,
		modelName:        fixture.supply.ModelName,
		relayMode:        relayInfo.RelayMode,
		relayInfo:        relayInfo,
		preConsumed:      300,
		entitlementId:    fixture.entitlement.Id,
		entitlementLotId: fixture.lot.Id,
		sellerId:         fixture.seller.Id,
		supplyAccountId:  fixture.supply.Id,
		listingId:        fixture.listing.Id,
		orderId:          fixture.order.Id,
		orderItemId:      fixture.lot.OrderItemId,
		channelIDs:       []int{fixture.channel.Id},
		promptTokens:     240,
		completionTokens: 60,
		totalTokens:      300,
	}
	if err := retryFunding.Settle(0); err != nil {
		t.Fatalf("expected mysql retry settle to be idempotent, got error: %v", err)
	}

	orderItem, err = getMarketOrderItemByID(fixture.lot.OrderItemId)
	if err != nil {
		t.Fatalf("failed to reload order item after mysql retry settle: %v", err)
	}
	if orderItem.UsedAmount != 300 {
		t.Fatalf("expected mysql retry settle not to double count order item usage, got %+v", orderItem)
	}

	var successLedgerCount int64
	if err := db.Model(&model.UsageLedger{}).Where("event_key = ?", relayInfo.RequestId+":success").Count(&successLedgerCount).Error; err != nil {
		t.Fatalf("failed to count mysql success ledgers: %v", err)
	}
	if successLedgerCount != 1 {
		t.Fatalf("expected exactly one mysql success usage ledger, got %d", successLedgerCount)
	}
}

func TestMarketplaceAcceptanceMySQLLockForUpdateUsesForUpdate(t *testing.T) {
	db := setupMarketplaceMySQLAcceptanceDB(t)
	query := lockForUpdate(db.Model(&model.SupplyAccount{}))
	value, exists := query.Get("gorm:query_option")
	if !exists {
		t.Fatalf("expected mysql query to set FOR UPDATE")
	}
	if got := strings.TrimSpace(fmt.Sprint(value)); got != "FOR UPDATE" {
		t.Fatalf("expected mysql query option FOR UPDATE, got %q", got)
	}
}
