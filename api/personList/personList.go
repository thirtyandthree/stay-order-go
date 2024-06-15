package personlist

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"gocode/first/config"
	"log"
	"net/http"
)

type Person struct {
	ID      int    `json:"id"`
	Email   string `json:"email"`
	Gender  string `json:"gender"`
	Address string `json:"address"`
	Area    string `json:"area"`
	Phone   string `json:"phone"`
}

var db *sql.DB // 全局数据库连接

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
func InsertPerson(w http.ResponseWriter, r *http.Request) {
	var person Person
	if err := json.NewDecoder(r.Body).Decode(&person); err != nil {
		fmt.Printf("Error decoding Person: %v\n", err) // 添加此行
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 插入数据
	query := `INSERT INTO pesonlist (email,gender,address,area,phone) VALUES (?, ?, ?, ?, ?)`
	_, err := db.Exec(query, person.Email, person.Gender, person.Address, person.Area, person.Phone)
	if err != nil {
		fmt.Printf("Error inserting Person: %v\n", err) // 添加此行
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	query = `UPDATE user SET address =?,phone=? WHERE email = ?`
	_, err = db.Exec(query, person.Address, person.Phone, person.Email)
	if err != nil {
		fmt.Printf("Error updating Person: %v\n", err) // 添加此行

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "person created successfully"})
}

func UpdatePerson(w http.ResponseWriter, r *http.Request) {
	var person Person
	if err := json.NewDecoder(r.Body).Decode(&person); err != nil {
		fmt.Printf("Error decoding Person: %v\n", err) // 添加此行
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	query := `UPDATE pesonlist SET gender =?,address=?,area=?,phone=? WHERE id = ?`
	_, err := db.Exec(query, person.Gender, person.Address, person.Area, person.Phone, person.ID)
	if err != nil {
		fmt.Printf("Error updating Person: %v\n", err) // 添加此行
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	query = `UPDATE user SET address =?,phone=? WHERE email = ?`
	_, err = db.Exec(query, person.Address, person.Phone, person.Email)
	if err != nil {
		fmt.Printf("Error updating Person: %v\n", err) // 添加此行

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "person updated successfully"})
}

func DeletePerson(w http.ResponseWriter, r *http.Request) {
	var person Person
	if err := json.NewDecoder(r.Body).Decode(&person); err != nil {
		fmt.Printf("Error decoding Person: %v\n", err) // 添加此行
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	query := `DELETE FROM pesonlist WHERE id = ?`
	_, err := db.Exec(query, person.ID)
	if err != nil {
		fmt.Printf("Error deleting Person: %v\n", err) // 添加此行
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "person deleted successfully"})
}

func GetPersonList(w http.ResponseWriter, r *http.Request) {
	var person Person
	if err := json.NewDecoder(r.Body).Decode(&person); err != nil {
		fmt.Printf("Error decoding Person: %v\n", err) // 添加此行
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	rows, err := db.Query("SELECT * FROM pesonlist where email=?", person.Email)
	if err != nil {
		fmt.Printf("Error querying Person: %v\n", err) // 添加此行
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	var persons []Person
	for rows.Next() {
		var person Person
		if err := rows.Scan(&person.ID, &person.Email, &person.Gender, &person.Address, &person.Area, &person.Phone); err != nil {
			fmt.Printf("Error scanning Person: %v\n", err) // 添加此行
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		persons = append(persons, person)
	}
	json.NewEncoder(w).Encode(persons)
}
