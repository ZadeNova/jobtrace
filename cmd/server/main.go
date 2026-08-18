package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/joho/godotenv"

	"github.com/ZadeNova/jobtrace/internal/config"
	jobtracedb "github.com/ZadeNova/jobtrace/internal/db"
	_ "github.com/lib/pq"
)

func healthzHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func readyzHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		err := db.PingContext(ctx)
		
		if err != nil {
			log.Printf("readyz: db ping failed %v", err)
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("not ready"))

		} else {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("ok"))
		}
	}
}

func main() {

	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, relying on real environment variables")
	}

	cfg, err := config.Load()

	if err != nil {
		log.Fatalf("Failed to retrieve credentials: %v", err)
	}

	dsn := fmt.Sprintf(
		"host=localhost port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.PostgresPort, cfg.PostgresUser, cfg.PostgresPassword, cfg.PostgresDB,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to open db: %v", err)
	}
	defer db.Close()

	err = jobtracedb.RunMigrations(db)
	if err != nil {
		log.Fatalf("failed to run migration: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", healthzHandler)
	mux.HandleFunc("GET /readyz", readyzHandler(db))

	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}

}

