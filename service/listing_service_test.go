package service

import (
	"testing"

	"github.com/QuantumNous/new-api/model"
)

func TestCreateListingWithSKUsRejectsUnreadySupply(t *testing.T) {
	db := setupMarketplaceServiceTestDB(t)
	user := seedMarketplaceServiceUser(t, db, "svc-listing-unready")
	seller, supply := seedMarketplaceServiceSupply(t, db, user, "paused", "pending", "token")

	_, _, err := CreateListingWithSKUs(CreateListingInput{
		Listing: model.Listing{
			SellerId:        seller.Id,
			SupplyAccountId: supply.Id,
			ListingCode:     "listing-unready-001",
			Title:           "Unready Listing",
		},
		SKUs: []model.ListingSKU{
			{
				SkuCode:        "sku-unready-001",
				PackageAmount:  1000,
				PackageUnit:    "token",
				UnitPriceMinor: 199,
				MinQuantity:    1,
				MaxQuantity:    2,
			},
		},
	})
	if err == nil {
		t.Fatalf("expected listing creation to reject unready supply")
	}
}

func TestCreateListingWithSKUsRejectsNonTokenPackageUnit(t *testing.T) {
	db := setupMarketplaceServiceTestDB(t)
	user := seedMarketplaceServiceUser(t, db, "svc-listing-unit")
	seller, supply := seedMarketplaceServiceSupply(t, db, user, "active", "success", "request")
	_, _ = seedMarketplaceChannelBinding(t, db, supply, "preloaded-runtime-key")
	_ = seedSellerSecretRecord(t, db, seller.Id, supply.Id, `{"alg":"aes-256-gcm","kid":"v1","nonce":"bm9uY2U=","ciphertext":"Y2lwaGVy"}`, "fp-token-unit", "active", "success")

	_, _, err := CreateListingWithSKUs(CreateListingInput{
		Listing: model.Listing{
			SellerId:        seller.Id,
			SupplyAccountId: supply.Id,
			ListingCode:     "listing-unit-001",
			Title:           "Wrong Unit Listing",
		},
		SKUs: []model.ListingSKU{
			{
				SkuCode:        "sku-unit-001",
				PackageAmount:  1000,
				PackageUnit:    "request",
				UnitPriceMinor: 199,
				MinQuantity:    1,
				MaxQuantity:    2,
			},
		},
	})
	if err == nil {
		t.Fatalf("expected listing creation to reject non-token unit supply")
	}
}

func TestCreateListingWithSKUsRejectsSupplyMismatchMetadata(t *testing.T) {
	db := setupMarketplaceServiceTestDB(t)
	user := seedMarketplaceServiceUser(t, db, "svc-listing-mismatch")
	seller, supply := seedMarketplaceServiceSupply(t, db, user, "active", "success", "token")
	_, _ = seedMarketplaceChannelBinding(t, db, supply, "preloaded-runtime-key")
	_ = seedSellerSecretRecord(t, db, seller.Id, supply.Id, `{"alg":"aes-256-gcm","kid":"v1","nonce":"bm9uY2U=","ciphertext":"Y2lwaGVy"}`, "fp-token-match", "active", "success")

	_, _, err := CreateListingWithSKUs(CreateListingInput{
		Listing: model.Listing{
			SellerId:        seller.Id,
			SupplyAccountId: supply.Id,
			ListingCode:     "listing-mismatch-001",
			Title:           "Mismatch Listing",
			VendorId:        99,
			ModelName:       "other-model",
		},
		SKUs: []model.ListingSKU{
			{
				SkuCode:        "sku-mismatch-001",
				PackageAmount:  1000,
				PackageUnit:    "token",
				UnitPriceMinor: 199,
				MinQuantity:    1,
				MaxQuantity:    2,
			},
		},
	})
	if err == nil {
		t.Fatalf("expected listing creation to reject mismatched vendor/model")
	}
}
