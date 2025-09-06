
# 支付 SDK 集成指南 (Alipay / WxPay)

本文档为新手开发者提供完整的支付 SDK 接入指南，涵盖准备工作、集成步骤、测试与上线。

---

## 一、准备工作

1. **注册并认证开发者账号**
   - 支付宝开放平台：https://open.alipay.com
   - 微信支付商户平台：https://pay.weixin.qq.com

2. **申请支付能力**
   - 提交营业执照、银行账户等企业资质。
   - 完成签约后，获取商户号（MchID）、应用 ID（AppID）、API 密钥。

3. **下载官方 SDK**
   - Alipay SDK: [Java / Python / PHP / Node.js]
   - WxPay SDK: [Java / Python / PHP / Node.js]

---

## 二、集成步骤

### 1. 安装 SDK

以 Python 为例：

```bash
pip install alipay-sdk-python wxpay-sdk-python
```

### 2. 配置参数

配置文件示例：

```yaml
alipay:
  app_id: "your-app-id"
  private_key_path: "./keys/alipay_private.pem"
  alipay_public_key_path: "./keys/alipay_public.pem"

wxpay:
  mch_id: "your-mch-id"
  app_id: "your-app-id"
  api_key: "your-api-key"
  cert_path: "./keys/apiclient_cert.pem"
  key_path: "./keys/apiclient_key.pem"
```

### 3. 调用支付接口

#### Alipay 示例

```python
from alipay import AliPay

alipay = AliPay(
    appid="your-app-id",
    app_notify_url="https://yourdomain.com/notify",
    app_private_key_string=open("./keys/alipay_private.pem").read(),
    alipay_public_key_string=open("./keys/alipay_public.pem").read(),
    sign_type="RSA2",
    debug=False
)

order_string = alipay.api_alipay_trade_page_pay(
    out_trade_no="20250905",
    total_amount=100.0,
    subject="测试订单",
    return_url="https://yourdomain.com/return",
    notify_url="https://yourdomain.com/notify"
)
```

#### WxPay 示例

```python
from wxpay import WeChatPay

wxpay = WeChatPay(
    mch_id="your-mch-id",
    appid="your-app-id",
    api_key="your-api-key",
    mch_cert="./keys/apiclient_cert.pem",
    mch_key="./keys/apiclient_key.pem"
)

order = wxpay.order.create(
    trade_type="JSAPI",
    body="测试订单",
    out_trade_no="20250905",
    total_fee=100,
    notify_url="https://yourdomain.com/notify",
    openid="用户openid"
)
```

---

## 三、测试与上线

1. 使用 **沙箱环境** 进行支付测试。
2. 验证 **回调通知** 是否能正确处理。
3. 确认 **签名验证** 成功，避免伪造请求。
4. 上线前检查证书是否过期。

---

## 四、安全注意事项

- 私钥文件务必放置在安全目录，不要上传到 Git。
- 所有支付请求必须使用 HTTPS。
- 建立日志审计，监控异常支付行为。
- 定期更换 API 密钥和证书。

---

## 五、参考文档

- 支付宝开放平台文档：https://opendocs.alipay.com/
- 微信支付开发文档：https://pay.weixin.qq.com/wiki/doc/api/index.html
