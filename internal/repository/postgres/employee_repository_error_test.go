package postgres

import (
	"context"
	"errors"
	"testing"

	"backend-api/internal/config"
	"backend-api/internal/database"
	"backend-api/internal/domain"
)

func TestEmployeeRepositoryDuplicateEmail(t *testing.T) {
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

	repo := NewEmployeeRepository(db)

	ctx := context.Background()

	email := "duplicate-test@example.com"

	employee1 := domain.Employee{
		Name:  "Duplicate Test 1",
		Email: email,
	}

	employee2 := domain.Employee{
		Name:  "Duplicate Test 2",
		Email: email,
	}

	created, err := repo.Create(ctx, employee1)
	if err != nil {
		t.Fatalf("failed to create first employee: %v", err)
	}

	defer repo.Delete(ctx, created.ID)

	_, err = repo.Create(ctx, employee2)

	if err == nil {
		t.Fatal("expected duplicate email error, got nil")
	}

	if !errors.Is(err, domain.ErrEmployeeAlreadyExists) {
		t.Fatalf(
			"expected ErrEmployeeAlreadyExists, got: %v",
			err,
		)
	}
}
