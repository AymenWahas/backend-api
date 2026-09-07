package postgres

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"backend-api/internal/config"
	"backend-api/internal/database"
	"backend-api/internal/domain"
)

func TestProjectOptimisticConcurrency(t *testing.T) {
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

	repo := NewProjectRepository(db)

	// Create a test employee to use as project owner.
	var ownerID int

	email := fmt.Sprintf(
		"optimistic-owner-%d@example.com",
		time.Now().UnixNano(),
	)

	result := db.Raw(`
		INSERT INTO employees (name, email)
		VALUES (?, ?)
		RETURNING id
	`, "Optimistic Owner", email).Scan(&ownerID)

	if result.Error != nil {
		t.Fatalf("failed to create test employee: %v", result.Error)
	}

	// Create a test project.
	project := &domain.Project{
		OwnerID:     ownerID,
		Name:        "Optimistic Test",
		Description: "original",
	}

	result = db.Raw(`
		INSERT INTO projects (name, description, owner_id, version)
		VALUES (?, ?, ?, 1)
		RETURNING id
	`, project.Name, project.Description, project.OwnerID).Scan(&project.ID)

	if result.Error != nil {
		t.Fatalf("failed to create project: %v", result.Error)
	}

	// Cleanup.
	defer db.Exec("DELETE FROM employees WHERE id = ?", ownerID)
	defer db.Exec("DELETE FROM projects WHERE id = ?", project.ID)
	// Read the project twice.
	project1, err := repo.GetByID(context.Background(), project.ID)
	if err != nil {
		t.Fatalf("failed to read project1: %v", err)
	}

	project2, err := repo.GetByID(context.Background(), project.ID)
	if err != nil {
		t.Fatalf("failed to read project2: %v", err)
	}

	t.Logf("TX1 version: %d", project1.Version)
	t.Logf("TX2 version: %d", project2.Version)

	// Both transactions have the same version.
	if project1.Version != 1 || project2.Version != 1 {
		t.Fatalf(
			"expected both projects to have version 1, got TX1=%d TX2=%d",
			project1.Version,
			project2.Version,
		)
	}

	// TX1 updates successfully.
	project1.Name = "Project A"

	if err := repo.Update(context.Background(), project1); err != nil {
		t.Fatalf("TX1 update failed: %v", err)
	}

	t.Logf(
		"TX1 update succeeded: version %d → %d",
		1,
		project1.Version,
	)

	// TX2 still has the old version.
	project2.Name = "Project B"

	err = repo.Update(context.Background(), project2)

	if !errors.Is(err, domain.ErrProjectConflict) {
		t.Fatalf(
			"expected ErrProjectConflict, got: %v",
			err,
		)
	}

	t.Log("TX2 update rejected: stale version detected")

	// Verify final database state.
	finalProject, err := repo.GetByID(context.Background(), project.ID)
	if err != nil {
		t.Fatalf("failed to read final project: %v", err)
	}

	t.Logf(
		"Final project: name=%q version=%d",
		finalProject.Name,
		finalProject.Version,
	)

	if finalProject.Name != "Project A" {
		t.Fatalf(
			"expected final name %q, got %q",
			"Project A",
			finalProject.Name,
		)
	}

	if finalProject.Version != 2 {
		t.Fatalf(
			"expected final version 2, got %d",
			finalProject.Version,
		)
	}

	t.Log("optimistic concurrency verified: lost update prevented")
}
