package repositories

import (
	"database/sql"
	"fmt"

	"gorm.io/driver/mysql" // GORM MySQL Driver
	"gorm.io/gorm"

	"openhouse-2025-api/internal/config"

	_ "github.com/go-sql-driver/mysql" // Standard SQL Driver
)

// This function will now create and return both connection types.
func NewDatabaseConnections(cfg *config.Config) (*sql.DB, *gorm.DB, error) {
	// 1. Create the standard *sql.DB connection pool
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4&loc=Local", cfg.DBUser, cfg.DBPass, cfg.DBHost, cfg.DBPort, cfg.DBName)

	sqlDB, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, nil, err
	}
	if err := sqlDB.Ping(); err != nil {
		return nil, nil, err
	}

	// 2. Create the *gorm.DB instance using the existing *sql.DB
	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn: sqlDB,
	}), &gorm.Config{})
	if err != nil {
		return nil, nil, err
	}

	// 3. Return both connections successfully
	return sqlDB, gormDB, nil
}
