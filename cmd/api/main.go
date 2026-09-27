package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/time/rate"

	"github.com/Shreyansh1032/payments-ledger-system/internal/config"
	"github.com/Shreyansh1032/payments-ledger-system/internal/db"
	"github.com/Shreyansh1032/payments-ledger-system/internal/handler"
	"github.com/Shreyansh1032/payments-ledger-system/internal/middleware"
	"github.com/Shreyansh1032/payments-ledger-system/internal/service"
)

// Version is set at build time via -ldflags "-X main.Version=v1.0.0"
var Version = "dev"

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
	accountHandler := handler.NewAccountHandler(accountService, authService)

	transferService := service.NewTransferService(pool)
	transferHandler := handler.NewTransferHandler(transferService)
	depositHandler := handler.NewDepositHandler(transferService)

	historyService := service.NewTransactionHistoryService(pool)
	transactionHandler := handler.NewTransactionHandler(historyService)

	userSearchService := service.NewUserSearchService(pool)
	userSearchHandler := handler.NewUserSearchHandler(userSearchService)

	favoriteService := service.NewFavoriteService(pool)
	favoriteHandler := handler.NewFavoriteHandler(favoriteService)

	paymentRequestService := service.NewPaymentRequestService(pool, transferService)
	paymentRequestHandler := handler.NewPaymentRequestHandler(paymentRequestService)

	insightsService := service.NewInsightsService(pool)
	insightsHandler := handler.NewInsightsHandler(insightsService)

	authLimiter := middleware.NewRateLimiter(rate.Every(time.Minute/5), 5)

	r := chi.NewRouter()
	r.Use(chimiddleware.RequestID)
	r.Use(middleware.StructuredLogger)
	r.Use(middleware.Metrics)
	r.Use(chimiddleware.Recoverer)

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{cfg.CORSOrigin},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Content-Type", "Authorization", "Idempotency-Key"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/health", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	r.Get("/version", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"version": Version})
	})

	r.Handle("/metrics", promhttp.Handler())

	fileServer := http.FileServer(http.Dir("./docs"))
	r.Handle("/docs/*", http.StripPrefix("/docs/", fileServer))
	r.Get("/docs", func(w http.ResponseWriter, req *http.Request) {
		http.ServeFile(w, req, "./docs/swagger.html")
	})

	r.Route("/api/v1/auth", func(r chi.Router) {
		r.With(authLimiter.Middleware).Post("/signup", authHandler.SignUp)
		r.With(authLimiter.Middleware).Post("/login", authHandler.Login)
		r.Post("/refresh", authHandler.Refresh)
		r.Post("/logout", authHandler.Logout)
	})

	r.Route("/api/v1/account", func(r chi.Router) {
		r.Use(middleware.Auth(cfg.JWTSecret))
		r.Get("/balance", accountHandler.Balance)
		r.Get("/profile", accountHandler.Profile)
		r.Put("/password", accountHandler.ChangePassword)
		r.Post("/deposit", depositHandler.Deposit)
	})

	r.Route("/api/v1/transfer", func(r chi.Router) {
		r.Use(middleware.Auth(cfg.JWTSecret))
		r.Post("/", transferHandler.Transfer)
	})

	r.Route("/api/v1/transactions", func(r chi.Router) {
		r.Use(middleware.Auth(cfg.JWTSecret))
		r.Get("/", transactionHandler.History)
	})

	r.Route("/api/v1/users", func(r chi.Router) {
		r.Use(middleware.Auth(cfg.JWTSecret))
		r.Get("/search", userSearchHandler.Search)
	})

	r.Route("/api/v1/favorites", func(r chi.Router) {
		r.Use(middleware.Auth(cfg.JWTSecret))
		r.Get("/", favoriteHandler.List)
		r.Post("/", favoriteHandler.Add)
		r.Delete("/{username}", favoriteHandler.Remove)
	})

	r.Route("/api/v1/requests", func(r chi.Router) {
		r.Use(middleware.Auth(cfg.JWTSecret))
		r.Post("/", paymentRequestHandler.Create)
		r.Post("/split", paymentRequestHandler.CreateSplit)
		r.Get("/incoming", paymentRequestHandler.ListIncoming)
		r.Get("/outgoing", paymentRequestHandler.ListOutgoing)
		r.Post("/{id}/approve", paymentRequestHandler.Approve)
		r.Post("/{id}/decline", paymentRequestHandler.Decline)
	})

	r.Route("/api/v1/insights", func(r chi.Router) {
		r.Use(middleware.Auth(cfg.JWTSecret))
		r.Get("/monthly", insightsHandler.Monthly)
	})

	log.Info().Str("port", cfg.Port).Str("version", Version).Msg("server starting")
	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		log.Fatal().Err(err).Msg("server failed")
	}
}
