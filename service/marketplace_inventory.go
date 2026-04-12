package service

import "github.com/QuantumNous/new-api/model"

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
