package email

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"gocode/first/config"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
)

type EmailConfig struct {
	LogoURL       string `json:"logo_url"`
	SMTP          string `json:"smtp"`
	Port          int    `json:"port"`
	Sender        string `json:"sender"`
	Password      string `json:"password"`
	ServerAddress string `json:"server_address"`
}

var db *sql.DB

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

func updateEmailConfig(ec *EmailConfig) error {
	query := `UPDATE email SET logo_url=?, smtp=?, port=?, password=?, server_address=? WHERE id=1`
	_, err := db.Exec(query, ec.LogoURL, ec.SMTP, ec.Port, ec.Password, ec.ServerAddress)
	return err
}

func GetEmailConfig() (*EmailConfig, error) {
	ec := &EmailConfig{}
	query := `SELECT logo_url, smtp, port, sender, password, server_address FROM email WHERE id=1`
	row := db.QueryRow(query)
	err := row.Scan(&ec.LogoURL, &ec.SMTP, &ec.Port, &ec.Sender, &ec.Password, &ec.ServerAddress)
	if err != nil {
		return nil, err
	}
	return ec, nil
}

// HandleCreateEmailConfig 处理创建邮箱配置的请求
func HandleUpdateEmailConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	var ec EmailConfig
	err := json.NewDecoder(r.Body).Decode(&ec)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = updateEmailConfig(&ec)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(ec)
}

// HandleGetEmailConfig 处理获取特定发送者邮箱配置的请求
func HandleGetEmailConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET method is allowed", http.StatusMethodNotAllowed)
		return
	}

	// 移除查询sender的逻辑
	ec, err := GetEmailConfig() // 直接调用getEmailConfig获取配置

	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(ec)
}
