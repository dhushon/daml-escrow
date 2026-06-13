// @title Stablecoin Escrow API
// @version 1.0
// @description API for managing privacy-preserving stablecoin escrows on DAML.
// @host localhost:8081
// @BasePath /api/v1
// @query.collection.format multi
package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"daml-escrow/internal/api"
	"daml-escrow/internal/config"
	"daml-escrow/internal/crypto"
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

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/riandyrn/otelchi"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var (
	cfgFile     string
	environment string
	authBypass  bool
	tlsDisabled bool
	certFile    string
	keyFile     string
	caCertFile  string
	port        int
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "escrow-api",
	Short: "Stablecoin Escrow Platform API",
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the escrow platform API server",
	Run: func(cmd *cobra.Command, args []string) {
		runServer()
	},
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./config/config.yaml)")
	rootCmd.PersistentFlags().StringVar(&environment, "env", "", "environment (dev, production)")
	rootCmd.PersistentFlags().BoolVar(&authBypass, "bypass", false, "bypass JWT authentication (dev only)")
	rootCmd.PersistentFlags().IntVar(&port, "port", 0, "port to listen on (overrides config)")

	// High-Assurance TLS Flags (Secure by Default)

	rootCmd.PersistentFlags().BoolVar(&tlsDisabled, "notls", false, "disable mTLS enforcement (dev only)")
	rootCmd.PersistentFlags().StringVar(&certFile, "tls-cert", "/etc/escrow/certs/tls.crt", "path to server certificate")
	rootCmd.PersistentFlags().StringVar(&keyFile, "tls-key", "/etc/escrow/certs/tls.key", "path to server private key")
	rootCmd.PersistentFlags().StringVar(&caCertFile, "tls-ca", "/etc/escrow/certs/ca.crt", "path to CA certificate for mTLS client validation")

	rootCmd.AddCommand(serveCmd)
}

func initConfig() {
	if environment != "" {
		viper.Set("auth.environment", environment)
	}
	if authBypass {
		viper.Set("auth.authBypass", true)
		viper.Set("auth.environment", "dev")
	}
}

