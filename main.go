package main

import (
	"github.com/joho/godotenv"
	"log"
	"os"
)

func main() {
	a := App{}

	err := godotenv.Load("properties.env")
	if err != nil {
		log.Fatal("Error loading properties.env")
	}

	a.Initialize(
		os.Getenv("APP_DB_USERNAME"),
		os.Getenv("APP_DB_PASSWORD"),
		os.Getenv("APP_DB_NAME"))

	a.Run(":8010")
}
