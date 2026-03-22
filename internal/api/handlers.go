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
}

func NewHandler(logger *zap.Logger, escrowService *services.EscrowService, metricsService *services.MetricsService) *Handler {
	return &Handler{
		logger:         logger,
		escrowService:  escrowService,
		metricsService: metricsService,
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
	json.NewEncoder(w).Encode(proposal)
}

// AcceptProposal handles POST /escrows/{id}/accept
func (h *Handler) AcceptProposal(w http.ResponseWriter, r *http.Request) {
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
	w.Write([]byte(`{"status":"accepted"}`))
}

// ListEscrows handles GET /escrows
func (h *Handler) ListEscrows(w http.ResponseWriter, r *http.Request) {
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
	json.NewEncoder(w).Encode(proposals)
}

// GetEscrow handles GET /escrows/{escrowID}
func (h *Handler) GetEscrow(w http.ResponseWriter, r *http.Request) {
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
	id := chi.URLParam(r, "escrowID")
	if err := h.escrowService.ReleaseFunds(r.Context(), id); err != nil {
		h.logger.Error("release funds failed", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

// RefundBuyer handles POST /escrows/{escrowID}/refund
func (h *Handler) RefundBuyer(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "escrowID")
	if err := h.escrowService.RefundBuyer(r.Context(), id); err != nil {
		h.logger.Error("refund buyer failed", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

// RefundBySeller handles POST /escrows/{escrowID}/refund-by-seller
func (h *Handler) RefundBySeller(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "escrowID")
	if err := h.escrowService.RefundBySeller(r.Context(), id); err != nil {
		h.logger.Error("seller refund failed", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

// ResolveDispute handles POST /escrows/{escrowID}/resolve
func (h *Handler) ResolveDispute(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "escrowID")
	var req struct {
		PayoutToBuyer  float64 `json:"payoutToBuyer"`
		PayoutToSeller float64 `json:"payoutToSeller"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if err := h.escrowService.ResolveDispute(r.Context(), id, req.PayoutToBuyer, req.PayoutToSeller); err != nil {
		h.logger.Error("resolve dispute failed", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

// GetMetrics handles GET /metrics
func (h *Handler) GetMetrics(w http.ResponseWriter, r *http.Request) {
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
	id := chi.URLParam(r, "settlementID")
	if err := h.escrowService.SettlePayment(r.Context(), id); err != nil {
		h.logger.Error("settle payment failed", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

// ListWallets handles GET /wallets (Mocked for Phase 4)
func (h *Handler) ListWallets(w http.ResponseWriter, r *http.Request) {
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
	json.NewEncoder(w).Encode(wallets)
}

// OracleMilestoneTrigger handles POST /webhooks/milestone
func (h *Handler) OracleMilestoneTrigger(w http.ResponseWriter, r *http.Request) {
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
	_, _ = w.Write([]byte("ok"))
}
