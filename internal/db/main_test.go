package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/joho/godotenv"

	"github.com/ZadeNova/jobtrace/internal/config"
	_ "github.com/lib/pq"
)

var testDB *sql.DB

// New Test function to apply migrations to the test database.
// This is done to ensure that both the database and test databasee sync.
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

	if err := RunMigrations(testDB, "../../migrations"); err != nil {
		log.Fatalf("failed to run migrations on test db: %v", err)
	}

	code := m.Run()

	testDB.Close()
	os.Exit(code)
}
