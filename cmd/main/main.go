package main

import (
	"log"
	"net/http"

	"github.com/isaacthajunior/url-shortener/internal/shortenUrl"
)

type config struct{}

func main() {
	mux := http.NewServeMux()
	port := "8080"

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		message := "Server is running and ready!"
		w.Write([]byte(message))
	})

	mux.HandleFunc("POST /shorten", shortenUrl.HandleShorten)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}
	log.Fatal(server.ListenAndServe())
}
