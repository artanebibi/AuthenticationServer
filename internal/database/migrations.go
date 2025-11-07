package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
)

func RunMigrations(db *sql.DB) {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	} else {
		//log.Print("Loading .env file")
	}

	var migrationsPath string

	migrationsPath, err = filepath.Abs("./internal/database/migrations")
	migrationsPath = filepath.ToSlash(migrationsPath)

	if err != nil {
		log.Fatalf("Failed to get absolute path: %v", err)
	}

	//if os.Getenv("BUILD_ENV") == "docker" {
	//	migrationsPath = "/app/internal/database/migrations"
	//} else {
	//	var err error = nil
	//	migrationsPath, err = filepath.Abs("./internal/database/migrations")
	//	if err != nil {
	//		log.Fatalf("Failed to get absolute path: %v", err)
	//	}
	//}

	if _, err := os.Stat(migrationsPath); os.IsNotExist(err) {
		//pwd, _ := os.Getwd()
		//dir, _ := os.ReadDir(fmt.Sprintf("%s/main", pwd))
		//log.Fatalf("You are here: %s, and the dir\n%s", pwd, dir)

		log.Fatalf("Migrations directory not found at: %s", migrationsPath)
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Fatalf("Failed to create migration driver: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", migrationsPath),
		"authdb",
		driver,
	)

	if err != nil {
		log.Fatalf("Failed to initialize migrations: %v", err)
	}
	if m == nil {
		log.Fatal("Migration instance is nil")
	}

	// Run migrations
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("Migration failed: %v", err)
	}

	log.Print("Migrations completed successfully")

}
