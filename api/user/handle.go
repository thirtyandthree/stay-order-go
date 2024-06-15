package user

import (
	"bytes"
	"io"
	"io/ioutil" // 如果你使用的Go版本低于1.16，可以使用ioutil.ReadAll代替io.ReadAll
	"net/http"
)

// HandleUserRequest 根据请求体内容选择适当的处理函数
func HandleUserRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	// 检查请求体是否为空
	body, err := ioutil.ReadAll(r.Body) // 或使用io.ReadAll，如果你的Go版本支持
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}

	if len(body) == 0 {
		// 请求体为空时获取用户数据
		PostGetUsers(w, r)
	} else {
		// 请求体不为空时更新用户数据
		// 需要重新设置请求体，因为ioutil.ReadAll已经读取过了
		r.Body = io.NopCloser(bytes.NewReader(body))
		UpdateUser(w, r)
	}
}
