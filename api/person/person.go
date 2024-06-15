// 文件位置: password/password.go
package person

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"gocode/first/config"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
)

// UpdatePasswordRequest 定义了接收密码更新请求的结构体
type UpdatePasswordRequest struct {
	Email    string `json:"email"`
	Password string `json:"pass"`
}

// UpdatePasswordHandler 是HTTP请求的处理器，用于处理密码更新请求
func UpdatePasswordHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Only POST requests are allowed", http.StatusMethodNotAllowed)
		return
	}

	var req UpdatePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := updatePasswordInDB(req.Email, req.Password); err != nil {
		http.Error(w, "Failed to update password", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Password updated successfully"))
}

func updatePasswordInDB(email, password string) error {
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
		config.DBConfig.Username, config.DBConfig.Password, config.DBConfig.Host, config.DBConfig.Port, config.DBConfig.Database))
	if err != nil {
		fmt.Println("Failed to connect to database:", err)
		return err
	}
	defer db.Close()

	_, err = db.Exec("UPDATE manager SET password = ? WHERE email = ?", password, email)
	return err
}

// UserInfo 结构体定义
type UserInfo struct {
	ID                int    `json:"id"`
	Avatar            string `json:"avatar"`
	Nickname          string `json:"nickname"`
	Age               int    `json:"age"`
	Email             string `json:"email"`
	MobilePhoneNumber string `json:"mobilePhoneNumber"`
	Sex               int    `json:"sex"`
	Account           string `json:"account"`
	Area              string `json:"area"`
	Hobby             string `json:"hobby"`
	Work              string `json:"work"`
	Design            string `json:"design"`
}

const dataSourceName = "root:123mysql@tcp(localhost:3306)/stay33"

// fetchUserInfo查询id为1的用户信息
func fetchUserInfo(db *sql.DB) (*UserInfo, error) {
	user := &UserInfo{}
	query := "SELECT * FROM userInfo WHERE id = 3"
	row := db.QueryRow(query)
	err := row.Scan(&user.ID, &user.Avatar, &user.Nickname, &user.Age, &user.Email, &user.MobilePhoneNumber, &user.Sex, &user.Account, &user.Area, &user.Hobby, &user.Work, &user.Design)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// UserInfoHandler处理HTTP请求
func UserInfoHandler(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
		config.DBConfig.Username, config.DBConfig.Password, config.DBConfig.Host, config.DBConfig.Port, config.DBConfig.Database))
	if err != nil {
		fmt.Println("Failed to connect to database:", err)
		return
	}
	defer db.Close()

	user, err := fetchUserInfo(db)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// UpdateUserInfoHandler 处理用户信息更新请求
func UpdateUserInfoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "只允许POST请求", http.StatusMethodNotAllowed)
		return
	}

	var userInfo UserInfo
	if err := json.NewDecoder(r.Body).Decode(&userInfo); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := updateUserInfoInDB(&userInfo); err != nil {
		log.Printf("更新用户信息失败: %v", err)
		http.Error(w, "服务器内部错误", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("用户信息更新成功"))
}

// updateUserInfoInDB 在数据库中更新用户信息
func updateUserInfoInDB(userInfo *UserInfo) error {
	db, err := sql.Open("mysql", dataSourceName)
	if err != nil {
		return err
	}
	defer db.Close()

	// 构造SQL语句，根据需要更新的字段进行调整
	query := `UPDATE userInfo SET avatar = ?, nickname = ?, age = ?, email = ?, 
        mobilePhoneNumber = ?, sex = ?, account = ?, area = ?, hobby = ?, work = ?, design = ? WHERE id = ?`
	_, err = db.Exec(query, userInfo.Avatar, userInfo.Nickname, userInfo.Age, userInfo.Email,
		userInfo.MobilePhoneNumber, userInfo.Sex, userInfo.Account, userInfo.Area, userInfo.Hobby, userInfo.Work,
		userInfo.Design, userInfo.ID)
	return err
}
