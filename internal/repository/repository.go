package repository

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

var DB *sql.DB

func InitDB() error {

	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "localhost"
	}
	port := os.Getenv("DB_PORT")
	if port == "" {
		port = "5432"
	}
	user := os.Getenv("DB_USER")
	if user == "" {
		user = "postgres"
	}
	password := os.Getenv("DB_PASSWORD")
	if password == "" {
		password = "postgres"
	}
	dbname := os.Getenv("DB_NAME")
	if dbname == "" {
		dbname = "neotchislyat"
	}

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		return err
	}

	return nil
}

type Subscription struct {
	ID          int       `json:"id"`
	ServiceName string    `json:"service_name"`
	Price       int       `json:"price"`
	UserID      uuid.UUID `json:"user_id"`
	StartDate   string    `json:"start_date"`
	EndDate     string    `json:"end_date,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

type CreateSubscriptionRequest struct {
	ServiceName string    `json:"service_name"`
	Price       int       `json:"price"`
	UserID      uuid.UUID `json:"user_id"`
	StartDate   string    `json:"start_date"`
	EndDate     string    `json:"end_date,omitempty"`
}

type UpdateRequest struct {
	ServiceName string `json:"service_name,omitempty"`
	Price       int    `json:"price,omitempty"`
	StartDate   string `json:"start_date,omitempty"`
	EndDate     string `json:"end_date,omitempty"`
}

func CreateSubscription(req CreateSubscriptionRequest) error {
	startDate, err := time.Parse("01-2006", req.StartDate)
	if err != nil {
		return fmt.Errorf("неверный формат start_date: %w", err)
	}

	var endDate sql.NullTime
	if req.EndDate != "" {
		dateEnd, err := time.Parse("01-2006", req.EndDate)
		if err != nil {
			log.Printf("неверный формат end_date: %w", err)
			return fmt.Errorf("неверный формат end_date: %w", err)
		}
		endDate = sql.NullTime{Time: dateEnd, Valid: true}
	}
	_, err = DB.Exec("INSERT INTO sub (user_id, service_name, price, start_date, end_date) VALUES ($1, $2, $3, $4, $5)",
		req.UserID, req.ServiceName, req.Price, startDate, endDate)
	if err != nil {
		return err
	}
	return nil
}

func UpdateSubscription(id int, req UpdateRequest) error {
	query := "UPDATE sub SET "
	args := []interface{}{}
	counter := 1

	if req.ServiceName != "" {
		query += fmt.Sprintf("service_name = $%d,", counter)
		args = append(args, req.ServiceName)
		counter++
	}
	if req.Price > 0 {
		query += fmt.Sprintf("price = $%d,", counter)
		args = append(args, req.Price)
		counter++
	}
	if req.StartDate != "" {
		startDate, err := time.Parse("01-2006", req.StartDate)
		if err != nil {
			return err
		}
		query += fmt.Sprintf("start_date = $%d,", counter)
		args = append(args, startDate)
		counter++
	}
	if req.EndDate != "" {
		endDate, err := time.Parse("01-2006", req.EndDate)
		if err != nil {
			return err
		}
		query += fmt.Sprintf("end_date = $%d,", counter)
		args = append(args, endDate)
		counter++
	}

	queryPart2 := fmt.Sprintf(" WHERE id = $%d", counter)
	if len(args) == 0 {
		return fmt.Errorf("нет полей для обновления")
	}
	query = query[:len(query)-1] + queryPart2
	args = append(args, id)

	_, err := DB.Exec(query, args...)
	if err != nil {
		return err
	}
	return nil
}

func DeleteSubscription(id int) error {
	_, err := DB.Exec("DELETE FROM sub WHERE id = $1", id)
	if err != nil {
		return err
	}
	return nil
}

func GetSub(id int) (Subscription, error) {
	var sub Subscription
	var endDate sql.NullTime
	var startDate time.Time
	var userID uuid.UUID

	err := DB.QueryRow(`SELECT id, service_name, price, user_id, start_date, end_date, created_at FROM sub WHERE id = $1`,
		id).Scan(&sub.ID, &sub.ServiceName, &sub.Price, &userID, &startDate, &endDate, &sub.CreatedAt)
	if err != nil {
		return Subscription{}, err
	}
	sub.UserID = userID

	sub.StartDate = startDate.Format("01-2006")
	if endDate.Valid {
		s := endDate.Time.Format("01-2006")
		sub.EndDate = s
	}

	return sub, nil
}

func GetSum(userID, serviceName, startDateStr, endDateStr string) (int, error) {
	startDate, err := time.Parse("01-2006", startDateStr)
	if err != nil {
		return 0, fmt.Errorf("неверный формат start_date")
	}

	endDate, err := time.Parse("01-2006", endDateStr)
	if err != nil {
		return 0, fmt.Errorf("неверный формат end_date")
	}
	endDate = endDate.AddDate(0, 1, -1)

	query := `SELECT SUM(price) FROM sub WHERE start_date >= $1 AND start_date <= $2`
	args := []interface{}{startDate, endDate}

	if userID != "" {
		userUUID, err := uuid.Parse(userID)
		if err != nil {
			return 0, fmt.Errorf("неверный формат user_id: %w", err)
		}
		query += " AND user_id = $" + strconv.Itoa(len(args)+1)
		args = append(args, userUUID)
	}
	if serviceName != "" {
		query += " AND service_name = $" + strconv.Itoa(len(args)+1)
		args = append(args, serviceName)
	}

	var sum int
	err = DB.QueryRow(query, args...).Scan(&sum)
	if err != nil {
		return 0, fmt.Errorf("ошибка запроса: %w", err)
	}

	return sum, nil
}

func ListSubscriptions(userId uuid.UUID) ([]Subscription, error) {
	rows, err := DB.Query(`SELECT id, service_name, price, user_id, start_date, end_date, created_at FROM sub WHERE user_id = $1 ORDER BY created_at DESC`, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Subscription
	for rows.Next() {
		var sub Subscription
		var endDate sql.NullTime
		var startDate time.Time

		err := rows.Scan(&sub.ID, &sub.ServiceName, &sub.Price, &sub.UserID, &startDate, &endDate, &sub.CreatedAt)
		if err != nil {
			return []Subscription{}, err
		}
		sub.StartDate = startDate.Format("01-2006")
		if endDate.Valid {
			s := endDate.Time.Format("01-2006")
			sub.EndDate = s
		}
		list = append(list, sub)
	}

	return list, nil
}
