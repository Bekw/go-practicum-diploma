package config

import (
	"flag"
	"os"
)

type Config struct {
	RunAddress        string
	DatabaseURI       string
	AccrualSystemAddr string
	AuthSecret        string
}

func Load() *Config {
	cfg := &Config{
		RunAddress:        "localhost:8080",
		DatabaseURI:       "",
		AccrualSystemAddr: "",
		AuthSecret:        "",
	}

	if v := os.Getenv("RUN_ADDRESS"); v != "" {
		cfg.RunAddress = v
	}
	if v := os.Getenv("DATABASE_URI"); v != "" {
		cfg.DatabaseURI = v
	}
	if v := os.Getenv("ACCRUAL_SYSTEM_ADDRESS"); v != "" {
		cfg.AccrualSystemAddr = v
	}
	if v := os.Getenv("AUTH_SECRET"); v != "" {
		cfg.AuthSecret = v
	}

	flag.StringVar(&cfg.RunAddress, "a", cfg.RunAddress, "server address")
	flag.StringVar(&cfg.DatabaseURI, "d", cfg.DatabaseURI, "database URI")
	flag.StringVar(&cfg.AccrualSystemAddr, "r", cfg.AccrualSystemAddr, "accrual system address")
	flag.StringVar(&cfg.AuthSecret, "s", cfg.AuthSecret, "auth secret key")

	flag.Parse()

	return cfg
}
