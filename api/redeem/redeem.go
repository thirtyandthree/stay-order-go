package redeem

import (
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"fmt"
	"gocode/first/config"
	"log"
	"net/http"
	"time"
)

type Redeem struct {
	ID          int       `json:"id"`
	Code        string    `json:"code"`
	ProductName string    `json:"productName"`
	CodeCount   int       `json:"codeCount"`
	StartTime   time.Time `json:"startTime"`
	EndTime     time.Time `json:"endTime"`
	Price       float64   `json:"price"`
	CodeTtatus  string    `json:"codeTtatus"` //状态
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
func InsertRedeem(w http.ResponseWriter, r *http.Request) {
	var redeem Redeem
	if err := json.NewDecoder(r.Body).Decode(&redeem); err != nil {
		fmt.Printf("Error decoding Redeem: %v\n", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	for i := 0; i < redeem.CodeCount; i++ {
		uniqueCode, err := generateUniqueCode()
		if err != nil {
			fmt.Printf("Error generating unique redeem code: %v\n", err)
			http.Error(w, "Failed to generate redeem code", http.StatusInternalServerError)
			return
		}

		// 将时间字段格式化为字符串
		startTimeStr := redeem.StartTime.Format("2006-01-02 15:04:05")
		endTimeStr := redeem.EndTime.Format("2006-01-02 15:04:05")

		query := `INSERT INTO redeem (code, productName, startTime, endTime, codeTtatus, price) VALUES (?, ?, ?, ?, ?, ?)`
		_, err = db.Exec(query, uniqueCode, redeem.ProductName, startTimeStr, endTimeStr, redeem.CodeTtatus, redeem.Price)
		if err != nil {
			fmt.Printf("Error inserting redeem: %v\n", err)
			http.Error(w, "Failed to insert redeem", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Redeem created successfully"})
}

func UpdateRedeem(w http.ResponseWriter, r *http.Request) {
	var redeem Redeem
	if err := json.NewDecoder(r.Body).Decode(&redeem); err != nil {
		fmt.Printf("Error decoding Redeem for update: %v\n", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	query := `UPDATE redeem SET code = ?, productName = ?, startTime=?, endTime=?, codeTtatus=?,price=? WHERE id = ?`
	_, err := db.Exec(query, redeem.Code, redeem.ProductName, redeem.StartTime, redeem.EndTime, redeem.CodeTtatus, redeem.Price, redeem.ID)
	if err != nil {
		fmt.Printf("Error updating redeem: %v\n", err)
		http.Error(w, "Failed to update redeem", http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Redeem updated successfully"})
}

func DeleteRedeem(w http.ResponseWriter, r *http.Request) {
	var redeem Redeem
	if err := json.NewDecoder(r.Body).Decode(&redeem); err != nil {
		fmt.Printf("Error decoding Redeem for deletion: %v\n", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	query := `DELETE FROM redeem WHERE id = ?`
	_, err := db.Exec(query, redeem.ID)
	if err != nil {
		fmt.Printf("Error deleting redeem: %v\n", err)
		http.Error(w, "Failed to delete redeem", http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Redeem deleted successfully"})
}
func CheckRedeemcode(w http.ResponseWriter, r *http.Request) {
	var redeem Redeem
	if err := json.NewDecoder(r.Body).Decode(&redeem); err != nil {
		fmt.Printf("Error decoding Redeem for redemption: %v\n", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 查询兑换码是否存在
	var existingCode string
	query := `SELECT code, codeTtatus FROM redeem WHERE code = ?`
	row := db.QueryRow(query, redeem.Code)
	if err := row.Scan(&existingCode, &redeem.CodeTtatus); err != nil {
		if err == sql.ErrNoRows {
			fmt.Println("Redeem code not found")
			http.Error(w, "Redeem code not found", http.StatusNotFound)
			return
		}
		fmt.Printf("Error checking redeem code: %v\n", err)
		http.Error(w, "Failed to check redeem code", http.StatusInternalServerError)
		return
	}

	// 检查兑换码状态
	if redeem.CodeTtatus != "待兑换" {
		fmt.Println("Redeem code has been already redeemed")
		http.Error(w, "Redeem code has been already redeemed", http.StatusBadRequest)
		return
	}

	// 更新兑换码状态为已兑换
	updateQuery := `UPDATE redeem SET codeTtatus = ? WHERE code = ?`
	if _, err := db.Exec(updateQuery, "已兑换", redeem.Code); err != nil {
		fmt.Printf("Error updating redeem code status: %v\n", err)
		http.Error(w, "Failed to update redeem code status", http.StatusInternalServerError)
		return
	}
	// 确保查询语句和Scan方法中的变量完全匹配
	row = db.QueryRow("SELECT id, productName, codeTtatus, price FROM redeem WHERE code = ?", redeem.Code)
	if err := row.Scan(&redeem.ID, &redeem.ProductName, &redeem.CodeTtatus, &redeem.Price); err != nil {
		// 处理错误
		fmt.Printf("Error fetching updated redeem code: %v\n", err)
		http.Error(w, "Failed to fetch updated redeem code", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(redeem)
}

func FetchRedeems(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, code, productName, startTime, endTime, codeTtatus, price FROM redeem")
	if err != nil {
		fmt.Printf("Error fetching redeems: %v\n", err)
		log.Fatal(err)
		http.Error(w, "Failed to fetch redeem", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	redeems := make([]Redeem, 0)
	for rows.Next() {
		var redeem Redeem
		var startTime, endTime string // 使用 string 类型来接收时间字段

		if err := rows.Scan(&redeem.ID, &redeem.Code, &redeem.ProductName, &startTime, &endTime, &redeem.CodeTtatus, &redeem.Price); err != nil {
			fmt.Printf("Error scanning redeem: %v\n", err)
			log.Fatal(err)
			http.Error(w, "Failed to fetch redeem", http.StatusInternalServerError)
			return
		}

		// 解析 startTime 和 endTime
		redeem.StartTime, err = time.Parse("2006-01-02 15:04:05", startTime)
		if err != nil {
			fmt.Printf("Error parsing startTime: %v\n", err)
			log.Fatal(err)
			http.Error(w, "Failed to parse startTime", http.StatusInternalServerError)
			return
		}
		redeem.EndTime, err = time.Parse("2006-01-02 15:04:05", endTime)
		if err != nil {
			fmt.Printf("Error parsing endTime: %v\n", err)
			log.Fatal(err)
			http.Error(w, "Failed to parse endTime", http.StatusInternalServerError)
			return
		}

		redeems = append(redeems, redeem)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(redeems)
}

func generateUniqueCode() (string, error) {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	codeLength := 30 // 现在设置为生成30位的兑换码

	b := make([]byte, codeLength)
	for {
		if _, err := rand.Read(b); err != nil {
			return "", err
		}
		for i := range b {
			b[i] = charset[b[i]%byte(len(charset))]
		}

		code := string(b)
		var exists bool
		err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM redeem WHERE code = ?)", code).Scan(&exists)
		if err != nil {
			return "", err
		}
		if !exists {
			return code, nil
		}
		// 如果生成的代码已存在，则循环继续尝试生成新代码
	}
}
