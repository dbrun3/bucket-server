package main

import (
	"bucket-serve/handler"
	"log"
	"net/http"
	"os"
)

func main() {
	log.Println("Starting server...")
	mux := http.NewServeMux()
	config, err := handler.ReadConfigFromFile(os.Getenv("CONFIG"))
	if err != nil {
		log.Fatal("Fatal error reading config: ", err)
	}
	handler := handler.NewHandler(*config)
	mux.Handle("/", handler)

	log.Println("Server Listening on 8080")
	http.ListenAndServe(":8080", mux)
}
