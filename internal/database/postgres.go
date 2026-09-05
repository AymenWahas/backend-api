package database

import (
	"database/sql"
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"backend-api/internal/config"
	"backend-api/internal/domain"
)

func NewPostgres(cfg config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBName,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql db: %w", err)
	}

	// Connection Pool
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping postgres: %w", err)
	}

	// Auto-migration
	if err := autoMigrate(db); err != nil {
		return nil, fmt.Errorf("failed to auto-migrate: %w", err)
	}

	return db, nil
}

func autoMigrate(db *gorm.DB) error {
	// ترتيب المigration مهم بسبب الـ Foreign Keys
	if err := db.AutoMigrate(&domain.Project{}); err != nil {
		return fmt.Errorf("failed to migrate Project: %w", err)
	}

	if err := db.AutoMigrate(&domain.Task{}); err != nil {
		return fmt.Errorf("failed to migrate Task: %w", err)
	}

	if err := db.AutoMigrate(&domain.Employee{}); err != nil {
		return fmt.Errorf("failed to migrate Employee: %w", err)
	}

	return nil
}

func Stats(db *gorm.DB) sql.DBStats {
	sqlDB, err := db.DB()
	if err != nil {
		return sql.DBStats{}
	}
	return sqlDB.Stats()
}
