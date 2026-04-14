package service

import (
	"testing"

	"github.com/QuantumNous/new-api/model"
)

func TestSyncMarketplaceInventoryMarksListingSoldOutWhenNoSKUCanBeSatisfied(t *testing.T) {
	db := setupMarketplaceServiceTestDB(t)
	user := seedMarketplaceServiceUser(t, db, "svc-inventory-soldout")
	seller, supply := seedMarketplaceServiceSupply(t, db, user, "active", "success", "token")
	_, _ = seedMarketplaceChannelBinding(t, db, supply, "runtime-key")
	_ = seedSellerSecretRecord(t, db, seller.Id, supply.Id, `{"alg":"aes-256-gcm","kid":"v1","nonce":"bm9uY2U=","ciphertext":"Y2lwaGVy"}`, "fp-inventory-soldout", "active", "success")
	listing, sku := seedMarketplaceListing(t, db, seller, supply, 1000, 199)

	if err := db.Model(&model.SupplyAccount{}).
		Where("id = ?", supply.Id).
		Updates(map[string]interface{}{
			"sellable_capacity": 900,
			"reserved_capacity": 0,
			"used_capacity":     0,
		}).Error; err != nil {
		t.Fatalf("failed to reduce sellable capacity: %v", err)
	}
	if err := db.Model(&model.InventorySnapshot{}).
		Where("supply_account_id = ?", supply.Id).
		Updates(map[string]interface{}{
			"available_amount":  999999,
			"frozen_amount":     0,
			"sold_amount":       0,
			"consumed_amount":   0,
			"risk_discount_bps": 10000,
			"health_score":      100,
			"sync_status":       "ok",
		}).Error; err != nil {
		t.Fatalf("failed to seed stale inventory snapshot: %v", err)
	}

	if err := SyncMarketplaceInventoryBySupplyAccountID(supply.Id, "test_sold_out"); err != nil {
		t.Fatalf("sync marketplace inventory returned error: %v", err)
	}

	snapshot, err := model.GetInventorySnapshotBySupplyAccountID(supply.Id)
	if err != nil {
		t.Fatalf("failed to reload inventory snapshot: %v", err)
	}
	if snapshot.AvailableAmount != 900 {
		t.Fatalf("expected available amount recalculated to 900, got %+v", snapshot)
	}

	reloadedListing, err := model.GetListingByID(listing.Id)
	if err != nil {
		t.Fatalf("failed to reload listing: %v", err)
	}
	if reloadedListing.Status != "sold_out" {
		t.Fatalf("expected listing to become sold_out, got %+v", reloadedListing)
	}

	reloadedSKU, err := model.GetListingSKUByID(sku.Id)
	if err != nil {
		t.Fatalf("failed to reload sku: %v", err)
	}
	if reloadedSKU.PackageAmount*int64(reloadedSKU.MinQuantity) <= snapshot.AvailableAmount {
		t.Fatalf("expected sku minimum package requirement to exceed available inventory")
	}
}

func TestSyncMarketplaceInventoryPausesListingWhenSupplyIsUnhealthy(t *testing.T) {
	db := setupMarketplaceServiceTestDB(t)
	user := seedMarketplaceServiceUser(t, db, "svc-inventory-paused")
	seller, supply := seedMarketplaceServiceSupply(t, db, user, "active", "success", "token")
	_, _ = seedMarketplaceChannelBinding(t, db, supply, "runtime-key")
	_ = seedSellerSecretRecord(t, db, seller.Id, supply.Id, `{"alg":"aes-256-gcm","kid":"v1","nonce":"bm9uY2U=","ciphertext":"Y2lwaGVy"}`, "fp-inventory-paused", "active", "success")
	listing, _ := seedMarketplaceListing(t, db, seller, supply, 1000, 199)

	if err := db.Model(&model.SupplyAccount{}).
		Where("id = ?", supply.Id).
		Updates(map[string]interface{}{
			"status":        "paused",
			"verify_status": "failed",
		}).Error; err != nil {
		t.Fatalf("failed to pause supply account: %v", err)
	}

	if err := SyncMarketplaceInventoryBySupplyAccountID(supply.Id, "test_pause"); err != nil {
		t.Fatalf("sync marketplace inventory returned error: %v", err)
	}

	snapshot, err := model.GetInventorySnapshotBySupplyAccountID(supply.Id)
	if err != nil {
		t.Fatalf("failed to reload inventory snapshot: %v", err)
	}
	if snapshot.SyncStatus != "error" {
		t.Fatalf("expected unhealthy supply to mark snapshot sync_status=error, got %+v", snapshot)
	}

	reloadedListing, err := model.GetListingByID(listing.Id)
	if err != nil {
		t.Fatalf("failed to reload listing: %v", err)
	}
	if reloadedListing.Status != "paused" {
		t.Fatalf("expected unhealthy supply to pause listing, got %+v", reloadedListing)
	}
}

func TestSyncMarketplaceInventoryDoesNotAutoReactivateSoldOutListing(t *testing.T) {
	db := setupMarketplaceServiceTestDB(t)
	user := seedMarketplaceServiceUser(t, db, "svc-inventory-manual-recover")
	seller, supply := seedMarketplaceServiceSupply(t, db, user, "active", "success", "token")
	_, _ = seedMarketplaceChannelBinding(t, db, supply, "runtime-key")
	_ = seedSellerSecretRecord(t, db, seller.Id, supply.Id, `{"alg":"aes-256-gcm","kid":"v1","nonce":"bm9uY2U=","ciphertext":"Y2lwaGVy"}`, "fp-inventory-manual-recover", "active", "success")
	listing, _ := seedMarketplaceListing(t, db, seller, supply, 1000, 199)

	if err := db.Model(&model.Listing{}).Where("id = ?", listing.Id).Update("status", "sold_out").Error; err != nil {
		t.Fatalf("failed to mark listing sold_out: %v", err)
	}
	if err := db.Model(&model.SupplyAccount{}).
		Where("id = ?", supply.Id).
		Updates(map[string]interface{}{
			"sellable_capacity": 5000,
			"reserved_capacity": 0,
			"used_capacity":     0,
		}).Error; err != nil {
		t.Fatalf("failed to expand sellable capacity: %v", err)
	}

	if err := SyncMarketplaceInventoryBySupplyAccountID(supply.Id, "test_manual_recover"); err != nil {
		t.Fatalf("sync marketplace inventory returned error: %v", err)
	}

	reloadedListing, err := model.GetListingByID(listing.Id)
	if err != nil {
		t.Fatalf("failed to reload listing: %v", err)
	}
	if reloadedListing.Status != "sold_out" {
		t.Fatalf("expected sold_out listing to stay sold_out until manual recovery, got %+v", reloadedListing)
	}
}
