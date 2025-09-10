package db

import (
	"os"
	"strconv"

	"gorm.io/gorm"
)

func GetDB() (*gorm.DB, error) {
	db_host := os.Getenv("DB_HOST")
	db_host_port := os.Getenv("DB_PORT")
	port, err := strconv.ParseUint(db_host_port, 10, 32)
	if err != nil {
		port = 5432 // Default PostgreSQL port
	}

	db_name := os.Getenv("DB_NAME")
	db_creds_secretID := os.Getenv("DB_SECRET_ID")
	return ConnectDataBase(uint(port), db_host, db_name, db_creds_secretID)
}
