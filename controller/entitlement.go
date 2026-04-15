package controller

import (
	"strconv"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/service"
	"github.com/gin-gonic/gin"
)

func GetMarketEntitlements(c *gin.Context) {
	pageInfo := common.GetPageQuery(c)
	entitlements, total, err := service.ListBuyerEntitlements(
		c.GetInt("id"),
		c.Query("model_name"),
		pageInfo.GetStartIdx(),
		pageInfo.GetPageSize(),
	)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	pageInfo.SetTotal(int(total))
	pageInfo.SetItems(entitlements)
	common.ApiSuccess(c, pageInfo)
}

func GetMarketEntitlementsAdmin(c *gin.Context) {
	pageInfo := common.GetPageQuery(c)
	buyerUserId, _ := strconv.Atoi(c.Query("buyer_user_id"))
	entitlements, total, err := service.ListEntitlementsAdmin(
		buyerUserId,
		c.Query("model_name"),
		c.Query("status"),
		pageInfo.GetStartIdx(),
		pageInfo.GetPageSize(),
	)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	pageInfo.SetTotal(int(total))
	pageInfo.SetItems(entitlements)
	common.ApiSuccess(c, pageInfo)
}
