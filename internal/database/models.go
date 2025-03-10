package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/viper"
)

type Models struct {
	Users  UserModel
	Points PointModel
	Gifts  GiftModel
	Shop   ItemModel
}

func New() *sql.DB {
	connStr := viper.GetString("bot.db")
	if connStr == "" {
		log.Fatal("bot.db env variable can't be empty!")
	}

	db, err := sql.Open("sqlite3", connStr)
	if err != nil {
		// This will not be a connection error, but a DSN parse error or
		// another initialization error.
		log.Fatal(err)
	}

	err = db.PingContext(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Connected to database!")
	return db
}

func NewModels(db *sql.DB) Models {
	return Models{
		Users:  UserModel{DB: db},
		Points: PointModel{DB: db},
		Gifts:  GiftModel{DB: db},
		Shop:   ItemModel{DB: db},
	}
}

// Migrate applies database migrations for an sqlite3 database.
// It reads migration files from the designated migration folder and
// ensures the database schema is updated accordingly.
func Migrate(db *sql.DB) error {
	driver, err := sqlite3.WithInstance(db, &sqlite3.Config{})
	if err != nil {
		return err
	}

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	migrationFilesPath := fmt.Sprintf("file://%s/migrations", cwd)

	m, err := migrate.NewWithDatabaseInstance(migrationFilesPath, viper.GetString("bot.db"), driver)
	if err != nil {
		return fmt.Errorf("migration initialization failed: %v", err)
	}

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		version, dirty, _ := m.Version()
		if dirty {
			log.Printf("Database is in a dirty state at version %d. Forcing reset...", version)
			_ = m.Force(int(version)) // Reset to last successful version

			//retry migration
			err = m.Up()
			if err != nil && err != migrate.ErrNoChange {
				return fmt.Errorf("migration failed after fixing dirty state: %v", err)
			}
			return nil
		}
		return fmt.Errorf("migration failed: %v", err)
	}

	return nil
}
