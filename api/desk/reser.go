package desk

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

// Reservation 结构体定义预订信息
type Reservation struct {
	ID              int     `json:"id"`
	Name            string  `json:"name"`
	NumOfPeople     int     `json:"numOfPeople"`
	ReservationTime string  `json:"reservationTime"`
	Status          string  `json:"status"`
	TableID         int     `json:"tableId"`
	Remarks         *string `json:"remarks"` // 将Remarks字段声明为指针类型
}

// GetAllReservations 查询所有预订信息
func GetAllReservations(db *sql.DB) ([]Reservation, error) {
	var reservations []Reservation

	rows, err := db.Query("SELECT * FROM reservationlist")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var reservation Reservation
		var remarks sql.NullString // 使用sql.NullString来处理可能为NULL的字符串

		if err := rows.Scan(&reservation.ID, &reservation.Name, &reservation.NumOfPeople, &reservation.ReservationTime, &reservation.Status, &reservation.TableID, &remarks); err != nil {
			return nil, err
		}

		// 检查remarks是否为有效值
		if remarks.Valid {
			reservation.Remarks = &remarks.String // 如果有效，则将其赋值给Reservation.Remarks
		} else {
			// 如果为NULL，则将Reservation.Remarks置为nil
			reservation.Remarks = nil
		}

		reservations = append(reservations, reservation)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return reservations, nil
}

func HandleGetAllReservations(w http.ResponseWriter, r *http.Request) {
	// 设置响应头为 JSON 格式
	w.Header().Set("Content-Type", "application/json")

	// 检查请求方法是否为 GET
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 调用 desk 包中的 GetAllReservations
	reservations, err := GetAllReservations(db)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 返回查询到的全部预订信息（此处应有适当的 JSON 格式化和错误处理）
	jsonReservations, err := json.Marshal(reservations)
	if err != nil {
		http.Error(w, "Failed to marshal JSON", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(jsonReservations)
}

// GetMaxReservationID 获取当前最大的预订ID
func GetMaxReservationID(db *sql.DB) (int, error) {
	var maxID int

	err := db.QueryRow("SELECT MAX(ID) FROM reservationlist").Scan(&maxID)
	if err != nil {
		// 如果没有预订信息或出现其他错误，假设最大ID为0
		return 0, err
	}

	return maxID, nil
}

// AddReservation 添加预订信息
func AddReservation(w http.ResponseWriter, r *http.Request) {
	// 解析请求体中的 JSON 数据
	var reservation Reservation
	if err := json.NewDecoder(r.Body).Decode(&reservation); err != nil {
		log.Println("Failed to decode JSON:", err)
		http.Error(w, "Failed to decode JSON", http.StatusBadRequest)
		return
	}

	// 获取当前最大的预订ID，并加一作为新的预订信息的ID
	maxID, err := GetMaxReservationID(db)
	if err != nil {
		log.Println("Failed to get max reservation ID:", err)
		http.Error(w, "Failed to get max reservation ID", http.StatusInternalServerError)
		return
	}
	reservation.ID = maxID + 1

	// 将字符串解析为时间
	parsedTime, err := time.Parse(time.RFC3339, reservation.ReservationTime)
	if err != nil {
		log.Println("Failed to parse reservation time:", err)
		http.Error(w, "Failed to parse reservation time", http.StatusBadRequest)
		return
	}

	// 将时间格式化为数据库所需的日期时间格式
	formattedTime := parsedTime.Format("2006-01-02 15:04:05")

	// 插入预订信息到数据库
	_, err = db.Exec("INSERT INTO reservationList (ID, Name, NumOfPeople, ReservationTime, Status, Table_ID, Remarks) VALUES (?, ?, ?, ?, ?, ?, ?)",
		reservation.ID, reservation.Name, reservation.NumOfPeople, formattedTime, reservation.Status, reservation.TableID, reservation.Remarks)
	if err != nil {
		log.Println("Failed to insert reservation into database:", err)
		http.Error(w, "Failed to insert reservation into database", http.StatusInternalServerError)
		return
	}

	// 打印预订成功信息
	log.Println("Reservation added successfully")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Reservation added successfully"))
}

// DeleteReservation 根据预订ID删除预订信息
func DeleteReservation(w http.ResponseWriter, r *http.Request) {
	// 设置响应头为 JSON 格式
	w.Header().Set("Content-Type", "application/json")

	// 检查请求方法是否为 POST
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 解析请求体中的 JSON 数据来获取 reservationId
	var data struct {
		ReservationId int `json:"reservationId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Failed to decode JSON", http.StatusBadRequest)
		return
	}

	// 在数据库中执行删除预订信息的操作
	_, err := db.Exec("DELETE FROM reservationList WHERE ID = ?", data.ReservationId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 返回成功响应
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Reservation deleted successfully"))
}
