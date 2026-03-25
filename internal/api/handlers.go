package api

import (
	"encoding/json"
	"net/http"

	"daml-escrow/internal/ledger"
	"daml-escrow/internal/services"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type Handler struct {
	logger         *zap.Logger
	escrowService  *services.EscrowService
	metricsService *services.MetricsService
	configService  *services.ConfigService
}

func NewHandler(logger *zap.Logger, escrowService *services.EscrowService, metricsService *services.MetricsService, configService *services.ConfigService) *Handler {
	return &Handler{
		logger:         logger,
		escrowService:  escrowService,
		metricsService: metricsService,
		configService:  configService,
	}
}

// GetConfig handles GET /config?user={id}&key={key}
func (h *Handler) GetConfig(w http.ResponseWriter, r *http.Request) {
	if !RequireScope(r.Context(), ScopeSystemAdmin) {
		http.Error(w, "insufficient scope", http.StatusForbidden)
		return
	}

	userID := r.URL.Query().Get("user")
	key := r.URL.Query().Get("key")
	if userID == "" || key == "" {
		http.Error(w, "missing user or key", http.StatusBadRequest)
		return
	}

	val, err := h.configService.GetConfig(userID, key)
	if err != nil {
		h.logger.Error("get config failed", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if val == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(val); err != nil {
		h.logger.Error("failed to write response", zap.Error(err))
	}
}

// SaveConfig handles POST /config?user={id}&key={key}
func (h *Handler) SaveConfig(w http.ResponseWriter, r *http.Request) {
	if !RequireScope(r.Context(), ScopeSystemAdmin) {
		http.Error(w, "insufficient scope", http.StatusForbidden)
		return
	}

	userID := r.URL.Query().Get("user")
	key := r.URL.Query().Get("key")
	if userID == "" || key == "" {
		http.Error(w, "missing user or key", http.StatusBadRequest)
		return
	}

	var val json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&val); err != nil {
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}

	if err := h.configService.SaveConfig(userID, key, val); err != nil {
		h.logger.Error("save config failed", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("ok")); err != nil {
		h.logger.Error("failed to write response", zap.Error(err))
	}
}

type CreateEscrowRequest struct {
	Buyer       string                `json:"buyer" example:"Buyer"`
	Seller      string                `json:"seller" example:"Seller"`
	Amount      float64               `json:"amount" example:"500.0"`
	Currency    string                `json:"currency" example:"USD"`
	Description string                `json:"description" example:"Software Project"`
	Milestones  []ledger.Milestone    `json:"milestones"`
	Metadata    ledger.EscrowMetadata `json:"metadata"`
}

type CreateInvitationRequest struct {
	InviterID    string             `json:"inviterId"`
	InviteeEmail string             `json:"inviteeEmail"`
	InviteeRole  string             `json:"inviteeRole"` // Buyer, Seller, Mediator
	InviteeType  string             `json:"inviteeType"` // Residential, Company
	Terms        ledger.EscrowTerms `json:"terms"`
}

type EscrowResponse struct {
	ID                    string                `json:"id"`
	Buyer                 string                `json:"buyer"`
	Seller                string                `json:"seller"`
	Amount                float64               `json:"amount"`
	Currency              string                `json:"currency"`
	State                 string                `json:"state"`
	Milestones            []ledger.Milestone    `json:"milestones"`
	CurrentMilestoneIndex int                   `json:"currentMilestoneIndex"`
	Metadata              ledger.EscrowMetadata `json:"metadata"`
}

// CreateEscrow handles POST /escrows
func (h *Handler) CreateEscrow(w http.ResponseWriter, r *http.Request) {
	if !RequireScope(r.Context(), ScopeEscrowWrite) {
		http.Error(w, "insufficient scope", http.StatusForbidden)
		return
	}

	var req CreateEscrowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("failed to decode request", zap.Error(err))
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	escrow, err := h.escrowService.CreateEscrow(r.Context(), ledger.CreateEscrowRequest{
		Buyer:       req.Buyer,
		Seller:      req.Seller,
		Amount:      req.Amount,
		Currency:    req.Currency,
		Description: req.Description,
		Milestones:  req.Milestones,
		Metadata:    req.Metadata,
	})
	if err != nil {
		h.logger.Error("create escrow failed", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(escrow); err != nil {
		h.logger.Error("failed to encode response", zap.Error(err))
	}
}

// ProposeEscrow handles POST /escrows/propose
func (h *Handler) ProposeEscrow(w http.ResponseWriter, r *http.Request) {
	if !RequireScope(r.Context(), ScopeEscrowWrite) {
		http.Error(w, "insufficient scope", http.StatusForbidden)
		return
	}

	var req CreateEscrowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	proposal, err := h.escrowService.ProposeEscrow(r.Context(), ledger.CreateEscrowRequest{
		Buyer:       req.Buyer,
		Seller:      req.Seller,
		Amount:      req.Amount,
		Currency:    req.Currency,
		Description: req.Description,
		Milestones:  req.Milestones,
		Metadata:    req.Metadata,
	})
	if err != nil {
		h.logger.Error("propose escrow failed", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(proposal); err != nil {
		h.logger.Error("failed to encode response", zap.Error(err))
	}
}

// AcceptProposal handles POST /escrows/{id}/accept
func (h *Handler) AcceptProposal(w http.ResponseWriter, r *http.Request) {
	if !RequireScope(r.Context(), ScopeEscrowAccept) {
		http.Error(w, "insufficient scope", http.StatusForbidden)
		return
	}

	id := chi.URLParam(r, "escrowID")
	sellerID := r.URL.Query().Get("user")
	if sellerID == "" {
		sellerID = "Seller"
	}

	if err := h.escrowService.AcceptProposal(r.Context(), id, sellerID); err != nil {
		h.logger.Error("accept proposal failed", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(`{"status":"accepted"}`)); err != nil {
		h.logger.Error("failed to write response", zap.Error(err))
	}
}

// CreateInvitation handles POST /invites
func (h *Handler) CreateInvitation(w http.ResponseWriter, r *http.Request) {
	if !RequireScope(r.Context(), ScopeEscrowWrite) {
		http.Error(w, "insufficient scope", http.StatusForbidden)
		return
	}

	var req CreateInvitationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	// Extract logical User ID from context (injected by AuthMiddleware)
	userID, ok := r.Context().Value(AuthSubKey).(string)
	if !ok {
		h.logger.Error("user ID not found in context")
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	invitation, err := h.escrowService.CreateInvitation(r.Context(), userID, req.InviteeEmail, req.InviteeRole, req.InviteeType, req.Terms)
	if err != nil {
		h.logger.Error("create invitation failed", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(invitation); err != nil {
		h.logger.Error("failed to encode response", zap.Error(err))
	}
}

// ClaimInvitation handles POST /invites/token/{token}/claim
func (h *Handler) ClaimInvitation(w http.ResponseWriter, r *http.Request) {
	if !RequireScope(r.Context(), ScopeEscrowAccept) {
		http.Error(w, "insufficient scope", http.StatusForbidden)
		return
	}

	token := chi.URLParam(r, "token")
	if token == "" {
		http.Error(w, "missing token", http.StatusBadRequest)
		return
	}

	// Extract logical User ID and Email from the JWT (authoritative)
	userID, _ := r.Context().Value(AuthSubKey).(string)
	userEmail, _ := r.Context().Value(EmailKey).(string)

	if userID == "" || userEmail == "" {
		http.Error(w, "unauthorized: missing identity claims", http.StatusUnauthorized)
		return
	}

	// 1. Resolve the Invitation by token
	invite, err := h.escrowService.GetInvitationByToken(r.Context(), token)
	if err != nil {
		h.logger.Error("invitation lookup failed", zap.Error(err))
		http.Error(w, "invitation not found", http.StatusNotFound)
		return
	}

	// 2. Authoritative Verification: JWT Email MUST match Invitation Email
	if invite.InviteeEmail != userEmail {
		h.logger.Warn("invitation email mismatch", 
			zap.String("jwtEmail", userEmail), 
			zap.String("inviteEmail", invite.InviteeEmail))
		http.Error(w, "forbidden: this invitation belongs to another email address", http.StatusForbidden)
		return
	}

	// 3. Claim the invitation as the authenticated user
	proposal, err := h.escrowService.ClaimInvitation(r.Context(), invite.ID, userID)
	if err != nil {
		h.logger.Error("claim invitation failed", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(proposal); err != nil {
		h.logger.Error("failed to encode response", zap.Error(err))
	}
}

// ListInvitations handles GET /invites
func (h *Handler) ListInvitations(w http.ResponseWriter, r *http.Request) {
	if !RequireScope(r.Context(), ScopeEscrowRead) {
		http.Error(w, "insufficient scope", http.StatusForbidden)
		return
	}

	userID := r.URL.Query().Get("user")
	if userID == "" {
		userID = "Buyer"
	}

	invitations, err := h.escrowService.ListInvitations(r.Context(), userID)
	if err != nil {
		h.logger.Error("list invitations failed", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(invitations); err != nil {
		h.logger.Error("failed to encode response", zap.Error(err))
	}
}

// GetInvitationByToken handles GET /invites/token/{token} (Anonymous)
func (h *Handler) GetInvitationByToken(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	h.logger.Info("anonymous token lookup", zap.String("token", token))
	if token == "" {
		http.Error(w, "missing token", http.StatusBadRequest)
		return
	}

	invitation, err := h.escrowService.GetInvitationByToken(r.Context(), token)
	if err != nil {
		h.logger.Error("get invitation by token failed", zap.Error(err))
		http.Error(w, "invitation not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(invitation); err != nil {
		h.logger.Error("failed to encode response", zap.Error(err))
	}
}

// ListEscrows handles GET /escrows
func (h *Handler) ListEscrows(w http.ResponseWriter, r *http.Request) {
	if !RequireScope(r.Context(), ScopeEscrowRead) {
		http.Error(w, "insufficient scope", http.StatusForbidden)
		return
	}

	userID := r.URL.Query().Get("user")
	if userID == "" {
		userID = "Buyer"
	}

	escrows, err := h.escrowService.ListEscrows(r.Context(), userID)
	if err != nil {
		h.logger.Error("list escrows failed", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(escrows); err != nil {
		h.logger.Error("failed to encode response", zap.Error(err))
	}
}

// ListProposals handles GET /escrows/proposals
func (h *Handler) ListProposals(w http.ResponseWriter, r *http.Request) {
	if !RequireScope(r.Context(), ScopeEscrowRead) {
		http.Error(w, "insufficient scope", http.StatusForbidden)
		return
	}

	userID := r.URL.Query().Get("user")
	if userID == "" {
		userID = "Buyer"
	}

	proposals, err := h.escrowService.ListProposals(r.Context(), userID)
	if err != nil {
		h.logger.Error("list proposals failed", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(proposals); err != nil {
		h.logger.Error("failed to encode response", zap.Error(err))
	}
}

// GetEscrow handles GET /escrows/{escrowID}
func (h *Handler) GetEscrow(w http.ResponseWriter, r *http.Request) {
	if !RequireScope(r.Context(), ScopeEscrowRead) {
		http.Error(w, "insufficient scope", http.StatusForbidden)
		return
	}

	id := chi.URLParam(r, "escrowID")
	userID := r.URL.Query().Get("user")
	if userID == "" {
		userID = "Buyer"
	}

	escrow, err := h.escrowService.GetEscrow(r.Context(), id, userID)
	if err != nil {
		h.logger.Error("get escrow failed", zap.Error(err), zap.String("id", id))
		http.Error(w, "escrow not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(escrow); err != nil {
		h.logger.Error("failed to encode response", zap.Error(err))
	}
}

// ReleaseFunds handles POST /escrows/{escrowID}/release
func (h *Handler) ReleaseFunds(w http.ResponseWriter, r *http.Request) {
	if !RequireScope(r.Context(), ScopeEscrowAccept) {
		http.Error(w, "insufficient scope", http.StatusForbidden)
		return
	}

	id := chi.URLParam(r, "escrowID")
	userID, _ := r.Context().Value(AuthSubKey).(string)

	if err := h.escrowService.ReleaseFunds(r.Context(), id, userID); err != nil {
		h.logger.Error("release funds failed", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("ok")); err != nil {
		h.logger.Error("failed to write response", zap.Error(err))
	}
}

// RefundBuyer handles POST /escrows/{escrowID}/refund
func (h *Handler) RefundBuyer(w http.ResponseWriter, r *http.Request) {
	if !RequireScope(r.Context(), ScopeEscrowWrite) {
		http.Error(w, "insufficient scope", http.StatusForbidden)
		return
	}

	id := chi.URLParam(r, "escrowID")
	if err := h.escrowService.RefundBuyer(r.Context(), id); err != nil {
		h.logger.Error("refund buyer failed", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("ok")); err != nil {
		h.logger.Error("failed to write response", zap.Error(err))
	}
}

// RefundBySeller handles POST /escrows/{escrowID}/refund-by-seller
func (h *Handler) RefundBySeller(w http.ResponseWriter, r *http.Request) {
	if !RequireScope(r.Context(), ScopeEscrowWrite) {
		http.Error(w, "insufficient scope", http.StatusForbidden)
		return
	}

	id := chi.URLParam(r, "escrowID")
	if err := h.escrowService.RefundBySeller(r.Context(), id); err != nil {
		h.logger.Error("seller refund failed", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("ok")); err != nil {
		h.logger.Error("failed to write response", zap.Error(err))
	}
}

// RaiseDispute handles POST /escrows/{escrowID}/dispute
func (h *Handler) RaiseDispute(w http.ResponseWriter, r *http.Request) {
	if !RequireScope(r.Context(), ScopeEscrowWrite) && !RequireScope(r.Context(), ScopeEscrowAccept) {
		http.Error(w, "insufficient scope", http.StatusForbidden)
		return
	}

	id := chi.URLParam(r, "escrowID")
	userID, _ := r.Context().Value(AuthSubKey).(string)

	disputeID, err := h.escrowService.RaiseDispute(r.Context(), id, userID)
	if err != nil {
		h.logger.Error("raise dispute failed", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"disputeId": disputeID})
}

// ResolveDispute handles POST /escrows/{escrowID}/resolve
func (h *Handler) ResolveDispute(w http.ResponseWriter, r *http.Request) {
	if !RequireScope(r.Context(), ScopeSystemAdmin) && !RequireScope(r.Context(), ScopeEscrowAccept) {
		http.Error(w, "insufficient scope", http.StatusForbidden)
		return
	}

	id := chi.URLParam(r, "escrowID")
	userID, _ := r.Context().Value(AuthSubKey).(string)

	var req struct {
		PayoutToBuyer  float64 `json:"payoutToBuyer"`
		PayoutToSeller float64 `json:"payoutToSeller"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if err := h.escrowService.ResolveDispute(r.Context(), id, req.PayoutToBuyer, req.PayoutToSeller, userID); err != nil {
		h.logger.Error("resolve dispute failed", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("ok")); err != nil {
		h.logger.Error("failed to write response", zap.Error(err))
	}
}

// GetHealth handles GET /health
func (h *Handler) GetHealth(w http.ResponseWriter, r *http.Request) {
	health := h.metricsService.GetHealth()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(health); err != nil {
		h.logger.Error("failed to encode health response", zap.Error(err))
	}
}

// GetIdentity handles GET /auth/me for JIT provisioning
func (h *Handler) GetIdentity(w http.ResponseWriter, r *http.Request) {
	sub, ok := r.Context().Value(AuthSubKey).(string)
	if !ok || sub == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	email, _ := r.Context().Value(EmailKey).(string)

	// Attempt to get existing identity
	identity, err := h.escrowService.GetIdentity(r.Context(), sub)
	if err != nil {
		h.logger.Error("failed to fetch identity", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	// If identity doesn't exist, JIT provision
	if identity == nil {
		h.logger.Info("identity not found, triggering JIT provisioning", zap.String("sub", sub))
		identity, err = h.escrowService.ProvisionUser(r.Context(), sub, email)
		if err != nil {
			h.logger.Error("JIT provisioning failed", zap.Error(err))
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(identity); err != nil {
		h.logger.Error("failed to encode response", zap.Error(err))
	}
}

// GetMetrics handles GET /metrics
func (h *Handler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	if !RequireScope(r.Context(), ScopeSystemAdmin) && !RequireScope(r.Context(), ScopeEscrowRead) {
		http.Error(w, "insufficient scope", http.StatusForbidden)
		return
	}

	userID := r.URL.Query().Get("user")
	if userID == "" {
		userID = ledger.CentralBankUser
	}

	metrics, err := h.escrowService.GetMetrics(r.Context(), userID)
	if err != nil {
		h.logger.Error("get metrics failed", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	// Override mock performance with real-time stats if we have the service
	if h.metricsService != nil {
		latency, reqs, errRate, mem, cpu, uptime := h.metricsService.GetSystemPerformance()
		metrics.SystemPerformance.APILatencyMS = latency
		metrics.SystemPerformance.RequestCount = reqs
		metrics.SystemPerformance.ErrorRate = errRate
		metrics.SystemPerformance.MemoryUsage = mem
		metrics.SystemPerformance.CPUUsage = cpu
		metrics.SystemPerformance.Uptime = uptime
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(metrics); err != nil {
		h.logger.Error("failed to encode response", zap.Error(err))
	}
}

// ListSettlements handles GET /settlements
func (h *Handler) ListSettlements(w http.ResponseWriter, r *http.Request) {
	if !RequireScope(r.Context(), ScopeSystemAdmin) && !RequireScope(r.Context(), ScopeEscrowRead) {
		http.Error(w, "insufficient scope", http.StatusForbidden)
		return
	}

	settlements, err := h.escrowService.ListSettlements(r.Context())
	if err != nil {
		h.logger.Error("list settlements failed", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(settlements); err != nil {
		h.logger.Error("failed to encode response", zap.Error(err))
	}
}

// SettlePayment handles POST /settlements/{settlementID}/settle
func (h *Handler) SettlePayment(w http.ResponseWriter, r *http.Request) {
	if !RequireScope(r.Context(), ScopeSystemAdmin) {
		http.Error(w, "insufficient scope", http.StatusForbidden)
		return
	}

	id := chi.URLParam(r, "settlementID")
	if err := h.escrowService.SettlePayment(r.Context(), id); err != nil {
		h.logger.Error("settle payment failed", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("ok")); err != nil {
		h.logger.Error("failed to write response", zap.Error(err))
	}
}

// ListWallets handles GET /wallets (Mocked for Phase 4)
func (h *Handler) ListWallets(w http.ResponseWriter, r *http.Request) {
	if !RequireScope(r.Context(), ScopeEscrowRead) {
		http.Error(w, "insufficient scope", http.StatusForbidden)
		return
	}

	userID := r.URL.Query().Get("user")
	if userID == "" {
		userID = "Buyer"
	}

	// Mocked stablecoin accounts
	wallets := []*ledger.Wallet{
		{ID: "wallet-usd-001", Owner: userID, Currency: "USD", Balance: 15000.50},
		{ID: "wallet-eur-002", Owner: userID, Currency: "EUR", Balance: 4200.00},
		{ID: "wallet-gbp-003", Owner: userID, Currency: "GBP", Balance: 850.75},
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(wallets); err != nil {
		h.logger.Error("failed to encode response", zap.Error(err))
	}
}

// OracleMilestoneTrigger handles POST /webhooks/milestone
func (h *Handler) OracleMilestoneTrigger(w http.ResponseWriter, r *http.Request) {
	// Webhooks typically use their own secret verification instead of OIDC
	var req ledger.OracleWebhookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if err := h.escrowService.ProcessOracleWebhook(r.Context(), req); err != nil {
		h.logger.Error("webhook processing failed", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("ok")); err != nil {
		h.logger.Error("failed to write response", zap.Error(err))
	}
}
