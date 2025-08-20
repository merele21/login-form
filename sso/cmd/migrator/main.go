package main

import (
	"errors"
	"flag"
	"fmt"

	// migrations library
	"github.com/golang-migrate/migrate/v4"
	// driver for run migrations postgresql
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	// driver for takes migrations from files
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	var dsn, migrationsPath string // migrationsTable для функциональных тестов

	flag.StringVar(&dsn, "storage-path", "", "postgres connection string")
	flag.StringVar(&migrationsPath, "migrations-path", "", "path to migrations")
	flag.Parse()

	if dsn == "" {
		panic("storage-path is required")
	}

	if migrationsPath == "" {
		panic("migrations-path is required")
	}

	m, err := migrate.New("file://"+migrationsPath, dsn)
	if err != nil {
		panic(err)
	}

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			fmt.Println("no migrations to apply")

			return
		}

		panic(err)
	}

	fmt.Println("migrations applied successfully")
}
