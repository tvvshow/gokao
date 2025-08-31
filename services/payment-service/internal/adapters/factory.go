package adapters

import (
	"fmt"

	"github.com/gaokaohub/payment-service/internal/config"
)

// paymentAdapterFactory 支付适配器工厂实现
type paymentAdapterFactory struct {
	config config.PaymentConfig
}

// NewPaymentAdapterFactory 创建支付适配器工厂
func NewPaymentAdapterFactory(config config.PaymentConfig) PaymentAdapterFactory {
	return &paymentAdapterFactory{
		config: config,
	}
}

// GetAdapter 获取支付适配器
func (f *paymentAdapterFactory) GetAdapter(channel string) (PaymentAdapter, error) {
	switch channel {
	case "alipay":
		return NewStubAlipayAdapter(AdapterConfig{
			AppID:     "stub_alipay_app_id",
			NotifyURL: "http://localhost:8080/notify/alipay",
			IsProd:    false,
			Debug:     true,
		}), nil

	case "wechat":
		return NewStubWeChatAdapter(AdapterConfig{
			AppID:     "stub_wechat_app_id",
			MchID:     "stub_wechat_mch_id",
			APIKey:    "stub_wechat_api_key",
			NotifyURL: "http://localhost:8080/notify/wechat",
			IsProd:    false,
			Debug:     true,
		}), nil

	case "qq":
		return NewStubQQAdapter(AdapterConfig{
			AppID:     "stub_qq_app_id",
			MchID:     "stub_qq_mch_id",
			APIKey:    "stub_qq_api_key",
			NotifyURL: "http://localhost:8080/notify/qq",
			IsProd:    false,
			Debug:     true,
		}), nil

	case "unionpay":
		return NewStubUnionPayAdapter(AdapterConfig{
			AppID:     "stub_unionpay_merchant_id",
			NotifyURL: "http://localhost:8080/notify/unionpay",
			IsProd:    false,
			Debug:     true,
		}), nil

	default:
		return nil, fmt.Errorf("unsupported payment channel: %s", channel)
	}
}
