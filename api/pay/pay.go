package pay

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"gocode/first/config"
	"log"
	"net/http"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/smartwalle/alipay/v3"
)

// AlipayConfig 用于存储支付宝配置信息的结构体
type AlipayConfig struct {
	AppID           string `json:"appId"`
	PrivateKey      string `json:"privateKey"`
	AlipayPublicKey string `json:"alipayPublicKey"`
}

// HandleAlipayConfig 处理支付宝配置的HTTP请求
func HandleAlipayConfig(w http.ResponseWriter, r *http.Request) {
	// 连接到数据库
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
		config.DBConfig.Username, config.DBConfig.Password, config.DBConfig.Host, config.DBConfig.Port, config.DBConfig.Database))
	if err != nil {
		fmt.Println("Failed to connect to database:", err)
		return
	}
	defer db.Close()

	switch r.Method {
	case "POST":
		var config AlipayConfig
		if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// 检查是否存在相同的app_id
		var exists int
		err = db.QueryRow("SELECT COUNT(*) FROM pay WHERE app_id = ?", config.AppID).Scan(&exists)
		if err != nil && err != sql.ErrNoRows {
			http.Error(w, "数据库查询错误", http.StatusInternalServerError)
			return
		}

		if exists > 0 {
			// 存在相同的app_id，执行更新操作
			_, err = db.Exec("UPDATE pay SET private_key = ?, alipay_public_key = ? WHERE app_id = ?", config.PrivateKey, config.AlipayPublicKey, config.AppID)
		} else {
			// 不存在相同的app_id，执行插入操作
			_, err = db.Exec("UPDATE pay SET private_key = ?, alipay_public_key = ?, app_id = ? WHERE id = 1 LIMIT 1", config.PrivateKey, config.AlipayPublicKey, config.AppID)
		}

		if err != nil {
			http.Error(w, "数据库操作错误", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "支付宝配置成功保存"})

	case "GET":
		// 查询最大id的配置
		var config AlipayConfig
		err := db.QueryRow("SELECT app_id, private_key, alipay_public_key FROM pay ORDER BY id ASC LIMIT 1").Scan(&config.AppID, &config.PrivateKey, &config.AlipayPublicKey)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "未找到配置", http.StatusNotFound)
			} else {
				http.Error(w, "数据库查询错误", http.StatusInternalServerError)
			}
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(config)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"message": "不允许的方法"})
	}
}

// 假设成功状态码为"10000"，请根据实际情况调整
const KSuccessCode = "10000"

func TestPayment(w http.ResponseWriter, r *http.Request) {
	var req struct {
		AppID           string  `json:"appId"`
		PrivateKey      string  `json:"privateKey"`
		AlipayPublicKey string  `json:"alipayPublicKey"`
		Amount          float64 `json:"amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "无效的请求", http.StatusBadRequest)
		return
	}

	// 将 false 改为 true 以使用沙箱模式
	client, err := alipay.New(req.AppID, req.PrivateKey, true)
	if err != nil {
		log.Printf("创建支付宝客户端失败: %v", err)
		http.Error(w, "创建支付宝客户端失败", http.StatusInternalServerError)
		return
	}
	client.LoadAliPayPublicKey(req.AlipayPublicKey)

	var p = alipay.TradePreCreate{}
	p.OutTradeNo = fmt.Sprintf("%d", time.Now().Unix())
	p.TotalAmount = fmt.Sprintf("%.2f", req.Amount)
	p.Subject = "测试订单"

	resp, err := client.TradePreCreate(p)
	if err != nil || resp.Code != KSuccessCode {
		log.Printf("生成支付宝支付二维码失败: %v, Code: %s", err, resp.Code)
		http.Error(w, "生成支付宝支付二维码失败", http.StatusInternalServerError)
		return
	}

	// 修改这里，添加 OutTradeNo 字段到响应中
	response := struct {
		QrCode     string `json:"qrCode"`
		OutTradeNo string `json:"outTradeNo"` // 添加这行
	}{
		QrCode:     resp.QRCode,
		OutTradeNo: p.OutTradeNo, // 这里返回订单号
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	go checkPaymentStatus(client, p.OutTradeNo)
}

var paymentSuccessful bool // 全局变量定义，无需在此处初始化为 false，Go 默认布尔值为 false
func PaymentStatusHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]bool{"paymentSuccessful": paymentSuccessful}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
func checkPaymentStatus(client *alipay.Client, outTradeNo string) {
	// 定义最大尝试查询次数
	const maxAttempts = 20
	attempt := 0
	paymentSuccessful = false
	for attempt < maxAttempts {
		resp, err := client.TradeQuery(alipay.TradeQuery{OutTradeNo: outTradeNo})
		if err != nil {
			log.Printf("查询支付状态失败: %v", err)
			break // 出错时结束循环
		}

		switch resp.TradeStatus {
		case alipay.TradeStatusSuccess:
			log.Println("支付成功")
			log.Printf("支付金额: %s", resp.TradeStatus)
			paymentSuccessful = true // 支付成功，设置全局变量
			return                   // 支付成功，结束循环
		case alipay.TradeStatusClosed:
			log.Println("支付失败或交易已关闭")
			return // 支付失败或已关闭，结束循环
		default:
			log.Printf("支付金额: %s", resp.TradeStatus)
			log.Println("支付状态未确定，将在10秒后再次查询")
			attempt++ // 尝试次数增加
		}

		time.Sleep(3 * time.Second) // 等待20秒后再次尝试
	}

	// 达到最大尝试次数，尝试取消交易
	if attempt == maxAttempts {
		_, err := client.TradeCancel(alipay.TradeCancel{OutTradeNo: outTradeNo})
		if err != nil {
			log.Printf("尝试取消交易失败: %v", err)
		} else {
			log.Printf("交易 %s 被成功取消", outTradeNo)
		}
	}
}

func EndTransactionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		AppID           string `json:"appId"`
		PrivateKey      string `json:"privateKey"`
		AlipayPublicKey string `json:"alipayPublicKey"`
		OutTradeNo      string `json:"outTradeNo"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// 根据请求动态创建支付宝客户端
	client, err := alipay.New(req.AppID, req.PrivateKey, true) // 假设使用沙箱模式
	if err != nil {
		log.Printf("创建支付宝客户端失败: %v", err)
		http.Error(w, "创建支付宝客户端失败", http.StatusInternalServerError)
		return
	}
	client.LoadAliPayPublicKey(req.AlipayPublicKey)

	// 首先查询当前的支付状态
	queryResp, queryErr := client.TradeQuery(alipay.TradeQuery{OutTradeNo: req.OutTradeNo})
	if queryErr != nil {
		log.Printf("查询交易 %s 状态失败: %v", req.OutTradeNo, queryErr)
		// 这里不直接返回错误，因为我们的主要目的是尝试取消交易
	} else {
		// 打印当前的支付状态
		log.Printf("当前交易 %s 的状态为: %s", req.OutTradeNo, queryResp.TradeStatus)
	}

	// 使用新创建的客户端尝试取消交易
	_, err = client.TradeCancel(alipay.TradeCancel{OutTradeNo: req.OutTradeNo})
	if err != nil {
		log.Printf("尝试取消交易失败: %v", err)
		http.Error(w, "取消交易失败", http.StatusInternalServerError)
		return
	}

	log.Printf("交易 %s 被成功取消", req.OutTradeNo)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "交易已被取消"})
}
