package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Calcium-Ion/go-epay/epay"
	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/QuantumNous/new-api/setting/system_setting"
	"github.com/shopspring/decimal"
	"github.com/stripe/stripe-go/v81"
	stripeSession "github.com/stripe/stripe-go/v81/checkout/session"
	waffo "github.com/waffo-com/waffo-go"
	"github.com/waffo-com/waffo-go/config"
	waffoOrder "github.com/waffo-com/waffo-go/types/order"
)

type marketProviderPaymentInitInput struct {
	Order                  *model.MarketOrder
	Buyer                  *model.User
	RequestedPaymentMethod string
}

type marketProviderPaymentInitResult struct {
	ProviderReference string
	RedirectURL       string
	FormURL           string
	FormParams        map[string]string
	ProviderPayload   string
}

type storedMarketPaymentIntent struct {
	PaymentMethod     string            `json:"payment_method"`
	ProviderReference string            `json:"provider_reference,omitempty"`
	RedirectURL       string            `json:"redirect_url,omitempty"`
	FormURL           string            `json:"form_url,omitempty"`
	FormParams        map[string]string `json:"form_params,omitempty"`
	ProviderPayload   string            `json:"provider_payload,omitempty"`
}

type marketCreemProduct struct {
	ProductID string  `json:"productId"`
	Name      string  `json:"name"`
	Price     float64 `json:"price"`
	Currency  string  `json:"currency"`
}

