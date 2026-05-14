package controller

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/logger"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"
	"github.com/QuantumNous/new-api/setting/operation_setting"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

// ============================================================================
// Alipay Official Payment Controller
// ============================================================================

const alipayGateway = "https://openapi.alipay.com/gateway.do"

// AlipayRequest 支付宝支付请求
type AlipayRequest struct {
	Amount        int64  `json:"amount"`
	PaymentMethod string `json:"payment_method"`
}

// getAlipayCommonParams 生成支付宝公共参数
func getAlipayCommonParams() map[string]string {
	return map[string]string{
		"app_id":    operation_setting.AlipayAppId,
		"format":    "JSON",
		"charset":   "utf-8",
		"sign_type": "RSA2",
		"timestamp": time.Now().Format("2006-01-02 15:04:05"),
		"version":   "1.0",
	}
}

// formatAlipayPrivateKey 格式化支付宝私钥
func formatAlipayPrivateKey(key string) string {
	key = strings.TrimSpace(key)
	if strings.Contains(key, "-----BEGIN") {
		return key
	}
	return fmt.Sprintf("-----BEGIN PRIVATE KEY-----\n%s\n-----END PRIVATE KEY-----", key)
}

// formatAlipayPublicKey 格式化支付宝公钥
func formatAlipayPublicKey(key string) string {
	key = strings.TrimSpace(key)
	if strings.Contains(key, "-----BEGIN") {
		return key
	}
	return fmt.Sprintf("-----BEGIN PUBLIC KEY-----\n%s\n-----END PUBLIC KEY-----", key)
}

// generateAlipaySign 生成支付宝 RSA2 签名
func generateAlipaySign(params map[string]string, privateKeyStr string) (string, error) {
	// 过滤空值和 sign
	filtered := make(map[string]string)
	for k, v := range params {
		if k == "sign" || strings.TrimSpace(v) == "" {
			continue
		}
		filtered[k] = v
	}

	// 按键排序
	keys := make([]string, 0, len(filtered))
	for k := range filtered {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 拼接待签名字符串
	var parts []string
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%s", k, filtered[k]))
	}
	signContent := strings.Join(parts, "&")

	// 加载私钥
	privateKeyStr = formatAlipayPrivateKey(privateKeyStr)
	block, _ := pem.Decode([]byte(privateKeyStr))
	if block == nil {
		return "", fmt.Errorf("failed to decode private key")
	}

	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		privateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return "", fmt.Errorf("failed to parse private key: %w", err)
		}
	}

	// RSA 签名
	hasher := sha256.New()
	hasher.Write([]byte(signContent))
	signature, err := rsa.SignPKCS1v15(nil, privateKey.(*rsa.PrivateKey), crypto.SHA256, hasher.Sum(nil))
	if err != nil {
		return "", fmt.Errorf("failed to sign: %w", err)
	}

	return base64.StdEncoding.EncodeToString(signature), nil
}

// verifyAlipaySign 验证支付宝签名
func verifyAlipaySign(params map[string]string, publicKeyStr string, sign string) error {
	// 过滤 sign 和 sign_type
	filtered := make(map[string]string)
	for k, v := range params {
		if k == "sign" || k == "sign_type" || strings.TrimSpace(v) == "" {
			continue
		}
		filtered[k] = v
	}

	// 按键排序
	keys := make([]string, 0, len(filtered))
	for k := range filtered {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 拼接待验签字符串
	var parts []string
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%s", k, filtered[k]))
	}
	signContent := strings.Join(parts, "&")

	// 加载公钥
	publicKeyStr = formatAlipayPublicKey(publicKeyStr)
	block, _ := pem.Decode([]byte(publicKeyStr))
	if block == nil {
		return fmt.Errorf("failed to decode public key")
	}

	publicKeyInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse public key: %w", err)
	}
	publicKey := publicKeyInterface.(*rsa.PublicKey)

	// 解码签名
	signature, err := base64.StdEncoding.DecodeString(sign)
	if err != nil {
		return fmt.Errorf("failed to decode signature: %w", err)
	}

	// RSA 验签
	hasher := sha256.New()
	hasher.Write([]byte(signContent))
	err = rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, hasher.Sum(nil), signature)
	if err != nil {
		return fmt.Errorf("signature verification failed: %w", err)
	}

	return nil
}

