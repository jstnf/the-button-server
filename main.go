package main

import (
	"github.com/jstnf/the-button-server/api"
	"github.com/jstnf/the-button-server/data"
	"log"
	"os"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3001"
	}

	store, err := data.NewPostgresStorage(os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Failed to create storage: %v", err)
	}
	defer store.Close()
	if err := store.Init(); err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	router := api.NewAPIServer(":"+port, store)

	log.Printf("Server is running on port %s", port)
	log.Fatal(router.Run())
}
