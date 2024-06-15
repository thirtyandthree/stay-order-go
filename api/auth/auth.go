package auth

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
)

const (
	appID     = "wx74f8de5e01fcf1e5"
	appSecret = "5dcb7234a89e950f447dc0beda68feb2"
	dbSource  = "root:123mysql@tcp(localhost:3306)/stay33"
)

type WechatAuthResponse struct {
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
}

type UserInfo struct {
	NickName  string `json:"nickName"`
	AvatarUrl string `json:"avatarUrl"`
	Gender    int    `json:"gender"` // 性别 0：未知、1：男、2：女
}

func AuthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Unsupported request method.", http.StatusMethodNotAllowed)
		return
	}

	var requestData struct {
		Code     string   `json:"code"`
		UserInfo UserInfo `json:"userInfo"`
	}
	err := json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		http.Error(w, "Error parsing request body", http.StatusBadRequest)
		return
	}

	url := fmt.Sprintf("https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code", appID, appSecret, requestData.Code)
	resp, err := http.Get(url)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var authResp WechatAuthResponse
	err = json.Unmarshal(body, &authResp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 插入数据库
	db, err := sql.Open("mysql", dbSource)
	if err != nil {
		http.Error(w, "Database connection failed", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// 检查是否存在用户记录
	var existingUser int
	err = db.QueryRow("SELECT COUNT(*) FROM users1 WHERE openid = ?", authResp.OpenID).Scan(&existingUser)
	if err != nil {
		http.Error(w, "Database query failed", http.StatusInternalServerError)
		return
	}

	if existingUser == 0 {
		_, err = db.Exec("INSERT INTO users1 (openid, session_key, nickname, avatar_url, gender) VALUES (?, ?, ?, ?, ?)",
			authResp.OpenID, authResp.SessionKey, requestData.UserInfo.NickName, requestData.UserInfo.AvatarUrl, requestData.UserInfo.Gender)
		if err != nil {
			http.Error(w, "Failed to insert user data", http.StatusInternalServerError)
			return
		}
	} else {
		_, err = db.Exec("UPDATE users1 SET session_key = ?, nickname = ?, avatar_url = ?, gender = ? WHERE openid = ?",
			authResp.SessionKey, requestData.UserInfo.NickName, requestData.UserInfo.AvatarUrl, requestData.UserInfo.Gender, authResp.OpenID)
		if err != nil {
			http.Error(w, "Failed to update user data", http.StatusInternalServerError)
			return
		}
	}

	log.Println("User info updated successfully.")

	// 返回用户信息
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(body)
}
