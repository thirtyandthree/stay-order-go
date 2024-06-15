package order

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"gocode/first/config"
	"log"
	"net/http"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

// OrderDetail 结构体代表订单详情
type OrderDetail struct {
	GoodsName       string  `json:"goods_name"`
	GoodsWeight     string  `json:"goods_weight"`
	GoodsNumber     int     `json:"goods_number"`
	GoodsPrice      float64 `json:"goods_price"`
	GoodsTotalPrice float64 `json:"goods_total_price"`
	URL             string  `json:"url"`
	GoodsStatus     string  `json:"goods_status"`
}

// Order 结构体代表一个订单及其详情
type Order struct {
	OrderID     int           `json:"order_id"`
	OrderNumber string        `json:"order_number"`
	OrderPrice  float64       `json:"order_price"`
	OrderUser   string        `json:"order_user"`
	PayStatus   string        `json:"pay_status"`
	IsSend      string        `json:"is_send"`
	CreateTime  string        `json:"create_time"`
	Detail      []OrderDetail `json:"detail"`
}

type UpdateOrderRequest struct {
	OrderNumber string `json:"order_number"`
	OrderPrice  string `json:"order_price"`
	PayStatus   string `json:"pay_status"`
	IsSend      string `json:"is_send"`
}
type BatchDeleteRequest struct {
	OrderIds []int `json:"orderIds"` // 假设前端发送的是订单ID的数组
}

var db *sql.DB // 全局数据库连接

