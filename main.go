package main

import (
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	apiKey := os.Getenv("WEATHER_API_KEY")
	if apiKey == "" {
		log.Fatal("WEATHER_API_KEY is required")
	}

	// Cloud Run injects PORT; 8080 is the local default.
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	client := &http.Client{Timeout: 10 * time.Second}
	s := &server{
		viacep:  &viaCEPClient{baseURL: "https://viacep.com.br", http: client},
		weather: &weatherClient{baseURL: "https://api.weatherapi.com", apiKey: apiKey, http: client},
		timeout: 10 * time.Second,
	}

	log.Printf("listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, s.routes()))
}
