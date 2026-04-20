package api

import (
	"daml-escrow/internal/ledger"
	"daml-escrow/internal/services"
	"encoding/json"
	"net/http"

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
	if len(val) == 0 {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	h.renderJSON(w, val)
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
	var req ledger.CreateEscrowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	proposal, err := h.escrowService.ProposeEscrow(r.Context(), req)
	if err != nil {
		h.logger.Error("failed to propose escrow", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	h.renderJSON(w, proposal)
}

func (h *Handler) SellerAccept(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "escrowID")
	userID, _ := r.Context().Value(AuthSubKey).(string)

	newID, err := h.escrowService.SellerAccept(r.Context(), id, userID)
	if err != nil {
		h.logger.Error("failed to accept escrow", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	h.renderJSON(w, map[string]string{"id": newID})
}

func (h *Handler) Fund(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "escrowID")
	userID, _ := r.Context().Value(AuthSubKey).(string)

	var args struct {
		CustodyRef string `json:"custodyRef"`
		HoldingCid string `json:"holdingCid"`
	}
	if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.escrowService.Fund(r.Context(), id, args.CustodyRef, args.HoldingCid, userID); err != nil {
		h.logger.Error("failed to fund escrow", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) Activate(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "escrowID")
	userID, _ := r.Context().Value(AuthSubKey).(string)

	if _, err := h.escrowService.Activate(r.Context(), id, []string{userID}); err != nil {
		h.logger.Error("failed to activate escrow", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) ConfirmConditions(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "escrowID")
	userID, _ := r.Context().Value(AuthSubKey).(string)

	if err := h.escrowService.ConfirmConditions(r.Context(), id, userID); err != nil {
		h.logger.Error("failed to confirm conditions", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) RaiseDispute(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "escrowID")
	userID, _ := r.Context().Value(AuthSubKey).(string)

	if err := h.escrowService.RaiseDispute(r.Context(), id, userID); err != nil {
		h.logger.Error("failed to raise dispute", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) ProposeSettlement(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "escrowID")
	userID, _ := r.Context().Value(AuthSubKey).(string)

	var settlement ledger.SettlementTerms
	if err := json.NewDecoder(r.Body).Decode(&settlement); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	newID, err := h.escrowService.ProposeSettlement(r.Context(), id, settlement, userID)
	if err != nil {
		h.logger.Error("failed to propose settlement", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	h.renderJSON(w, map[string]string{"id": newID})
}

func (h *Handler) RatifySettlement(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "escrowID")
	userID, _ := r.Context().Value(AuthSubKey).(string)

	newID, err := h.escrowService.RatifySettlement(r.Context(), id, userID)
	if err != nil {
		h.logger.Error("failed to ratify settlement", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	h.renderJSON(w, map[string]string{"id": newID})
}

func (h *Handler) FinalizeSettlement(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "escrowID")
	userID, _ := r.Context().Value(AuthSubKey).(string)

	newID, err := h.escrowService.FinalizeSettlement(r.Context(), id, userID)
	if err != nil {
		h.logger.Error("failed to finalize settlement", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	h.renderJSON(w, map[string]string{"id": newID})
}

func (h *Handler) Disburse(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "escrowID")
	userID, _ := r.Context().Value(AuthSubKey).(string)

	if err := h.escrowService.Disburse(r.Context(), id, []string{userID}); err != nil {
		h.logger.Error("failed to disburse escrow", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) ListEscrows(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(AuthSubKey).(string)
	escrows, err := h.escrowService.ListEscrows(r.Context(), userID)
	if err != nil {
		h.logger.Error("failed to list escrows", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	h.renderJSON(w, escrows)
}

func (h *Handler) GetEscrow(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "escrowID")
	userID, _ := r.Context().Value(AuthSubKey).(string)

	escrow, err := h.escrowService.GetEscrow(r.Context(), id, userID)
	if err != nil {
		h.logger.Error("failed to get escrow", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	h.renderJSON(w, escrow)
}

func (h *Handler) GetEscrowLifecycle(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "escrowID")
	userID, _ := r.Context().Value(AuthSubKey).(string)

	// Fetch current state to drive lifecycle visualization
	escrow, err := h.escrowService.GetEscrow(r.Context(), id, userID)
	if err != nil {
		h.logger.Error("failed to fetch escrow for lifecycle", zap.Error(err))
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	events, err := h.analyticsService.GetEscrowLifecycle(r.Context(), id, escrow.State)
	if err != nil {
		h.logger.Error("failed to fetch lifecycle", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	h.renderJSON(w, events)
}

func (h *Handler) OracleMilestoneTrigger(w http.ResponseWriter, r *http.Request) {
	var body struct {
		EscrowID       string `json:"escrowId"`
		MilestoneIndex int    `json:"milestoneIndex"`
		Event          string `json:"event"`
		Signature      string `json:"signature"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.escrowService.OracleMilestoneTrigger(r.Context(), body.EscrowID, body.MilestoneIndex, body.Event, body.Signature); err != nil {
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
		h.logger.Error("failed to fetch metrics", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	h.renderJSON(w, metrics)
}

func (h *Handler) ListSettlements(w http.ResponseWriter, r *http.Request) {
	settlements, err := h.escrowService.ListSettlements(r.Context())
	if err != nil {
		h.logger.Error("failed to list settlements", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	h.renderJSON(w, settlements)
}

func (h *Handler) SettlePayment(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "settlementID")
	if err := h.escrowService.SettlePayment(r.Context(), id); err != nil {
		h.logger.Error("failed to settle payment", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) ListWallets(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(AuthSubKey).(string)
	wallets, err := h.escrowService.ListWallets(r.Context(), userID)
	if err != nil {
		h.logger.Error("failed to list wallets", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	h.renderJSON(w, wallets)
}

// ---------------------------------------------------------------------------
// Identity Handlers
// ---------------------------------------------------------------------------

func (h *Handler) GetIdentity(w http.ResponseWriter, r *http.Request) {
	sub, ok := r.Context().Value(AuthSubKey).(string)
	if !ok || sub == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	email, _ := r.Context().Value(EmailKey).(string)

	identity, err := h.escrowService.GetIdentity(r.Context(), sub)
	if err != nil {
		h.logger.Error("failed to get identity", zap.String("sub", sub), zap.Error(err))
		http.Error(w, "internal error: identity lookup failed", http.StatusInternalServerError)
		return
	}

	if identity == nil {
		scopes, _ := r.Context().Value(ScopesKey).([]string)
		identity, err = h.escrowService.ProvisionUser(r.Context(), sub, email, scopes)
		if err != nil {
			h.logger.Error("failed to provision user", zap.String("sub", sub), zap.Error(err))
			http.Error(w, "internal error: user provisioning failed", http.StatusInternalServerError)
			return
		}
	}

	h.renderJSON(w, identity)
}

func (h *Handler) ListIdentities(w http.ResponseWriter, r *http.Request) {
	identities, err := h.escrowService.ListIdentities(r.Context())
	if err != nil {
		h.logger.Error("list identities failed", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	h.renderJSON(w, identities)
}

func (h *Handler) GetHealth(w http.ResponseWriter, r *http.Request) {
	health := h.metricsService.GetHealth(
		h.configService,
		h.escrowService.GetLedgerClient(),
		h.escrowService.GetOracleSecret(),
	)
	h.renderJSON(w, health)
}

func (h *Handler) DiscoverAuth(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	if email == "" {
		http.Error(w, "missing email parameter", http.StatusBadRequest)
		return
	}

	idp := h.identityService.DiscoverProvider(r.Context(), email)
	h.renderJSON(w, idp)
}

func (h *Handler) CreateInvitation(w http.ResponseWriter, r *http.Request) {
	inviterID, _ := r.Context().Value(AuthSubKey).(string)

	var req struct {
		InviteeEmail string             `json:"inviteeEmail"`
		InviteeRole  string             `json:"inviteeRole"`
		InviteeType  string             `json:"inviteeType"`
		Asset        ledger.Asset       `json:"asset"`
		Terms        ledger.EscrowTerms `json:"terms"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	invitation, err := h.escrowService.CreateInvitation(r.Context(), inviterID, req.InviteeEmail, req.InviteeRole, req.InviteeType, req.Asset, req.Terms)
	if err != nil {
		h.logger.Error("failed to create invitation", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	h.renderJSON(w, invitation)
}

func (h *Handler) ListInvitations(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(AuthSubKey).(string)
	invites, err := h.escrowService.ListInvitations(r.Context(), userID)
	if err != nil {
		h.logger.Error("failed to list invitations", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	h.renderJSON(w, invites)
}

func (h *Handler) GetInvitationByToken(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	invitation, err := h.escrowService.GetInvitationByToken(r.Context(), token)
	if err != nil {
		h.logger.Warn("invitation lookup failed", zap.String("token", token), zap.Error(err))
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	h.renderJSON(w, invitation)
}

func (h *Handler) ClaimInvitation(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	claimantID, _ := r.Context().Value(AuthSubKey).(string)
	claimantEmail, _ := r.Context().Value(EmailKey).(string)

	inv, err := h.escrowService.GetInvitationByToken(r.Context(), token)
	if err != nil {
		http.Error(w, "invitation not found", http.StatusNotFound)
		return
	}

	if inv.InviteeEmail != claimantEmail {
		h.logger.Warn("invitation claim blocked: email mismatch", zap.String("expected", inv.InviteeEmail), zap.String("got", claimantEmail))
		http.Error(w, "unauthorized: email mismatch", http.StatusForbidden)
		return
	}

	_, err = h.escrowService.GetIdentity(r.Context(), claimantID)
	if err != nil {
		scopes := []string{"escrow:read", "escrow:accept"}
		_, err = h.escrowService.ProvisionUser(r.Context(), claimantID, claimantEmail, scopes)
		if err != nil {
			h.logger.Error("jit provisioning failed during claim", zap.Error(err))
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
	}

	proposal, err := h.escrowService.ClaimInvitation(r.Context(), inv.ID, claimantID)
	if err != nil {
		h.logger.Error("ledger claim failed", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	h.renderJSON(w, proposal)
}

func (h *Handler) renderJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("failed to render JSON", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}
