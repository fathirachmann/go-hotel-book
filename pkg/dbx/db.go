package dbx

import (
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func OpenDatabase(schemaEnv string) *gorm.DB {
	dsn := os.Getenv("DB_DSN")
	schema := os.Getenv(schemaEnv)

	if schema != "" {
		dsn = dsn + "&search_path=" + schema
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		panic(err)
	}

	return db
}
