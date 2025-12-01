package main

import (
	"log"
	"net/http"

	"github.com/Bekw/go-practicum-diploma/internal/config"
	apphttp "github.com/Bekw/go-practicum-diploma/internal/http"
	"github.com/Bekw/go-practicum-diploma/internal/storage"
)

func main() {
	cfg := config.Load()

	if cfg.DatabaseURI == "" {
		log.Fatal("DATABASE_URI (или флаг -d) не задан")
	}

	store, err := storage.New(cfg.DatabaseURI)
	if err != nil {
		log.Fatalf("failed to init storage: %v", err)
	}
	defer store.Close()

	log.Printf("starting on %s", cfg.RunAddress)

	r := apphttp.NewRouter(store)

	if err := http.ListenAndServe(cfg.RunAddress, r); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
