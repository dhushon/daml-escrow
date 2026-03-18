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
// @host localhost:8080
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
		fmt.Sscanf(port, "%d", &ledgerPort)
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

	escrowService := services.NewEscrowService(
		logger,
		ledgerClient,
	)

	handler := api.NewHandler(
		logger,
		escrowService,
	)

	router := chi.NewRouter()
	router.Use(api.LoggingMiddleware(logger))

	router.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL(fmt.Sprintf("http://localhost:%d/swagger/doc.json", cfg.Server.Port)),
	))

	router.Post("/escrows", handler.CreateEscrow)
	router.Get("/escrows", handler.ListEscrows)
	router.Get("/escrows/{escrowID}", handler.GetEscrow)
	router.Post("/escrows/{escrowID}/release", handler.ReleaseFunds)
	router.Post("/escrows/{escrowID}/refund", handler.RefundBuyer)
	router.Post("/escrows/{escrowID}/refund-by-seller", handler.RefundBySeller)
	router.Post("/escrows/{escrowID}/resolve", handler.ResolveDispute)

	router.Get("/metrics", handler.GetMetrics)
	router.Get("/settlements", handler.ListSettlements)
	router.Post("/settlements/{settlementID}/settle", handler.SettlePayment)

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
