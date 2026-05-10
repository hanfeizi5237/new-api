package service

import (
	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"gorm.io/gorm"
)

// lockForUpdate wraps a GORM query with FOR UPDATE, skipping it for SQLite
// which doesn't support row-level locking.
func lockForUpdate(query *gorm.DB) *gorm.DB {
	if common.UsingSQLite {
		// SQLite doesn't support FOR UPDATE; callers rely on atomic UPDATE
		// patterns there and row locks everywhere else.
		return query
	}
	return query.Set("gorm:query_option", "FOR UPDATE")
}

func normalizeRiskDiscountBps(snapshot *model.InventorySnapshot) int64 {
	if snapshot == nil || snapshot.RiskDiscountBps <= 0 {
		return 10000
	}
	return int64(snapshot.RiskDiscountBps)
}

func recomputeInventoryAvailableAmount(supply *model.SupplyAccount, snapshot *model.InventorySnapshot) int64 {
	if supply == nil || snapshot == nil {
		return 0
	}
	rawRemaining := supply.SellableCapacity - supply.ReservedCapacity - supply.UsedCapacity - snapshot.FrozenAmount
	if rawRemaining < 0 {
		rawRemaining = 0
	}
	return rawRemaining * normalizeRiskDiscountBps(snapshot) / 10000
}
