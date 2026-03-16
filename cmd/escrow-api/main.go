package main

import (
	"context"
	"daml-escrow/internal/api"
	"daml-escrow/internal/ledger"
	"daml-escrow/internal/services"
	"daml-escrow/pkg/logging"
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

	ledgerClient := ledger.NewDamlClient(logger)

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
		httpSwagger.URL("http://localhost:8080/swagger/doc.json"),
	))

	router.Post("/escrows", handler.CreateEscrow)
	router.Get("/escrows/{escrowID}", handler.GetEscrow)
	router.Post("/escrows/{escrowID}/release", handler.ReleaseEscrow)
	router.Post("/escrows/{escrowID}/refund", handler.RefundEscrow)

	server := &http.Server{
		Addr:         ":8080",
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		logger.Info("starting escrow-api", zap.String("port", "8080"))
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
