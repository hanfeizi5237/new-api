package controller

import (
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/Calcium-Ion/go-epay/epay"
	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/service"
	"github.com/QuantumNous/new-api/setting"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/webhook"
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
	LockOrder(verifyInfo.ServiceTradeNo)
	defer UnlockOrder(verifyInfo.ServiceTradeNo)
	payableAmountMinor, err := parseMarketMoneyToMinor(verifyInfo.Money)
	if err != nil {
		_, _ = c.Writer.Write([]byte("fail"))
		return
	}
	if verifyInfo.TradeStatus == epay.StatusTradeSuccess {
		if _, err := service.CompleteMarketOrderPayment(service.CompleteMarketOrderPaymentInput{
			OrderNo:            verifyInfo.ServiceTradeNo,
			PaymentMethod:      "epay",
			PaymentTradeNo:     verifyInfo.TradeNo,
			Currency:           "CNY",
			PayableAmountMinor: payableAmountMinor,
			ProviderPayload:    common.GetJsonString(params),
		}); err != nil {
			log.Printf("Marketplace 易支付回调处理失败: %v, order=%s", err, verifyInfo.ServiceTradeNo)
			_, _ = c.Writer.Write([]byte("fail"))
			return
		}
		_, _ = c.Writer.Write([]byte("success"))
		return
	}
	if _, err := service.FailMarketOrderPayment(service.FailMarketOrderPaymentInput{
		OrderNo:            verifyInfo.ServiceTradeNo,
		PaymentMethod:      "epay",
		PaymentTradeNo:     verifyInfo.TradeNo,
		Currency:           "CNY",
		PayableAmountMinor: payableAmountMinor,
		FailureReason:      "epay_" + strings.ToLower(verifyInfo.TradeStatus),
		ProviderPayload:    common.GetJsonString(params),
	}); err != nil {
		log.Printf("Marketplace 易支付失败回调处理失败: %v, order=%s", err, verifyInfo.ServiceTradeNo)
		_, _ = c.Writer.Write([]byte("fail"))
		return
	}
	_, _ = c.Writer.Write([]byte("success"))
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
	if event.Type == stripe.EventTypeCheckoutSessionExpired {
		orderNo := event.GetObjectValue("client_reference_id")
		if orderNo == "" {
			c.Status(http.StatusOK)
			return
		}
		LockOrder(orderNo)
		defer UnlockOrder(orderNo)
		if _, err := service.FailMarketOrderPayment(service.FailMarketOrderPaymentInput{
			OrderNo:         orderNo,
			PaymentMethod:   PaymentMethodStripe,
			FailureReason:   "stripe_session_expired",
			ProviderPayload: string(payload),
		}); err != nil {
			log.Printf("Marketplace Stripe 过期回调处理失败: %v, order=%s", err, orderNo)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		c.Status(http.StatusOK)
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
	amountMinor, err := strconv.ParseInt(event.GetObjectValue("amount_total"), 10, 64)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
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
	if event.EventType != "checkout.completed" {
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
	if event.Object.Order.Status != "paid" {
		if _, err := service.FailMarketOrderPayment(service.FailMarketOrderPaymentInput{
			OrderNo:            orderNo,
			PaymentMethod:      "creem",
			PaymentTradeNo:     event.Object.Order.Id,
			Currency:           strings.ToUpper(event.Object.Order.Currency),
			PayableAmountMinor: int64(event.Object.Order.AmountDue),
			FailureReason:      "creem_" + strings.ToLower(event.Object.Order.Status),
			ProviderPayload:    string(bodyBytes),
		}); err != nil {
			log.Printf("Marketplace Creem 失败回调处理失败: %v, order=%s", err, orderNo)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		c.Status(http.StatusOK)
		return
	}
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
		payableAmountMinor, err := parseMarketMoneyToMinor(payload.Result.OrderAmount)
		if err != nil {
			sendWaffoWebhookResponse(c, wh, false, "invalid amount")
			return
		}
		if _, err := service.FailMarketOrderPayment(service.FailMarketOrderPaymentInput{
			OrderNo:            payload.Result.MerchantOrderID,
			PaymentMethod:      "waffo",
			PaymentTradeNo:     payload.Result.PaymentRequestID,
			Currency:           strings.ToUpper(payload.Result.OrderCurrency),
			PayableAmountMinor: payableAmountMinor,
			FailureReason:      "waffo_" + strings.ToLower(payload.Result.OrderStatus),
			ProviderPayload:    string(bodyBytes),
		}); err != nil {
			log.Printf("Marketplace Waffo 失败回调处理失败: %v, order=%s", err, payload.Result.MerchantOrderID)
			sendWaffoWebhookResponse(c, wh, false, err.Error())
			return
		}
		sendWaffoWebhookResponse(c, wh, true, "")
		return
	}
	orderNo := payload.Result.MerchantOrderID
	payableAmountMinor, err := parseMarketMoneyToMinor(payload.Result.OrderAmount)
	if err != nil {
		sendWaffoWebhookResponse(c, wh, false, "invalid amount")
		return
	}
	LockOrder(orderNo)
	defer UnlockOrder(orderNo)
	if _, err := service.CompleteMarketOrderPayment(service.CompleteMarketOrderPaymentInput{
		OrderNo:            orderNo,
		PaymentMethod:      "waffo",
		PaymentTradeNo:     payload.Result.PaymentRequestID,
		Currency:           strings.ToUpper(payload.Result.OrderCurrency),
		PayableAmountMinor: payableAmountMinor,
		ProviderPayload:    string(bodyBytes),
	}); err != nil {
		log.Printf("Marketplace Waffo 回调处理失败: %v, order=%s", err, orderNo)
		sendWaffoWebhookResponse(c, wh, false, err.Error())
		return
	}
	sendWaffoWebhookResponse(c, wh, true, "")
}

func parseMarketMoneyToMinor(amount string) (int64, error) {
	value, err := decimal.NewFromString(strings.TrimSpace(amount))
	if err != nil {
		return 0, err
	}
	return value.Mul(decimal.NewFromInt(100)).Round(0).IntPart(), nil
}
