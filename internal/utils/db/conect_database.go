package db

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func ConnectDataBase(port uint, host, dbname, secretID string) (*gorm.DB, error) {
	sslDisabled := os.Getenv("DB_SSL_MODE_DISABLE")
	var sslMode string
	if sslDisabled == "true" {
		sslMode = " sslmode=disable"
	}
	username, password := retrieveCredentials(secretID)
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d%s", host, username, password, dbname, port, sslMode)
	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Error),
	})
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	return database, nil
}
