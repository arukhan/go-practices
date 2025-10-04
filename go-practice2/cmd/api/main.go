package main

import (
	"log"
	"net/http"

	"go-practice2/internal/handlers"
	"go-practice2/internal/middleware"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/user", handlers.UserHandler)

	handler := middleware.AuthMiddleware(mux)

	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", handler); err != nil {
		log.Fatal(err)
	}
}
