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

func NewHandler(
	logger *zap.Logger,
	escrowService *services.EscrowService,
) *Handler {
	return &Handler{
		logger:        logger,
		escrowService: escrowService,
	}
}

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

	json.NewEncoder(w).Encode(escrow)
}

func (h *Handler) GetEscrow(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "escrowID")

	escrow, err := h.escrowService.GetEscrow(r.Context(), id)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(escrow)
}

func (h *Handler) ReleaseEscrow(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "escrowID")

	err := h.escrowService.ReleaseEscrow(r.Context(), id)
	if err != nil {
		http.Error(w, "release failed", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) RefundEscrow(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "escrowID")

	err := h.escrowService.RefundEscrow(r.Context(), id)
	if err != nil {
		http.Error(w, "refund failed", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
