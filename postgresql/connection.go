package postgresql

import (
	"fmt"
	"time"

	"github.com/nesiler/cestx/common"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)


func NewPostgreSQLDB(cfg *Config) (*gorm.DB, error) {
	// Construct the database connection string
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName)

	// Connect to the database using GORM
	db, err := gorm.Open(postgres.Open(connStr), &gorm.Config{})
	if err != nil {
		return nil, common.Err("Failed to connect to PostgreSQL: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, common.Err("Failed to get SQL DB instance: %v", err)
	}

	sqlDB.SetMaxOpenConns(10)                 // Maximum number of open connections to the database.
	sqlDB.SetMaxIdleConns(5)                  // Maximum number of connections in the idle connection pool.
	sqlDB.SetConnMaxLifetime(time.Minute * 5) // Maximum lifetime of a connection.

	common.Ok("Connected to PostgreSQL database successfully!")
	return db, nil
}
