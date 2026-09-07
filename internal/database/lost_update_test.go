package database

import (
	"testing"

	"backend-api/internal/config"
)

func TestPostgreSQLLostUpdate(t *testing.T) {
	cfg := config.Config{
		DBHost:     "localhost",
		DBPort:     "5434",
		DBUser:     "postgres",
		DBPassword: "postgrespassword",
		DBName:     "employee_db",
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

	// Create test project.
	var projectID uint

	err = db.Raw(`
		INSERT INTO projects (name, description, owner_id)
		VALUES ('Lost Update Test', 'test', NULL)
		RETURNING id
	`).Scan(&projectID).Error

	if err != nil {
		t.Fatalf("failed to create test project: %v", err)
	}

	defer db.Exec("DELETE FROM projects WHERE id = ?", projectID)

	// Transaction 1.
	tx1 := db.Begin()
	if tx1.Error != nil {
		t.Fatalf("failed to begin tx1: %v", tx1.Error)
	}

	// Transaction 2.
	tx2 := db.Begin()
	if tx2.Error != nil {
		tx1.Rollback()
		t.Fatalf("failed to begin tx2: %v", tx2.Error)
	}

	var name1 string
	var name2 string

	// Both transactions read the same old value.
	if err := tx1.Raw(
		"SELECT name FROM projects WHERE id = ?",
		projectID,
	).Scan(&name1).Error; err != nil {
		tx1.Rollback()
		tx2.Rollback()
		t.Fatalf("tx1 read failed: %v", err)
	}

	if err := tx2.Raw(
		"SELECT name FROM projects WHERE id = ?",
		projectID,
	).Scan(&name2).Error; err != nil {
		tx1.Rollback()
		tx2.Rollback()
		t.Fatalf("tx2 read failed: %v", err)
	}

	t.Logf("TX1 read: %q", name1)
	t.Logf("TX2 read: %q", name2)

	// TX1 updates using the value it previously read.
	if err := tx1.Exec(
		"UPDATE projects SET name = ? WHERE id = ?",
		"Project A",
		projectID,
	).Error; err != nil {
		tx1.Rollback()
		tx2.Rollback()
		t.Fatalf("tx1 update failed: %v", err)
	}

	if err := tx1.Commit().Error; err != nil {
		tx2.Rollback()
		t.Fatalf("tx1 commit failed: %v", err)
	}

	// TX2 still has stale data and writes its own value.
	if err := tx2.Exec(
		"UPDATE projects SET name = ? WHERE id = ?",
		"Project B",
		projectID,
	).Error; err != nil {
		tx2.Rollback()
		t.Fatalf("tx2 update failed: %v", err)
	}

	if err := tx2.Commit().Error; err != nil {
		t.Fatalf("tx2 commit failed: %v", err)
	}

	// Check final value.
	var finalName string

	if err := db.Raw(
		"SELECT name FROM projects WHERE id = ?",
		projectID,
	).Scan(&finalName).Error; err != nil {
		t.Fatalf("failed to read final value: %v", err)
	}

	t.Logf("Final database value: %q", finalName)

	if finalName != "Project B" {
		t.Fatalf(
			"expected final value %q, got %q",
			"Project B",
			finalName,
		)
	}

	t.Log("lost update verified: TX2 overwrote TX1")
}