type marketCreemCheckoutRequest struct {
	ProductId string `json:"product_id"`
	RequestId string `json:"request_id"`
	Customer  struct {
		Email string `json:"email"`
	} `json:"customer"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

type marketCreemCheckoutResponse struct {
	CheckoutURL string `json:"checkout_url"`
	ID          string `json:"id"`
}

var (
	marketStripePaymentIntentCreator = createMarketStripePaymentIntent
	marketCreemPaymentIntentCreator  = createMarketCreemPaymentIntent
	marketWaffoPaymentIntentCreator  = createMarketWaffoPaymentIntent
	marketEpayPaymentIntentCreator   = createMarketEpayPaymentIntent
)

func decodeStoredMarketPaymentIntent(order *model.MarketOrder, paymentMethod string) (*MarketPaymentIntent, bool) {
	if order == nil || strings.TrimSpace(order.ProviderPayload) == "" {
		return nil, false
	}
	var stored storedMarketPaymentIntent
	if err := common.UnmarshalJsonStr(order.ProviderPayload, &stored); err != nil {
		return nil, false
	}
	if stored.PaymentMethod == "" || stored.PaymentMethod != strings.TrimSpace(paymentMethod) {
		return nil, false
	}
	return &MarketPaymentIntent{
		OrderID:            order.Id,
		OrderNo:            order.OrderNo,
		PaymentMethod:      stored.PaymentMethod,
		Currency:           order.Currency,
		PayableAmountMinor: order.PayableAmountMinor,
		OrderStatus:        order.OrderStatus,
		PaymentStatus:      order.PaymentStatus,
		ProviderReference:  stored.ProviderReference,
		RedirectURL:        stored.RedirectURL,
		FormURL:            stored.FormURL,
		FormParams:         stored.FormParams,
	}, true
}

func encodeStoredMarketPaymentIntent(paymentMethod string, result *marketProviderPaymentInitResult) (string, error) {
	if result == nil {
		return "", errors.New("market provider init result is required")
	}
	payload, err := common.Marshal(storedMarketPaymentIntent{
		PaymentMethod:     strings.TrimSpace(paymentMethod),
		ProviderReference: result.ProviderReference,
		RedirectURL:       result.RedirectURL,
		FormURL:           result.FormURL,
		FormParams:        result.FormParams,
		ProviderPayload:   result.ProviderPayload,
	})
	if err != nil {
		return "", err
	}
	return string(payload), nil
}

func createMarketProviderPaymentIntent(input marketProviderPaymentInitInput) (*marketProviderPaymentInitResult, error) {
	if input.Order == nil || input.Buyer == nil {
		return nil, errors.New("market payment init requires order and buyer")
	}
	paymentMethod := strings.TrimSpace(input.RequestedPaymentMethod)
	switch {
	case paymentMethod == "stripe":
		return marketStripePaymentIntentCreator(input)
	case paymentMethod == "creem":
		return marketCreemPaymentIntentCreator(input)
	case paymentMethod == "waffo":
		return marketWaffoPaymentIntentCreator(input)
	case paymentMethod == "epay", operation_setting.ContainsPayMethod(paymentMethod):
		return marketEpayPaymentIntentCreator(input)
	default:
		return nil, fmt.Errorf("unsupported payment method: %s", paymentMethod)
	}
}

func createMarketStripePaymentIntent(input marketProviderPaymentInitInput) (*marketProviderPaymentInitResult, error) {
	if !strings.HasPrefix(setting.StripeApiSecret, "sk_") && !strings.HasPrefix(setting.StripeApiSecret, "rk_") {
		return nil, errors.New("invalid stripe api secret")
	}
	stripe.Key = setting.StripeApiSecret

	currency := strings.ToLower(strings.TrimSpace(input.Order.Currency))
	if currency == "" {
		currency = "cny"
	}
	successURL := getMarketplaceConsoleReturnURL("payment_return=1")
	cancelURL := getMarketplaceConsoleReturnURL("payment_cancel=1")
	params := &stripe.CheckoutSessionParams{
		ClientReferenceID:   stripe.String(input.Order.OrderNo),
		SuccessURL:          stripe.String(successURL),
		CancelURL:           stripe.String(cancelURL),
		Mode:                stripe.String(string(stripe.CheckoutSessionModePayment)),
		AllowPromotionCodes: stripe.Bool(setting.StripePromotionCodesEnabled),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Quantity: stripe.Int64(1),
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency:   stripe.String(currency),
					UnitAmount: stripe.Int64(input.Order.PayableAmountMinor),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name: stripe.String("Marketplace Order " + input.Order.OrderNo),
					},
				},
			},
		},
	}
	if strings.TrimSpace(input.Buyer.StripeCustomer) != "" {
		params.Customer = stripe.String(strings.TrimSpace(input.Buyer.StripeCustomer))
	} else if strings.TrimSpace(input.Buyer.Email) != "" {
		params.CustomerEmail = stripe.String(strings.TrimSpace(input.Buyer.Email))
		params.CustomerCreation = stripe.String(string(stripe.CheckoutSessionCustomerCreationAlways))
	}
	result, err := stripeSession.New(params)
	if err != nil {
		return nil, err
	}
	return &marketProviderPaymentInitResult{
		ProviderReference: result.ID,
		RedirectURL:       result.URL,
		ProviderPayload: common.GetJsonString(map[string]any{
			"client_reference_id": input.Order.OrderNo,
			"session_id":          result.ID,
			"url":                 result.URL,
		}),
	}, nil
}

func createMarketCreemPaymentIntent(input marketProviderPaymentInitInput) (*marketProviderPaymentInitResult, error) {
	if strings.TrimSpace(setting.CreemApiKey) == "" {
		return nil, errors.New("creem api key is not configured")
	}
	product, err := selectMarketCreemProduct(input.Order)
	if err != nil {
		return nil, err
	}
	apiURL := "https://api.creem.io/v1/checkouts"
	if setting.CreemTestMode {
		apiURL = "https://test-api.creem.io/v1/checkouts"
	}
	requestData := marketCreemCheckoutRequest{
		ProductId: product.ProductID,
		RequestId: input.Order.OrderNo,
		Metadata: map[string]string{
			"market_order_no": input.Order.OrderNo,
			"product_name":    product.Name,
		},
	}
	requestData.Customer.Email = strings.TrimSpace(input.Buyer.Email)
	jsonData, err := common.Marshal(requestData)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", setting.CreemApiKey)
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("creem api http status %d", resp.StatusCode)
	}
	var checkoutResp marketCreemCheckoutResponse
	if err := common.Unmarshal(body, &checkoutResp); err != nil {
		return nil, err
	}
	if strings.TrimSpace(checkoutResp.CheckoutURL) == "" {
		return nil, errors.New("creem checkout_url is empty")
	}
	providerReference := strings.TrimSpace(checkoutResp.ID)
	if providerReference == "" {
		providerReference = input.Order.OrderNo
	}
	return &marketProviderPaymentInitResult{
		ProviderReference: providerReference,
		RedirectURL:       checkoutResp.CheckoutURL,
		ProviderPayload: common.GetJsonString(map[string]any{
			"request_id":   input.Order.OrderNo,
			"product_id":   product.ProductID,
			"checkout_url": checkoutResp.CheckoutURL,
			"id":           providerReference,
		}),
	}, nil
}

func createMarketWaffoPaymentIntent(input marketProviderPaymentInitInput) (*marketProviderPaymentInitResult, error) {
	sdk, err := getMarketWaffoSDK()
	if err != nil {
		return nil, err
	}
	callbackAddr := GetCallbackAddress()
	notifyURL := callbackAddr + "/api/market/payment/waffo/webhook"
	if setting.WaffoNotifyUrl != "" {
		notifyURL = setting.WaffoNotifyUrl
	}
	returnURL := getMarketplaceConsoleReturnURL("payment_return=1")
	if setting.WaffoReturnUrl != "" {
		returnURL = setting.WaffoReturnUrl
	}
	currency := strings.ToUpper(strings.TrimSpace(input.Order.Currency))
	if currency == "" {
		currency = getMarketWaffoCurrency()
	}
	createParams := &waffoOrder.CreateOrderParams{
		PaymentRequestID: input.Order.OrderNo,
		MerchantOrderID:  input.Order.OrderNo,
		OrderAmount:      formatMarketMinorAmount(input.Order.PayableAmountMinor, currency),
		OrderCurrency:    currency,
		OrderDescription: fmt.Sprintf("Marketplace Order %s", input.Order.OrderNo),
		OrderRequestedAt: time.Now().UTC().Format("2006-01-02T15:04:05.000Z"),
		NotifyURL:        notifyURL,
		MerchantInfo: &waffoOrder.MerchantInfo{
			MerchantID: setting.WaffoMerchantId,
		},
		UserInfo: &waffoOrder.UserInfo{
			UserID:       fmt.Sprintf("%d", input.Buyer.Id),
			UserEmail:    getMarketWaffoUserEmail(input.Buyer),
			UserTerminal: "WEB",
		},
		PaymentInfo: &waffoOrder.PaymentInfo{
			ProductName: "ONE_TIME_PAYMENT",
		},
		SuccessRedirectURL: returnURL,
		FailedRedirectURL:  returnURL,
	}
	resp, err := sdk.Order().Create(context.Background(), createParams, nil)
	if err != nil {
		return nil, err
	}
	if !resp.IsSuccess() {
		return nil, fmt.Errorf("waffo create order failed: [%s] %s", resp.Code, resp.Message)
	}
	orderData := resp.GetData()
	redirectURL := orderData.FetchRedirectURL()
	if redirectURL == "" {
		redirectURL = orderData.OrderAction
	}
	return &marketProviderPaymentInitResult{
		ProviderReference: input.Order.OrderNo,
		RedirectURL:       redirectURL,
		ProviderPayload: common.GetJsonString(map[string]any{
			"payment_request_id": input.Order.OrderNo,
			"merchant_order_id":  input.Order.OrderNo,
			"payment_url":        redirectURL,
		}),
	}, nil
}

func createMarketEpayPaymentIntent(input marketProviderPaymentInitInput) (*marketProviderPaymentInitResult, error) {
	client := getMarketEpayClient()
	if client == nil {
		return nil, errors.New("epay is not configured")
	}
	payType := resolveMarketEpayType(input.RequestedPaymentMethod)
	if payType == "" {
		return nil, errors.New("epay payment type is not configured")
	}
	callbackAddr := GetCallbackAddress()
	returnURL, err := url.Parse(getMarketplaceConsoleReturnURL("payment_return=1"))
	if err != nil {
		return nil, err
	}
	notifyURL, err := url.Parse(strings.TrimRight(callbackAddr, "/") + "/api/market/payment/epay/notify")
	if err != nil {
		return nil, err
	}
	uri, params, err := client.Purchase(&epay.PurchaseArgs{
		Type:           payType,
		ServiceTradeNo: input.Order.OrderNo,
		Name:           "Marketplace " + input.Order.OrderNo,
		Money:          formatMarketMinorAmount(input.Order.PayableAmountMinor, strings.ToUpper(strings.TrimSpace(input.Order.Currency))),
		Device:         epay.PC,
		NotifyUrl:      notifyURL,
		ReturnUrl:      returnURL,
	})
	if err != nil {
		return nil, err
	}
	return &marketProviderPaymentInitResult{
		ProviderReference: input.Order.OrderNo,
		FormURL:           uri,
		FormParams:        params,
		ProviderPayload: common.GetJsonString(map[string]any{
			"service_trade_no": input.Order.OrderNo,
			"epay_type":        payType,
			"url":              uri,
			"params":           params,
		}),
	}, nil
}

func selectMarketCreemProduct(order *model.MarketOrder) (*marketCreemProduct, error) {
	var products []marketCreemProduct
	if err := common.UnmarshalJsonStr(setting.CreemProducts, &products); err != nil {
		return nil, err
	}
	for i := range products {
		product := &products[i]
		productCurrency := strings.ToUpper(strings.TrimSpace(product.Currency))
		orderCurrency := strings.ToUpper(strings.TrimSpace(order.Currency))
		if productCurrency != "" && orderCurrency != "" && productCurrency != orderCurrency {
			continue
		}
		priceMinor := decimal.NewFromFloat(product.Price).Mul(decimal.NewFromInt(100)).Round(0).IntPart()
		if priceMinor == order.PayableAmountMinor {
			return product, nil
		}
	}
	return nil, fmt.Errorf("no creem product configured for %s %d", strings.ToUpper(strings.TrimSpace(order.Currency)), order.PayableAmountMinor)
}

func getMarketEpayClient() *epay.Client {
	if operation_setting.PayAddress == "" || operation_setting.EpayId == "" || operation_setting.EpayKey == "" {
		return nil
	}
	client, err := epay.NewClient(&epay.Config{
		PartnerID: operation_setting.EpayId,
		Key:       operation_setting.EpayKey,
	}, operation_setting.PayAddress)
	if err != nil {
		return nil
	}
	return client
}

func resolveMarketEpayType(requested string) string {
	requested = strings.TrimSpace(requested)
	if operation_setting.ContainsPayMethod(requested) {
		return requested
	}
	for _, payMethod := range operation_setting.PayMethods {
		if candidate := strings.TrimSpace(payMethod["type"]); candidate != "" {
			return candidate
		}
	}
	return ""
}

func getMarketWaffoSDK() (*waffo.Waffo, error) {
	env := config.Sandbox
	apiKey := setting.WaffoSandboxApiKey
	privateKey := setting.WaffoSandboxPrivateKey
	publicKey := setting.WaffoSandboxPublicCert
	if !setting.WaffoSandbox {
		env = config.Production
		apiKey = setting.WaffoApiKey
		privateKey = setting.WaffoPrivateKey
		publicKey = setting.WaffoPublicCert
	}
	builder := config.NewConfigBuilder().
		APIKey(apiKey).
		PrivateKey(privateKey).
		WaffoPublicKey(publicKey).
		Environment(env)
	if setting.WaffoMerchantId != "" {
		builder = builder.MerchantID(setting.WaffoMerchantId)
	}
	cfg, err := builder.Build()
	if err != nil {
		return nil, err
	}
	return waffo.New(cfg), nil
}

func getMarketWaffoUserEmail(user *model.User) string {
	if user != nil && strings.TrimSpace(user.Email) != "" {
		return strings.TrimSpace(user.Email)
	}
	if user == nil {
		return "market@example.com"
	}
	return fmt.Sprintf("%d@examples.com", user.Id)
}

func getMarketWaffoCurrency() string {
	if strings.TrimSpace(setting.WaffoCurrency) != "" {
		return strings.ToUpper(strings.TrimSpace(setting.WaffoCurrency))
	}
	return "USD"
}

func formatMarketMinorAmount(amountMinor int64, currency string) string {
	currency = strings.ToUpper(strings.TrimSpace(currency))
	if currency == "JPY" || currency == "KRW" || currency == "VND" || currency == "IDR" {
		return fmt.Sprintf("%d", amountMinor)
	}
	return decimal.NewFromInt(amountMinor).Div(decimal.NewFromInt(100)).StringFixed(2)
}

func getMarketplaceConsoleReturnURL(query string) string {
	base := strings.TrimRight(system_setting.ServerAddress, "/") + "/console/marketplace"
	if strings.TrimSpace(query) == "" {
		return base
	}
	return base + "?" + strings.TrimLeft(strings.TrimSpace(query), "?")
}
