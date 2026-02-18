package main

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/vizucode/e-wallet-system/internal/app"
)

func main() {
	if err := godotenv.Load("configs/.env"); err != nil {
		log.Println("Error loading .env file")
	}

	app.Run()
}
