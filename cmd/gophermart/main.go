package main

import (
	"log"
	"net/http"

	"github.com/Bekw/go-practicum-diploma/internal/config"
	apphttp "github.com/Bekw/go-practicum-diploma/internal/http"
)

func main() {
	cfg := config.Load()

	log.Printf("starting on %s", cfg.RunAddress)

	r := apphttp.NewRouter()

	if err := http.ListenAndServe(cfg.RunAddress, r); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
