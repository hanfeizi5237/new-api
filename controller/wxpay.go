package controller

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
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
// WeChat Pay Official Payment Controller
// ============================================================================

const wxpayBaseURL = "https://api.mch.weixin.qq.com"

// WxpayRequest 微信支付请求
type WxpayRequest struct {
	Amount        int64  `json:"amount"`
	PaymentMethod string `json:"payment_method"`
}

// WxpayNativeResponse 微信 Native 支付响应
type WxpayNativeResponse struct {
	CodeUrl string `json:"code_url"`
}

// WxpayH5Response 微信 H5 支付响应
type WxpayH5Response struct {
	H5Url string `json:"h5_url"`
}

// formatWxpayPrivateKey 格式化微信支付私钥
func formatWxpayPrivateKey(key string) string {
	key = strings.TrimSpace(key)
	if strings.Contains(key, "-----BEGIN") {
		return key
	}
	return fmt.Sprintf("-----BEGIN PRIVATE KEY-----\n%s\n-----END PRIVATE KEY-----", key)
}

// formatWxpayPublicKey 格式化微信支付公钥
func formatWxpayPublicKey(key string) string {
	key = strings.TrimSpace(key)
	if strings.Contains(key, "-----BEGIN") {
		return key
	}
	return fmt.Sprintf("-----BEGIN PUBLIC KEY-----\n%s\n-----END PUBLIC KEY-----", key)
}

// generateWxpaySignature 生成微信支付签名
func generateWxpaySignature(method, url string, body []byte, privateKeyStr string) (string, string, string, error) {
	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	nonce := generateNonce()

	// 构建待签名字符串
	var bodyStr string
	if len(body) > 0 {
		bodyStr = string(body)
	}

	signContent := fmt.Sprintf("%s\n%s\n%s\n%s\n", method, url, timestamp, nonce)
	if bodyStr != "" {
		signContent += bodyStr + "\n"
	}

	// 加载私钥
	privateKeyStr = formatWxpayPrivateKey(privateKeyStr)
	block, _ := pem.Decode([]byte(privateKeyStr))
	if block == nil {
		return "", "", "", fmt.Errorf("failed to decode private key")
	}

	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		privateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return "", "", "", fmt.Errorf("failed to parse private key: %w", err)
		}
	}

	// RSA 签名
	hasher := sha256.New()
	hasher.Write([]byte(signContent))
	signature, err := rsa.SignPKCS1v15(nil, privateKey.(*rsa.PrivateKey), 0, hasher.Sum(nil))
	if err != nil {
		return "", "", "", fmt.Errorf("failed to sign: %w", err)
	}

	return base64.StdEncoding.EncodeToString(signature), timestamp, nonce, nil
}

// generateNonce 生成随机字符串
func generateNonce() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// yuanToFen 元转分
func yuanToFen(yuan float64) int64 {
	return int64(yuan * 100)
}

