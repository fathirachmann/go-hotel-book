package dbx

import (
	"errors"
	"fmt"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func InitDatabase(schemaEnv string) (*gorm.DB, error) {
	dsn := os.Getenv("DB_DSN")

	fmt.Println(dsn, "dsn")

	if dsn == "" {
		return nil, errors.New("DB_DSN env is missing")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix: os.Getenv("DB_SCHEMA"),
		},
	})

	if err != nil {
		return nil, err
	}

	return db, nil
}
