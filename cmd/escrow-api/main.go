package main

import (
	"context"
	"daml-escrow/internal/api"
	"daml-escrow/internal/config"
	"daml-escrow/internal/ledger"
	"daml-escrow/internal/services"
	"daml-escrow/pkg/logging"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "daml-escrow/docs"

	chi "github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger/v2"
	"go.uber.org/zap"
)

// @title Stablecoin Escrow API
// @version 1.0
// @description API for managing privacy-preserving stablecoin escrows on DAML.
// @host localhost:8081
// @BasePath /
func main() {
	logger := logging.NewLogger()
	defer func() {
		_ = logger.Sync()
	}()

	cfg, err := config.LoadConfig("config/config.yaml")
	if err != nil {
		logger.Fatal("failed to load config", zap.Error(err))
	}

	// Environment variable overrides
	ledgerHost := cfg.Ledger.Host
	if host := os.Getenv("LEDGER_HOST"); host != "" {
		ledgerHost = host
	}
	ledgerPort := cfg.Ledger.Port
	if port := os.Getenv("LEDGER_PORT"); port != "" {
		if _, err := fmt.Sscanf(port, "%d", &ledgerPort); err != nil {
			logger.Warn("failed to parse LEDGER_PORT, using default", zap.String("port", port), zap.Int("default", ledgerPort))
		}
	}

	var ledgerClient ledger.Client
	ledgerType := os.Getenv("LEDGER_TYPE")

	if ledgerType == "grpc" {
		logger.Info("using gRPC ledger client", zap.String("host", ledgerHost), zap.Int("port", ledgerPort))
		ledgerClient = ledger.NewDamlClient(logger, ledgerHost, ledgerPort)
	} else {
		// Default to JSON API for better dynamic binding support
		logger.Info("using JSON ledger client", zap.String("host", ledgerHost), zap.Int("port", ledgerPort))
		ledgerClient = ledger.NewJsonLedgerClient(logger, ledgerHost, ledgerPort)
	}

	// Perform dynamic discovery (resolve Package and Party IDs)
	if err := ledgerClient.Discover(context.Background()); err != nil {
		logger.Error("ledger discovery failed (continuing with defaults)", zap.Error(err))
	}

	// Initialize core services
	metricsService := services.NewMetricsService()
	
	configService, err := services.NewConfigService(cfg.UserConfig.DSN)
	if err != nil {
		logger.Fatal("failed to initialize config service", zap.Error(err))
	}
	defer configService.Close()

	escrowService := services.NewEscrowService(
		logger,
		ledgerClient,
		cfg.Oracle.WebhookSecret,
	)

	handler := api.NewHandler(
		logger,
		escrowService,
		metricsService,
		configService,
	)

	router := chi.NewRouter()
	
	router.Use(api.LoggingMiddleware(logger))
	router.Use(api.MetricsMiddleware(metricsService))
	router.Use(api.AuthMiddleware(cfg.Auth.Issuer, cfg.Auth.ClientID, cfg.Auth.Audience, logger))

	router.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL(fmt.Sprintf("http://localhost:%d/swagger/doc.json", cfg.Server.Port)),
	))

	// API Routes
	router.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", handler.GetHealth)
		r.Get("/auth/me", handler.GetIdentity)
		r.Get("/config", handler.GetConfig)
		r.Post("/config", handler.SaveConfig)

		r.Get("/invites", handler.ListInvitations)
		r.Get("/invites/token/{token}", handler.GetInvitationByToken)
		r.Post("/invites", handler.CreateInvitation)
		r.Post("/invites/token/{token}/claim", handler.ClaimInvitation)
		r.Post("/escrows", handler.CreateEscrow)
		r.Post("/escrows/propose", handler.ProposeEscrow)
		r.Post("/escrows/{escrowID}/accept", handler.AcceptProposal)
		r.Get("/escrows", handler.ListEscrows)
		r.Get("/escrows/proposals", handler.ListProposals)
		r.Get("/escrows/{escrowID}", handler.GetEscrow)
		r.Post("/escrows/{escrowID}/release", handler.ReleaseFunds)
		r.Post("/escrows/{escrowID}/refund", handler.RefundBuyer)
		r.Post("/escrows/{escrowID}/refund-by-seller", handler.RefundBySeller)
		r.Post("/escrows/{escrowID}/resolve", handler.ResolveDispute)

		r.Post("/webhooks/milestone", handler.OracleMilestoneTrigger)

		r.Get("/metrics", handler.GetMetrics)
		r.Get("/settlements", handler.ListSettlements)
		r.Post("/settlements/{settlementID}/settle", handler.SettlePayment)
		r.Get("/wallets", handler.ListWallets)
	})

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		logger.Info("starting escrow-api", zap.Int("port", cfg.Server.Port))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("server failed", zap.Error(err))
		}
	}()

	waitForShutdown(server, logger)
}

func waitForShutdown(server *http.Server, logger *zap.Logger) {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
	<-stop

	logger.Info("shutdown signal received")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("server shutdown failed", zap.Error(err))
	}
}
