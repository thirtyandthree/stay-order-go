package cheapgoods

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// HandleProducts 根据请求方法返回所有产品或根据POST请求的内容更新或添加产品
func HandleProducts(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		// 检查请求体是否为空
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Error reading request body", http.StatusBadRequest)
			return
		}

		if len(body) == 0 {
			// 请求体为空，假设是获取并返回所有产品
			products, err := FetchProducts()
			if err != nil {
				http.Error(w, "Server error", http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(products)
		} else {
			// 请求体不为空，解析产品信息进行添加或更新
			var p Product
			if err := json.Unmarshal(body, &p); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			if p.ID > 0 {
				// 更新产品
				err := UpdateProduct(p)
				if err != nil {
					http.Error(w, "Failed to update product", http.StatusInternalServerError)
					return
				}
			} else {
				// 添加新产品
				fmt.Println(p)

				err := AddProduct(p)

				if err != nil {
					http.Error(w, "Failed to add product", http.StatusInternalServerError)
					return
				}
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(p)
		}

	default:
		// 如果请求不是POST，返回405 Method Not Allowed
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}
