package user

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func HandleAddUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = AddUser(user)
	if err != nil {
		http.Error(w, "Failed to add user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("User added successfully"))
}

func AddUser(user User) error {
	stmt, err := db.Prepare("INSERT INTO user (date, name, address, email, gender, phone) VALUES (?, ?, ?, ?, ?, ?)")
	if err != nil {
		log.Printf("Error preparing statement: %v", err)
		return fmt.Errorf("error preparing insert statement: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(user.Date, user.Name, user.Address, user.Email, user.Gender, user.Phone)
	if err != nil {
		log.Printf("Error executing statement: %v", err)
		return fmt.Errorf("error executing insert statement: %w", err)
	}

	return nil
}

// HandleDeleteUser 处理删除用户的 HTTP 请求
func HandleDeleteUser(w http.ResponseWriter, r *http.Request) {
	// 仅允许 POST 方法
	if r.Method != "POST" {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	// 解析请求体获取用户信息
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 根据用户名删除用户，这里使用 Name 字段
	err = DeleteUserByName(user.Name)
	if err != nil {
		http.Error(w, "Failed to delete user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("User deleted successfully"))
}

// DeleteUserByName 从数据库中删除一个用户
func DeleteUserByName(name string) error {
	// 准备 SQL 语句，这里假设你已经有了数据库连接 `db`
	stmt, err := db.Prepare("DELETE FROM user WHERE name = ?")
	if err != nil {
		return fmt.Errorf("error preparing delete statement: %w", err)
	}
	defer stmt.Close()

	// 执行 SQL 语句
	_, err = stmt.Exec(name)
	if err != nil {
		return fmt.Errorf("error executing delete statement: %w", err)
	}

	return nil
}