func runServer() {
	logger := logging.NewLogger()
	defer func() { _ = logger.Sync() }()

	ctx := context.Background()

	cfg, err := config.LoadConfig(cfgFile)
	if err != nil {
		logger.Fatal("failed to load config", zap.Error(err))
	}

	// Flag Overrides
	if port != 0 {
		cfg.Server.Port = port
	}

	if err := config.ResolveSecrets(ctx, cfg); err != nil {
		logger.Warn("failed to resolve cloud secrets, falling back to environment/file", zap.Error(err))
	}

	logger.Info("configuration loaded", 
		zap.String("env", cfg.Auth.Environment), 
		zap.Bool("bypass", cfg.Auth.AuthBypass),
		zap.Bool("tls", !tlsDisabled))

	// 1. Initialize high-assurance telemetry
	telemetry, err := api.InitTelemetry(ctx, "escrow-api", cfg.Auth.Environment)
	if err != nil {
		logger.Warn("failed to initialize telemetry, continuing without tracing", zap.Error(err))
	} else {
		defer func() { _ = telemetry.Shutdown(context.Background()) }()
	}

	// 2. Initialize high-assurance signer for Oracle Verification
	var oracleSigner crypto.HighAssuranceSigner
	if cfg.GCPProjectID != "" {
		logger.Info("initializing cloud KMS oracle signer", zap.String("project", cfg.GCPProjectID))
		kmsSigner, err := crypto.NewCloudKMSSigner(ctx, "projects/"+cfg.GCPProjectID+"/locations/"+cfg.Region+"/keyRings/escrow-keyring-"+cfg.Auth.Environment+"/cryptoKeys/oracle-signer-key-"+cfg.Auth.Environment)
		if err != nil {
			logger.Fatal("failed to initialize cloud KMS signer", zap.Error(err))
		}
		oracleSigner = kmsSigner
	} else {
		logger.Info("using local ephemeral oracle signer")
		oracleSigner, _ = crypto.NewLocalSigner()
	}

	var ledgerClient ledger.Client
	if cfg.Ledger.ParticipantID != "" {
		node, ok := cfg.Ledger.Nodes[cfg.Ledger.ParticipantID]
		// High-Assurance: Authoritatively fallback to cfg.Ledger.Host if node map is stale or missing
		host := cfg.Ledger.Host
		if envHost := os.Getenv("LEDGER_HOST"); envHost != "" {
			host = envHost
		} else if ok && node.Host != "" && node.Host != "canton" {
			host = node.Host
		}

		logger.Info("starting in isolated participant mode",
			zap.String("name", cfg.Ledger.ParticipantID),
			zap.String("host", host),
			zap.Int("port", cfg.Ledger.Port))

		ledgerClient = ledger.NewJsonLedgerClient(logger, host, cfg.Ledger.Port, cfg.Ledger.Packages.Implementation, cfg.Ledger.Packages.Interfaces)
	} else if len(cfg.Ledger.Nodes) > 0 {
		clients := make(map[string]ledger.Client)
		envHost := os.Getenv("LEDGER_HOST")
		for name, node := range cfg.Ledger.Nodes {
			host := node.Host
			if envHost != "" {
				host = envHost
			}
			clients[name] = ledger.NewJsonLedgerClient(logger, host, node.Port, cfg.Ledger.Packages.Implementation, cfg.Ledger.Packages.Interfaces)
		}
		ledgerClient = ledger.NewMultiLedgerClient(logger, clients)
	} else {
		host := cfg.Ledger.Host
		if envHost := os.Getenv("LEDGER_HOST"); envHost != "" {
			host = envHost
		}
		ledgerClient = ledger.NewJsonLedgerClient(logger, host, cfg.Ledger.Port, cfg.Ledger.Packages.Implementation, cfg.Ledger.Packages.Interfaces)
	}
	if err := ledgerClient.Discover(ctx, true); err != nil {
		logger.Error("ledger discovery failed", zap.Error(err))
	}

	metricsService := services.NewMetricsService()
	configService, err := services.NewConfigService(cfg.UserConfig.DSN)
	if err != nil {
		logger.Fatal("failed to config service", zap.Error(err))
	}
	defer func() { _ = configService.Close() }()

	factory := ledger.NewStablecoinFactory(logger)
	stablecoinProvider, err := factory.CreateProvider(cfg, ledgerClient)
	if err != nil {
		logger.Fatal("failed to stablecoin provider", zap.Error(err))
	}

	complianceService := services.NewMockCompliance()
	analyticsService := services.NewAnalyticsService(logger)
	schemaService, err := services.NewSchemaService("architecture/schemas")
	if err != nil {
		logger.Fatal("failed to initialize schema service", zap.Error(err))
	}
	aiService, err := services.NewAIService(ctx)
	if err != nil {
		logger.Warn("AI service disabled (missing API key or config)", zap.Error(err))
	}
	identityService, err := services.NewIdentityService("config/identity_providers.yaml", configService.GetDB())
	if err != nil {
		logger.Fatal("failed to identity service", zap.Error(err))
	}
	storageService, err := services.NewStorageService(ctx, os.Getenv("STORAGE_BUCKET"))
	if err != nil {
		logger.Warn("storage service disabled (missing config)", zap.Error(err))
	}
	ingestService := services.NewIngestService(logger, aiService, schemaService, identityService, storageService)
	escrowService := services.NewEscrowService(logger, ledgerClient, stablecoinProvider, complianceService, cfg.Oracle.WebhookSecret, oracleSigner, storageService)

	handler := api.NewHandler(logger, escrowService, metricsService, configService, analyticsService, identityService, schemaService, ingestService, storageService)

	router := chi.NewRouter()

	if telemetry != nil {
		router.Use(otelchi.Middleware("escrow-api", otelchi.WithChiRoutes(router)))
	}
	router.Use(api.LoggingMiddleware(logger))
	router.Use(api.MetricsMiddleware(metricsService))

	router.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"http://localhost:4321", "http://localhost:8080", "https://*.vdatacloudai.com"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "HEAD"},
		AllowedHeaders: []string{"Accept", "Authorization", "Content-Type", "X-Dev-User", "X-Assumed-Role", "X-Requested-With", "Origin"},
	}))
	var verifier api.TokenVerifier
	var oktaVer api.TokenVerifier

	if cfg.Auth.Environment != "dev" || !cfg.Auth.AuthBypass {
		oktaVer, err = api.NewRealVerifier(ctx, cfg.Auth.Issuer, cfg.Auth.Audience)
		if err != nil {
			logger.Fatal("failed to OIDC verifier", zap.Error(err))
		}
	}

	// High-Assurance: Wrap with UnifiedTokenVerifier to allow local wallet JWT verification
	jwtSecret := []byte("platform-jwt-signing-secret-key-32-bytes!")
	verifier = api.NewUnifiedTokenVerifier(oktaVer, jwtSecret)

	router.Use(api.AuthMiddleware(cfg.Auth, verifier, logger))
	
	router.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", handler.GetHealth)
		r.Get("/auth/me", handler.GetIdentity)
		r.Post("/auth/discover", handler.DiscoverAuth)
		r.Get("/auth/nonce", handler.GetNonce)
		r.Post("/auth/wallet/verify", handler.VerifyWallet)
		r.Get("/identities", handler.ListIdentities)
		r.Get("/config", handler.GetConfig)
		r.Post("/config", handler.SaveConfig)

		// --- Phase 11 & 13: Drafting & Ingest ---
		r.Post("/ingest", handler.IngestContract)
		r.Post("/drafts", handler.SaveDraft)
		r.Get("/drafts", handler.ListDrafts)
		r.Get("/drafts/{draftID}", handler.GetDraft)
		r.Post("/drafts/{draftID}/amend", handler.AmendDraft)
		r.Post("/drafts/{draftID}/approve", handler.ApproveDraft)
		r.Post("/drafts/{draftID}/promote", handler.PromoteToLedger)

		// --- Invitations ---
		r.Post("/invites", handler.CreateInvitation)
		r.Get("/invites", handler.ListInvitations)
		r.Get("/invites/token/{token}", handler.GetInvitationByToken)
		r.Post("/invites/token/{token}/claim", handler.ClaimInvitation)

		// --- Escrow Lifecycle ---
		r.Get("/escrows", handler.ListEscrows)
		r.Post("/escrows", handler.ProposeEscrow)
		r.Get("/escrows/{id}", handler.GetEscrow)
		r.Get("/escrows/{id}/lifecycle", handler.GetEscrowLifecycle)
		r.Post("/escrows/{id}/fund", handler.Fund)
		r.Post("/escrows/{id}/activate", handler.Activate)
		r.Post("/escrows/{id}/confirm", handler.ConfirmConditions)
		r.Post("/escrows/{id}/dispute", handler.RaiseDispute)
		r.Post("/escrows/{id}/ratify", handler.RatifySettlement)
		r.Post("/escrows/{id}/finalize", handler.FinalizeSettlement)
		r.Post("/escrows/{id}/disburse", handler.Disburse)

		// --- Settlements & Wallets ---
		r.Get("/settlements", handler.ListSettlements)
		r.Post("/settlements/{settlementID}/settle", handler.SettlePayment)
		r.Get("/wallets", handler.ListWallets)

		// --- System ---
		r.Post("/webhooks/milestone", handler.OracleMilestoneTrigger)
		r.Get("/metrics", handler.GetMetrics)
	})
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// --- High-Assurance mTLS Enforcement (Secure by Default) ---
	if !tlsDisabled {
		caCert, err := os.ReadFile(caCertFile)
		if err != nil {
			logger.Fatal("failed to load CA cert (TLS enabled by default, use --notls to disable)", zap.Error(err))
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)

		server.TLSConfig = &tls.Config{
			ClientCAs:    caCertPool,
			ClientAuth:   tls.RequireAndVerifyClientCert,
			MinVersion:   tls.VersionTLS13,
			CipherSuites: []uint16{tls.TLS_AES_128_GCM_SHA256, tls.TLS_AES_256_GCM_SHA384},
		}
		
		go func() {
			logger.Info("starting high-assurance mTLS escrow-api", zap.Int("port", cfg.Server.Port))
			if err := server.ListenAndServeTLS(certFile, keyFile); err != nil && err != http.ErrServerClosed {
				logger.Fatal("mTLS server failed", zap.Error(err))
			}
		}()
	} else {
		go func() {
			logger.Info("starting unencrypted escrow-api (--notls active)", zap.Int("port", cfg.Server.Port))
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				logger.Fatal("server failed", zap.Error(err))
			}
		}()
	}

	waitForShutdown(server, logger)
}

func waitForShutdown(server *http.Server, logger *zap.Logger) {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
	<-stop
	logger.Info("shutdown signal received")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = server.Shutdown(ctx)
}
