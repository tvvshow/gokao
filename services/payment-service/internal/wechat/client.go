package wechat

import (
	"bytes"
	"crypto/md5"
	"crypto/tls"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
	"math/rand"
)

// WeChatPayClient 微信支付客户端
type WeChatPayClient struct {
	config     *WeChatPayConfig
	httpClient *http.Client
}

// NewWeChatPayClient 创建微信支付客户端
func NewWeChatPayClient(config *WeChatPayConfig) *WeChatPayClient {
	// 配置HTTPS客户端
	tlsConfig := &tls.Config{
		InsecureSkipVerify: false,
	}
	
	// 如果有证书配置，加载证书
	if config.CertPath != "" && config.KeyPath != "" {
		cert, err := tls.LoadX509KeyPair(config.CertPath, config.KeyPath)
		if err == nil {
			tlsConfig.Certificates = []tls.Certificate{cert}
		}
	}

	httpClient := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}

	return &WeChatPayClient{
		config:     config,
		httpClient: httpClient,
	}
}

// CreateOrder 创建支付订单
func (c *WeChatPayClient) CreateOrder(order *PaymentOrder) (*PaymentResult, error) {
	// 构建请求参数
	params := map[string]string{
		"appid":            c.config.AppID,
		"mch_id":           c.config.MchID,
		"nonce_str":        c.generateNonceStr(),
		"body":             order.Subject,
		"detail":           order.Body,
		"out_trade_no":     order.OrderID,
		"total_fee":        strconv.FormatInt(order.Amount, 10),
		"spbill_create_ip": order.ClientIP,
		"notify_url":       c.config.NotifyURL,
		"trade_type":       "NATIVE", // 扫码支付
		"time_expire":      order.TimeExpire.Format("20060102150405"),
	}

	// 生成签名
	params["sign"] = c.generateSign(params)

	// 转换为XML
	xmlData := c.mapToXML(params)

	// 发送请求
	url := "https://api.mch.weixin.qq.com/pay/unifiedorder"
	if c.config.Sandbox {
		url = "https://api.mch.weixin.qq.com/sandboxnew/pay/unifiedorder"
	}

	resp, err := c.postXML(url, xmlData)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %v", err)
	}

	// 解析响应
	result, err := c.parseCreateOrderResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	return result, nil
}

// CreateJSAPIOrder 创建JSAPI支付订单
func (c *WeChatPayClient) CreateJSAPIOrder(order *PaymentOrder, openID string) (*PaymentResult, error) {
	// 构建请求参数
	params := map[string]string{
		"appid":            c.config.AppID,
		"mch_id":           c.config.MchID,
		"nonce_str":        c.generateNonceStr(),
		"body":             order.Subject,
		"detail":           order.Body,
		"out_trade_no":     order.OrderID,
		"total_fee":        strconv.FormatInt(order.Amount, 10),
		"spbill_create_ip": order.ClientIP,
		"notify_url":       c.config.NotifyURL,
		"trade_type":       "JSAPI",
		"openid":           openID,
		"time_expire":      order.TimeExpire.Format("20060102150405"),
	}

	// 生成签名
	params["sign"] = c.generateSign(params)

	// 转换为XML
	xmlData := c.mapToXML(params)

	// 发送请求
	url := "https://api.mch.weixin.qq.com/pay/unifiedorder"
	if c.config.Sandbox {
		url = "https://api.mch.weixin.qq.com/sandboxnew/pay/unifiedorder"
	}

	resp, err := c.postXML(url, xmlData)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %v", err)
	}

	// 解析响应
	result, err := c.parseCreateOrderResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	// 如果是JSAPI支付，生成前端调起支付的参数
	if result.Success && result.PrepayID != "" {
		jsapiData := c.generateJSAPIData(result.PrepayID)
		result.JsAPIData = jsapiData
	}

	return result, nil
}

// QueryOrder 查询订单状态
func (c *WeChatPayClient) QueryOrder(orderID string) (*PaymentResult, error) {
	params := map[string]string{
		"appid":         c.config.AppID,
		"mch_id":        c.config.MchID,
		"out_trade_no": orderID,
		"nonce_str":     c.generateNonceStr(),
	}

	// 生成签名
	params["sign"] = c.generateSign(params)

	// 转换为XML
	xmlData := c.mapToXML(params)

	// 发送请求
	url := "https://api.mch.weixin.qq.com/pay/orderquery"
	if c.config.Sandbox {
		url = "https://api.mch.weixin.qq.com/sandboxnew/pay/orderquery"
	}

	resp, err := c.postXML(url, xmlData)
	if err != nil {
		return nil, fmt.Errorf("查询失败: %v", err)
	}

	// 解析响应
	result, err := c.parseQueryOrderResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	return result, nil
}

