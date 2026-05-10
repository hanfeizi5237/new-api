package controller

import (
	"errors"
	"strconv"
	"strings"

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

func isSupportedMarketPaymentMethod(method string) bool {
	switch strings.TrimSpace(method) {
	case model.PaymentMethodStripe, model.PaymentMethodCreem, model.PaymentMethodWaffo, "epay":
		return true
	default:
		return false
	}
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
	if id <= 0 {
		common.ApiError(c, errors.New("invalid listing id"))
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
	if req.ListingId <= 0 || req.SkuId <= 0 || req.Quantity <= 0 {
		common.ApiError(c, errors.New("listing_id, sku_id, and quantity must be positive"))
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
	buyerUserId := 0
	if rawBuyerUserID := strings.TrimSpace(c.Query("buyer_user_id")); rawBuyerUserID != "" {
		parsedBuyerUserID, err := strconv.Atoi(rawBuyerUserID)
		if err != nil || parsedBuyerUserID <= 0 {
			common.ApiError(c, errors.New("invalid buyer_user_id"))
			return
		}
		buyerUserId = parsedBuyerUserID
	}
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
	if id <= 0 {
		common.ApiError(c, errors.New("invalid order id"))
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
	if id <= 0 {
		common.ApiError(c, errors.New("invalid order id"))
		return
	}
	var req PayMarketOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiError(c, err)
		return
	}
	if !isSupportedMarketPaymentMethod(req.PaymentMethod) {
		common.ApiError(c, errors.New("invalid payment_method"))
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
