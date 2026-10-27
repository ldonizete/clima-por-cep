package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"regexp"
	"time"
)

var zipcodeRe = regexp.MustCompile(`^[0-9]{8}$`)

type server struct {
	viacep  *viaCEPClient
	weather *weatherClient
	timeout time.Duration
}

type weatherResponse struct {
	TempC float64 `json:"temp_C"`
	TempF float64 `json:"temp_F"`
	TempK float64 `json:"temp_K"`
}

func (s *server) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /weather/{cep}", s.handleWeather)
	mux.HandleFunc("GET /weather", s.handleWeather)
	return mux
}

// handleWeather accepts the CEP as /weather/{cep} or /weather?cep={cep}.
func (s *server) handleWeather(w http.ResponseWriter, r *http.Request) {
	cep := r.PathValue("cep")
	if cep == "" {
		cep = r.URL.Query().Get("cep")
	}
	if !zipcodeRe.MatchString(cep) {
		http.Error(w, "invalid zipcode", http.StatusUnprocessableEntity)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), s.timeout)
	defer cancel()

	loc, err := s.viacep.Lookup(ctx, cep)
	if errors.Is(err, errZipcodeNotFound) {
		http.Error(w, "can not find zipcode", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("viacep lookup %s: %v", cep, err)
		http.Error(w, "upstream error", http.StatusBadGateway)
		return
	}

	tempC, err := s.weather.CurrentC(ctx, loc)
	if errors.Is(err, errCityNotFound) {
		http.Error(w, "can not find zipcode", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("weather lookup %s (%s/%s): %v", cep, loc.City, loc.State, err)
		http.Error(w, "upstream error", http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(weatherResponse{
		TempC: round2(tempC),
		TempF: CelsiusToFahrenheit(tempC),
		TempK: CelsiusToKelvin(tempC),
	})
}
