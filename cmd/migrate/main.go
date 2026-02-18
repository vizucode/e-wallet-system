package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
	"github.com/vizucode/e-wallet-system/configs/database"
	"github.com/vizucode/e-wallet-system/internal/app/database/seeder"
)

func main() {

	runSeeder := flag.Bool("seed", false, "Run database seeder")
	flag.Parse()

	err := godotenv.Load("configs/.env")
	if err != nil {
		log.Println("Error loading .env file")
	}

	var (
		host     = os.Getenv("DB_HOST")
		port     = os.Getenv("DB_PORT")
		user     = os.Getenv("DB_USER")
		password = os.Getenv("DB_PASSWORD")
		dbname   = os.Getenv("DB_NAME")
	)

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, password, host, port, dbname)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	fmt.Println("🚀 Running migrations...")

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Fatal(err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://internal/app/database/migrations",
		"postgres",
		driver,
	)
	if err != nil {
		log.Fatal(err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatal(err)
	}

	fmt.Println("✅ Migration completed!")

	if *runSeeder {
		fmt.Println("🚀 Running seeder...")
		gormDB, err := database.NewDatabaseConnection(dsn)
		if err != nil {
			log.Fatal(err)
		}

		sqlDB, err := gormDB.DB()
		if err != nil {
			log.Fatal(err)
		}
		defer sqlDB.Close()

		if err := seeder.Seed(gormDB); err != nil {
			log.Fatal("Seeder failed!", err)
		}

		fmt.Println("✅ Seeding completed!")
	}
}
