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
	logger           *zap.Logger
	escrowService    *services.EscrowService
	metricsService   *services.MetricsService
	configService    *services.ConfigService
	analyticsService *services.AnalyticsService
	identityService  *services.IdentityService
}

func NewHandler(
	logger *zap.Logger,
	escrowService *services.EscrowService,
	metricsService *services.MetricsService,
	configService *services.ConfigService,
	analyticsService *services.AnalyticsService,
	identityService *services.IdentityService,
) *Handler {
	return &Handler{
		logger:           logger,
		escrowService:    escrowService,
		metricsService:   metricsService,
		configService:    configService,
		analyticsService: analyticsService,
		identityService:  identityService,
	}
}

func (h *Handler) GetHealth(w http.ResponseWriter, r *http.Request) {
	h.logger.Debug("health check requested")
	health := h.metricsService.GetHealth(h.configService, h.escrowService.GetLedgerClient(), h.escrowService.GetOracleSecret())
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(health)
}

func (h *Handler) GetIdentity(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(AuthSubKey).(string)
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized"})
		return
	}
	email, _ := r.Context().Value(EmailKey).(string)

	identity, err := h.identityService.GetOrCreateIdentity(r.Context(), userID, email, h.escrowService.GetLedgerClient())
	if err != nil {
		h.logger.Error("failed to resolve identity", zap.Error(err))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Internal Error: " + err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(identity)
}

