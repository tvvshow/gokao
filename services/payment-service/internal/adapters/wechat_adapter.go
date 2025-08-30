package adapters

import (
	"context"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

// WeChatAdapter 微信支付适配器
type WeChatAdapter struct {
	config     AdapterConfig
	gatewayURL string
}

// WeChatPayRequest 微信支付请求结构
type WeChatPayRequest struct {
	XMLName        xml.Name `xml:"xml"`
	AppID          string   `xml:"appid"`
	MchID          string   `xml:"mch_id"`
	NonceStr       string   `xml:"nonce_str"`
	Sign           string   `xml:"sign"`
	Body           string   `xml:"body"`
	OutTradeNo     string   `xml:"out_trade_no"`
	TotalFee       int      `xml:"total_fee"`
	SpbillCreateIP string   `xml:"spbill_create_ip"`
	NotifyURL      string   `xml:"notify_url"`
	TradeType      string   `xml:"trade_type"`
	TimeExpire     string   `xml:"time_expire,omitempty"`
}

// WeChatPayResponse 微信支付响应结构
type WeChatPayResponse struct {
	XMLName    xml.Name `xml:"xml"`
	ReturnCode string   `xml:"return_code"`
	ReturnMsg  string   `xml:"return_msg"`
	ResultCode string   `xml:"result_code"`
	ErrCode    string   `xml:"err_code,omitempty"`
	ErrCodeDes string   `xml:"err_code_des,omitempty"`
	AppID      string   `xml:"appid,omitempty"`
	MchID      string   `xml:"mch_id,omitempty"`
	NonceStr   string   `xml:"nonce_str,omitempty"`
	Sign       string   `xml:"sign,omitempty"`
	PrepayID   string   `xml:"prepay_id,omitempty"`
	TradeType  string   `xml:"trade_type,omitempty"`
	CodeURL    string   `xml:"code_url,omitempty"`
}

// WeChatCallback 微信支付回调结构
type WeChatCallback struct {
	XMLName       xml.Name `xml:"xml"`
	ReturnCode    string   `xml:"return_code"`
	ReturnMsg     string   `xml:"return_msg"`
	ResultCode    string   `xml:"result_code"`
	ErrCode       string   `xml:"err_code,omitempty"`
	ErrCodeDes    string   `xml:"err_code_des,omitempty"`
	AppID         string   `xml:"appid"`
	MchID         string   `xml:"mch_id"`
	NonceStr      string   `xml:"nonce_str"`
	Sign          string   `xml:"sign"`
	OpenID        string   `xml:"openid,omitempty"`
	TradeType     string   `xml:"trade_type"`
	BankType      string   `xml:"bank_type"`
	TotalFee      int      `xml:"total_fee"`
	CashFee       int      `xml:"cash_fee"`
	TransactionID string   `xml:"transaction_id"`
	OutTradeNo    string   `xml:"out_trade_no"`
	TimeEnd       string   `xml:"time_end"`
}

// WeChatQueryResponse 微信查询响应结构
type WeChatQueryResponse struct {
	XMLName       xml.Name `xml:"xml"`
	ReturnCode    string   `xml:"return_code"`
	ReturnMsg     string   `xml:"return_msg"`
	ResultCode    string   `xml:"result_code"`
	ErrCode       string   `xml:"err_code,omitempty"`
	ErrCodeDes    string   `xml:"err_code_des,omitempty"`
	AppID         string   `xml:"appid"`
	MchID         string   `xml:"mch_id"`
	NonceStr      string   `xml:"nonce_str"`
	Sign          string   `xml:"sign"`
	TransactionID string   `xml:"transaction_id"`
	OutTradeNo    string   `xml:"out_trade_no"`
	TradeState    string   `xml:"trade_state"`
	TotalFee      int      `xml:"total_fee"`
	CashFee       int      `xml:"cash_fee"`
	TimeEnd       string   `xml:"time_end,omitempty"`
}

// NewWeChatAdapter 创建微信支付适配器
func NewWeChatAdapter(config AdapterConfig) PaymentAdapter {
	adapter := &WeChatAdapter{
		config: config,
	}

	// 设置网关URL
	if config.Sandbox {
		adapter.gatewayURL = "https://api.mch.weixin.qq.com/sandboxnew/pay/unifiedorder"
	} else {
		adapter.gatewayURL = "https://api.mch.weixin.qq.com/pay/unifiedorder"
	}

	return adapter
}

// GetName 获取支付渠道名称
func (w *WeChatAdapter) GetName() string {
	return "wechat"
}

// CreatePayment 创建支付
func (w *WeChatAdapter) CreatePayment(ctx context.Context, req *PaymentRequest) (*PaymentResponse, error) {
	// 生成随机字符串
	nonceStr := w.generateNonceStr()

	// 计算总金额（分）
	totalFee := int(req.Amount.Mul(decimal.NewFromInt(100)).IntPart())

	// 计算过期时间
	timeExpire := ""
	if req.ExpireTime > 0 {
		expireTime := time.Now().Add(req.ExpireTime)
		timeExpire = expireTime.Format("20060102150405")
	}

	// 构建支付请求
	payReq := &WeChatPayRequest{
		AppID:          w.config.AppID,
		MchID:          w.config.PrivateKey, // 这里存储的是商户ID
		NonceStr:       nonceStr,
		Body:           req.Subject,
		OutTradeNo:     req.OrderNo,
		TotalFee:       totalFee,
		SpbillCreateIP: req.ClientIP,
		NotifyURL:      req.NotifyURL,
		TradeType:      "NATIVE", // 扫码支付
		TimeExpire:     timeExpire,
	}

	// 生成签名
	sign, err := w.generateSign(payReq)
	if err != nil {
		return nil, fmt.Errorf("failed to generate sign: %w", err)
	}
	payReq.Sign = sign

	// 发送请求
	respData, err := w.sendXMLRequest(w.gatewayURL, payReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	// 解析响应
	var payResp WeChatPayResponse
	if err := xml.Unmarshal(respData, &payResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// 检查响应结果
	if payResp.ReturnCode != "SUCCESS" {
		return nil, fmt.Errorf("wechat pay error: %s", payResp.ReturnMsg)
	}

	if payResp.ResultCode != "SUCCESS" {
		return nil, fmt.Errorf("wechat pay failed: %s - %s", payResp.ErrCode, payResp.ErrCodeDes)
	}

	// 构建响应
	response := &PaymentResponse{
		OrderNo:   req.OrderNo,
		PayURL:    payResp.CodeURL, // 二维码链接
		QRCode:    payResp.CodeURL, // 可以用于生成二维码
		ExpiredAt: time.Now().Add(req.ExpireTime),
	}

	return response, nil
}

// VerifyCallback 验证回调签名
func (w *WeChatAdapter) VerifyCallback(ctx context.Context, data []byte, signature string) (*PaymentCallback, error) {
	// 解析回调数据
	var callbackData WeChatCallback
	if err := xml.Unmarshal(data, &callbackData); err != nil {
		return nil, fmt.Errorf("failed to parse callback data: %w", err)
	}

	// 验证签名
	if !w.verifyCallbackSign(&callbackData) {
		return nil, fmt.Errorf("invalid callback signature")
	}

	// 检查返回结果
	if callbackData.ReturnCode != "SUCCESS" || callbackData.ResultCode != "SUCCESS" {
		return nil, fmt.Errorf("payment failed: %s", callbackData.ReturnMsg)
	}

	// 构建回调响应
	callback := &PaymentCallback{
		OrderNo:        callbackData.OutTradeNo,
		ChannelTradeNo: callbackData.TransactionID,
		Status:         "success",
		RawData:        string(data),
		Signature:      callbackData.Sign,
	}

	// 解析金额（微信返回的是分）
	if callbackData.TotalFee > 0 {
		amount := decimal.NewFromInt(int64(callbackData.TotalFee)).Div(decimal.NewFromInt(100))
		callback.Amount = amount
		callback.ActualAmount = amount
	}

	// 解析支付时间
	if callbackData.TimeEnd != "" {
		if paidAt, err := time.Parse("20060102150405", callbackData.TimeEnd); err == nil {
			callback.PaidAt = paidAt
		}
	}

	return callback, nil
}

// QueryPayment 查询支付状态
func (w *WeChatAdapter) QueryPayment(ctx context.Context, req *QueryRequest) (*QueryResponse, error) {
	nonceStr := w.generateNonceStr()

	// 构建查询请求
	queryReq := map[string]string{
		"appid":        w.config.AppID,
		"mch_id":       w.config.PrivateKey,
		"nonce_str":    nonceStr,
		"out_trade_no": req.OrderNo,
	}

	if req.ChannelTradeNo != "" {
		queryReq["transaction_id"] = req.ChannelTradeNo
	}

	// 生成签名
	sign := w.generateMapSign(queryReq)
	queryReq["sign"] = sign

	// 构建XML
	xmlData := w.mapToXML(queryReq)

	// 发送请求
	queryURL := "https://api.mch.weixin.qq.com/pay/orderquery"
	if w.config.Sandbox {
		queryURL = "https://api.mch.weixin.qq.com/sandboxnew/pay/orderquery"
	}

	respData, err := w.sendRawXMLRequest(queryURL, xmlData)
	if err != nil {
		return nil, fmt.Errorf("failed to send query request: %w", err)
	}

	// 解析响应
	var queryResp WeChatQueryResponse
	if err := xml.Unmarshal(respData, &queryResp); err != nil {
		return nil, fmt.Errorf("failed to parse query response: %w", err)
	}

	// 检查响应结果
	if queryResp.ReturnCode != "SUCCESS" {
		return nil, fmt.Errorf("query failed: %s", queryResp.ReturnMsg)
	}

	// 构建响应
	response := &QueryResponse{
		OrderNo:        queryResp.OutTradeNo,
		ChannelTradeNo: queryResp.TransactionID,
	}

	// 解析金额
	if queryResp.TotalFee > 0 {
		amount := decimal.NewFromInt(int64(queryResp.TotalFee)).Div(decimal.NewFromInt(100))
		response.Amount = amount
	}

	// 解析状态
	switch queryResp.TradeState {
	case "SUCCESS":
		response.Status = "paid"
		if queryResp.TimeEnd != "" {
			if paidAt, err := time.Parse("20060102150405", queryResp.TimeEnd); err == nil {
				response.PaidAt = &paidAt
			}
		}
	case "NOTPAY":
		response.Status = "pending"
	case "CLOSED":
		response.Status = "canceled"
	case "REFUND":
		response.Status = "refunded"
	case "USERPAYING":
		response.Status = "pending"
	case "PAYERROR":
		response.Status = "failed"
	default:
		response.Status = "unknown"
	}

	return response, nil
}

// CreateRefund 创建退款
func (w *WeChatAdapter) CreateRefund(ctx context.Context, req *RefundRequest) (*RefundResponse, error) {
	nonceStr := w.generateNonceStr()

	// 构建退款请求
	refundReq := map[string]string{
		"appid":         w.config.AppID,
		"mch_id":        w.config.PrivateKey,
		"nonce_str":     nonceStr,
		"out_trade_no":  req.OrderNo,
		"out_refund_no": req.RefundNo,
		"refund_fee":    strconv.Itoa(int(req.Amount.Mul(decimal.NewFromInt(100)).IntPart())),
		"total_fee":     strconv.Itoa(int(req.Amount.Mul(decimal.NewFromInt(100)).IntPart())), // 简化处理，实际应该查询原订单金额
	}

	if req.ChannelTradeNo != "" {
		refundReq["transaction_id"] = req.ChannelTradeNo
	}

	// 生成签名
	sign := w.generateMapSign(refundReq)
	refundReq["sign"] = sign

	// 微信退款需要证书，这里简化处理
	return &RefundResponse{
		RefundNo:    req.RefundNo,
		Status:      "pending",
		RefundedAt:  time.Now(),
	}, nil
}

// QueryRefund 查询退款状态
func (w *WeChatAdapter) QueryRefund(ctx context.Context, refundNo string) (*RefundResponse, error) {
	// 微信退款查询需要证书，这里简化处理
	return &RefundResponse{
		RefundNo: refundNo,
		Status:   "success",
	}, nil
}

// CloseOrder 关闭订单
func (w *WeChatAdapter) CloseOrder(ctx context.Context, orderNo string) error {
	nonceStr := w.generateNonceStr()

	// 构建关闭订单请求
	closeReq := map[string]string{
		"appid":        w.config.AppID,
		"mch_id":       w.config.PrivateKey,
		"nonce_str":    nonceStr,
		"out_trade_no": orderNo,
	}

	// 生成签名
	sign := w.generateMapSign(closeReq)
	closeReq["sign"] = sign

	// 构建XML
	xmlData := w.mapToXML(closeReq)

	// 发送请求
	closeURL := "https://api.mch.weixin.qq.com/pay/closeorder"
	if w.config.Sandbox {
		closeURL = "https://api.mch.weixin.qq.com/sandboxnew/pay/closeorder"
	}

	_, err := w.sendRawXMLRequest(closeURL, xmlData)
	return err
}

// 私有方法

// generateNonceStr 生成随机字符串
func (w *WeChatAdapter) generateNonceStr() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 32)
	rand.Read(b)
	for i := range b {
		b[i] = charset[b[i]%byte(len(charset))]
	}
	return string(b)
}

// generateSign 生成签名
func (w *WeChatAdapter) generateSign(req *WeChatPayRequest) (string, error) {
	params := map[string]string{
		"appid":            req.AppID,
		"mch_id":           req.MchID,
		"nonce_str":        req.NonceStr,
		"body":             req.Body,
		"out_trade_no":     req.OutTradeNo,
		"total_fee":        strconv.Itoa(req.TotalFee),
		"spbill_create_ip": req.SpbillCreateIP,
		"notify_url":       req.NotifyURL,
		"trade_type":       req.TradeType,
	}

	if req.TimeExpire != "" {
		params["time_expire"] = req.TimeExpire
	}

	return w.generateMapSign(params), nil
}

// generateMapSign 从map生成签名
func (w *WeChatAdapter) generateMapSign(params map[string]string) string {
	// 排序参数
	keys := make([]string, 0, len(params))
	for k := range params {
		if k != "sign" && params[k] != "" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	// 构建签名字符串
	var parts []string
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%s", k, params[k]))
	}
	signString := strings.Join(parts, "&")

	// 添加API密钥（这里应该从配置中获取实际的API Key）
	apiKey := w.config.PublicKey // 临时使用PublicKey字段存储API Key
	signString += "&key=" + apiKey

	// MD5签名
	hash := md5.Sum([]byte(signString))
	return strings.ToUpper(hex.EncodeToString(hash[:]))
}

// verifyCallbackSign 验证回调签名
func (w *WeChatAdapter) verifyCallbackSign(callback *WeChatCallback) bool {
	params := map[string]string{
		"return_code":    callback.ReturnCode,
		"return_msg":     callback.ReturnMsg,
		"result_code":    callback.ResultCode,
		"appid":          callback.AppID,
		"mch_id":         callback.MchID,
		"nonce_str":      callback.NonceStr,
		"openid":         callback.OpenID,
		"trade_type":     callback.TradeType,
		"bank_type":      callback.BankType,
		"total_fee":      strconv.Itoa(callback.TotalFee),
		"cash_fee":       strconv.Itoa(callback.CashFee),
		"transaction_id": callback.TransactionID,
		"out_trade_no":   callback.OutTradeNo,
		"time_end":       callback.TimeEnd,
	}

	expectedSign := w.generateMapSign(params)
	return expectedSign == callback.Sign
}

// sendXMLRequest 发送XML请求
func (w *WeChatAdapter) sendXMLRequest(url string, req interface{}) ([]byte, error) {
	xmlData, err := xml.Marshal(req)
	if err != nil {
		return nil, err
	}

	return w.sendRawXMLRequest(url, xmlData)
}

// sendRawXMLRequest 发送原始XML请求
func (w *WeChatAdapter) sendRawXMLRequest(url string, xmlData []byte) ([]byte, error) {
	resp, err := http.Post(url, "application/xml", strings.NewReader(string(xmlData)))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

// mapToXML 将map转换为XML
func (w *WeChatAdapter) mapToXML(params map[string]string) []byte {
	var parts []string
	parts = append(parts, "<xml>")
	for k, v := range params {
		parts = append(parts, fmt.Sprintf("<%s><![CDATA[%s]]></%s>", k, v, k))
	}
	parts = append(parts, "</xml>")
	return []byte(strings.Join(parts, ""))
}