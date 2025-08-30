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
		return NewAlipayAdapter(AdapterConfig{
			AppID:      f.config.Alipay.AppID,
			PrivateKey: f.config.Alipay.PrivateKey,
			PublicKey:  f.config.Alipay.PublicKey,
			NotifyURL:  f.config.Alipay.NotifyURL,
			ReturnURL:  f.config.Alipay.ReturnURL,
			SignType:   f.config.Alipay.SignType,
			Sandbox:    f.config.Alipay.Sandbox,
		}), nil

	case "wechat":
		return NewWeChatAdapter(AdapterConfig{
			AppID:      f.config.WeChat.AppID,
			PrivateKey: f.config.WeChat.APIKey,
			NotifyURL:  f.config.WeChat.NotifyURL,
			Sandbox:    f.config.WeChat.Sandbox,
		}), nil

	case "unionpay":
		return NewUnionPayAdapter(AdapterConfig{
			AppID:      f.config.UnionPay.MerchantID,
			PrivateKey: f.config.UnionPay.PrivateKey,
			PublicKey:  f.config.UnionPay.PublicKey,
			NotifyURL:  f.config.UnionPay.NotifyURL,
			ReturnURL:  f.config.UnionPay.ReturnURL,
			Sandbox:    f.config.UnionPay.Sandbox,
		}), nil

	default:
		return nil, fmt.Errorf("unsupported payment channel: %s", channel)
	}
}