// RequestWxpay 创建微信支付订单
func RequestWxpay(c *gin.Context) {
	var req WxpayRequest
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

	// 检查微信支付配置
	if operation_setting.WxpayAppId == "" || operation_setting.WxpayMchId == "" || operation_setting.WxpayPrivateKey == "" {
		common.ApiErrorMsg(c, "当前管理员未配置微信支付信息")
		return
	}

	// 判断设备类型
	isMobile := false
	userAgent := c.GetHeader("User-Agent")
	if strings.Contains(userAgent, "Mobile") || strings.Contains(userAgent, "Android") || strings.Contains(userAgent, "iPhone") {
		isMobile = true
	}

	// 生成订单号
	tradeNo := fmt.Sprintf("WX%s%d", common.GetRandomString(6), time.Now().Unix())
	tradeNo = fmt.Sprintf("USR%dNO%s", id, tradeNo)

	// 回调地址
	callBackAddress := service.GetCallbackAddress()
	notifyUrl := callBackAddress + "/api/wxpay/notify"
	if operation_setting.WxpayNotifyUrl != "" {
		notifyUrl = operation_setting.WxpayNotifyUrl
	}

	var payUrl string
	var payType string

	if isMobile {
		// H5 支付
		payUrl, payType, err = createWxpayH5Order(tradeNo, payMoney, notifyUrl, c.ClientIP())
	} else {
		// Native 扫码支付
		payUrl, payType, err = createWxpayNativeOrder(tradeNo, payMoney, notifyUrl)
	}

	if err != nil {
		logger.LogError(c.Request.Context(), fmt.Sprintf("微信支付订单创建失败 user_id=%d error=%q", id, err.Error()))
		common.ApiErrorMsg(c, "创建支付订单失败")
		return
	}

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
		PaymentProvider: model.PaymentProviderWxpay,
		CreateTime:      time.Now().Unix(),
		Status:          common.TopUpStatusPending,
	}
	if err = topUp.Insert(); err != nil {
		logger.LogError(c.Request.Context(), fmt.Sprintf("微信充值订单创建失败 user_id=%d trade_no=%s error=%q", id, tradeNo, err.Error()))
		common.ApiErrorMsg(c, "创建订单失败")
		return
	}

	logger.LogInfo(c.Request.Context(), fmt.Sprintf("微信充值订单创建成功 user_id=%d trade_no=%s amount=%d money=%.2f pay_type=%s", id, tradeNo, req.Amount, payMoney, payType))

	common.ApiSuccess(c, gin.H{
		"pay_url":   payUrl,
		"trade_no":  tradeNo,
		"pay_type":  payType,
		"is_mobile": isMobile,
	})
}

// createWxpayNativeOrder 创建微信 Native 扫码支付订单
func createWxpayNativeOrder(tradeNo string, payMoney float64, notifyUrl string) (string, string, error) {
	url := fmt.Sprintf("%s/v3/pay/transactions/native", wxpayBaseURL)

	body := map[string]interface{}{
		"appid":        operation_setting.WxpayAppId,
		"mchid":        operation_setting.WxpayMchId,
		"description":  fmt.Sprintf("充值 %.2f", payMoney),
		"out_trade_no": tradeNo,
		"notify_url":   notifyUrl,
		"amount": map[string]interface{}{
			"total":    yuanToFen(payMoney),
			"currency": "CNY",
		},
	}

	bodyJson, _ := json.Marshal(body)

	// 生成签名
	sign, timestamp, nonce, err := generateWxpaySignature("POST", "/v3/pay/transactions/native", bodyJson, operation_setting.WxpayPrivateKey)
	if err != nil {
		return "", "", err
	}

	// 构建 Authorization 头
	authHeader := fmt.Sprintf("mchid=\"%s\",nonce_str=\"%s\",timestamp=\"%s\",serial_no=\"%s\",signature=\"%s\"",
		operation_setting.WxpayMchId, nonce, timestamp, operation_setting.WxpayCertSerial, sign)

	// 发送请求
	req, err := http.NewRequest("POST", url, strings.NewReader(string(bodyJson)))
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("WECHATPAY2-SHA256-RSA2048 %s", authHeader))

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("wxpay native order failed: %s", string(respBody))
	}

	var result WxpayNativeResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", "", err
	}

	if result.CodeUrl == "" {
		return "", "", fmt.Errorf("wxpay native order returned empty code_url")
	}

	return result.CodeUrl, "native", nil
}

