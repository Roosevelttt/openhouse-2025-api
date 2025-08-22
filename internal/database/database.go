package database

import (
	"database/sql"
	"fmt"
	"log"

	"openhouse-2025-api/internal/config"

	_ "github.com/go-sql-driver/mysql" // Standard SQL Driver
	"gorm.io/driver/mysql"             // GORM MySQL Driver
	"gorm.io/gorm"
)

var (
	// SqlDB will hold the standard SQL connection for raw queries
	SqlDB *sql.DB
	// GormDB will hold the GORM instance for ORM operations
	GormDB *gorm.DB
)

// Connect initializes both the standard SQL and GORM database connections.
func Connect(cfg *config.Config) {
	// 1. Create the standard *sql.DB connection pool
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4&loc=Local",
		cfg.DBUser, cfg.DBPass, cfg.DBHost, cfg.DBPort, cfg.DBName)

	sqlDb, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Failed to open SQL connection: %v", err)
	}

	if err := sqlDb.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	SqlDB = sqlDb // Assign to the global variable
	log.Println("Standard SQL database connection successful.")

	// 2. Create the *gorm.DB instance using the existing *sql.DB connection
	gormDb, err := gorm.Open(mysql.New(mysql.Config{
		Conn: SqlDB,
	}), &gorm.Config{})

	if err != nil {
		log.Fatalf("Failed to initialize GORM: %v", err)
	}

	GormDB = gormDb // Assign to the global variable
	log.Println("GORM database connection successful.")
}
