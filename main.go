package main

import (
	"fmt"
	"gocode/first/api/annocement"
	"gocode/first/api/auth"
	"gocode/first/api/cheapgoods"
	"gocode/first/api/desk"
	"gocode/first/api/email"
	"gocode/first/api/login"
	"gocode/first/api/openId"
	orderHandlers "gocode/first/api/order"
	"gocode/first/api/pay"
	person "gocode/first/api/person"
	personlist "gocode/first/api/personList"
	"gocode/first/api/product"
	"gocode/first/api/redeem"
	"gocode/first/api/register"
	"gocode/first/api/user"
	"gocode/first/api/userChart"
	"gocode/first/config"
	"gocode/first/utils"
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func main() {
	config.Load("stay.yaml")
	utils.Client = utils.NewWeChatClient()
	r := mux.NewRouter()
	//fmt.Println("API URL:", config.APIUrl)
	// 使用login包中定义的CORS中间件
	r.Use(login.CORSMiddleware)

	r.HandleFunc("/api/openId", openId.Login).Methods("POST")

	r.HandleFunc("/payment", pay.Payment)
	// 登录和注册路由
	r.HandleFunc("/api/login", login.HandleLogin).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/register", register.RegisterHandler).Methods("POST", "OPTIONS")

	// 产品相关路由
	r.HandleFunc("/api/products", product.HandleProducts).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/reset", product.HandleResetProductInfo).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/delete", product.HandleDeleteProduct).Methods("POST", "OPTIONS")
	//特惠商品路由
	r.HandleFunc("/api/cheapgoods", cheapgoods.HandleProducts).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/cheapgoods/reset", cheapgoods.HandleResetProductInfo).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/cheapgoods/delete", cheapgoods.HandleDeleteProduct).Methods("POST", "OPTIONS")
	// 用户路由
	r.HandleFunc("/api/user", user.HandleUserRequest).Methods("POST") // 使用HandleUserRequest处理POST请求
	r.HandleFunc("/api/user/delete", user.HandleDeleteUser).Methods("POST")
	r.HandleFunc("/api/user/checkEmail", user.CheckEmailExists).Methods("POST")
	r.HandleFunc("/api/user/countUser", user.CheckUser).Methods("POST")
	// 在 main 包中的 main 函数里
	r.HandleFunc("/api/user/add", user.HandleAddUser).Methods("POST")
	//公告相关
	r.HandleFunc("/api/announcement", annocement.FetchAnnouncements).Methods("POST")
	r.HandleFunc("/api/announcement/add", annocement.InsertAnnouncement).Methods("POST")
	r.HandleFunc("/api/announcement/update", annocement.UpdateAnnouncement).Methods("POST")
	r.HandleFunc("/api/announcement/delete", annocement.DeleteAnnouncement).Methods("POST")
	// pesonList路由配置
	r.HandleFunc("/api/personList", personlist.GetPersonList).Methods("POST")
	r.HandleFunc("/api/personList/add", personlist.InsertPerson).Methods("POST")
	r.HandleFunc("/api/personList/update", personlist.UpdatePerson).Methods("POST")
	r.HandleFunc("/api/personList/delete", personlist.DeletePerson).Methods("POST")
	//订单路由配置
	r.HandleFunc("/api/orders", orderHandlers.GetOrders).Methods("GET")
	r.HandleFunc("/api/order/check", orderHandlers.CheckOrder).Methods("POST")
	r.HandleFunc("/api/order/detail/{order_id}", orderHandlers.GetOrderDetail).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/orders/update/{order_id}", orderHandlers.UpdateOrder).Methods("PUT")
	r.HandleFunc("/api/orders/add", orderHandlers.AddOrder).Methods("POST")
	r.HandleFunc("/api/orders/delete", orderHandlers.BatchDeleteOrders).Methods("POST")
	r.HandleFunc("/api/orders/getSpec", orderHandlers.GetSpecificOrder).Methods("POST")
	// 添加餐桌数据处理路由
	r.HandleFunc("/api/desk", desk.HandleTableData).Methods("POST") // 修改此处为HandleTableData
	r.HandleFunc("/api/desk/delete", desk.HandleDeskData).Methods("POST")
	r.HandleFunc("/api/desk/update", desk.HandleUpdateDeskData).Methods("POST")
	r.HandleFunc("/api/desk/add", desk.HandleAddTable).Methods("POST")
	// 添加预订信息处理路由
	r.HandleFunc("/api/reservation", desk.HandleGetAllReservations).Methods("GET")
	r.HandleFunc("/api/reservation/add", desk.AddReservation).Methods("POSt")
	r.HandleFunc("/api/reservation/delete", desk.DeleteReservation).Methods("POSt")
	// 添加用户图表数据处理路由
	r.HandleFunc("/api/userChart", userChart.ChartDataHandler).Methods("GET")
	r.HandleFunc("/api/userChart/info", userChart.UserRecordsHandler).Methods("GET")
	// Inside your main.go where you setup your routes.
	r.HandleFunc("/api/pay", pay.HandleAlipayConfig).Methods("POST", "GET", "OPTIONS")
	r.HandleFunc("/api/pay/test", pay.TestPayment).Methods("POST")
	r.HandleFunc("/api/pay/cancel", pay.EndTransactionHandler).Methods("POST")
	r.HandleFunc("/api/wpay", pay.HandleWeixinConfig).Methods("POST", "GET", "OPTIONS")
	// 添加支付状态处理路由
	r.HandleFunc("/api/payment-status", pay.PaymentStatusHandler).Methods("GET")
	// 在合适的位置添加路由，确保路径和方法匹配
	//微信支付
	// 写个人中心
	r.HandleFunc("/api/update-password", person.UpdatePasswordHandler)
	r.HandleFunc("/api/personInfo", person.UserInfoHandler)
	r.HandleFunc("/api/personUpdate", person.UpdateUserInfoHandler)
	r.HandleFunc("/api/UploadFile", person.UploadFile)
	// 上传图片的图库
	r.HandleFunc("/api/images", person.GetImages)
	//兑换码对应路由
	r.HandleFunc("/api/redeem", redeem.FetchRedeems).Methods("POST")
	r.HandleFunc("/api/redeem/add", redeem.InsertRedeem).Methods("POST")
	r.HandleFunc("/api/redeem/update", redeem.UpdateRedeem).Methods("POST")
	r.HandleFunc("/api/redeem/delete", redeem.DeleteRedeem).Methods("POST")
	r.HandleFunc("/api/redeem/check", redeem.CheckRedeemcode).Methods("POST")
	// 添加邮箱配置路由
	r.HandleFunc("/api/updateEmail", email.HandleUpdateEmailConfig).Methods("POST")
	r.HandleFunc("/api/getEmail", email.HandleGetEmailConfig).Methods("GET")
	r.HandleFunc("/api/sendEmail", email.SendTestEmailHandler).Methods("POST")
	r.HandleFunc("/api/verifyCode", email.VerifyCodeHandler).Methods("POST")
	r.HandleFunc("/api/auth", auth.AuthHandler).Methods("POST")
	// 应用CORS，允许来自前端的跨源请求
	headersOk := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"})
	originsOk := handlers.AllowedOrigins([]string{config.C.Wechat.Domain})
	methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "DELETE", "OPTIONS"})
	credentialsOk := handlers.AllowCredentials() // 允许携带凭证

	corsHandler := handlers.CORS(headersOk, originsOk, methodsOk, credentialsOk)(r)

	// 启动HTTP服务器
	port := config.C.Wechat.Port
	if port[0] != ':' {
		port = ":" + port // 确保端口格式正确
	}
	fmt.Println("Server running on port", port)

	http.ListenAndServe(port, corsHandler)
}
