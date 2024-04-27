package main

import (
	"github.com/jstnf/the-button-server/api"
	"log"
	"os"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3001"
	}

	router := api.NewAPIServer(":"+port, nil)

	log.Printf("Server is running on port %s", port)
	log.Fatal(router.Run())
}