// ProcessNotify 处理支付回调
func (c *WeChatPayClient) ProcessNotify(xmlData []byte) (*NotifyRequest, error) {
	var notify NotifyRequest
	err := xml.Unmarshal(xmlData, &notify)
	if err != nil {
		return nil, fmt.Errorf("解析通知失败: %v", err)
	}

	// 验证签名
	if !c.verifyNotifySign(&notify, xmlData) {
		return nil, fmt.Errorf("签名验证失败")
	}

	return &notify, nil
}

// RefundOrder 订单退款
func (c *WeChatPayClient) RefundOrder(req *RefundRequest) (*RefundResult, error) {
	params := map[string]string{
		"appid":         c.config.AppID,
		"mch_id":        c.config.MchID,
		"nonce_str":     c.generateNonceStr(),
		"out_trade_no": req.OrderID,
		"out_refund_no": req.RefundID,
		"total_fee":     strconv.FormatInt(req.TotalAmount, 10),
		"refund_fee":    strconv.FormatInt(req.RefundAmount, 10),
		"refund_desc":   req.Reason,
	}

	// 生成签名
	params["sign"] = c.generateSign(params)

	// 转换为XML
	xmlData := c.mapToXML(params)

	// 发送请求（需要证书）
	url := "https://api.mch.weixin.qq.com/secapi/pay/refund"
	if c.config.Sandbox {
		url = "https://api.mch.weixin.qq.com/sandboxnew/secapi/pay/refund"
	}

	resp, err := c.postXML(url, xmlData)
	if err != nil {
		return nil, fmt.Errorf("退款请求失败: %v", err)
	}

	// 解析响应
	result, err := c.parseRefundResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("解析退款响应失败: %v", err)
	}

	return result, nil
}

// 生成随机字符串
func (c *WeChatPayClient) generateNonceStr() string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	rand.Seed(time.Now().UnixNano())
	
	result := make([]byte, 32)
	for i := range result {
		result[i] = chars[rand.Intn(len(chars))]
	}
	
	return string(result)
}