func init() {
	var err error

	// 从配置文件中获取数据库连接信息
	dbc := config.DBConfig
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
		dbc.Username, dbc.Password, dbc.Host, dbc.Port, dbc.Database)

	// 使用配置信息打开数据库连接
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}

	// 检查与数据库的连接
	err = db.Ping()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
}
func CheckOrder(w http.ResponseWriter, r *http.Request) {
	// Prepare and execute the SQL queries
	var orderCount int
	err := db.QueryRow("SELECT COUNT(*) FROM orders").Scan(&orderCount)
	if err != nil {
		http.Error(w, "Failed to query total number of orders", http.StatusInternalServerError)
		return
	}

	var totalPrice float64
	err = db.QueryRow("SELECT SUM(order_price) FROM orders").Scan(&totalPrice)
	if err != nil {
		http.Error(w, "Failed to query total price of orders", http.StatusInternalServerError)
		return
	}

	// Create a response structure
	response := struct {
		OrderCount int     `json:"order_count"`
		TotalPrice float64 `json:"total_price"`
	}{
		OrderCount: orderCount,
		TotalPrice: totalPrice,
	}
	// Set the header and encode the response as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
func GetSpecificOrder(w http.ResponseWriter, r *http.Request) {
	var order Order
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	query := `SELECT order_id, order_number, order_price, pay_status, is_send, create_time FROM orders WHERE order_user = ?`
	rows, err := db.Query(query, order.OrderUser)
	if err != nil {
		fmt.Println("Error:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var orders []Order
	for rows.Next() {
		var o Order
		if err := rows.Scan(&o.OrderID, &o.OrderNumber, &o.OrderPrice, &o.PayStatus, &o.IsSend, &o.CreateTime); err != nil {
			fmt.Println("Error:", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		orders = append(orders, o)
	}

	if len(orders) == 0 {
		fmt.Println("No order found for the given user")
		http.Error(w, "No order found for the given user", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}

// 用于从数据库获取订单列表
func GetOrders(w http.ResponseWriter, r *http.Request) {
	orders := []Order{}
	rows, err := db.Query(`SELECT order_id, order_number, order_price, order_user, pay_status, is_send, create_time FROM orders`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var o Order
		if err := rows.Scan(&o.OrderID, &o.OrderNumber, &o.OrderPrice, &o.OrderUser, &o.PayStatus, &o.IsSend, &o.CreateTime); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		orders = append(orders, o)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}

// 用于从数据库获取指定订单的详情
func GetOrderDetail(w http.ResponseWriter, r *http.Request) {
	// 使用mux.Vars获取路径参数
	vars := mux.Vars(r)
	orderID := vars["order_id"] // 获取路径中的order_id参数
	if orderID == "" {
		http.Error(w, "Missing order_id", http.StatusBadRequest)
		return
	}

	details := []OrderDetail{}
	query := `SELECT goods_name, goods_weight, goods_number, goods_price, goods_total_price, url, goods_status FROM orderDetails WHERE order_id = ?`
	rows, err := db.Query(query, orderID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var d OrderDetail
		if err := rows.Scan(&d.GoodsName, &d.GoodsWeight, &d.GoodsNumber, &d.GoodsPrice, &d.GoodsTotalPrice, &d.URL, &d.GoodsStatus); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		details = append(details, d)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(details)
}

func AddOrder(w http.ResponseWriter, r *http.Request) {
	// 解析请求体到 Order 结构体
	var newOrder Order
	if err := json.NewDecoder(r.Body).Decode(&newOrder); err != nil {
		log.Printf("Error decoding request body: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 开始数据库事务
	tx, err := db.Begin()
	if err != nil {
		log.Printf("Error starting database transaction: %v", err)
		http.Error(w, "Failed to start transaction", http.StatusInternalServerError)
		return
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			log.Printf("Error rolling back transaction: %v", err)
		}
	}()

	// 插入订单基本信息
	query := `INSERT INTO orders(order_number, order_price, order_user, pay_status, is_send, create_time) VALUES(?, ?, ?, ?, ?, ?)`
	res, err := tx.Exec(query, newOrder.OrderNumber, newOrder.OrderPrice, newOrder.OrderUser, newOrder.PayStatus, newOrder.IsSend, newOrder.CreateTime)
	if err != nil {
		log.Printf("Error inserting order: %v", err)
		http.Error(w, "Failed to insert order", http.StatusInternalServerError)
		return
	}

	// 获取新插入订单的ID
	lastId, err := res.LastInsertId()
	if err != nil {
		log.Printf("Error retrieving last insert ID: %v", err)
		http.Error(w, "Failed to get last insert ID", http.StatusInternalServerError)
		return
	}

	// 插入订单详情
	detailQuery := `INSERT INTO orderDetails(order_id, goods_name, goods_weight, goods_number, goods_price, goods_total_price, url, goods_status) VALUES(?, ?, ?, ?, ?, ?, ?, ?)`
	for _, detail := range newOrder.Detail {
		_, err := tx.Exec(detailQuery, lastId, detail.GoodsName, detail.GoodsWeight, detail.GoodsNumber, detail.GoodsPrice, detail.GoodsTotalPrice, detail.URL, detail.GoodsStatus)
		if err != nil {
			log.Printf("Error inserting order detail: %v", err)
			http.Error(w, "Failed to insert order detail", http.StatusInternalServerError)
			return
		}
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		log.Printf("Error committing transaction: %v", err)
		http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
		return
	}

	// 发送成功响应
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{"message": "New order added successfully"}); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}

// UpdateOrder 处理更新订单请求的函数
func UpdateOrder(w http.ResponseWriter, r *http.Request) {
	// 解析URL路径中的order_id参数
	vars := mux.Vars(r)
	orderID := vars["order_id"]

	// 解析请求体到UpdateOrderRequest结构体
	var req UpdateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 将字符串转换为适当的数据类型
	orderPrice, err := strconv.ParseFloat(req.OrderPrice, 64)
	if err != nil {
		http.Error(w, "Invalid order price", http.StatusBadRequest)
		return
	}
	payStatus, err := strconv.ParseBool(req.PayStatus)
	if err != nil {
		http.Error(w, "Invalid pay status", http.StatusBadRequest)
		return
	}
	isSend, err := strconv.ParseBool(req.IsSend)
	if err != nil {
		http.Error(w, "Invalid is send status", http.StatusBadRequest)
		return
	}

	// 构造SQL更新语句
	// 注意安全性，避免SQL注入
	query := `UPDATE orders SET order_number = ?, order_price = ?, pay_status = ?, is_send = ? WHERE order_id = ?`

	// 执行SQL更新语句
	_, err = db.Exec(query, req.OrderNumber, orderPrice, payStatus, isSend, orderID)
	if err != nil {
		http.Error(w, "Failed to update order", http.StatusInternalServerError)
		return
	}

	// 返回成功响应
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Order updated successfully"})
}

// BatchDeleteOrders 处理批量删除订单的请求
func BatchDeleteOrders(w http.ResponseWriter, r *http.Request) {
	var req BatchDeleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 开始数据库事务
	tx, err := db.Begin()
	if err != nil {
		http.Error(w, "Failed to start transaction", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	query := "DELETE FROM orders WHERE order_id = ?"
	stmt, err := tx.Prepare(query)
	if err != nil {
		http.Error(w, "Failed to prepare delete statement", http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	// 遍历订单ID数组，执行删除操作
	for _, orderId := range req.OrderIds {
		if _, err := stmt.Exec(orderId); err != nil {
			http.Error(w, "Failed to delete order", http.StatusInternalServerError)
			return
		}
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
		return
	}

	// 发送成功响应
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Orders deleted successfully"})
}
