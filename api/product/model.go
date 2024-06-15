package product

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"gocode/first/config"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

func init() {
	var err error

	// 从配置文件中获取数据库连接信息
	dbc := config.DBConfig
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
		dbc.Username, dbc.Password, dbc.Host, dbc.Port, dbc.Database)

	// 使用配置信息打开数据库连接
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}

	// 检查与数据库的连接
	err = db.Ping()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
}

type Product struct {
	ID           int      `json:"id"`
	Name         string   `json:"name"`
	Price        float64  `json:"price"`
	ImageURL     string   `json:"imageUrl"`
	Sizes        []string `json:"sizes"`
	Temperatures []string `json:"temperatures"`
	Addons       []string `json:"addons"`
	Stock        int      `json:"stock"`
	Category     string   `json:"category"`
}

func FetchProducts() ([]Product, error) {
	rows, err := db.Query("SELECT id, name, price, imageUrl, sizes, temperatures, addons, stock, category FROM products")
	if err != nil {
		log.Printf("Failed to execute query: %v", err)
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	defer rows.Close()

	var products []Product
	for rows.Next() {
		var p Product
		var sizes, temperatures, addons string

		if err := rows.Scan(&p.ID, &p.Name, &p.Price, &p.ImageURL, &sizes, &temperatures, &addons, &p.Stock, &p.Category); err != nil {
			log.Printf("Failed to scan product data: %v", err)
			return nil, fmt.Errorf("failed to scan product data: %w", err)
		}

		if err := json.Unmarshal([]byte(sizes), &p.Sizes); err != nil {
			log.Printf("Error unmarshalling sizes for product ID %d: %v", p.ID, err)
			continue // Optionally continue processing other products, ignoring the current one
		}
		if err := json.Unmarshal([]byte(temperatures), &p.Temperatures); err != nil {
			log.Printf("Error unmarshalling temperatures for product ID %d: %v", p.ID, err)
			continue // Optionally continue processing other products, ignoring the current one
		}
		if err := json.Unmarshal([]byte(addons), &p.Addons); err != nil {
			log.Printf("Error unmarshalling addons for product ID %d: %v", p.ID, err)
			continue // Optionally continue processing other products, ignoring the current one
		}

		products = append(products, p)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error after scanning products: %v", err)
		return nil, fmt.Errorf("error after scanning products: %w", err)
	}

	return products, nil
}
func AddProduct(p Product) error {
	sizes, err := json.Marshal(p.Sizes)
	if err != nil {
		return err
	}

	temperatures, err := json.Marshal(p.Temperatures)
	if err != nil {
		return err
	}

	addons, err := json.Marshal(p.Addons)
	if err != nil {
		return err
	}

	pID, err := GetMaxID()
	if err != nil {
		return err
	}

	stmt, err := db.Prepare("INSERT INTO products(id, name, price, imageUrl, sizes, temperatures, addons, stock, category) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(pID, p.Name, p.Price, p.ImageURL, sizes, temperatures, addons, p.Stock, p.Category)
	return err
}

func UpdateProduct(p Product) error {
	sizes, err := json.Marshal(p.Sizes)
	if err != nil {
		return err
	}

	temperatures, err := json.Marshal(p.Temperatures)
	if err != nil {
		return err
	}

	addons, err := json.Marshal(p.Addons)
	if err != nil {
		return err
	}

	stmt, err := db.Prepare("UPDATE products SET name=?, price=?, imageUrl=?, sizes=?, temperatures=?, addons=?, stock=?, category=? WHERE id=?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(p.Name, p.Price, p.ImageURL, sizes, temperatures, addons, p.Stock, p.Category, p.ID)
	return err
}

func GetMaxID() (int, error) {
	var maxID sql.NullInt64
	err := db.QueryRow("SELECT MAX(id) FROM products").Scan(&maxID)
	if err != nil {
		if err == sql.ErrNoRows {
			// 如果找不到记录，则返回默认值 1
			return 1, nil
		}
		return 0, err
	}

	// 如果 maxID 为 NULL，则返回默认值 1
	if !maxID.Valid {
		return 1, nil
	}

	return int(maxID.Int64) + 1, nil
}
