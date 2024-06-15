// 文件路径: /product/reset.go
package cheapgoods

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// ResetProductInfo 更新数据库中的商品信息
func ResetProductInfo(productData Product) error {
	// 构造一个更新SQL语句，包括category字段
	query := `UPDATE cheapgoods SET price = ?, sizes = ?, temperatures = ?, addons = ?, stock = ?, ImageUrl = ?, category = ? WHERE name = ?`

	// 将slices转换为JSON字符串存储
	sizes, err := json.Marshal(productData.Sizes)
	if err != nil {
		return fmt.Errorf("error marshalling sizes: %w", err)
	}

	temperatures, err := json.Marshal(productData.Temperatures)
	if err != nil {
		return fmt.Errorf("error marshalling temperatures: %w", err)
	}

	addons, err := json.Marshal(productData.Addons)
	if err != nil {
		return fmt.Errorf("error marshalling addons: %w", err)
	}

	// 执行SQL语句，包括category
	_, err = db.Exec(query, productData.Price, sizes, temperatures, addons, productData.Stock, productData.ImageURL, productData.Category, productData.Name)
	if err != nil {
		return fmt.Errorf("updating product: %w", err)
	}

	return nil
}

// HandleResetProductInfo 处理商品信息重置请求
func HandleResetProductInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	var productData Product
	// 解码请求体到productData结构体
	if err := json.NewDecoder(r.Body).Decode(&productData); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 调用ResetProductInfo进行商品信息更新
	if err := ResetProductInfo(productData); err != nil {
		http.Error(w, fmt.Sprintf("Error resetting product: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode("Product info reset successfully")
}
