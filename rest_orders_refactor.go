// Refactor
// GET /orders

package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

type OrderResponse struct {
	ID       int       `json:"id"`
	Datetime time.Time `json:"datetime"`
	AN       int       `json:"an"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

// var db *sql.DB
type server struct {
	db     *sql.DB
	logger *slog.Logger
}

func main() {
	// var err error
	// Структурированный логгер
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	connStr := os.Getenv("DATABASE_URL")

	fmt.Println(connStr)

	if connStr == "" {
		connStr = "postgres://user:password@localhost:5432/postgres"
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		// log.Fatalf("Error connect to DB: %v", err)
		logger.Error("Failed to open DB connection", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// Настройка пула соединений
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Проверка связи с таймаутом
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err = db.PingContext(ctx); err != nil {
		logger.Error("DB ping failed", "error", err)
		os.Exit(1)
	}

	// if err = db.Ping(); err != nil {
	// 	log.Fatalf("DB ping error: %v", err)
	// }

	srv := &server{
		db:     db,
		logger: logger,
	}

	// Роутер
	r := mux.NewRouter()

	// Маршруты
	// r.HandleFunc("/orders", orderHandler).Methods(http.MethodGet)
	// Явно передаем зависимости через метод структуры
	r.HandleFunc("/orders", srv.orderHandler).Methods(http.MethodGet)

	// log.Println("Start server")
	// if err := http.ListenAndServe(":8080", r); err != nil {
	// 	log.Fatalf("Error started server: %v", err)
	// }

	httpSrv := &http.Server{
		Addr:         ":8080",
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful Shutdown (Плавная остановка сервера)
	go func() {
		logger.Info("Starting server on :8080")
		if err := httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Info("Server standart listen error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := httpSrv.Shutdown(shutdownCtx); err != nil {
		logger.Info("Server forced to shutdown", "error", err)
	}
	logger.Info("Server exited gracefully")

}

func (s *server) orderHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	// Параметр
	anStr := r.URL.Query().Get("an")
	log.Println("URL param `an` ->>>", anStr, "<<<-")
	if anStr == "" {
		s.respondWithError(w, http.StatusBadRequest, "Пропущен обязательный параметр 'an'")
		return
	}

	// Валидация: проверяем, что 'an' — это число, раз в БД мы пишем его в int (res.AN)
	an, err := strconv.Atoi(anStr)
	if err != nil {
		s.respondWithError(w, http.StatusBadRequest, "Параметр 'an' должен быть числом")
		return
	}

	// Использование Context для отмены тяжелых запросов со стороны клиента
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	query := `SELECT ord_id, ord_datetime, ord_an FROM "Orders" WHERE ord_an = $1 LIMIT 1`

	var res OrderResponse

	err = s.db.QueryRowContext(ctx, query, an).Scan(&res.ID, &res.Datetime, &res.AN)

	if errors.Is(err, sql.ErrNoRows) {
		s.respondWithError(w, http.StatusNotFound, "Ничего не найдено")
		return
	} else if err != nil {
		s.logger.Error("Database query error", "error", err, "an", an)
		s.respondWithError(w, http.StatusInternalServerError, "Внутренняя ошибка сервера")
		return
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(res); err != nil {
		s.logger.Error("Failed to encode response", "error", err)
	}
}

// Хелпер для стандартизации ответов об ошибках
func (s *server) respondWithError(w http.ResponseWriter, code int, message string) {
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(ErrorResponse{Error: message})
}

// func orderHandler(w http.ResponseWriter, r *http.Request) {
// 	var res OrderResponse

// 	w.Header().Set("Content-Type", "application/json")

// 	// Параметр
// 	an := r.URL.Query().Get("an")

// 	log.Println("URL param `an` ->>>", an, "<<<-")

// 	// Сделаем параметр обязательным
// 	if an == "" {
// 		w.WriteHeader(http.StatusBadRequest)
// 		json.NewEncoder(w).Encode(ErrorResponse{Error: "Пропущен обязательный параметр"})
// 		return
// 	}

// 	// Точное совпадение
// 	query := "Select * from \"Orders\" where ord_an = $1 limit 100"
// 	// важен порядок в Scan
// 	err := db.QueryRow(query, an).Scan(&res.ID, &res.Datetime, &res.AN)

// 	if err == sql.ErrNoRows {
// 		w.WriteHeader(http.StatusNotFound)
// 		json.NewEncoder(w).Encode(ErrorResponse{Error: "Ничего не найдено"})
// 		log.Println(err)
// 		return
// 	} else if err != nil {
// 		w.WriteHeader(http.StatusInternalServerError)
// 		json.NewEncoder(w).Encode(ErrorResponse{Error: "Ошибка сервера БД"})
// 		log.Println(err)
// 		return
// 	}

// 	w.WriteHeader(http.StatusOK)
// 	json.NewEncoder(w).Encode(res)
// }
