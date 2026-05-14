package controller

import (
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/stretchr/testify/assert"
)

func TestCalcSubscriptionQuotaCost_CeilsFractionalQuota(t *testing.T) {
	originalQuotaPerUnit := common.QuotaPerUnit
	common.QuotaPerUnit = 3
	t.Cleanup(func() {
		common.QuotaPerUnit = originalQuotaPerUnit
	})

	assert.Equal(t, 3, calcSubscriptionQuotaCost(1))
	assert.Equal(t, 4, calcSubscriptionQuotaCost(1.01))
}