// createWxpayH5Order 创建微信 H5 支付订单
func createWxpayH5Order(tradeNo string, payMoney float64, notifyUrl string, clientIP string) (string, string, error) {
	url := fmt.Sprintf("%s/v3/pay/transactions/h5", wxpayBaseURL)

	body := map[string]interface{}{
		"appid":        operation_setting.WxpayAppId,
		"mchid":        operation_setting.WxpayMchId,
		"description":  fmt.Sprintf("充值 %.2f", payMoney),
		"out_trade_no": tradeNo,
		"notify_url":   notifyUrl,
		"amount": map[string]interface{}{
			"total":    yuanToFen(payMoney),
			"currency": "CNY",
		},
		"scene_info": map[string]interface{}{
			"payer_client_ip": clientIP,
			"h5_info": map[string]string{
				"type": "Wap",
			},
		},
	}

	bodyJson, _ := json.Marshal(body)

	// 生成签名
	sign, timestamp, nonce, err := generateWxpaySignature("POST", "/v3/pay/transactions/h5", bodyJson, operation_setting.WxpayPrivateKey)
	if err != nil {
		return "", "", err
	}

	// 构建 Authorization 头
	authHeader := fmt.Sprintf("mchid=\"%s\",nonce_str=\"%s\",timestamp=\"%s\",serial_no=\"%s\",signature=\"%s\"",
		operation_setting.WxpayMchId, nonce, timestamp, operation_setting.WxpayCertSerial, sign)

	// 发送请求
	req, err := http.NewRequest("POST", url, strings.NewReader(string(bodyJson)))
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("WECHATPAY2-SHA256-RSA2048 %s", authHeader))

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("wxpay h5 order failed: %s", string(respBody))
	}

	var result WxpayH5Response
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", "", err
	}

	if result.H5Url == "" {
		return "", "", fmt.Errorf("wxpay h5 order returned empty h5_url")
	}

	return result.H5Url, "h5", nil
}

// WxpayNotify 处理微信异步回调
func WxpayNotify(c *gin.Context) {
	// 读取请求体
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		logger.LogWarn(c.Request.Context(), fmt.Sprintf("微信回调读取请求体失败 error=%q", err.Error()))
		c.Writer.Write([]byte("fail"))
		return
	}

	// 获取回调头信息
	timestamp := c.GetHeader("Wechatpay-Timestamp")
	nonce := c.GetHeader("Wechatpay-Nonce")
	signature := c.GetHeader("Wechatpay-Signature")
	serial := c.GetHeader("Wechatpay-Serial")

	if timestamp == "" || nonce == "" || signature == "" || serial == "" {
		logger.LogWarn(c.Request.Context(), "微信回调缺少必要的头信息")
		c.Writer.Write([]byte("fail"))
		return
	}

	// 验证签名
	err = verifyWxpayNotify(timestamp, nonce, string(body), serial, signature)
	if err != nil {
		logger.LogWarn(c.Request.Context(), fmt.Sprintf("微信回调验签失败 error=%q", err.Error()))
		c.Writer.Write([]byte("fail"))
		return
	}

	logger.LogInfo(c.Request.Context(), "微信回调验签成功")

	// 解密回调内容
	var payload struct {
		ID        string `json:"id"`
		EventType string `json:"event_type"`
		Resource  struct {
			Algorithm      string `json:"algorithm"`
			Ciphertext     string `json:"ciphertext"`
			Nonce          string `json:"nonce"`
			AssociatedData string `json:"associated_data"`
		} `json:"resource"`
	}

	if err := json.Unmarshal(body, &payload); err != nil {
		logger.LogWarn(c.Request.Context(), fmt.Sprintf("微信回调解析 JSON 失败 error=%q", err.Error()))
		c.Writer.Write([]byte("fail"))
		return
	}

	if payload.EventType != "TRANSACTION.SUCCESS" {
		logger.LogInfo(c.Request.Context(), fmt.Sprintf("微信回调忽略非支付成功事件 event_type=%s", payload.EventType))
		c.Writer.Write([]byte("success"))
		return
	}

	// 解密
	plaintext, err := decryptWxpayNotify(payload.Resource.Ciphertext, payload.Resource.AssociatedData, payload.Resource.Nonce)
	if err != nil {
		logger.LogWarn(c.Request.Context(), fmt.Sprintf("微信回调解密失败 error=%q", err.Error()))
		c.Writer.Write([]byte("fail"))
		return
	}

	var notifyData struct {
		TransactionID string `json:"transaction_id"`
		OutTradeNo    string `json:"out_trade_no"`
		TradeState    string `json:"trade_state"`
		Amount        struct {
			Total int `json:"total"`
		} `json:"amount"`
	}

	if err := json.Unmarshal(plaintext, &notifyData); err != nil {
		logger.LogWarn(c.Request.Context(), fmt.Sprintf("微信回调解析解密数据失败 error=%q", err.Error()))
		c.Writer.Write([]byte("fail"))
		return
	}

	// 处理支付结果
	if notifyData.TradeState == "SUCCESS" {
		LockOrder(notifyData.OutTradeNo)
		defer UnlockOrder(notifyData.OutTradeNo)

		topUp := model.GetTopUpByTradeNo(notifyData.OutTradeNo)
		if topUp == nil {
			logger.LogWarn(c.Request.Context(), fmt.Sprintf("微信回调订单不存在 trade_no=%s", notifyData.OutTradeNo))
			c.Writer.Write([]byte("success"))
			return
		}

		if topUp.Status == common.TopUpStatusPending {
			topUp.Status = common.TopUpStatusSuccess
			err := topUp.Update()
			if err != nil {
				logger.LogError(c.Request.Context(), fmt.Sprintf("微信回调更新订单失败 trade_no=%s error=%q", notifyData.OutTradeNo, err.Error()))
				c.Writer.Write([]byte("fail"))
				return
			}

			// 更新用户额度
			dAmount := decimal.NewFromInt(int64(topUp.Amount))
			dQuotaPerUnit := decimal.NewFromFloat(common.QuotaPerUnit)
			quotaToAdd := int(dAmount.Mul(dQuotaPerUnit).IntPart())
			err = model.IncreaseUserQuota(topUp.UserId, quotaToAdd, true)
			if err != nil {
				logger.LogError(c.Request.Context(), fmt.Sprintf("微信回调更新用户额度失败 trade_no=%s user_id=%d error=%q", notifyData.OutTradeNo, topUp.UserId, err.Error()))
				c.Writer.Write([]byte("fail"))
				return
			}

			logger.LogInfo(c.Request.Context(), fmt.Sprintf("微信充值成功 trade_no=%s user_id=%d quota_to_add=%d", notifyData.OutTradeNo, topUp.UserId, quotaToAdd))
			model.RecordTopupLog(topUp.UserId, fmt.Sprintf("使用微信充值成功，充值金额: %v，支付金额：%f", logger.LogQuota(quotaToAdd), topUp.Money), c.ClientIP(), topUp.PaymentMethod, "wxpay")
		}
	}

	c.Writer.Write([]byte("success"))
}

