package desk

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"gocode/first/config"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
)

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

// Table 表示餐桌信息
type Table struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Capacity int    `json:"capacity"`
	Status   string `json:"status"`
	Image    string `json:"image"`
}

// GetMaxTableID 查询当前最大的餐桌ID
func GetMaxTableID() (int, error) {
	var maxID int
	query := "SELECT MAX(id) FROM tableList"
	err := db.QueryRow(query).Scan(&maxID)
	if err != nil {
		return 0, err
	}
	return maxID, nil
}

// handleTableData 请求处理函数，用于处理对餐桌数据的请求
func HandleTableData(w http.ResponseWriter, r *http.Request) {
	// 设置响应头为 JSON 格式
	w.Header().Set("Content-Type", "application/json")

	// 检查请求方法是否为 POST
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 执行查询语句
	rows, err := db.Query("SELECT id, name, capacity, status, image FROM tablelist")
	if err != nil {
		http.Error(w, "Failed to execute query", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// 遍历查询结果并将数据存入切片
	var tables []Table
	for rows.Next() {
		var table Table
		err := rows.Scan(&table.ID, &table.Name, &table.Capacity, &table.Status, &table.Image)
		if err != nil {
			http.Error(w, "Failed to scan rows", http.StatusInternalServerError)
			return
		}
		tables = append(tables, table)
	}

	// 将切片转换为 JSON 格式并返回
	err = json.NewEncoder(w).Encode(tables)
	if err != nil {
		http.Error(w, "Failed to encode JSON", http.StatusInternalServerError)
		return
	}
}

// handleDeskData 请求处理函数，用于处理对餐桌数据的请求
func HandleDeskData(w http.ResponseWriter, r *http.Request) {
	// 设置响应头为 JSON 格式
	w.Header().Set("Content-Type", "application/json")

	// 检查请求方法是否为 POST
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 解析请求体中的 JSON 数据
	var requestData struct {
		TableName string `json:"tableName"`
	}
	err := json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		http.Error(w, "Failed to decode JSON", http.StatusBadRequest)
		return
	}
	fmt.Println("Table Name:", requestData.TableName)

	// 根据请求中的餐桌名称执行删除操作
	_, err = db.Exec("DELETE FROM tablelist WHERE name = ?", requestData.TableName)
	if err != nil {
		http.Error(w, "Failed to delete table", http.StatusInternalServerError)
		return
	}

	// 返回删除成功的响应
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Table deleted successfully"))
}

// handleUpdateTableData 请求处理函数，用于更新餐桌信息
func HandleUpdateDeskData(w http.ResponseWriter, r *http.Request) {
	// 设置响应头为 JSON 格式
	w.Header().Set("Content-Type", "application/json")

	// 检查请求方法是否为 POST
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 解析请求体中的 JSON 数据
	var updateData Table
	err := json.NewDecoder(r.Body).Decode(&updateData)
	if err != nil {
		http.Error(w, "Failed to decode JSON", http.StatusBadRequest)
		return
	}

	// 根据解析出的数据更新数据库中的记录
	_, err = db.Exec("UPDATE tableList SET name = ?, capacity = ?, status = ?, image = ? WHERE id = ?",
		updateData.Name, updateData.Capacity, updateData.Status, updateData.Image, updateData.ID)
	if err != nil {
		http.Error(w, "Failed to update table", http.StatusInternalServerError)
		return
	}

	// 返回更新成功的响应
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Table updated successfully"))
}

// HandleAddTable 处理添加餐桌的请求
func HandleAddTable(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var newTable Table
	err := json.NewDecoder(r.Body).Decode(&newTable)
	if err != nil {
		http.Error(w, "Failed to decode JSON", http.StatusBadRequest)
		return
	}

	// 查询当前最大的餐桌ID
	maxID, err := GetMaxTableID()
	if err != nil {
		http.Error(w, "Failed to get max table ID", http.StatusInternalServerError)
		return
	}
	newTableID := maxID + 1 // 新餐桌ID为当前最大ID加一

	// 插入数据到数据库，使用newTableID作为新餐桌的ID
	query := `INSERT INTO tableList (id, name, capacity, status, image) VALUES (?, ?, ?, ?, ?)`
	_, err = db.Exec(query, newTableID, newTable.Name, newTable.Capacity, newTable.Status, newTable.Image)
	if err != nil {
		http.Error(w, "Failed to insert new table", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Table added successfully", "newID": fmt.Sprintf("%d", newTableID)})
}
