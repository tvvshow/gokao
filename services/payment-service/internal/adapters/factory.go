package adapters

import (
	"fmt"

	"github.com/gaokaohub/gaokao/services/payment-service/internal/config"
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
		alipayConfig := NewAlipayConfigFromAdapterConfig(AdapterConfig{
			AppID:      f.config.Alipay.AppID,
			PrivateKey: f.config.Alipay.PrivateKey,
			PublicKey:  f.config.Alipay.PublicKey,
			NotifyURL:  f.config.Alipay.NotifyURL,
			ReturnURL:  f.config.Alipay.ReturnURL,
			IsProd:     !f.config.Alipay.Sandbox,
		})
		adapter, err := NewAlipayAdapter(alipayConfig)
		if err != nil {
			return nil, err
		}
		return adapter, nil

	case "wechat":
		wechatConfig := NewWechatPayConfigFromAdapterConfig(AdapterConfig{
			AppID:        f.config.WeChat.AppID,
			MchID:        f.config.WeChat.MchID,
			APIKey:       f.config.WeChat.APIKey,
			CertPath:     f.config.WeChat.CertPath,
			KeyPath:      f.config.WeChat.KeyPath,
			NotifyURL:    f.config.WeChat.NotifyURL,
			SerialNumber: "",
		})
		adapter, err := NewWechatPayAdapter(wechatConfig)
		if err != nil {
			return nil, err
		}
		return adapter, nil

	case "unionpay":
		adapter := NewUnionPayAdapter(AdapterConfig{
			AppID:      f.config.UnionPay.MerchantID,
			PrivateKey: f.config.UnionPay.PrivateKey,
			PublicKey:  f.config.UnionPay.PublicKey,
			NotifyURL:  f.config.UnionPay.NotifyURL,
			ReturnURL:  f.config.UnionPay.ReturnURL,
			IsProd:     !f.config.UnionPay.Sandbox,
		})
		return adapter, nil

	default:
		return nil, fmt.Errorf("unsupported payment channel: %s", channel)
	}
}
