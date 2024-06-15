package register

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"gocode/first/config"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql" // 导入MySQL驱动
)

// User 结构体
type User struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"` // 在实际应用中，密码应该是经过加密的
}

var db *sql.DB

// RegisterHandler 处理注册请求
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	// 仅接受POST方法
	if r.Method != "POST" {
		http.Error(w, "Only POST method is accepted", http.StatusMethodNotAllowed)
		return
	}

	var user User
	// 解析请求体中的JSON数据到User结构体
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 连接到数据库
	// 注意替换 "username:password@/dbname" 为你的实际数据库连接信息
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

	// 准备插入语句
	stmt, err := db.Prepare("INSERT INTO manager(username, email, password) VALUES(?, ?, ?)")
	if err != nil {
		http.Error(w, "Failed to prepare database statement", http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	// 执行插入语句
	result, err := stmt.Exec(user.Username, user.Email, user.Password)
	if err != nil {
		http.Error(w, "Failed to insert user into database", http.StatusInternalServerError)
		return
	}

	// 检查插入是否成功
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "No rows affected, user not inserted", http.StatusInternalServerError)
		return
	}

	// 设置响应头和发送成功消息
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := map[string]string{"status": "success", "message": "Registration successful"}
	json.NewEncoder(w).Encode(response)
}
