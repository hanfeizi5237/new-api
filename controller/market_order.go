package controller

import (
	"strconv"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"
	"github.com/gin-gonic/gin"
)

type CreateMarketOrderRequest struct {
	ListingId      int    `json:"listing_id"`
	SkuId          int    `json:"sku_id"`
	Quantity       int    `json:"quantity"`
	IdempotencyKey string `json:"idempotency_key"`
	Currency       string `json:"currency,omitempty"`
}

type PayMarketOrderRequest struct {
	PaymentMethod string `json:"payment_method"`
}

type MarketOrderView struct {
	Order *model.MarketOrder      `json:"order"`
	Items []model.MarketOrderItem `json:"items"`
}

func GetMarketListings(c *gin.Context) {
	pageInfo := common.GetPageQuery(c)
	listings, total, err := service.ListPublicMarketListings(
		c.Query("keyword"),
		pageInfo.GetStartIdx(),
		pageInfo.GetPageSize(),
	)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	items := make([]service.MarketListingDetail, 0, len(listings))
	for _, listing := range listings {
		detail, detailErr := service.GetPublicMarketListingDetail(listing.Id)
		if detailErr != nil {
			common.ApiError(c, detailErr)
			return
		}
		items = append(items, *detail)
	}
	pageInfo.SetTotal(int(total))
	pageInfo.SetItems(items)
	common.ApiSuccess(c, pageInfo)
}

func GetMarketListing(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		common.ApiError(c, err)
		return
	}
	detail, err := service.GetPublicMarketListingDetail(id)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, detail)
}

func CreateMarketOrder(c *gin.Context) {
	var req CreateMarketOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiError(c, err)
		return
	}
	order, items, err := service.CreateMarketOrder(service.CreateMarketOrderInput{
		BuyerUserID:    c.GetInt("id"),
		ListingID:      req.ListingId,
		SkuID:          req.SkuId,
		Quantity:       req.Quantity,
		IdempotencyKey: req.IdempotencyKey,
		Currency:       req.Currency,
	})
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, MarketOrderView{
		Order: order,
		Items: items,
	})
}

func GetMarketOrders(c *gin.Context) {
	userId := c.GetInt("id")
	pageInfo := common.GetPageQuery(c)
	orders, total, err := service.ListBuyerMarketOrders(userId, pageInfo.GetStartIdx(), pageInfo.GetPageSize())
	if err != nil {
		common.ApiError(c, err)
		return
	}
	items := make([]MarketOrderView, 0, len(orders))
	for _, order := range orders {
		detail, detailErr := service.GetBuyerMarketOrderDetail(order.Id, userId)
		if detailErr != nil {
			common.ApiError(c, detailErr)
			return
		}
		items = append(items, MarketOrderView{
			Order: detail.Order,
			Items: detail.Items,
		})
	}
	pageInfo.SetTotal(int(total))
	pageInfo.SetItems(items)
	common.ApiSuccess(c, pageInfo)
}

func GetMarketOrdersAdmin(c *gin.Context) {
	pageInfo := common.GetPageQuery(c)
	buyerUserId, _ := strconv.Atoi(c.Query("buyer_user_id"))
	orders, total, err := service.ListMarketOrdersAdmin(
		c.Query("keyword"),
		buyerUserId,
		c.Query("order_status"),
		c.Query("payment_status"),
		c.Query("entitlement_status"),
		pageInfo.GetStartIdx(),
		pageInfo.GetPageSize(),
	)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	items := make([]MarketOrderView, 0, len(orders))
	for _, order := range orders {
		orderItems, itemErr := model.GetMarketOrderItemsByOrderID(order.Id)
		if itemErr != nil {
			common.ApiError(c, itemErr)
			return
		}
		items = append(items, MarketOrderView{
			Order: order,
			Items: orderItems,
		})
	}
	pageInfo.SetTotal(int(total))
	pageInfo.SetItems(items)
	common.ApiSuccess(c, pageInfo)
}

func GetMarketOrder(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		common.ApiError(c, err)
		return
	}
	detail, err := service.GetBuyerMarketOrderDetail(id, c.GetInt("id"))
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, MarketOrderView{
		Order: detail.Order,
		Items: detail.Items,
	})
}

func PayMarketOrder(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		common.ApiError(c, err)
		return
	}
	var req PayMarketOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiError(c, err)
		return
	}
	intent, err := service.PrepareMarketOrderPayment(service.PrepareMarketOrderPaymentInput{
		OrderID:       id,
		BuyerUserID:   c.GetInt("id"),
		PaymentMethod: req.PaymentMethod,
	})
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, intent)
}