// RequestAlipay 创建支付宝支付订单
func RequestAlipay(c *gin.Context) {
	var req AlipayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiErrorMsg(c, "参数错误")
		return
	}

	if req.Amount < getMinTopup() {
		common.ApiErrorMsg(c, fmt.Sprintf("充值数量不能小于 %d", getMinTopup()))
		return
	}

	id := c.GetInt("id")
	group, err := model.GetUserGroup(id, true)
	if err != nil {
		common.ApiErrorMsg(c, "获取用户分组失败")
		return
	}

	payMoney := getPayMoney(req.Amount, group)
	if payMoney < 0.01 {
		common.ApiErrorMsg(c, "充值金额过低")
		return
	}

	// 检查支付宝配置
	if operation_setting.AlipayAppId == "" || operation_setting.AlipayPrivateKey == "" {
		common.ApiErrorMsg(c, "当前管理员未配置支付宝支付信息")
		return
	}

	// 判断设备类型
	isMobile := false
	userAgent := c.GetHeader("User-Agent")
	if strings.Contains(userAgent, "Mobile") || strings.Contains(userAgent, "Android") || strings.Contains(userAgent, "iPhone") {
		isMobile = true
	}

	// 生成订单号
	tradeNo := fmt.Sprintf("ALI%s%d", common.GetRandomString(6), time.Now().Unix())
	tradeNo = fmt.Sprintf("USR%dNO%s", id, tradeNo)

	// 构建支付参数
	bizContent := map[string]interface{}{
		"out_trade_no": tradeNo,
		"total_amount": fmt.Sprintf("%.2f", payMoney),
		"subject":      fmt.Sprintf("充值 %d", req.Amount),
		"product_code": "FAST_INSTANT_TRADE_PAY",
	}

	if isMobile {
		bizContent["product_code"] = "QUICK_WAP_WAY"
	}

	bizContentJson, _ := json.Marshal(bizContent)

	params := getAlipayCommonParams()
	if isMobile {
		params["method"] = "alipay.trade.wap.pay"
	} else {
		params["method"] = "alipay.trade.page.pay"
	}
	params["biz_content"] = string(bizContentJson)

	// 回调地址
	callBackAddress := service.GetCallbackAddress()
	notifyUrl := callBackAddress + "/api/alipay/notify"
	returnUrl := callBackAddress + "/console/log"

	if operation_setting.AlipayNotifyUrl != "" {
		notifyUrl = operation_setting.AlipayNotifyUrl
	}
	if operation_setting.AlipayReturnUrl != "" {
		returnUrl = operation_setting.AlipayReturnUrl
	}

	params["notify_url"] = notifyUrl
	params["return_url"] = returnUrl

	// 生成签名
	sign, err := generateAlipaySign(params, operation_setting.AlipayPrivateKey)
	if err != nil {
		logger.LogError(c.Request.Context(), fmt.Sprintf("支付宝签名生成失败 user_id=%d error=%q", id, err.Error()))
		common.ApiErrorMsg(c, "签名生成失败")
		return
	}
	params["sign"] = sign

	// 构建支付 URL
	payUrl := alipayGateway + "?" + buildQueryString(params)

	// 创建充值订单
	amount := req.Amount
	if operation_setting.GetQuotaDisplayType() == operation_setting.QuotaDisplayTypeTokens {
		dAmount := decimal.NewFromInt(int64(amount))
		dQuotaPerUnit := decimal.NewFromFloat(common.QuotaPerUnit)
		amount = dAmount.Div(dQuotaPerUnit).IntPart()
	}
	topUp := &model.TopUp{
		UserId:          id,
		Amount:          amount,
		Money:           payMoney,
		TradeNo:         tradeNo,
		PaymentMethod:   req.PaymentMethod,
		PaymentProvider: model.PaymentProviderAlipay,
		CreateTime:      time.Now().Unix(),
		Status:          common.TopUpStatusPending,
	}
	if err = topUp.Insert(); err != nil {
		logger.LogError(c.Request.Context(), fmt.Sprintf("支付宝充值订单创建失败 user_id=%d trade_no=%s error=%q", id, tradeNo, err.Error()))
		common.ApiErrorMsg(c, "创建订单失败")
		return
	}

	logger.LogInfo(c.Request.Context(), fmt.Sprintf("支付宝充值订单创建成功 user_id=%d trade_no=%s amount=%d money=%.2f", id, tradeNo, req.Amount, payMoney))

	common.ApiSuccess(c, gin.H{
		"pay_url":  payUrl,
		"trade_no": tradeNo,
	})
}

