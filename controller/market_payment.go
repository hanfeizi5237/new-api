package controller

import (
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/service"
	"github.com/QuantumNous/new-api/setting"
	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/webhook"
	"github.com/Calcium-Ion/go-epay/epay"
	"github.com/waffo-com/waffo-go/core"
)

func MarketEpayNotify(c *gin.Context) {
	var params map[string]string
	if c.Request.Method == http.MethodPost {
		if err := c.Request.ParseForm(); err != nil {
			log.Println("Marketplace 易支付回调 POST 解析失败:", err)
			_, _ = c.Writer.Write([]byte("fail"))
			return
		}
		params = map[string]string{}
		for key := range c.Request.PostForm {
			params[key] = c.Request.PostForm.Get(key)
		}
	} else {
		params = map[string]string{}
		for key := range c.Request.URL.Query() {
			params[key] = c.Request.URL.Query().Get(key)
		}
	}
	client := GetEpayClient()
	if client == nil {
		_, _ = c.Writer.Write([]byte("fail"))
		return
	}
	verifyInfo, err := client.Verify(params)
	if err != nil || !verifyInfo.VerifyStatus {
		_, _ = c.Writer.Write([]byte("fail"))
		return
	}
	_, _ = c.Writer.Write([]byte("success"))
	if verifyInfo.TradeStatus != epay.StatusTradeSuccess {
		return
	}
	LockOrder(verifyInfo.ServiceTradeNo)
	defer UnlockOrder(verifyInfo.ServiceTradeNo)
	if _, err := service.CompleteMarketOrderPayment(service.CompleteMarketOrderPaymentInput{
		OrderNo:        verifyInfo.ServiceTradeNo,
		PaymentMethod:  "epay",
		PaymentTradeNo: verifyInfo.TradeNo,
		Currency:       "CNY",
		ProviderPayload: common.GetJsonString(params),
	}); err != nil {
		log.Printf("Marketplace 易支付回调处理失败: %v, order=%s", err, verifyInfo.ServiceTradeNo)
	}
}

func MarketStripeWebhook(c *gin.Context) {
	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.AbortWithStatus(http.StatusServiceUnavailable)
		return
	}
	event, err := webhook.ConstructEventWithOptions(
		payload,
		c.GetHeader("Stripe-Signature"),
		setting.StripeWebhookSecret,
		webhook.ConstructEventOptions{IgnoreAPIVersionMismatch: true},
	)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	if event.Type != stripe.EventTypeCheckoutSessionCompleted {
		c.Status(http.StatusOK)
		return
	}
	orderNo := event.GetObjectValue("client_reference_id")
	if orderNo == "" {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	amountMinor, _ := strconv.ParseInt(event.GetObjectValue("amount_total"), 10, 64)
	LockOrder(orderNo)
	defer UnlockOrder(orderNo)
	if _, err := service.CompleteMarketOrderPayment(service.CompleteMarketOrderPaymentInput{
		OrderNo:            orderNo,
		PaymentMethod:      PaymentMethodStripe,
		PaymentTradeNo:     event.GetObjectValue("payment_intent"),
		Currency:           strings.ToUpper(event.GetObjectValue("currency")),
		PayableAmountMinor: amountMinor,
		ProviderPayload:    string(payload),
	}); err != nil {
		log.Printf("Marketplace Stripe 回调处理失败: %v, order=%s", err, orderNo)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	c.Status(http.StatusOK)
}

func MarketCreemWebhook(c *gin.Context) {
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	signature := c.GetHeader("creem-signature")
	if !verifyCreemSignature(string(bodyBytes), signature, setting.CreemWebhookSecret) {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	var event CreemWebhookEvent
	if err := common.Unmarshal(bodyBytes, &event); err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	if event.EventType != "checkout.completed" || event.Object.Order.Status != "paid" {
		c.Status(http.StatusOK)
		return
	}
	orderNo := event.Object.RequestId
	if orderNo == "" {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	LockOrder(orderNo)
	defer UnlockOrder(orderNo)
	if _, err := service.CompleteMarketOrderPayment(service.CompleteMarketOrderPaymentInput{
		OrderNo:            orderNo,
		PaymentMethod:      "creem",
		PaymentTradeNo:     event.Object.Order.Id,
		Currency:           strings.ToUpper(event.Object.Order.Currency),
		PayableAmountMinor: int64(event.Object.Order.AmountPaid),
		ProviderPayload:    string(bodyBytes),
	}); err != nil {
		log.Printf("Marketplace Creem 回调处理失败: %v, order=%s", err, orderNo)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	c.Status(http.StatusOK)
}

func MarketWaffoWebhook(c *gin.Context) {
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	sdk, err := getWaffoSDK()
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	wh := sdk.Webhook()
	if !wh.VerifySignature(string(bodyBytes), c.GetHeader("X-SIGNATURE")) {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	var event core.WebhookEvent
	if err := common.Unmarshal(bodyBytes, &event); err != nil {
		sendWaffoWebhookResponse(c, wh, false, "invalid payload")
		return
	}
	if event.EventType != core.EventPayment {
		sendWaffoWebhookResponse(c, wh, true, "")
		return
	}
	var payload webhookPayloadWithSubInfo
	if err := common.Unmarshal(bodyBytes, &payload); err != nil {
		sendWaffoWebhookResponse(c, wh, false, "invalid payment payload")
		return
	}
	if payload.Result.OrderStatus != "PAY_SUCCESS" {
		sendWaffoWebhookResponse(c, wh, true, "")
		return
	}
	orderNo := payload.Result.MerchantOrderID
	LockOrder(orderNo)
	defer UnlockOrder(orderNo)
	if _, err := service.CompleteMarketOrderPayment(service.CompleteMarketOrderPaymentInput{
		OrderNo:         orderNo,
		PaymentMethod:   "waffo",
		PaymentTradeNo:  payload.Result.PaymentRequestID,
		ProviderPayload: string(bodyBytes),
	}); err != nil {
		log.Printf("Marketplace Waffo 回调处理失败: %v, order=%s", err, orderNo)
		sendWaffoWebhookResponse(c, wh, false, err.Error())
		return
	}
	sendWaffoWebhookResponse(c, wh, true, "")
}