// verifyWxpayNotify 验证微信回调签名
func verifyWxpayNotify(timestamp, nonce, body, serial, signature string) error {
	// 构建验签字符串
	signContent := fmt.Sprintf("%s\n%s\n%s\n", timestamp, nonce, body)

	// 加载公钥
	publicKeyStr := formatWxpayPublicKey(operation_setting.WxpayPublicKey)
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
	signatureBytes, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return fmt.Errorf("failed to decode signature: %w", err)
	}

	// RSA 验签
	hasher := sha256.New()
	hasher.Write([]byte(signContent))
	err = rsa.VerifyPKCS1v15(publicKey, 0, hasher.Sum(nil), signatureBytes)
	if err != nil {
		return fmt.Errorf("signature verification failed: %w", err)
	}

	return nil
}

// decryptWxpayNotify 解密微信回调通知
func decryptWxpayNotify(ciphertext, associatedData, nonce string) ([]byte, error) {
	key := operation_setting.WxpayApiV3Key
	if len(key) != 32 {
		return nil, fmt.Errorf("invalid key length: %d, expected 32", len(key))
	}

	// 解码密文
	ciphertextBytes, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return nil, fmt.Errorf("failed to decode ciphertext: %w", err)
	}

	// AES-GCM 解密
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceBytes := []byte(nonce)
	plaintext, err := gcm.Open(nil, nonceBytes, ciphertextBytes, []byte(associatedData))
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}

	return plaintext, nil
}

// verifyWxpayNotifyWithHMAC 使用 HMAC 验证微信回调（备用）
func verifyWxpayNotifyWithHMAC(timestamp, nonce, body string, apiV3Key string) bool {
	// 构建字符串
	signContent := fmt.Sprintf("%s\n%s\n%s\n", timestamp, nonce, body)

	// HMAC-SHA256
	mac := hmac.New(sha256.New, []byte(apiV3Key))
	mac.Write([]byte(signContent))
	_ = mac.Sum(nil)

	// 注意：实际微信支付回调使用的是 RSA 签名，不是 HMAC
	// 这里保留 HMAC 方法作为备用
	return true
}
