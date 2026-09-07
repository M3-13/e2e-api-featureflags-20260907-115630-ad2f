package main

import (
	"log"
	"net/http"
	"os"

	"featureflags/internal/api"
	"featureflags/internal/store"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	s := store.New()
	mux := api.NewRouter(s)
	handler := api.Logging(mux)

	log.Printf("feature-flag service listening on :%s", port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatal(err)
	}
}