// buildQueryString 构建 URL 查询字符串
func buildQueryString(params map[string]string) string {
	values := url.Values{}
	for k, v := range params {
		values.Set(k, v)
	}
	return values.Encode()
}

// AlipayNotify 处理支付宝异步回调
func AlipayNotify(c *gin.Context) {
	// 解析回调参数
	params := make(map[string]string)
	if c.Request.Method == "POST" {
		c.Request.ParseForm()
		for k, v := range c.Request.PostForm {
			if len(v) > 0 {
				params[k] = v[0]
			}
		}
	} else {
		for k, v := range c.Request.URL.Query() {
			if len(v) > 0 {
				params[k] = v[0]
			}
		}
	}

	logger.LogInfo(c.Request.Context(), fmt.Sprintf("支付宝回调收到请求 params=%q", common.GetJsonString(params)))

	if len(params) == 0 {
		logger.LogWarn(c.Request.Context(), "支付宝回调参数为空")
		c.Writer.Write([]byte("fail"))
		return
	}

	// 验证签名
	sign := params["sign"]
	if sign == "" {
		logger.LogWarn(c.Request.Context(), "支付宝回调缺少签名")
		c.Writer.Write([]byte("fail"))
		return
	}

	err := verifyAlipaySign(params, operation_setting.AlipayPublicKey, sign)
	if err != nil {
		logger.LogWarn(c.Request.Context(), fmt.Sprintf("支付宝回调验签失败 error=%q", err.Error()))
		c.Writer.Write([]byte("fail"))
		return
	}

	logger.LogInfo(c.Request.Context(), "支付宝回调验签成功")

	// 处理支付结果
	tradeStatus := params["trade_status"]
	tradeNo := params["out_trade_no"]

	if tradeStatus == "TRADE_SUCCESS" || tradeStatus == "TRADE_FINISHED" {
		LockOrder(tradeNo)
		defer UnlockOrder(tradeNo)

		topUp := model.GetTopUpByTradeNo(tradeNo)
		if topUp == nil {
			logger.LogWarn(c.Request.Context(), fmt.Sprintf("支付宝回调订单不存在 trade_no=%s", tradeNo))
			c.Writer.Write([]byte("success"))
			return
		}

		if topUp.Status == common.TopUpStatusPending {
			topUp.Status = common.TopUpStatusSuccess
			err := topUp.Update()
			if err != nil {
				logger.LogError(c.Request.Context(), fmt.Sprintf("支付宝回调更新订单失败 trade_no=%s error=%q", tradeNo, err.Error()))
				c.Writer.Write([]byte("fail"))
				return
			}

			// 更新用户额度
			dAmount := decimal.NewFromInt(int64(topUp.Amount))
			dQuotaPerUnit := decimal.NewFromFloat(common.QuotaPerUnit)
			quotaToAdd := int(dAmount.Mul(dQuotaPerUnit).IntPart())
			err = model.IncreaseUserQuota(topUp.UserId, quotaToAdd, true)
			if err != nil {
				logger.LogError(c.Request.Context(), fmt.Sprintf("支付宝回调更新用户额度失败 trade_no=%s user_id=%d error=%q", tradeNo, topUp.UserId, err.Error()))
				c.Writer.Write([]byte("fail"))
				return
			}

			logger.LogInfo(c.Request.Context(), fmt.Sprintf("支付宝充值成功 trade_no=%s user_id=%d quota_to_add=%d", tradeNo, topUp.UserId, quotaToAdd))
			model.RecordTopupLog(topUp.UserId, fmt.Sprintf("使用支付宝充值成功，充值金额: %v，支付金额：%f", logger.LogQuota(quotaToAdd), topUp.Money), c.ClientIP(), topUp.PaymentMethod, "alipay")
		}
	}

	c.Writer.Write([]byte("success"))
}