func (h *Handler) DiscoverAuth(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	if email == "" {
		http.Error(w, "email parameter required", http.StatusBadRequest)
		return
	}

	config, err := h.identityService.GetIdPConfigForEmail(email)
	if err != nil {
		http.Error(w, "provider not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(config)
}

func (h *Handler) ListIdentities(w http.ResponseWriter, r *http.Request) {
	identities, err := h.escrowService.GetLedgerClient().ListIdentities(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(identities)
}

func (h *Handler) GetConfig(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(AuthSubKey).(string)
	cfg, err := h.configService.GetConfig(userID, "user-preferences")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(cfg)
}

func (h *Handler) SaveConfig(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(AuthSubKey).(string)
	var body interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	val, _ := json.Marshal(body)
	if err := h.configService.SaveConfig(userID, "user-preferences", json.RawMessage(val)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ListInvitations(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(AuthSubKey).(string)
	invites, err := h.escrowService.GetLedgerClient().ListInvitations(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(invites)
}

func (h *Handler) GetInvitationByToken(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	invite, err := h.escrowService.GetLedgerClient().GetInvitationByToken(r.Context(), token)
	if err != nil {
		http.Error(w, "invitation not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(invite)
}

func (h *Handler) CreateInvitation(w http.ResponseWriter, r *http.Request) {
	var req CreateInvitationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	if err := req.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	userID, _ := r.Context().Value(AuthSubKey).(string)
	asset, terms := req.ToLedgerAssetAndTerms()
	invite, err := h.escrowService.GetLedgerClient().CreateInvitation(r.Context(), userID, req.InviteeEmail, req.InviteeRole, req.InviteeType, asset, terms)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(invite)
}

func (h *Handler) ClaimInvitation(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	userID, _ := r.Context().Value(AuthSubKey).(string)

	_, err := h.escrowService.GetLedgerClient().ClaimInvitation(r.Context(), token, userID)
	if err != nil {
		h.logger.Error("failed to claim invitation", zap.Error(err))
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ProposeEscrow(w http.ResponseWriter, r *http.Request) {
	var req ProposeEscrowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("failed to decode propose escrow request", zap.Error(err))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid json: " + err.Error()})
		return
	}
	if err := req.Validate(); err != nil {
		h.logger.Error("validation failed for propose escrow request", zap.Error(err))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid request: " + err.Error()})
		return
	}

	userID, _ := r.Context().Value(AuthSubKey).(string)
	email, _ := r.Context().Value(EmailKey).(string)
	
	// High-Assurance: Resolve institutional role to handle bilateral initiation
	identity, err := h.identityService.GetOrCreateIdentity(r.Context(), userID, email, h.escrowService.GetLedgerClient())
	if err != nil {
		h.logger.Error("failed to resolve identity for proposal", zap.Error(err))
		http.Error(w, "Identity resolution failed", http.StatusInternalServerError)
		return
	}

	ledgerReq := req.ToLedgerRequest()

	// High-Assurance Sanitization: Ensure all tripartite identities are Daml-compliant
	sanitizedCounterparty, err := ledger.SanitizeID(req.Counterparty)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Counterparty error: " + err.Error()})
		return
	}
	
	sanitizedMediator := ""
	if req.Mediator != "" {
		sanitizedMediator, err = ledger.SanitizeID(req.Mediator)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "Mediator error: " + err.Error()})
			return
		}
	}

	// Bilateral Logic: Map initiator to correct role
	if identity.Role == "Seller" {
		ledgerReq.Seller = identity.DamlUserID
		ledgerReq.Buyer = sanitizedCounterparty
	} else {
		ledgerReq.Buyer = identity.DamlUserID
		ledgerReq.Seller = sanitizedCounterparty
	}
	ledgerReq.Mediator = sanitizedMediator

	// High-Assurance: Authoritatively enforce role exclusivity and mandatory parties
	if err := ledger.ValidateTripartiteRoles(ledgerReq.Buyer, ledgerReq.Seller, ledgerReq.Mediator); err != nil {
		h.logger.Error("institutional mandate violation", zap.Error(err))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	proposal, err := h.escrowService.ProposeEscrow(r.Context(), ledgerReq)
	if err != nil {
		h.logger.Error("escrow proposal failed", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(proposal)
}

func (h *Handler) Fund(w http.ResponseWriter, r *http.Request) {
	escrowID := chi.URLParam(r, "escrowID")
	var req FundEscrowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	userID, _ := r.Context().Value(AuthSubKey).(string)
	err := h.escrowService.FundEscrow(r.Context(), escrowID, userID, req.HoldingCid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) Activate(w http.ResponseWriter, r *http.Request) {
	escrowID := chi.URLParam(r, "escrowID")
	userID, _ := r.Context().Value(AuthSubKey).(string)
	id, err := h.escrowService.ActivateEscrow(r.Context(), escrowID, userID, []string{userID})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, _ = w.Write([]byte(id))
}

func (h *Handler) ConfirmConditions(w http.ResponseWriter, r *http.Request) {
	escrowID := chi.URLParam(r, "escrowID")
	userID, _ := r.Context().Value(AuthSubKey).(string)
	err := h.escrowService.GetLedgerClient().ConfirmConditions(r.Context(), escrowID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) RaiseDispute(w http.ResponseWriter, r *http.Request) {
	escrowID := chi.URLParam(r, "escrowID")
	userID, _ := r.Context().Value(AuthSubKey).(string)
	err := h.escrowService.RaiseDispute(r.Context(), escrowID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ProposeSettlement(w http.ResponseWriter, r *http.Request) {
	escrowID := chi.URLParam(r, "escrowID")
	var req ProposeSettlementRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	userID, _ := r.Context().Value(AuthSubKey).(string)
	id, err := h.escrowService.ProposeSettlement(r.Context(), escrowID, userID, req.BuyerReturn)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, _ = w.Write([]byte(id))
}

func (h *Handler) RatifySettlement(w http.ResponseWriter, r *http.Request) {
	escrowID := chi.URLParam(r, "escrowID")
	userID, _ := r.Context().Value(AuthSubKey).(string)
	id, err := h.escrowService.GetLedgerClient().RatifySettlement(r.Context(), escrowID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, _ = w.Write([]byte(id))
}

func (h *Handler) FinalizeSettlement(w http.ResponseWriter, r *http.Request) {
	escrowID := chi.URLParam(r, "escrowID")
	userID, _ := r.Context().Value(AuthSubKey).(string)
	id, err := h.escrowService.GetLedgerClient().FinalizeSettlement(r.Context(), escrowID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, _ = w.Write([]byte(id))
}

func (h *Handler) Disburse(w http.ResponseWriter, r *http.Request) {
	escrowID := chi.URLParam(r, "escrowID")
	userID, _ := r.Context().Value(AuthSubKey).(string)
	err := h.escrowService.DisburseEscrow(r.Context(), escrowID, userID, []string{userID})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ListEscrows(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(AuthSubKey).(string)
	escrows, err := h.escrowService.ListEscrows(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(escrows)
}

func (h *Handler) GetEscrow(w http.ResponseWriter, r *http.Request) {
	escrowID := chi.URLParam(r, "escrowID")
	userID, _ := r.Context().Value(AuthSubKey).(string)
	escrow, err := h.escrowService.GetEscrow(r.Context(), escrowID, userID)
	if err != nil {
		http.Error(w, "escrow not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(escrow)
}

func (h *Handler) GetEscrowLifecycle(w http.ResponseWriter, r *http.Request) {
	escrowID := chi.URLParam(r, "escrowID")
	userID, _ := r.Context().Value(AuthSubKey).(string)
	lifecycle, err := h.analyticsService.GetEscrowLifecycle(r.Context(), escrowID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(lifecycle)
}

func (h *Handler) OracleMilestoneTrigger(w http.ResponseWriter, r *http.Request) {
	var body OracleWebhookRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.escrowService.OracleMilestoneTrigger(r.Context(), body.EscrowID, body.MilestoneIndex, body.Event, body.Signature, body.Asymmetric); err != nil {
		h.logger.Error("oracle trigger failed", zap.Error(err))
		http.Error(w, "trigger rejected", http.StatusForbidden)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(AuthSubKey).(string)
	metrics, err := h.escrowService.GetMetrics(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(metrics)
}

func (h *Handler) ListSettlements(w http.ResponseWriter, r *http.Request) {
	settlements, err := h.escrowService.GetLedgerClient().ListSettlements(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(settlements)
}

func (h *Handler) SettlePayment(w http.ResponseWriter, r *http.Request) {
	settlementID := chi.URLParam(r, "settlementID")
	err := h.escrowService.SettleEscrow(r.Context(), settlementID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ListWallets(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(AuthSubKey).(string)
	wallets, err := h.escrowService.GetLedgerClient().ListWallets(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(wallets)
}
