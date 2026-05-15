package controller

import (
	"fmt"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

// SubscriptionQuotaPayRequest 钱包余额支付订阅套餐请求
type SubscriptionQuotaPayRequest struct {
	PlanId int `json:"plan_id"`
}

// SubscriptionQuotaPayResponse 钱包余额支付返回结果
type SubscriptionQuotaPayResponse struct {
	TradeNo   string `json:"trade_no"`
	QuotaCost int    `json:"quota_cost"`
}

// SubscriptionRequestQuotaPay 处理用户使用钱包余额支付订阅套餐的请求
//
// 流程：
//  1. 校验开关、套餐、购买上限
//  2. 计算应扣 quota（PriceAmount × QuotaPerUnit）
//  3. 调用 model.PaySubscriptionByQuota 在事务内一体化完成扣费 + 订单 + 订阅
func SubscriptionRequestQuotaPay(c *gin.Context) {
	if !setting.EnableQuotaPayForSubscription {
		common.ApiErrorMsg(c, "管理员未开启钱包余额支付订阅功能")
		return
	}

	var req SubscriptionQuotaPayRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.PlanId <= 0 {
		common.ApiErrorMsg(c, "参数错误")
		return
	}

	plan, err := model.GetSubscriptionPlanById(req.PlanId)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	if !plan.Enabled {
		common.ApiErrorMsg(c, "套餐未启用")
		return
	}
	if plan.PriceAmount < 0.01 {
		common.ApiErrorMsg(c, "套餐金额过低")
		return
	}

	userId := c.GetInt("id")
	if userId <= 0 {
		common.ApiErrorMsg(c, "用户身份无效")
		return
	}

	// 购买上限预检查（事务内 CreateUserSubscriptionFromPlanTx 还会再校验一次，作为最终一致性保证）
	if plan.MaxPurchasePerUser > 0 {
		count, err := model.CountUserSubscriptionsByPlan(userId, plan.Id)
		if err != nil {
			common.ApiError(c, err)
			return
		}
		if count >= int64(plan.MaxPurchasePerUser) {
			common.ApiErrorMsg(c, "已达到该套餐购买上限")
			return
		}
	}

	quotaCost := calcSubscriptionQuotaCost(plan.PriceAmount)
	if quotaCost <= 0 {
		common.ApiErrorMsg(c, "套餐扣费额度计算异常")
		return
	}

	// 提前给一个友好提示（事务内还会再校验一次防止并发）
	user, err := model.GetUserById(userId, true)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	if user == nil {
		common.ApiErrorMsg(c, "用户不存在")
		return
	}
	if user.Quota < quotaCost {
		common.ApiErrorMsg(c, "钱包余额不足")
		return
	}

	// 串行化同一用户同一套餐的余额支付，避免并发重复扣
	lockKey := fmt.Sprintf("sub-quota-pay-u%d-p%d", userId, plan.Id)
	LockOrder(lockKey)
	defer UnlockOrder(lockKey)

	order, err := model.PaySubscriptionByQuota(userId, plan, quotaCost)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"message": "success",
		"data": SubscriptionQuotaPayResponse{
			TradeNo:   order.TradeNo,
			QuotaCost: quotaCost,
		},
	})
}

func calcSubscriptionQuotaCost(priceAmount float64) int {
	// 余额扣费必须向上取整，避免小数换算被截断导致少扣 1 quota。
	return int(decimal.NewFromFloat(priceAmount).
		Mul(decimal.NewFromFloat(common.QuotaPerUnit)).
		Ceil().
		IntPart())
}
