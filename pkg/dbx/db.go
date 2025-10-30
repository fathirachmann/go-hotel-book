package dbx

import (
	"errors"
	"fmt"
	"os"
	"strings"

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

	schemaName := os.Getenv("DB_SCHEMA")
	schemaPrefix := ""
	if strings.TrimSpace(schemaName) != "" {
		// Use schemaName as Postgres schema by prefixing with dot
		schemaPrefix = schemaName + "."
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix: schemaPrefix,
		},
	})
	if err != nil {
		return nil, err
	}

	// Ensure schema exists (safe if already exists)
	if schemaName != "" {
		dbExec := db.Exec("CREATE SCHEMA IF NOT EXISTS \"" + schemaName + "\"")
		if dbExec.Error != nil {
			return nil, dbExec.Error
		}
	}

	return db, nil
}
