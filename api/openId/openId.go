package openId

import (
	"encoding/json"
	"gocode/first/config"
	"io"
	"net/http"
)

type LoginRequest struct {
	Code string `json:"code"`
}

type WechatResponse struct {
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

func Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	// 解析JSON请求体
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "请求数据解析失败："+err.Error(), http.StatusBadRequest)
		return
	}
	wechat := config.C.Wechat

	// 构建微信API URL
	requestURL := "https://api.weixin.qq.com/sns/jscode2session?appid=" + wechat.AppID + "&secret=" + wechat.AppSecret + "&js_code=" + req.Code + "&grant_type=authorization_code"

	// 向微信服务器发送请求
	resp, err := http.Get(requestURL)
	if err != nil {
		http.Error(w, "从微信请求openid失败", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "读取微信响应失败", http.StatusInternalServerError)
		return
	}

	var wechatResp WechatResponse
	if err := json.Unmarshal(body, &wechatResp); err != nil {
		http.Error(w, "微信响应解析失败", http.StatusInternalServerError)
		return
	}

	if wechatResp.ErrCode != 0 {
		http.Error(w, "微信API错误: "+wechatResp.ErrMsg, http.StatusBadRequest)
		return
	}

	// 发送openid作为JSON响应
	response := map[string]string{"status": "success", "openid": wechatResp.OpenID}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
