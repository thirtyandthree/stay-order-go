package pay

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"gocode/first/config"
	"gocode/first/utils"
	"log"
	"net/http"
	"time"

	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/jsapi"
)

type Wpay struct {
	AppID      string `json:"appId"`
	MchID      string `json:"mchId"`
	AppSecret  string `json:"appSecret"`
	MchKey     string `json:"mchKey"`
	MchNumber  string `json:"mchNumber"`
	PrivateKey string `json:"privateKey"`
}
type PaymentData struct {
	OpenId string  `json:"openId"`
	Amount float64 `json:"amount"`
}

func HandleWeixinConfig(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
		config.DBConfig.Username, config.DBConfig.Password, config.DBConfig.Host, config.DBConfig.Port, config.DBConfig.Database))
	if err != nil {
		fmt.Println("Failed to connect to database:", err)
		return
	}
	defer db.Close()
	switch r.Method {
	case "POST":
		var wpay Wpay
		//

		if err := json.NewDecoder(r.Body).Decode(&wpay); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		fmt.Println(wpay.AppSecret)
		// 检查是否存在相同的app_id
		var exists int
		err = db.QueryRow("SELECT COUNT(*) FROM wpay WHERE AppID = ?", wpay.AppID).Scan(&exists)
		if err != nil && err != sql.ErrNoRows {
			http.Error(w, "数据库查询错误", http.StatusInternalServerError)
			return
		}
		if exists > 0 {
			// 存在相同的app_id，执行更新操作
			_, err = db.Exec("UPDATE wpay SET AppSecret = ?, Mchid = ?,Mchkey=?, MchNumber=?,PrivateKey=? ,AppID = ? LIMIT 1", wpay.AppSecret, wpay.MchID, wpay.MchKey, wpay.MchNumber, wpay.PrivateKey, wpay.AppID)
		} else {
			// 不存在相同的app_id，执行插入操作
			_, err = db.Exec("UPDATE wpay SET AppSecret = ?, Mchid = ?,Mchkey=?, MchNumber=?,PrivateKey=? ,AppID = ? LIMIT 1", wpay.AppSecret, wpay.MchID, wpay.MchKey, wpay.MchNumber, wpay.PrivateKey, wpay.AppID)
		}

		if err != nil {
			http.Error(w, fmt.Sprintf("数据库操作错误: %s", err.Error()), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "微信配置成功保存"})
	case "GET":
		// 查询最大id的配置
		var wpay Wpay
		err := db.QueryRow("SELECT AppID,AppSecret,MchId, MchKey,MchNumber,PrivateKey FROM wpay LIMIT 1").Scan(&wpay.AppID, &wpay.AppSecret, &wpay.MchID, &wpay.MchKey, &wpay.MchNumber, &wpay.PrivateKey)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "未找到配置", http.StatusNotFound)
			} else {
				http.Error(w, "数据库查询错误", http.StatusInternalServerError)
			}
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(wpay)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"message": "不允许的方法"})
	}

}
func Payment(w http.ResponseWriter, r *http.Request) {
	// 测试支付
	w.Header().Set("Content-Type", "application/json")
	var paymentData PaymentData
	err := json.NewDecoder(r.Body).Decode(&paymentData)
	fmt.Println(paymentData)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"msg":  "请求数据错误",
			"code": 400,
		})
		log.Printf("解析请求数据出错: %v", err)
		return
	}
	orderNumber := utils.GetOrderNo()
	resp, _, err := orderPaymentPrepayData(paymentData.OpenId, orderNumber, paymentData.Amount, "支付测试", "", time.Now().Unix()+60*30, "/payment/notify")
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"msg": "下单错误",
		})
		log.Fatalf("下单出错:%v", err.Error())
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"msg":  "生成信息成功",
		"data": resp,
		"code": 0,
	})
}

// 用户openid,订单编号,下单金额,备注,详情,超时时间,回调结果地址
func orderPaymentPrepayData(openId, tradeNo string, amount float64, body, attach string, timeExpire int64, notifyUrl string) (resp *jsapi.PrepayWithRequestPaymentResponse, result *core.APIResult, err error) {
	wechat := config.C.Wechat
	svc := jsapi.JsapiApiService{Client: utils.Client}

	// 得到prepay_id，以及调起支付所需的参数和签名
	duration := time.Second * time.Duration(timeExpire)
	timeData, _ := time.Parse(time.RFC3339, time.Unix(int64(duration), 0).Format(time.RFC3339))
	return svc.PrepayWithRequestPayment(context.TODO(),
		jsapi.PrepayRequest{
			Appid:       core.String(wechat.AppID),
			Mchid:       core.String(wechat.MchId),
			Description: core.String(body),
			OutTradeNo:  core.String(tradeNo),
			Attach:      core.String(attach),
			NotifyUrl:   core.String(wechat.Domain + notifyUrl),
			TimeExpire:  core.Time(timeData),
			Amount: &jsapi.Amount{
				Total:    core.Int64(int64(amount * 100)),
				Currency: core.String("CNY"),
			},
			Payer: &jsapi.Payer{
				Openid: core.String(openId),
			},
		},
	)
}
