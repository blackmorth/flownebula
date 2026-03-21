package main

import (
	"flownebula/server/internal/api"
	"log"
)

func main() {
	app := api.New()

	log.Println("Nebula server running on :8080")
	if err := app.Listen(":8080"); err != nil {
		log.Fatal(err)
	}
}
