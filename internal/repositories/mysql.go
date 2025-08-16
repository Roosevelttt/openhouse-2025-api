package repositories

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"openhouse-2025-api/internal/config"
)

func MustConnectMySQL(cfg *config.Config) *sql.DB {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4&loc=Local", cfg.DBUser, cfg.DBPass, cfg.DBHost, cfg.DBPort, cfg.DBName)
	db, err := sql.Open("mysql", dsn)
	if err != nil { panic(err) }
	if err := db.Ping(); err != nil { panic(err) }
	return db
}

