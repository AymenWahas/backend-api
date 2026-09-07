package usecase

import (
	"context"
	"fmt"
	"testing"
	"time"

	"backend-api/internal/config"
	"backend-api/internal/database"
	"backend-api/internal/domain"
	"backend-api/internal/repository/postgres"
)

func TestPostgresCreateProjectWithTaskRollback(t *testing.T) {
	cfg := config.Config{
		DBHost:     "localhost",
		DBPort:     "5434",
		DBUser:     "postgres",
		DBPassword: "postgrespassword",
		DBName:     "employee_db",
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

	// Create a real employee for the project owner.
	var ownerID int

	email := fmt.Sprintf(
		"transaction-test-owner-%d@example.com",
		time.Now().UnixNano(),
	)

	result := db.Raw(`
		INSERT INTO employees (name, email)
		VALUES (?, ?)
		RETURNING id
	`, "Transaction Test Owner", email).Scan(&ownerID)

	if result.Error != nil {
		t.Fatalf("failed to create test employee: %v", result.Error)
	}

	// Create repositories and transaction manager.
	projectRepo := postgres.NewProjectRepository(db)
	taskRepo := postgres.NewTaskRepository(db)
	txManager := database.NewTransactionManager(db)

	u := NewProjectTaskUsecase(
		projectRepo,
		taskRepo,
		txManager,
	)

	project := &domain.Project{
		OwnerID:     ownerID,
		Name:        "Atomic Project",
		Description: "should rollback",
	}

	// Cleanup runs in reverse order:
	// project first, then employee.
	defer db.Exec("DELETE FROM employees WHERE id = ?", ownerID)
	defer db.Exec("DELETE FROM projects WHERE id = ?", project.ID)

	// Duplicate task ID will force the task INSERT to fail.
	task := &domain.Task{
		ID:     1,
		Title:  "Duplicate Task",
		Status: "pending",
	}

	err = u.CreateProjectWithTask(
		context.Background(),
		project,
		task,
	)

	if err == nil {
		t.Fatal("expected transaction to fail")
	}

	t.Logf("transaction failed as expected: %v", err)

	// The project must NOT exist because the transaction must rollback.
	var count int64

	result = db.Raw(`
		SELECT COUNT(*)
		FROM projects
		WHERE id = ?
	`, project.ID).Scan(&count)

	if result.Error != nil {
		t.Fatalf("failed to verify rollback: %v", result.Error)
	}

	if count != 0 {
		t.Fatalf(
			"expected project to be rolled back, but found %d row(s)",
			count,
		)
	}

	t.Log("rollback verified: project was not committed")
}
