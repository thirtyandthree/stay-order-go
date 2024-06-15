package login

import (
	"crypto/rand"
	"database/sql" // 引入数据库sql包
	"encoding/base64"
	"encoding/json"
	"fmt"
	"gocode/first/config"
	"net/http"
	"time"

	_ "github.com/go-sql-driver/mysql" // 使用MySQL驱动
)

type Request struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Response struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Token   string `json:"token,omitempty"`
}

func GenerateToken() string {
	b := make([]byte, 48)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	token := base64.URLEncoding.EncodeToString(b) + fmt.Sprintf("%d", time.Now().Unix())
	return token
}

func checkCredentials(username, password string) bool {
	// 使用配置中的数据库连接信息
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
		config.DBConfig.Username, config.DBConfig.Password, config.DBConfig.Host, config.DBConfig.Port, config.DBConfig.Database))
	if err != nil {
		fmt.Println("Failed to connect to database:", err)
		return false
	}
	defer db.Close()

	var id int
	err = db.QueryRow("SELECT id FROM manager WHERE username = ? AND password = ?", username, password).Scan(&id)
	if err != nil || id == 0 {
		return false
	}

	return true
}

func HandleLogin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if checkCredentials(req.Username, req.Password) {
		token := GenerateToken()
		response := Response{"success", "Login successful", token}
		json.NewEncoder(w).Encode(response)
	} else {
		response := Response{Status: "error", Message: "Invalid username or password", Token: ""}
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(response)
	}
}

func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// w.Header().Set("Access-Control-Allow-Origin", "http://localhost:8080")
		// // 允许使用的方法
		// w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		// w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		// w.Header().Set("Access-Control-Allow-Credentials", "true")
		// if r.Method == "OPTIONS" {
		// 	w.WriteHeader(http.StatusOK)
		// 	return
		// }
		next.ServeHTTP(w, r)
	})
}
