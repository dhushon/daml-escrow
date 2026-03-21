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
	Amount      float64               `json:"amount" example:"100.0"`
	Currency    string                `json:"currency" example:"USD"`
	Description string                `json:"description" example:"Payment for goods"`
	Milestones  []ledger.Milestone    `json:"milestones,omitempty"`
	Metadata    ledger.EscrowMetadata `json:"metadata,omitempty"`
}

type ResolveDisputeRequest struct {
	PayoutToBuyer  float64 `json:"payoutToBuyer" example:"50.0"`
	PayoutToSeller float64 `json:"payoutToSeller" example:"50.0"`
}

type EscrowResponse struct {
	ID                    string                `json:"id"`
	Buyer                 string                `json:"buyer"`
	Seller                string                `json:"seller"`
	Issuer                string                `json:"issuer"`
	Mediator              string                `json:"mediator"`
	Amount                float64               `json:"amount"`
	Currency              string                `json:"currency"`
	State                 string                `json:"state"`
	Milestones            []ledger.Milestone    `json:"milestones"`
	CurrentMilestoneIndex int                   `json:"currentMilestoneIndex"`
	Metadata              ledger.EscrowMetadata `json:"metadata"`
}

type SettlementResponse struct {
	ID        string  `json:"id"`
	Issuer    string  `json:"issuer"`
	Recipient string  `json:"recipient"`
	Amount    float64 `json:"amount"`
	Currency  string  `json:"currency"`
	Status    string  `json:"status"`
}

// CreateEscrow handles POST /escrows
func (h *Handler) CreateEscrow(w http.ResponseWriter, r *http.Request) {
	var req CreateEscrowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.logger.Info("received create escrow request", zap.Any("request", req))

	// Map API request to service request
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
	if err := json.NewEncoder(w).Encode(mapToResponse(escrow)); err != nil {
		h.logger.Error("failed to encode response", zap.Error(err))
	}
}

// GetEscrow handles GET /escrows/{id}
func (h *Handler) GetEscrow(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	escrow, err := h.escrowService.GetEscrow(r.Context(), id)
	if err != nil {
		h.logger.Error("get escrow failed", zap.Error(err))
		http.Error(w, "escrow not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(mapToResponse(escrow)); err != nil {
		h.logger.Error("failed to encode response", zap.Error(err))
	}
}

// ListEscrows handles GET /escrows
func (h *Handler) ListEscrows(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user")
	if userID == "" {
		userID = ledger.BuyerUser // Default to Buyer view
	}

	escrows, err := h.escrowService.ListEscrows(r.Context(), userID)
	if err != nil {
		h.logger.Error("list escrows failed", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	var resp []EscrowResponse
	for _, e := range escrows {
		resp = append(resp, mapToResponse(e))
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Error("failed to encode response", zap.Error(err))
	}
}

// ReleaseFunds handles POST /escrows/{id}/release
func (h *Handler) ReleaseFunds(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	err := h.escrowService.ReleaseFunds(r.Context(), id)
	if err != nil {
		h.logger.Error("release funds failed", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// RefundBuyer handles POST /escrows/{id}/refund
func (h *Handler) RefundBuyer(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	err := h.escrowService.RefundBuyer(r.Context(), id)
	if err != nil {
		h.logger.Error("refund failed", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// RefundBySeller handles POST /escrows/{id}/refund-by-seller
func (h *Handler) RefundBySeller(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	err := h.escrowService.RefundBySeller(r.Context(), id)
	if err != nil {
		h.logger.Error("seller refund failed", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ResolveDispute handles POST /escrows/{id}/resolve
func (h *Handler) ResolveDispute(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req ResolveDisputeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err := h.escrowService.ResolveDispute(r.Context(), id, req.PayoutToBuyer, req.PayoutToSeller)
	if err != nil {
		h.logger.Error("resolve dispute failed", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
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

	var resp []SettlementResponse
	for _, s := range settlements {
		resp = append(resp, SettlementResponse{
			ID:        s.ID,
			Issuer:    s.Issuer,
			Recipient: s.Recipient,
			Amount:    s.Amount,
			Currency:  s.Currency,
			Status:    s.Status,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Error("failed to encode response", zap.Error(err))
	}
}

// SettlePayment handles POST /settlements/{id}/settle
func (h *Handler) SettlePayment(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	err := h.escrowService.SettlePayment(r.Context(), id)
	if err != nil {
		h.logger.Error("settle failed", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func mapToResponse(e *ledger.EscrowContract) EscrowResponse {
	return EscrowResponse{
		ID:                    e.ID,
		Buyer:                 e.Buyer,
		Seller:                e.Seller,
		Issuer:                e.Issuer,
		Mediator:              e.Mediator,
		Amount:                e.Amount,
		Currency:              e.Currency,
		State:                 e.State,
		Milestones:            e.Milestones,
		CurrentMilestoneIndex: e.CurrentMilestoneIndex,
		Metadata:              e.Metadata,
	}
}

// OracleMilestoneTrigger handles POST /webhooks/milestone
func (h *Handler) OracleMilestoneTrigger(w http.ResponseWriter, r *http.Request) {
	var req ledger.OracleWebhookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("failed to decode webhook", zap.Error(err))
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	// Any:Any metadata logging for visibility
	if len(req.Metadata) > 0 {
		h.logger.Info("webhook metadata received", zap.Any("metadata", req.Metadata))
	}

	if err := h.escrowService.ProcessOracleWebhook(r.Context(), req); err != nil {
		h.logger.Error("webhook processing failed", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}
