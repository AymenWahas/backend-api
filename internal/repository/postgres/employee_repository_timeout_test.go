package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	"backend-api/internal/config"
	"backend-api/internal/database"
)

func TestPostgreSQLContextTimeout(t *testing.T) {
	cfg := config.Config{
		DBHost:       "localhost",
		DBPort:       "5434",
		DBUser:       "postgres",
		DBPassword:   "postgrespassword",
		DBName:       "employee_db",
		MaxOpenConns: 5,
		MaxIdleConns: 2,
	}

	db, err := database.NewPostgres(cfg)
	if err != nil {
		t.Fatalf("failed to connect to PostgreSQL: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("failed to get sql.DB: %v", err)
	}
	defer sqlDB.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	var result int

	err = db.WithContext(ctx).
		Raw("SELECT 1 FROM pg_sleep(2)").
		Scan(&result).Error

	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}

	if !errors.Is(err, context.DeadlineExceeded) &&
		!errors.Is(ctx.Err(), context.DeadlineExceeded) {
		t.Fatalf("expected context deadline exceeded, got: %v", err)
	}
}
