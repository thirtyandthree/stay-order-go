package utils

import (
	"context"
	"gocode/first/config"
	"log"

	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/core/option"
	"github.com/wechatpay-apiv3/wechatpay-go/utils"
)

var (
	Client *core.Client = nil
)

func NewWeChatClient() *core.Client {
	mini := config.C.Wechat
	mchPrivateKey, err := utils.LoadPrivateKeyWithPath(mini.PrivateKey)

	if err != nil {
		log.Printf("加载私钥文件错误" + err.Error())
		return nil
	}

	ctx := context.Background()
	// 使用商户私钥等初始化 client，并使它具有自动定时获取微信支付平台证书的能力
	opts := []core.ClientOption{
		option.WithWechatPayAutoAuthCipher(mini.MchId, mini.MchNumber, mchPrivateKey, mini.MchKey),
	}
	client, err := core.NewClient(ctx, opts...)
	if err != nil {
		log.Printf("创建微信支付客户端失败,原因:%v", err)
	}

	return client
}
