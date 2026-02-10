package database

import (
	"errors"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"

	// source/file import is required for migration files to read
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"

	// load pq as database driver
	_ "github.com/lib/pq"
)

var (
	Todo *sqlx.DB
)

type SSLMode string

const (
	SSLModeDisabled SSLMode = "disable"
	SSLModeEnabled  SSLMode = "enable"
)

func CreateAndMigrate(host, port, user, password, dbname string, sslmode SSLMode) error {
	connStr := fmt.Sprintf("host =%s port = %s user =%s password =%s dbname =%s sslmode=%s ", host, port, user, password, dbname, sslmode)
	DataBase, err := sqlx.Open("postgres", connStr)
	if err != nil {
		return err
	}
	err = DataBase.Ping()
	if err != nil {
		return err
	}
	Todo = DataBase
	return migrateUp(DataBase)
}
func migrateUp(db *sqlx.DB) error {
	driver, err := postgres.WithInstance(db.DB, &postgres.Config{})
	if err != nil {
		return err
	}
	m, err := migrate.NewWithDatabaseInstance(
		"file://database/migration",
		"postgres", driver)

	if err != nil {
		return err
	}
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return nil
}
func Tx(fn func(tx *sqlx.Tx) error) error {
	tx, err := Todo.Beginx()
	if err != nil {
		return fmt.Errorf("failed to begin tx: %v", err)
	}
	defer func() {
		if err != nil {
			if rollBackErr := tx.Rollback(); rollBackErr != nil {
				log.Printf("failed to rollback tx: %v", rollBackErr)
			}
			return
		}
		if err := tx.Commit(); err != nil {
			log.Printf("failed to commit tx: %v", err)
		}
	}()
	err = fn(tx)
	return err
}
