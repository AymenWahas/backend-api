package database

import (
	"context"
	"testing"
	"time"

	"backend-api/internal/config"
)

func TestPostgreSQLRowLock(t *testing.T) {
	cfg := config.Config{
		DBHost:     "localhost",
		DBPort:     "5434",
		DBUser:     "postgres",
		DBPassword: "postgrespassword",
		DBName:     "employee_db",

		MaxOpenConns: 5,
		MaxIdleConns: 2,
	}

	db, err := NewPostgres(cfg)
	if err != nil {
		t.Fatalf("failed to connect to PostgreSQL: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("failed to get sql.DB: %v", err)
	}
	defer sqlDB.Close()

	ctx := context.Background()

	var projectID uint

	// Create a test project.
	err = db.WithContext(ctx).Exec(`
		INSERT INTO projects
			(owner_id, name, description, created_at, updated_at)
		VALUES
			(NULL, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id
	`, "Row Lock Test", "Testing PostgreSQL row locking").Error

	if err != nil {
		t.Fatalf("failed to create test project: %v", err)
	}

	err = db.WithContext(ctx).
		Raw(`
			SELECT id
			FROM projects
			WHERE name = ?
			ORDER BY id DESC
			LIMIT 1
		`, "Row Lock Test").
		Scan(&projectID).Error

	if err != nil {
		t.Fatalf("failed to get project ID: %v", err)
	}

	defer db.Exec(
		"DELETE FROM projects WHERE id = ?",
		projectID,
	)

	tx1 := db.Begin()
	if tx1.Error != nil {
		t.Fatalf("failed to begin transaction 1: %v", tx1.Error)
	}

	// Transaction 1 locks the row.
	var lockedID uint

	err = tx1.Raw(`
		SELECT id
		FROM projects
		WHERE id = ?
		FOR UPDATE
	`, projectID).Scan(&lockedID).Error

	if err != nil {
		tx1.Rollback()
		t.Fatalf("failed to lock project: %v", err)
	}

	if lockedID != projectID {
		tx1.Rollback()
		t.Fatalf("expected project ID %d, got %d", projectID, lockedID)
	}

	// Transaction 2 tries to update the locked row.
	tx2 := db.Begin()
	if tx2.Error != nil {
		tx1.Rollback()
		t.Fatalf("failed to begin transaction 2: %v", tx2.Error)
	}

	done := make(chan error, 1)

	go func() {
		err := tx2.Model(&struct{}{}).
			Exec(
				"UPDATE projects SET name = ? WHERE id = ?",
				"Blocked Update",
				projectID,
			).Error

		done <- err
	}()

	// Give transaction 2 time to reach the lock.
	select {
	case err := <-done:
		tx2.Rollback()
		tx1.Rollback()

		if err != nil {
			t.Fatalf("transaction 2 failed unexpectedly: %v", err)
		}

		t.Fatal("expected transaction 2 to be blocked by row lock")
	case <-time.After(200 * time.Millisecond):
		// Expected: transaction 2 is waiting for transaction 1.
	}

	// Release the row lock.
	if err := tx1.Commit().Error; err != nil {
		tx2.Rollback()
		t.Fatalf("failed to commit transaction 1: %v", err)
	}

	// Transaction 2 should now continue.
	select {
	case err := <-done:
		if err != nil {
			tx2.Rollback()
			t.Fatalf("transaction 2 failed after lock release: %v", err)
		}

		if err := tx2.Commit().Error; err != nil {
			t.Fatalf("failed to commit transaction 2: %v", err)
		}

	case <-time.After(2 * time.Second):
		tx2.Rollback()
		t.Fatal("transaction 2 remained blocked after transaction 1 committed")
	}

	t.Logf(
		"row lock verified: transaction 2 waited for project %d",
		projectID,
	)
}
