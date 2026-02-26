package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/isaacthajunior/url-shortener/internal/database"
	"github.com/isaacthajunior/url-shortener/internal/shortenUrl"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error Loading .env file", err)
	}
	dbURL := os.Getenv("DB_URL")

	mux := http.NewServeMux()
	port := "8080"
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Error connecting to db")
	}
	dbQueries := database.New(db)

	shortenHandler := shortenUrl.Handler{
		DB: dbQueries,
	}

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		message := "Server is running and ready!"
		w.Write([]byte(message))
	})

	mux.HandleFunc("POST /shorten", shortenHandler.HandleShorten)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}
	log.Fatal(server.ListenAndServe())
}
