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

	// Create a project and task that will be used only to reserve
	// an existing task ID for the duplicate-key failure.
	seedProject := &domain.Project{
		OwnerID:     ownerID,
		Name:        "Seed Project",
		Description: "used by transaction rollback test",
	}

	if err := db.Create(seedProject).Error; err != nil {
		t.Fatalf("failed to create seed project: %v", err)
	}

	seedTask := &domain.Task{
		ProjectID: seedProject.ID,
		Title:     "Existing Task",
		Status:    "pending",
	}

	if err := db.Create(seedTask).Error; err != nil {
		t.Fatalf("failed to create seed task: %v", err)
	}

	// Cleanup runs in reverse order:
	// seed task -> seed project -> test project -> employee.
	defer db.Exec("DELETE FROM employees WHERE id = ?", ownerID)
	defer db.Exec("DELETE FROM projects WHERE id = ?", seedProject.ID)
	defer db.Exec("DELETE FROM tasks WHERE id = ?", seedTask.ID)

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

	// Reuse an existing task ID.
	// The transaction will try to create another task with this ID,
	// causing a duplicate primary-key error.
	task := &domain.Task{
		ID:     seedTask.ID,
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
