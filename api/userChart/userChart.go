package userChart

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"gocode/first/config"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
)

// ChartData 结构体用于存放查询结果
type ChartData struct {
	ColumnDate     []string `json:"columnDate"`
	ColumnDateType []string `json:"columnDateType"`
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

// getChartData 从数据库查询最近七天的数据并返回
func getChartData() (ChartData, error) {
	var data ChartData
	query := `
        SELECT DATE_FORMAT(date, '%Y-%m-%d') as formattedDate, 
               CASE 
                   WHEN DAYOFWEEK(date) IN (1, 7) THEN '双休日'
                   ELSE '工作日'
               END as dateType
        FROM userChart
        WHERE date > CURDATE() - INTERVAL 7 DAY
        ORDER BY date DESC
    `
	rows, err := db.Query(query)
	if err != nil {
		return data, err
	}
	defer rows.Close()

	for rows.Next() {
		var formattedDate string
		var dateType string
		if err := rows.Scan(&formattedDate, &dateType); err != nil {
			return data, err
		}
		data.ColumnDate = append(data.ColumnDate, formattedDate)
		data.ColumnDateType = append(data.ColumnDateType, dateType)
	}

	return data, nil
}

// ChartDataHandler 是一个HTTP handler函数，它从数据库获取数据并以JSON格式返回
func ChartDataHandler(w http.ResponseWriter, r *http.Request) {
	data, err := getChartData()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

type UserRecord struct {
	Date          string  `json:"date"`
	UserName      string  `json:"userName"`
	TotalSpending float64 `json:"totalSpending"`
	Gender        string  `json:"gender"`
}

// getUserRecords 查询并返回UserRecord的切片
func getUserRecords(db *sql.DB) ([]UserRecord, error) {
	var records []UserRecord
	query := `
        SELECT date, user_name, total_spending, gender
        FROM UserChart
        ORDER BY total_spending DESC
    `
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var record UserRecord
		if err := rows.Scan(&record.Date, &record.UserName, &record.TotalSpending, &record.Gender); err != nil {
			return nil, err
		}
		records = append(records, record)
	}

	return records, nil
}

// UserRecordsHandler 返回用户记录的HTTP Handler
func UserRecordsHandler(w http.ResponseWriter, r *http.Request) {
	// 使用全局数据库连接变量db，而不是每次都打开新的连接
	records, err := getUserRecords(db)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(records)
}
