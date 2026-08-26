// GET /orders

package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

type OrderResponse struct {
	ID       int       `json:"id"`
	Datetime time.Time `json:"datetime"`
	AN       int
}

type ErrorResponse struct {
	Error string `json:"error"`
}

var db *sql.DB

func main() {
	var err error

	connStr := os.Getenv("DATABASE_URL")

	fmt.Println(connStr)

	if connStr == "" {
		connStr = "postgres://user:password@localhost:5432/postgres"
	}

	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Error connect to DB: %v", err)
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		log.Fatalf("DB ping error: %v", err)
	}

	// Роутер
	r := mux.NewRouter()

	// Маршруты
	r.HandleFunc("/orders", orderHandler).Methods(http.MethodGet)
	log.Println("Start server")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatalf("Error started server: %v", err)
	}

}

func orderHandler(w http.ResponseWriter, r *http.Request) {
	var res OrderResponse

	w.Header().Set("Content-Type", "application/json")

	// Параметр
	an := r.URL.Query().Get("an")

	log.Println("URL param `an` ->>>", an, "<<<-")

	// Сделаем параметр обязательным
	if an == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Пропущен обязательный параметр"})
		return
	}

	// Точное совпадение
	query := "Select * from \"Orders\" where ord_an = $1 limit 100"
	// важен порядок в Scan
	err := db.QueryRow(query, an).Scan(&res.ID, &res.Datetime, &res.AN)

	if err == sql.ErrNoRows {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Ничего не найдено"})
		log.Println(err)
		return
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Ошибка сервера БД"})
		log.Println(err)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}
