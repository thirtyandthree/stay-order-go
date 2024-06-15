package product

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// HandleDeleteProduct 处理删除商品的HTTP请求
func HandleDeleteProduct(w http.ResponseWriter, r *http.Request) {
	// 仅允许POST方法
	if r.Method != "POST" {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	// 假设请求的body包含要删除的商品信息，例如 {"name": "ProductName"}
	var product Product
	err := json.NewDecoder(r.Body).Decode(&product)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = DeleteProductByName(product.Name)
	if err != nil {
		http.Error(w, "Failed to delete product", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Product deleted successfully"))
}

// DeleteProductByName 从数据库中删除一个产品
func DeleteProductByName(name string) error {
	// 准备SQL语句
	stmt, err := db.Prepare("DELETE FROM products WHERE name = ?")
	if err != nil {
		return fmt.Errorf("error preparing delete statement: %w", err)
	}
	defer stmt.Close()

	// 执行SQL语句
	_, err = stmt.Exec(name)
	if err != nil {
		return fmt.Errorf("error executing delete statement: %w", err)
	}

	return nil
}
