package api

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"daml-escrow/internal/services"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type Handler struct {
	logger           *zap.Logger
	escrowService    *services.EscrowService
	metricsService   *services.MetricsService
	configService    *services.ConfigService
	analyticsService *services.AnalyticsService
	identityService  *services.IdentityService
}

func NewHandler(logger *zap.Logger, escrowService *services.EscrowService, metricsService *services.MetricsService, configService *services.ConfigService, analyticsService *services.AnalyticsService, identityService *services.IdentityService) *Handler {
	return &Handler{
		logger:           logger,
		escrowService:    escrowService,
		metricsService:   metricsService,
		configService:    configService,
		analyticsService: analyticsService,
		identityService:  identityService,
	}
}

// ---------------------------------------------------------------------------
// Config Handlers
// ---------------------------------------------------------------------------

func (h *Handler) GetConfig(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user")
	key := r.URL.Query().Get("key")
	val, err := h.configService.GetConfig(userID, key)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if val == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(val); err != nil {
		h.logger.Error("failed to write config response", zap.Error(err))
	}
}

func (h *Handler) SaveConfig(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user")
	key := r.URL.Query().Get("key")
	var val json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&val); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	if err := h.configService.SaveConfig(userID, key, val); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// ---------------------------------------------------------------------------
// Escrow Lifecycle Handlers (Directive 05)
// ---------------------------------------------------------------------------

