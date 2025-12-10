package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Bekw/go-practicum-diploma/internal/accrual"
	"github.com/Bekw/go-practicum-diploma/internal/auth"
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

	auth.Init(cfg.AuthSecret)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if cfg.AccrualSystemAddr != "" {
		p := accrual.NewProcessor(cfg.AccrualSystemAddr, store)
		go p.Run(ctx)
	} else {
		log.Println("ACCRUAL_SYSTEM_ADDRESS не задан, обновление начислений отключено")
	}

	r := apphttp.NewRouter(store)

	srv := &http.Server{
		Addr:    cfg.RunAddress,
		Handler: r,
	}

	go func() {
		log.Printf("starting on %s", cfg.RunAddress)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server stopped: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("server shutdown error: %v", err)
	}

	log.Println("server stopped")
}
