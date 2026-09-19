package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/Shreyansh1032/payments-ledger-system/internal/config"
	"github.com/Shreyansh1032/payments-ledger-system/internal/db"
	"github.com/Shreyansh1032/payments-ledger-system/internal/handler"
	"github.com/Shreyansh1032/payments-ledger-system/internal/middleware"
	"github.com/Shreyansh1032/payments-ledger-system/internal/service"
)

func main() {
	log.Logger = zerolog.New(os.Stdout).With().Timestamp().Logger()

	cfg := config.Load()

	ctx := context.Background()
	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal().Err(err).Msg("db connection failed")
	}
	defer pool.Close()

	authService := service.NewAuthService(pool)
	refreshTokenService := service.NewRefreshTokenService(pool)
	authHandler := handler.NewAuthHandler(authService, refreshTokenService, cfg.JWTSecret)

	accountService := service.NewAccountService(pool)
	accountHandler := handler.NewAccountHandler(accountService)

	transferService := service.NewTransferService(pool)
	transferHandler := handler.NewTransferHandler(transferService)

	r := chi.NewRouter()
	r.Use(chimiddleware.RequestID)
	r.Use(middleware.StructuredLogger)
	r.Use(middleware.Metrics)
	r.Use(chimiddleware.Recoverer)

	r.Get("/health", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	r.Handle("/metrics", promhttp.Handler())

	fileServer := http.FileServer(http.Dir("./docs"))
	r.Handle("/docs/*", http.StripPrefix("/docs/", fileServer))
	r.Get("/docs", func(w http.ResponseWriter, req *http.Request) {
		http.ServeFile(w, req, "./docs/swagger.html")
	})

	r.Route("/api/v1/auth", func(r chi.Router) {
		r.Post("/signup", authHandler.SignUp)
		r.Post("/login", authHandler.Login)
		r.Post("/refresh", authHandler.Refresh)
		r.Post("/logout", authHandler.Logout)
	})

	r.Route("/api/v1/account", func(r chi.Router) {
		r.Use(middleware.Auth(cfg.JWTSecret))
		r.Get("/balance", accountHandler.Balance)
	})

	r.Route("/api/v1/transfer", func(r chi.Router) {
		r.Use(middleware.Auth(cfg.JWTSecret))
		r.Post("/", transferHandler.Transfer)
	})

	log.Info().Str("port", cfg.Port).Msg("server starting")
	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		log.Fatal().Err(err).Msg("server failed")
	}
}
