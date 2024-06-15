// handlers/userHandler.go

package user

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"gocode/first/config"
	"log"
	"net/http"
)

var db *sql.DB // 假设这是在其他地方初始化的数据库连接

type User struct {
	Date    string `json:"date"`
	Name    string `json:"name"`
	Address string `json:"address"`
	Email   string `json:"email"`
	Gender  string `json:"gender"`
	Phone   string `json:"phone"`
}

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
func CheckUser(w http.ResponseWriter, r *http.Request) {
	// Prepare and execute the SQL queries
	var userCount int
	err := db.QueryRow("SELECT COUNT(*) FROM user").Scan(&userCount)
	if err != nil {
		http.Error(w, "Failed to query total number of users", http.StatusInternalServerError)
		return
	}

	var totalAnnouncements int
	err = db.QueryRow("SELECT COUNT(*) FROM announcements").Scan(&totalAnnouncements)
	if err != nil {
		http.Error(w, "Failed to query total number of announcements", http.StatusInternalServerError)
		return
	}

	// Create a response structure
	response := struct {
		UserCount          int `json:"userCount"`
		TotalAnnouncements int `json:"totalAnnouncements"`
	}{
		UserCount:          userCount,
		TotalAnnouncements: totalAnnouncements,
	}
	// Set the header and encode the response as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func UpdateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if user.Gender == "男" {
		user.Gender = "male"
	} else if user.Gender == "女" {
		user.Gender = "female"
	}
	query := `UPDATE user SET date = ?, address = ?, name = ?, gender = ?, phone = ? WHERE email = ?`
	_, err := db.Exec(query, user.Date, user.Address, user.Name, user.Gender, user.Phone, user.Email)
	if err != nil {
		http.Error(w, "Failed to update user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "User updated successfully"})

}
func CheckEmailExists(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	query := "SELECT EXISTS(SELECT 1 FROM user WHERE email = ?)"
	var exists bool
	err := db.QueryRow(query, req.Email).Scan(&exists)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		log.Println(err)
		return
	}

	response := map[string]bool{"exists": exists}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
func PostGetUsers(w http.ResponseWriter, r *http.Request) {
	// 确保请求方法为POST
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	var users []User
	query := `SELECT date, name, address, email, gender, phone FROM user`
	rows, err := db.Query(query)
	if err != nil {
		http.Error(w, "Failed to query user data", http.StatusInternalServerError)
		log.Println(err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var user User
		if err := rows.Scan(&user.Date, &user.Name, &user.Address, &user.Email, &user.Gender, &user.Phone); err != nil {
			http.Error(w, "Failed to parse user data", http.StatusInternalServerError)
			log.Println(err)
			return
		}
		// 在存储到数组之前，将性别从英文转换为中文
		switch user.Gender {
		case "male":
			user.Gender = "男"
		case "female":
			user.Gender = "女"
		}
		users = append(users, user)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}