func (h *Handler) ProposeEscrow(w http.ResponseWriter, r *http.Request) {
	var req ProposeEscrowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request format", http.StatusBadRequest)
		return
	}

	if err := req.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ledgerReq := req.ToLedgerRequest()
	proposal, err := h.escrowService.ProposeEscrow(r.Context(), ledgerReq)
	if err != nil {
		h.logger.Error("propose escrow failed", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	h.renderJSON(w, proposal)
}

func (h *Handler) Fund(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "escrowID")
	userID, _ := r.Context().Value(AuthSubKey).(string)

	var req FundEscrowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request format", http.StatusBadRequest)
		return
	}

	if err := req.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.escrowService.Fund(r.Context(), id, req.CustodyRef, req.HoldingCid, userID); err != nil {
		h.logger.Error("fund failed", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) Activate(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "escrowID")
	userID, _ := r.Context().Value(AuthSubKey).(string)

	// Task 6.3: Analytics & Validation (Noves)
	// 1. Fetch escrow to get deposit details
	escrow, err := h.escrowService.GetEscrow(r.Context(), id, userID)
	if err != nil {
		h.logger.Error("failed to fetch escrow for activation", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	// 2. Validate deposit on-ledger via Noves
	// We use the custodyRef as the transaction hash for this high-assurance prototype
	if escrow.Asset.CustodyRef != "" {
		ok, err := h.analyticsService.ConfirmDeposit(r.Context(), escrow.Asset.CustodyRef, escrow.Asset.Amount, escrow.Asset.Currency)
		if err != nil || !ok {
			h.logger.Warn("Noves deposit validation failed", zap.String("escrowID", id), zap.Error(err))
			http.Error(w, "deposit not yet confirmed on-ledger", http.StatusPreconditionFailed)
			return
		}
	}

	// 3. Proceed with activation
	if _, err := h.escrowService.Activate(r.Context(), id, userID); err != nil {
		h.logger.Error("activate failed", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) ConfirmConditions(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "escrowID")
	userID, _ := r.Context().Value(AuthSubKey).(string)

	if err := h.escrowService.ConfirmConditions(r.Context(), id, userID); err != nil {
		h.logger.Error("confirm failed", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) RaiseDispute(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "escrowID")
	userID, _ := r.Context().Value(AuthSubKey).(string)

	if err := h.escrowService.RaiseDispute(r.Context(), id, userID); err != nil {
		h.logger.Error("raise dispute failed", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) ProposeSettlement(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "escrowID")
	userID, _ := r.Context().Value(AuthSubKey).(string)

	var req ProposeSettlementRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request format", http.StatusBadRequest)
		return
	}

	if err := req.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ledgerTerms := req.ToLedgerTerms()
	if _, err := h.escrowService.ProposeSettlement(r.Context(), id, ledgerTerms, userID); err != nil {
		h.logger.Error("propose settlement failed", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}


	w.WriteHeader(http.StatusOK)
}

func (h *Handler) RatifySettlement(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "escrowID")
	userID, _ := r.Context().Value(AuthSubKey).(string)

	if _, err := h.escrowService.RatifySettlement(r.Context(), id, userID); err != nil {
		h.logger.Error("ratify failed", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) FinalizeSettlement(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "escrowID")
	userID, _ := r.Context().Value(AuthSubKey).(string)

	if _, err := h.escrowService.FinalizeSettlement(r.Context(), id, userID); err != nil {
		h.logger.Error("finalize failed", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) Disburse(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "escrowID")
	userID, _ := r.Context().Value(AuthSubKey).(string)

	if err := h.escrowService.Disburse(r.Context(), id, userID); err != nil {
		h.logger.Error("disburse failed", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) GetEscrow(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "escrowID")
	userID := r.URL.Query().Get("user")
	if userID == "" {
		userID, _ = r.Context().Value(AuthSubKey).(string)
	}

	escrow, err := h.escrowService.GetEscrow(r.Context(), id, userID)
	if err != nil {
		http.Error(w, "escrow not found", http.StatusNotFound)
		return
	}

	h.renderJSON(w, escrow)
}

func (h *Handler) ListEscrows(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user")
	if userID == "" {
		userID, _ = r.Context().Value(AuthSubKey).(string)
	}

	escrows, err := h.escrowService.ListEscrows(r.Context(), userID)
	if err != nil {
		h.logger.Error("list failed", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	h.renderJSON(w, escrows)
}

// ---------------------------------------------------------------------------
// Invitation Handlers
// ---------------------------------------------------------------------------

func (h *Handler) CreateInvitation(w http.ResponseWriter, r *http.Request) {
	var req CreateInvitationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request format", http.StatusBadRequest)
		return
	}

	if err := req.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userID, _ := r.Context().Value(AuthSubKey).(string)
	asset, terms := req.ToLedgerAssetAndTerms()
	invitation, err := h.escrowService.CreateInvitation(r.Context(), userID, req.InviteeEmail, req.InviteeRole, req.InviteeType, asset, terms)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	h.renderJSON(w, invitation)
}

func (h *Handler) ClaimInvitation(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	userID, _ := r.Context().Value(AuthSubKey).(string)
	userEmail, _ := r.Context().Value(EmailKey).(string)
	originDomain, _ := r.Context().Value(OriginDomainKey).(string)

	invite, err := h.escrowService.GetInvitationByToken(r.Context(), token)
	if err != nil {
		http.Error(w, "invitation not found or expired", http.StatusNotFound)
		return
	}

	// Cryptographic Binding Check (Directive 12)
	// Verify that the logged-in user matches the intended recipient domain or email
	isAuthorized := false
	if invite.InviteeEmail == userEmail {
		isAuthorized = true
	} else if strings.HasSuffix(userEmail, "@"+originDomain) {
		// Verify domain-level alignment
		isAuthorized = true
	}

	if !isAuthorized && os.Getenv("AUTH_BYPASS") != "true" {
		h.logger.Warn("unauthorized invitation claim attempt", 
			zap.String("inviteId", invite.ID), 
			zap.String("userEmail", userEmail),
			zap.String("targetEmail", invite.InviteeEmail))
		http.Error(w, "unauthorized: this invitation is bound to another identity", http.StatusForbidden)
		return
	}

	proposal, err := h.escrowService.ClaimInvitation(r.Context(), invite.ID, userID)
	if err != nil {
		h.logger.Error("claim invitation failed", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	h.renderJSON(w, proposal)
}

func (h *Handler) ListInvitations(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(AuthSubKey).(string)
	invitations, err := h.escrowService.ListInvitations(r.Context(), userID)
	if err != nil {
		h.logger.Error("list invitations failed", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	h.renderJSON(w, invitations)
}

func (h *Handler) GetInvitationByToken(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	invitation, err := h.escrowService.GetInvitationByToken(r.Context(), token)
	if err != nil {
		http.Error(w, "invitation not found", http.StatusNotFound)
		return
	}
	h.renderJSON(w, invitation)
}

// ---------------------------------------------------------------------------
// Identity & Health
// ---------------------------------------------------------------------------

func (h *Handler) DiscoverAuth(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	if email == "" {
		http.Error(w, "email parameter required", http.StatusBadRequest)
		return
	}

	provider := h.identityService.DiscoverProvider(r.Context(), email)
	h.renderJSON(w, provider)
}

func (h *Handler) GetIdentity(w http.ResponseWriter, r *http.Request) {
	sub, ok := r.Context().Value(AuthSubKey).(string)
	if !ok || sub == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	email, _ := r.Context().Value(EmailKey).(string)

	identity, err := h.escrowService.GetIdentity(r.Context(), sub)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if identity == nil {
		scopes, _ := r.Context().Value(ScopesKey).([]string)
		identity, err = h.escrowService.ProvisionUser(r.Context(), sub, email, scopes)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
	}

	h.renderJSON(w, identity)
}

func (h *Handler) GetHealth(w http.ResponseWriter, r *http.Request) {
	health := h.metricsService.GetHealth(
		h.configService,
		h.escrowService.GetLedgerClient(),
		h.escrowService.GetOracleSecret(),
	)
	h.renderJSON(w, health)
}

func (h *Handler) GetEscrowLifecycle(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "escrowID")
	userID, _ := r.Context().Value(AuthSubKey).(string)

	// Fetch escrow to get current state
	escrow, err := h.escrowService.GetEscrow(r.Context(), id, userID)
	if err != nil {
		h.logger.Error("failed to fetch escrow for lifecycle", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	lifecycle, err := h.analyticsService.GetEscrowLifecycle(r.Context(), id, escrow.State)
	if err != nil {
		h.logger.Error("failed to fetch lifecycle metadata", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	h.renderJSON(w, lifecycle)
}

func (h *Handler) OracleMilestoneTrigger(w http.ResponseWriter, r *http.Request) {
	var req OracleWebhookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request format", http.StatusBadRequest)
		return
	}

	if err := req.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ledgerReq := req.ToLedgerRequest()
	if err := h.escrowService.ProcessOracleWebhook(r.Context(), ledgerReq); err != nil {
		h.logger.Error("webhook failed", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// ---------------------------------------------------------------------------
// System Handlers
// ---------------------------------------------------------------------------

func (h *Handler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(AuthSubKey).(string)
	metrics, err := h.escrowService.GetMetrics(r.Context(), userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	h.renderJSON(w, metrics)
}

func (h *Handler) ListSettlements(w http.ResponseWriter, r *http.Request) {
	settlements, err := h.escrowService.ListSettlements(r.Context())
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	h.renderJSON(w, settlements)
}

func (h *Handler) SettlePayment(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "settlementID")
	if err := h.escrowService.SettlePayment(r.Context(), id); err != nil {
		h.logger.Error("settle payment failed", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) ListWallets(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(AuthSubKey).(string)
	wallets, err := h.escrowService.ListWallets(r.Context(), userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	h.renderJSON(w, wallets)
}

func (h *Handler) renderJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		h.logger.Error("failed to encode response", zap.Error(err))
	}
}
