package email

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/smtp"
	"time"
)

// EmailTestRequest 定义了接收前端请求的结构体
type EmailTestRequest struct {
	Email string `json:"email"`
}

// 假设有一个全局映射，用于存储电子邮件地址和对应的验证码
// 在实际应用中，这应该替换为数据库或其他持久化存储解决方案
var codeMap = make(map[string]string)
var codeExpiry = make(map[string]time.Time)

// VerifyCodeRequest 定义了接收前端验证码验证请求的结构体
type VerifyCodeRequest struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

func verifyCode(email, code string) bool {
	// 检查验证码是否存在且未过期
	storedCode, exists := codeMap[email]
	if !exists {
		return false
	}

	expiry := codeExpiry[email] // 直接获取过期时间，不使用空标识符
	if time.Now().After(expiry) {
		// 验证码已过期
		delete(codeMap, email)
		delete(codeExpiry, email)
		return false
	}

	return code == storedCode
}

// verifyCodeHandler 处理验证验证码的 HTTP 请求
func VerifyCodeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "只支持POST方法", http.StatusMethodNotAllowed)
		return
	}

	var request VerifyCodeRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	isValid := verifyCode(request.Email, request.Code)
	if !isValid {
		http.Error(w, "验证码错误或已过期", http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("验证码正确"))
}

// sendEmail 发送邮件，使用 TLS 加密
func sendEmail(recipient, code string) error {
	// 从数据库获取邮箱配置
	emailConfig, err := GetEmailConfig() // 假设 GetEmailConfig 是已经实现的函数
	if err != nil {
		log.Printf("Error getting email configuration: %v", err)
		return err
	}

	// 打印获取的邮箱配置
	log.Printf("SMTP Server: %s, Port: %d, Sender: %s", emailConfig.SMTP, emailConfig.Port, emailConfig.Sender)

	// 设置SMTP认证信息
	auth := smtp.PlainAuth("", emailConfig.Sender, emailConfig.Password, emailConfig.SMTP)

	// 替换动态内容
	blogURL := "https://www.staykoi.asia" // 示例URL
	blogEngName := "stay点餐"               // 示例博客名
	message := fmt.Sprintf(`
		<p style="font-family: 'Lucida Handwriting', cursive; color: #555; font-size: 16px; line-height: 1.6; text-align: center;">
		<span style="display: block; margin-bottom: 20px; font-size: 14px; color: #B94A5A;">花开再美，怎如初见</span>
		<span style="background-color: #FADADD; color: #D6336C; padding: 5px 10px; border-radius: 15px; box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1); text-shadow: 0.5px 0.5px 1px #FADADD; font-family: 'Comic Sans MS', cursive, sans-serif; font-size: 18px;">
			你的验证码是 :
		</span> 
		<strong style="display: inline-block; background-color: #FFE0F0; border: 1px dashed #FFAEC0; padding: 12px 20px; margin-top: 10px; border-radius: 25px; font-size: 20px; color: #D6336C; box-shadow: 0 4px 8px rgba(0, 0, 0, 0.1); text-shadow: 1px 1px 2px pink;">
			%s
		</strong>
		</p>
		`, code)

	logoURL := emailConfig.LogoURL // 从数据库获取的Logo URL

	// 构建邮件内容
	content := fmt.Sprintf(`
<div style="display: flex; align-items: center; padding: 15px; color: #666; font-size: 14px; line-height: 1.5; word-break: break-all; background-color: #FFF0F5;">
    <div style="overflow: hidden; width: 500px; margin: 0 auto; box-sizing: border-box; border: 1px solid #FADADD; box-shadow: 0px 0px 20px #F0B7C3; border-radius: 10px;">
        <div>
            <img style="display: block; width: 100%%;" src="%s">
            <div style="display: inline-block; font-size:20px;margin-left: 20px; padding: 10px 25px; background: #FADADD; color: #FFFFFF; text-align: center; box-shadow: 3px 3px 5px rgba(0, 0, 0, 0.3); border-radius: 5px; font-family: 'Lucida Handwriting', cursive; transform: translateY(-25px);">Dear</div>
        </div>
        <div style="padding: 35px 25px; font-family: 'Comic Sans MS', cursive, sans-serif;">%s</div>
        <div style="display: flex; flex-direction: column; align-items: center; margin-top: 50px;">
            <a style="padding: 10px 25px; background: #FADADD; color: #FFFFFF; text-decoration: none; box-shadow: 3px 3px 5px rgba(0, 0, 0, 0.3); border-radius: 5px; font-family: 'Comic Sans MS', cursive, sans-serif;" href="%s" target="_blank">访问Stay</a>
            <div style="margin-top: 30px; text-align: center; font-size: 12px; color: #8B008B;">本邮件为系统自动发出<br>Staykoi.asia © %d<br><a style="color: #DA70D6; text-decoration: none;" href="%s" target="_blank">%s</a>. 桃李春风一杯酒</div>
        </div>
    </div>
</div>
`, logoURL, message, blogURL, time.Now().Year(), blogURL, blogEngName)

	// 构建邮件头
	header := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: Verification Code\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n", emailConfig.Sender, recipient)
	// 在发送邮件前，存储验证码及其过期时间
	codeMap[recipient] = code
	codeExpiry[recipient] = time.Now().Add(10 * time.Minute) // 假设验证码10分钟后过期
	// 使用 tls 包的 Dial 函数创建到 SMTP 服务器的 SSL 连接
	smtpAddr := fmt.Sprintf("%s:%d", emailConfig.SMTP, emailConfig.Port)
	tlsConfig := &tls.Config{
		ServerName: emailConfig.SMTP,
	}
	conn, err := tls.Dial("tcp", smtpAddr, tlsConfig)
	if err != nil {
		log.Printf("Failed to dial: %v", err)
		return err
	}
	defer conn.Close()

	// 使用 conn 创建一个 SMTP 客户端
	c, err := smtp.NewClient(conn, emailConfig.SMTP)
	if err != nil {
		log.Printf("Failed to create SMTP client: %v", err)
		return err
	}
	defer c.Close()

	// 设置 SMTP 认证信息
	if err = c.Auth(auth); err != nil {
		log.Printf("SMTP Auth error: %v", err)
		return err
	}

	// 设置发送者和接收者
	if err = c.Mail(emailConfig.Sender); err != nil {
		log.Printf("Failed to set sender: %v", err)
		return err
	}
	if err = c.Rcpt(recipient); err != nil {
		log.Printf("Failed to set recipient: %v", err)
		return err
	}

	// 获取写入器来写入邮件内容
	wc, err := c.Data()
	if err != nil {
		log.Printf("Failed to get data writer: %v", err)
		return err
	}
	defer wc.Close()

	// 写入邮件内容
	_, err = wc.Write([]byte(header + content))
	if err != nil {
		log.Printf("Failed to send email: %v", err)
		return err
	}

	log.Printf("Email sent successfully to %s", recipient)
	return nil
}

// GenerateCode 生成六位数字的验证码
func GenerateCode() string {
	code := fmt.Sprintf("%06d", rand.Intn(1000000))
	return code
}

// sendTestEmailHandler 处理发送测试邮件的 HTTP 请求
func SendTestEmailHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "只支持POST方法", http.StatusMethodNotAllowed)
		return
	}

	var request EmailTestRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	code := GenerateCode()                // 调用 email 包中的方法生成验证码
	err := sendEmail(request.Email, code) // 发送邮件
	if err != nil {
		http.Error(w, "发送邮件失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("测试邮件已发送"))
}
