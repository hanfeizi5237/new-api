package model

import (
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestMarketplaceMigration(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)

	require.NoError(t, db.AutoMigrate(
		&SellerProfile{},
		&SupplyAccount{},
		&SellerSecret{},
		&SellerSecretAudit{},
		&SupplyChannelBinding{},
		&Listing{},
		&ListingSKU{},
		&InventorySnapshot{},
		&MarketOrder{},
		&MarketOrderItem{},
		&BuyerEntitlement{},
		&EntitlementLot{},
		&UsageLedger{},
	))

	assert.True(t, db.Migrator().HasTable(&SellerProfile{}))
	assert.True(t, db.Migrator().HasTable(&SupplyAccount{}))
	assert.True(t, db.Migrator().HasTable(&SellerSecret{}))
	assert.True(t, db.Migrator().HasTable(&Listing{}))
	assert.True(t, db.Migrator().HasTable(&MarketOrder{}))
	assert.True(t, db.Migrator().HasTable(&BuyerEntitlement{}))
	assert.True(t, db.Migrator().HasTable(&UsageLedger{}))

	assert.True(t, db.Migrator().HasIndex(&SellerProfile{}, "ux_seller_profiles_user_id"))
	assert.True(t, db.Migrator().HasIndex(&SellerProfile{}, "ux_seller_profiles_seller_code"))
	assert.True(t, db.Migrator().HasIndex(&SupplyAccount{}, "ux_supply_accounts_supply_code"))
	assert.True(t, db.Migrator().HasIndex(&SellerSecret{}, "ux_seller_secrets_supply_fingerprint"))
	assert.True(t, db.Migrator().HasIndex(&SupplyChannelBinding{}, "ux_supply_channel_bindings_supply_channel"))
	assert.True(t, db.Migrator().HasIndex(&Listing{}, "ux_listings_listing_code"))
	assert.True(t, db.Migrator().HasIndex(&ListingSKU{}, "ux_listing_skus_sku_code"))
	assert.True(t, db.Migrator().HasIndex(&InventorySnapshot{}, "ux_inventory_snapshots_supply_account_id"))
	assert.True(t, db.Migrator().HasIndex(&MarketOrder{}, "ux_market_orders_order_no"))
	assert.True(t, db.Migrator().HasIndex(&MarketOrder{}, "ux_market_orders_idempotency_key"))
	assert.True(t, db.Migrator().HasIndex(&BuyerEntitlement{}, "ux_buyer_entitlements_owner_model"))
	assert.True(t, db.Migrator().HasIndex(&EntitlementLot{}, "ux_entitlement_lots_source_event_key"))
	assert.True(t, db.Migrator().HasIndex(&UsageLedger{}, "ux_usage_ledgers_event_key"))
}
