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
	logger        *zap.Logger
	escrowService *services.EscrowService
}

func NewHandler(logger *zap.Logger, escrowService *services.EscrowService) *Handler {
	return &Handler{
		logger:        logger,
		escrowService: escrowService,
	}
}

type CreateEscrowRequest struct {
	Buyer       string             `json:"buyer" example:"Buyer"`
	Seller      string             `json:"seller" example:"Seller"`
	Amount      float64            `json:"amount" example:"100.0"`
	Currency    string             `json:"currency" example:"USD"`
	Description string             `json:"description" example:"Payment for goods"`
	Milestones  []ledger.Milestone `json:"milestones,omitempty"`
}

type ResolveDisputeRequest struct {
	PayoutToBuyer  float64 `json:"payoutToBuyer" example:"50.0"`
	PayoutToSeller float64 `json:"payoutToSeller" example:"50.0"`
}

type EscrowResponse struct {
	ID                    string             `json:"id"`
	Buyer                 string             `json:"buyer"`
	Seller                string             `json:"seller"`
	Amount                float64            `json:"amount"`
	Currency              string             `json:"currency"`
	State                 string             `json:"state"`
	Milestones            []ledger.Milestone `json:"milestones"`
	CurrentMilestoneIndex int                `json:"currentMilestoneIndex"`
}

// CreateEscrow handles POST /escrows
// @Summary Create a new escrow contract
// @Description Initiate a new escrow agreement between buyer and seller
// @Tags escrows
// @Accept json
// @Produce json
// @Param request body CreateEscrowRequest true "Escrow Creation Request"
// @Success 201 {object} EscrowResponse
// @Failure 400 {string} string "invalid request"
// @Failure 500 {string} string "internal error"
// @Router /escrows [post]
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
	})
	if err != nil {
		h.logger.Error("create escrow failed", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(mapToResponse(escrow))
}

// GetEscrow handles GET /escrows/{id}
// @Summary Get escrow details by ID
// @Description Retrieve information about a specific escrow contract
// @Tags escrows
// @Produce json
// @Param id path string true "Escrow ID"
// @Success 200 {object} EscrowResponse
// @Failure 404 {string} string "escrow not found"
// @Router /escrows/{id} [get]
func (h *Handler) GetEscrow(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	escrow, err := h.escrowService.GetEscrow(r.Context(), id)
	if err != nil {
		h.logger.Error("get escrow failed", zap.Error(err))
		http.Error(w, "escrow not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(mapToResponse(escrow))
}

// ReleaseFunds handles POST /escrows/{id}/release
// @Summary Release funds for the current milestone
// @Description Approve and release funds for the active milestone
// @Tags escrows
// @Param id path string true "Escrow ID"
// @Success 200 {string} string "ok"
// @Failure 500 {string} string "release failed"
// @Router /escrows/{id}/release [post]
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
// @Summary Refund funds to the buyer
// @Description Cancel escrow and return funds (requires mediator)
// @Tags escrows
// @Param id path string true "Escrow ID"
// @Success 200 {string} string "ok"
// @Failure 500 {string} string "refund failed"
// @Router /escrows/{id}/refund [post]
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

// ResolveDispute handles POST /escrows/{id}/resolve
// @Summary Resolve a disputed escrow
// @Description Mediator resolve dispute by splitting payout
// @Tags escrows
// @Accept json
// @Param id path string true "Escrow ID"
// @Param request body ResolveDisputeRequest true "Dispute Resolution Request"
// @Success 200 {string} string "ok"
// @Failure 400 {string} string "invalid request"
// @Failure 500 {string} string "resolution failed"
// @Router /escrows/{id}/resolve [post]
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

func mapToResponse(e *ledger.EscrowContract) EscrowResponse {
	return EscrowResponse{
		ID:                    e.ID,
		Buyer:                 e.Buyer,
		Seller:                e.Seller,
		Amount:                e.Amount,
		Currency:              e.Currency,
		State:                 e.State,
		Milestones:            e.Milestones,
		CurrentMilestoneIndex: e.CurrentMilestoneIndex,
	}
}
