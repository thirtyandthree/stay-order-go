package person

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
)

// UploadFile 处理文件上传
func UploadFile(w http.ResponseWriter, r *http.Request) {
	// 限制上传文件的大小（例如：10MB）
	r.ParseMultipartForm(10 << 20) // 10MB

	file, handler, err := r.FormFile("image")
	if err != nil {
		fmt.Println("Error Retrieving the File:", err)
		http.Error(w, "Error Retrieving the File", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// 安全的文件名检查
	safeFileName := regexp.MustCompile(`[^a-zA-Z0-9.-]+`).ReplaceAllString(handler.Filename, "")
	if safeFileName == "" {
		http.Error(w, "Invalid file name", http.StatusBadRequest)
		return
	}

	// 确定文件保存路径，例如 "./uploads/"
	uploadPath := filepath.Join(".", "./uploads") // 更改为相对当前目录的路径
	if err := os.MkdirAll(uploadPath, os.ModePerm); err != nil {
		fmt.Println("Error creating upload directory:", err)
		http.Error(w, "Error processing the upload", http.StatusInternalServerError)
		return
	}

	// 创建或打开文件
	filePath := filepath.Join(uploadPath, safeFileName)
	dst, err := os.Create(filePath)
	if err != nil {
		fmt.Println("Error creating file:", err)
		http.Error(w, "Unable to create the file for writing", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	// 读取上传的文件内容
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		fmt.Println("Error reading file content:", err)
		http.Error(w, "Error reading file content", http.StatusInternalServerError)
		return
	}

	// 将内容写入新文件
	if _, err := dst.Write(fileBytes); err != nil {
		fmt.Println("Error writing file to disk:", err)
		http.Error(w, "Error writing file to disk", http.StatusInternalServerError)
		return
	}

	// 返回上传成功的响应
	fmt.Fprintf(w, "File %s Uploaded Successfully", safeFileName)
}

// GetImages 用于处理获取图库中所有图片的请求
func GetImages(w http.ResponseWriter, r *http.Request) {
	uploadPath := filepath.Join(".", "./uploads") // 与UploadFile中使用的路径相同
	baseURL := "http://localhost:8081/uploads/"   // 基础URL，根据服务器配置适当修改

	var imageUrls []string

	// 遍历uploads目录，收集所有文件的完整URL
	err := filepath.Walk(uploadPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 确保只添加文件，不添加目录
		if !info.IsDir() {
			// 构建完整的URL路径
			imageUrl := baseURL + filepath.Base(path)
			imageUrls = append(imageUrls, imageUrl)
		}
		return nil
	})

	if err != nil {
		fmt.Println("Error reading upload directory:", err)
		http.Error(w, "Failed to list images", http.StatusInternalServerError)
		return
	}

	// 设置响应类型为JSON
	w.Header().Set("Content-Type", "application/json")
	// 返回图片URL列表的JSON表示
	json.NewEncoder(w).Encode(imageUrls)
}
