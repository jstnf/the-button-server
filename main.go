package main

import (
	"github.com/jstnf/the-button-server/api"
	"github.com/jstnf/the-button-server/data"
	"log"
	"os"
	"strconv"
)

func main() {
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "3001"
	}

	expiryStr := os.Getenv("BUTTON_EXPIRY")
	var expiry int64
	if expiryStr == "" {
		expiry = 0
	} else {
		parsed, err := strconv.ParseInt(expiryStr, 10, 64)
		if err != nil {
			log.Fatalf("Failed to parse BUTTON_EXPIRY: %v", err)
		}
		expiry = parsed
	}

	millisPerPressStr := os.Getenv("MILLIS_DEDUCTED_PER_PRESS")
	var millisPerPress int64
	if millisPerPressStr == "" {
		millisPerPress = 0
	} else {
		parsed, err := strconv.ParseInt(millisPerPressStr, 10, 64)
		if err != nil {
			log.Fatalf("Failed to parse MILLIS_DEDUCTED_PER_PRESS: %v", err)
		}
		millisPerPress = parsed
	}

	store, err := data.NewPostgresStorage(os.Getenv("SERVER_DATABASE_URL"))
	if err != nil {
		log.Fatalf("Failed to create storage: %v", err)
	}
	defer store.Close()
	if err := store.Init(); err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	users := data.NewLocalUserStorage()
	if err := users.Init(); err != nil {
		log.Fatalf("Failed to initialize user storage: %v", err)
	}

	router := api.NewAPIServer(":"+port, expiry, millisPerPress, store, users)

	log.Printf("Server is running on port %s", port)
	log.Fatal(router.Run())
}
