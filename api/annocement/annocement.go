package annocement

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"gocode/first/config"
	"log"
	"net/http"
)

type Announcement struct {
	ID       int    `json:"id"`
	Title    string `json:"title"`
	Content  string `json:"content"`
	CoverImg string `json:"coverImg"`
	Date     string `json:"date"`
}

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
func InsertAnnouncement(w http.ResponseWriter, r *http.Request) {
	var annocement Announcement
	if err := json.NewDecoder(r.Body).Decode(&annocement); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	query := `INSERT INTO announcements (title, content, coverImg,date) VALUES (?, ?, ?,?)`
	_, err := db.Exec(query, annocement.Title, annocement.Content, annocement.CoverImg, annocement.Date)
	if err != nil {
		log.Printf("Failed to insert annocement: %v", err)
		http.Error(w, "Failed to insert annocement", http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Announcement created successfully"})
}
func UpdateAnnouncement(w http.ResponseWriter, r *http.Request) {
	var annocement Announcement
	if err := json.NewDecoder(r.Body).Decode(&annocement); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	query := `UPDATE announcements SET title = ?, content = ?, coverImg = ?,date=? WHERE id = ?`
	_, err := db.Exec(query, annocement.Title, annocement.Content, annocement.CoverImg, annocement.Date, annocement.ID)
	if err != nil {
		log.Printf("Failed to update annocement: %v", err)
		http.Error(w, "Failed to update annocement", http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Announcement updated successfully"})
}
func DeleteAnnouncement(w http.ResponseWriter, r *http.Request) {
	var annocement Announcement
	if err := json.NewDecoder(r.Body).Decode(&annocement); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	query := `DELETE FROM announcements WHERE id = ?`
	_, err := db.Exec(query, annocement.ID)
	if err != nil {
		log.Printf("Failed to delete annocement: %v", err)
		http.Error(w, "Failed to delete annocement", http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Announcement deleted successfully"})
}

func FetchAnnouncements(w http.ResponseWriter, r *http.Request) {
	query := `SELECT id, title, content, coverImg,date FROM announcements`
	rows, err := db.Query(query)
	if err != nil {
		log.Printf("Query error: %s", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var announcements []Announcement
	for rows.Next() {
		var ann Announcement
		err := rows.Scan(&ann.ID, &ann.Title, &ann.Content, &ann.CoverImg, &ann.Date)
		if err != nil {
			log.Printf("Scan error: %s", err)
			http.Error(w, "Error reading data", http.StatusInternalServerError)
			return
		}
		announcements = append(announcements, ann)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error while scanning rows: %s", err)
		http.Error(w, "Error processing data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(announcements)
}
