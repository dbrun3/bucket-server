package main

import (
	"bucket-serve/handler"
	"errors"
	"log"
	"net/http"
	"os"
)

func main() {
	log.Println("Starting server...")
	mux := http.NewServeMux()
	config, err := initConfig()
	if err != nil {
		log.Fatal("Fatal error reading config: ", err)
	}

	handler := handler.NewHandler(*config)
	mux.Handle("/", handler)

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	log.Println("Server Listening on 8080")
	http.ListenAndServe(":8080", mux)
}

func initConfig() (*handler.Config, error) {
	configFilepath := os.Getenv("CONFIG_FILE")
	if configFilepath != "" {
		return handler.ReadConfigFromFile(configFilepath)
	}
	configString := os.Getenv("CONFIG")
	if configString != "" {
		return handler.ReadConfigFromString(configString)
	}
	return nil, errors.New("no config")
}
