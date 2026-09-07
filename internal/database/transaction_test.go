package database

import (
	"context"
	"testing"

	"backend-api/internal/config"
)

func TestTransactionRollbackPostgreSQL(t *testing.T) {
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

	txManager := NewTransactionManager(db)

	var projectID uint

	err = txManager.WithinTransaction(
		context.Background(),
		func(ctx context.Context) error {
			tx := DBFromContext(ctx, db)

			// Step 1: Create Project with owner_id = NULL.
			result := tx.WithContext(ctx).Exec(`
				INSERT INTO projects
					(owner_id, name, description, created_at, updated_at)
				VALUES
					(NULL, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
				RETURNING id
			`, "Transaction Rollback Test", "Must not remain after rollback")

			if result.Error != nil {
				return result.Error
			}

			if err := tx.WithContext(ctx).
				Raw(`
					SELECT id
					FROM projects
					WHERE name = ?
					ORDER BY id DESC
					LIMIT 1
				`, "Transaction Rollback Test").
				Scan(&projectID).Error; err != nil {
				return err
			}

			// Step 2: Force a foreign-key violation.
			invalidTask := struct {
				ProjectID uint
				Title     string
				Status    string
			}{
				ProjectID: 999999999,
				Title:     "This task must fail",
				Status:    "pending",
			}

			if err := tx.WithContext(ctx).Exec(`
				INSERT INTO tasks
					(project_id, title, status, created_at, updated_at)
				VALUES
					(?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
			`,
				invalidTask.ProjectID,
				invalidTask.Title,
				invalidTask.Status,
			).Error; err != nil {
				return err
			}

			return nil
		},
	)

	if err == nil {
		t.Fatal("expected transaction to fail, got nil")
	}

	var count int64

	result := db.Table("projects").
		Where("id = ?", projectID).
		Count(&count)

	if result.Error != nil {
		t.Fatalf("failed to verify rollback: %v", result.Error)
	}

	if count != 0 {
		t.Fatalf(
			"rollback failed: project with id %d still exists",
			projectID,
		)
	}

	t.Logf(
		"rollback verified: project %d was removed after transaction failure",
		projectID,
	)
}
