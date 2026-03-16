package api

import (
	"encoding/json"
	"net/http"

	"daml-escrow/internal/ledger"
	"daml-escrow/internal/services"

	chi "github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type Handler struct {
	logger        *zap.Logger
	escrowService *services.EscrowService
}

func NewHandler(
	logger *zap.Logger,
	escrowService *services.EscrowService,
) *Handler {
	return &Handler{
		logger:        logger,
		escrowService: escrowService,
	}
}

// CreateEscrow handles creation of a new escrow
// @Summary Create a new escrow contract
// @Description Initiate a new escrow agreement between buyer and seller
// @Tags escrows
// @Accept json
// @Produce json
// @Param request body ledger.CreateEscrowRequest true "Escrow Creation Request"
// @Success 201 {object} ledger.EscrowContract
// @Failure 400 {string} string "invalid request"
// @Failure 500 {string} string "internal error"
// @Router /escrows [post]
func (h *Handler) CreateEscrow(w http.ResponseWriter, r *http.Request) {

	var req ledger.CreateEscrowRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	escrow, err := h.escrowService.CreateEscrow(r.Context(), req)
	if err != nil {

		h.logger.Error("create escrow failed", zap.Error(err))

		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(escrow); err != nil {
		h.logger.Error("encode escrow failed", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
}

// GetEscrow handles retrieving escrow details
// @Summary Get escrow details by ID
// @Description Retrieve information about a specific escrow contract
// @Tags escrows
// @Produce json
// @Param escrowID path string true "Escrow ID"
// @Success 200 {object} ledger.EscrowContract
// @Failure 404 {string} string "not found"
// @Router /escrows/{escrowID} [get]
func (h *Handler) GetEscrow(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "escrowID")

	escrow, err := h.escrowService.GetEscrow(r.Context(), id)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	if err := json.NewEncoder(w).Encode(escrow); err != nil {
		h.logger.Error("encode escrow failed", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
}

// ReleaseEscrow handles fund release
// @Summary Release funds for the current milestone
// @Description Approve and release funds for the active milestone
// @Tags escrows
// @Param escrowID path string true "Escrow ID"
// @Success 200 {string} string "ok"
// @Failure 500 {string} string "release failed"
// @Router /escrows/{escrowID}/release [post]
func (h *Handler) ReleaseEscrow(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "escrowID")

	err := h.escrowService.ReleaseEscrow(r.Context(), id)
	if err != nil {
		http.Error(w, "release failed", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// RefundEscrow handles refund process
// @Summary Refund funds to the buyer
// @Description Cancel escrow and return funds (requires mediator)
// @Tags escrows
// @Param escrowID path string true "Escrow ID"
// @Success 200 {string} string "ok"
// @Failure 500 {string} string "refund failed"
// @Router /escrows/{escrowID}/refund [post]
func (h *Handler) RefundEscrow(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "escrowID")

	err := h.escrowService.RefundEscrow(r.Context(), id)
	if err != nil {
		http.Error(w, "refund failed", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