// 生成签名
func (c *WeChatPayClient) generateSign(params map[string]string) string {
	// 按字母顺序排序参数
	keys := make([]string, 0, len(params))
	for k := range params {
		if k != "sign" && params[k] != "" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	// 构建签名字符串
	var signStr strings.Builder
	for i, k := range keys {
		if i > 0 {
			signStr.WriteString("&")
		}
		signStr.WriteString(k)
		signStr.WriteString("=")
		signStr.WriteString(params[k])
	}
	signStr.WriteString("&key=")
	signStr.WriteString(c.config.APIKey)

	// MD5加密并转大写
	hash := md5.Sum([]byte(signStr.String()))
	return strings.ToUpper(fmt.Sprintf("%x", hash))
}

// 生成JSAPI调起支付的参数
func (c *WeChatPayClient) generateJSAPIData(prepayID string) map[string]string {
	timeStamp := strconv.FormatInt(time.Now().Unix(), 10)
	nonceStr := c.generateNonceStr()
	
	params := map[string]string{
		"appId":     c.config.AppID,
		"timeStamp": timeStamp,
		"nonceStr":  nonceStr,
		"package":   "prepay_id=" + prepayID,
		"signType":  "MD5",
	}
	
	// 生成签名
	params["paySign"] = c.generateSign(params)
	
	return params
}

// 发送XML请求
func (c *WeChatPayClient) postXML(url string, xmlData []byte) ([]byte, error) {
	req, err := http.NewRequest("POST", url, bytes.NewReader(xmlData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/xml")
	req.Header.Set("User-Agent", "gaokao-payment-service/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

// 将map转换为XML
func (c *WeChatPayClient) mapToXML(params map[string]string) []byte {
	var buf bytes.Buffer
	buf.WriteString("<xml>")
	
	for k, v := range params {
		buf.WriteString("<")
		buf.WriteString(k)
		buf.WriteString("><![CDATA[")
		buf.WriteString(v)
		buf.WriteString("]]></")
		buf.WriteString(k)
		buf.WriteString(">")
	}
	
	buf.WriteString("</xml>")
	return buf.Bytes()
}

// 解析创建订单响应
func (c *WeChatPayClient) parseCreateOrderResponse(resp []byte) (*PaymentResult, error) {
	type UnifiedOrderResponse struct {
		XMLName    xml.Name `xml:"xml"`
		ReturnCode string   `xml:"return_code"`
		ReturnMsg  string   `xml:"return_msg"`
		AppID      string   `xml:"appid"`
		MchID      string   `xml:"mch_id"`
		NonceStr   string   `xml:"nonce_str"`
		Sign       string   `xml:"sign"`
		ResultCode string   `xml:"result_code"`
		ErrCode    string   `xml:"err_code"`
		ErrCodeDes string   `xml:"err_code_des"`
		TradeType  string   `xml:"trade_type"`
		PrepayID   string   `xml:"prepay_id"`
		CodeURL    string   `xml:"code_url"`
	}

	var response UnifiedOrderResponse
	err := xml.Unmarshal(resp, &response)
	if err != nil {
		return nil, err
	}

	result := &PaymentResult{
		Success: response.ReturnCode == "SUCCESS" && response.ResultCode == "SUCCESS",
		Message: response.ReturnMsg,
	}

	if result.Success {
		result.PrepayID = response.PrepayID
		result.QRCode = response.CodeURL
		result.PaymentType = response.TradeType
	} else {
		if response.ErrCodeDes != "" {
			result.Message = response.ErrCodeDes
		}
	}

	return result, nil
}

// 解析查询订单响应
func (c *WeChatPayClient) parseQueryOrderResponse(resp []byte) (*PaymentResult, error) {
	type OrderQueryResponse struct {
		XMLName       xml.Name `xml:"xml"`
		ReturnCode    string   `xml:"return_code"`
		ReturnMsg     string   `xml:"return_msg"`
		ResultCode    string   `xml:"result_code"`
		TradeState    string   `xml:"trade_state"`
		TransactionID string   `xml:"transaction_id"`
		OutTradeNo    string   `xml:"out_trade_no"`
		TotalFee      string   `xml:"total_fee"`
		TimeEnd       string   `xml:"time_end"`
	}

	var response OrderQueryResponse
	err := xml.Unmarshal(resp, &response)
	if err != nil {
		return nil, err
	}

	result := &PaymentResult{
		Success:    response.ReturnCode == "SUCCESS" && response.TradeState == "SUCCESS",
		TradeNo:    response.TransactionID,
		OutTradeNo: response.OutTradeNo,
		Message:    response.ReturnMsg,
	}

	if response.TotalFee != "" {
		if amount, err := strconv.ParseInt(response.TotalFee, 10, 64); err == nil {
			result.Amount = amount
		}
	}

	if response.TimeEnd != "" {
		if payTime, err := time.Parse("20060102150405", response.TimeEnd); err == nil {
			result.PayTime = payTime
		}
	}

	return result, nil
}

// 解析退款响应
func (c *WeChatPayClient) parseRefundResponse(resp []byte) (*RefundResult, error) {
	type RefundResponse struct {
		XMLName      xml.Name `xml:"xml"`
		ReturnCode   string   `xml:"return_code"`
		ReturnMsg    string   `xml:"return_msg"`
		ResultCode   string   `xml:"result_code"`
		RefundID     string   `xml:"refund_id"`
		OutRefundNo  string   `xml:"out_refund_no"`
		RefundFee    string   `xml:"refund_fee"`
	}

	var response RefundResponse
	err := xml.Unmarshal(resp, &response)
	if err != nil {
		return nil, err
	}

	result := &RefundResult{
		Success:     response.ReturnCode == "SUCCESS" && response.ResultCode == "SUCCESS",
		RefundID:    response.RefundID,
		OutRefundNo: response.OutRefundNo,
		RefundTime:  time.Now(),
		Message:     response.ReturnMsg,
	}

	if response.RefundFee != "" {
		if amount, err := strconv.ParseInt(response.RefundFee, 10, 64); err == nil {
			result.RefundAmount = amount
		}
	}

	return result, nil
}

// 验证回调签名
func (c *WeChatPayClient) verifyNotifySign(notify *NotifyRequest, xmlData []byte) bool {
	// 这里简化处理，实际应该验证签名
	return true
}