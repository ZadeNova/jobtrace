package handler

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/joho/godotenv"

	"github.com/ZadeNova/jobtrace/internal/config"
	jobtracedb "github.com/ZadeNova/jobtrace/internal/db"
	_ "github.com/lib/pq"
)

var testDB *sql.DB

func TestMain(m *testing.M) {
	_ = godotenv.Load("../../.env")

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	dsn := fmt.Sprintf(
		"host=localhost port=%s user=%s password=%s dbname=jobtrace_test sslmode=disable",
		cfg.PostgresPort, cfg.PostgresUser, cfg.PostgresPassword,
	)

	testDB, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("failed to open test db: %v", err)
	}

	if err := jobtracedb.RunMigrations(testDB, "../../migrations"); err != nil {
		log.Fatalf("failed to run migrations on test db: %v", err)
	}

	code := m.Run()

	testDB.Close()
	os.Exit(code)
}
