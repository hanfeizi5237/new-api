package controller

import (
	"strconv"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"
	"github.com/gin-gonic/gin"
)

type CreateSellerAdminRequest struct {
	Seller        model.SellerProfile          `json:"seller"`
	SupplyAccount model.SupplyAccount          `json:"supply_account"`
	Bindings      []model.SupplyChannelBinding `json:"bindings"`
}

type UpdateSellerStatusRequest struct {
	Status string `json:"status"`
	Remark string `json:"remark"`
}

func GetSellerAdmin(c *gin.Context) {
	pageInfo := common.GetPageQuery(c)
	sellers, total, err := service.ListSellers(
		c.Query("keyword"),
		c.Query("status"),
		pageInfo.GetStartIdx(),
		pageInfo.GetPageSize(),
	)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	pageInfo.SetTotal(int(total))
	pageInfo.SetItems(sellers)
	common.ApiSuccess(c, pageInfo)
}

func CreateSellerAdmin(c *gin.Context) {
	var req CreateSellerAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiError(c, err)
		return
	}
	seller, supply, err := service.CreateSellerWithSupply(service.CreateSellerInput{
		Seller:        req.Seller,
		SupplyAccount: req.SupplyAccount,
		Bindings:      req.Bindings,
	})
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, gin.H{
		"seller":         seller,
		"supply_account": supply,
	})
}

func UpdateSellerAdminStatus(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		common.ApiError(c, err)
		return
	}
	var req UpdateSellerStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiError(c, err)
		return
	}
	if err := service.UpdateSellerStatus(id, req.Status, req.Remark); err != nil {
		common.ApiError(c, err)
		return
	}
	updated, err := model.GetSellerByID(id)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, updated)
}
