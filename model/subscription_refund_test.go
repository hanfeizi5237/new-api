package model

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRefundSubscriptionPreConsume_SQLiteIdempotent(t *testing.T) {
	truncateTables(t)
	require.NoError(t, DB.AutoMigrate(&SubscriptionPreConsumeRecord{}))
	t.Cleanup(func() {
		DB.Exec("DELETE FROM subscription_pre_consume_records")
	})

	sub := &UserSubscription{
		Id:          1001,
		UserId:      2001,
		AmountTotal: 1000,
		AmountUsed:  300,
		Status:      "active",
		StartTime:   time.Now().Unix(),
		EndTime:     time.Now().Add(24 * time.Hour).Unix(),
	}
	require.NoError(t, DB.Create(sub).Error)

	record := &SubscriptionPreConsumeRecord{
		RequestId:          "req-refund-sqlite",
		UserId:             sub.UserId,
		UserSubscriptionId: sub.Id,
		PreConsumed:        300,
		Status:             "consumed",
	}
	require.NoError(t, DB.Create(record).Error)

	require.NoError(t, RefundSubscriptionPreConsume(record.RequestId))

	var refreshedSub UserSubscription
	require.NoError(t, DB.First(&refreshedSub, sub.Id).Error)
	require.Equal(t, int64(0), refreshedSub.AmountUsed)

	var refreshedRecord SubscriptionPreConsumeRecord
	require.NoError(t, DB.Where("request_id = ?", record.RequestId).First(&refreshedRecord).Error)
	require.Equal(t, "refunded", refreshedRecord.Status)

	require.NoError(t, RefundSubscriptionPreConsume(record.RequestId))

	require.NoError(t, DB.First(&refreshedSub, sub.Id).Error)
	require.Equal(t, int64(0), refreshedSub.AmountUsed)
}
