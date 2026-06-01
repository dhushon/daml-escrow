package api

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"daml-escrow/internal/crypto"
	"daml-escrow/internal/ledger"
	"daml-escrow/internal/services"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Handler struct {
	logger           *zap.Logger
	escrowService    *services.EscrowService
	metricsService   *services.MetricsService
	configService    *services.ConfigService
	analyticsService *services.AnalyticsService
	identityService  *services.IdentityService
	schemaService    *services.SchemaService
	ingestService    *services.IngestService
	storageService   *services.StorageService
}

func NewHandler(
	logger *zap.Logger,
	escrowService *services.EscrowService,
	metricsService *services.MetricsService,
	configService *services.ConfigService,
	analyticsService *services.AnalyticsService,
	identityService *services.IdentityService,
	schemaService *services.SchemaService,
	ingestService *services.IngestService,
	storageService *services.StorageService,
) *Handler {
	return &Handler{
		logger:           logger,
		escrowService:    escrowService,
		metricsService:   metricsService,
		configService:    configService,
		analyticsService: analyticsService,
		identityService:  identityService,
		schemaService:    schemaService,
		ingestService:    ingestService,
		storageService:   storageService,
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

func (h *Handler) GetNonce(w http.ResponseWriter, r *http.Request) {
	nonceID := uuid.New().String()
	challenge := fmt.Sprintf("Sign this challenge to authenticate: %s", nonceID)

	// Persist the challenge in the DB to prevent replays
	if err := h.configService.CreateNonce(r.Context(), challenge); err != nil {
		h.logger.Error("failed to create nonce", zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"nonce": challenge})
}

type VerifyWalletRequest struct {
	Nonce       string `json:"nonce"`
	Signature   string `json:"signature"`
	PublicKey   string `json:"publicKey"`
	DamlPartyId string `json:"damlPartyId"`
}

func (h *Handler) VerifyWallet(w http.ResponseWriter, r *http.Request) {
	var req VerifyWalletRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// 1. Authoritatively verify and consume the nonce to prevent replay attacks
	valid, err := h.configService.VerifyAndConsumeNonce(r.Context(), req.Nonce)
	if err != nil || !valid {
		h.logger.Warn("invalid or expired nonce verification attempt", zap.Error(err))
		http.Error(w, "invalid or expired nonce", http.StatusUnauthorized)
		return
	}

	// 2. Decode signature and public key from hex
	sig, err := hex.DecodeString(req.Signature)
	if err != nil {
		http.Error(w, "invalid signature encoding", http.StatusBadRequest)
		return
	}

	pubKey, err := hex.DecodeString(req.PublicKey)
	if err != nil {
		// Fallback: If not hex, try using raw bytes directly (typically for PEM-based keys)
		pubKey = []byte(req.PublicKey)
	}

	// 3. Cryptographically verify the signature
	verified, err := crypto.VerifySignature(pubKey, []byte(req.Nonce), sig)
	if err != nil || !verified {
		h.logger.Warn("cryptographic signature verification failed", zap.Error(err))
		http.Error(w, "invalid signature", http.StatusUnauthorized)
		return
	}

	// 4. Resolve the user's institutional identity
	// Sync the wallet party as a managed identity in Postgres
	oktaSub := "wallet:" + req.DamlPartyId
	// Extract simple name prefix from party name (e.g., Depositor::1220... -> Depositor)
	roleName := "Depositor"
	parts := strings.Split(req.DamlPartyId, "::")
	if len(parts) > 0 {
		roleName = parts[0]
	}
	email := strings.ToLower(roleName) + "@wallet.devlocal"

	identity, err := h.identityService.GetOrCreateIdentity(r.Context(), oktaSub, email, h.escrowService.GetLedgerClient())
	if err != nil {
		h.logger.Error("failed to resolve identity for wallet", zap.Error(err))
		http.Error(w, "identity synchronization failed", http.StatusInternalServerError)
		return
	}

	// 5. Generate secure Platform JWT Session Token
	jwtSecret := []byte("platform-jwt-signing-secret-key-32-bytes!") // Default fallback
	
	claims := jwt.MapClaims{
		"sub":           oktaSub,
		"email":         email,
		"scp":           []string{ScopeEscrowRead, ScopeEscrowWrite, ScopeEscrowAccept},
		"origin_domain": "wallet.devlocal",
		"auth_method":   "wallet",
		"iss":           "daml-escrow-platform",
		"aud":           "daml-escrow",
		"exp":           time.Now().Add(24 * time.Hour).Unix(),
		"iat":           time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(jwtSecret)
	if err != nil {
		h.logger.Error("failed to sign platform token", zap.Error(err))
		http.Error(w, "token generation failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"token":    tokenStr,
		"identity": identity,
	})
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

// --- Phase 11: Draft & Negotiation Handlers ---

func (h *Handler) SaveDraft(w http.ResponseWriter, r *http.Request) {
	var body struct {
		BeneficiaryEmail string          `json:"beneficiaryEmail"`
		ContractType     string          `json:"contractType"`
		Amount           float64         `json:"amount"`
		Currency         string          `json:"currency"`
		Terms            json.RawMessage `json:"terms"`
		Metadata         json.RawMessage `json:"metadata"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	if body.ContractType == "" {
		body.ContractType = "Corporate"
	}

	userID, _ := r.Context().Value(AuthSubKey).(string)
	// High-Assurance: Resolve initiator role from context/identity
	email, _ := r.Context().Value(EmailKey).(string)
	identity, _ := h.identityService.GetOrCreateIdentity(r.Context(), userID, email, h.escrowService.GetLedgerClient())
	
	role := "Depositor"
	if identity != nil && identity.Role != "" {
		role = identity.Role
	}

	draft, err := h.configService.CreateDraft(userID, role, body.ContractType, body.BeneficiaryEmail, body.Amount, body.Currency, body.Terms, body.Metadata)
	if err != nil {
		h.logger.Error("failed to create draft", zap.Error(err))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Internal Error: " + err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(draft)
}

func (h *Handler) ListDrafts(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(AuthSubKey).(string)
	email, _ := r.Context().Value(EmailKey).(string)

	drafts, err := h.configService.ListDraftsForUser(userID, email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(drafts)
}

func (h *Handler) GetDraft(w http.ResponseWriter, r *http.Request) {
	draftID := chi.URLParam(r, "draftID")
	draft, err := h.configService.GetLatestDraft(draftID) // Assuming draftID here is root_id for simplicity or we fetch by ID
	if err != nil {
		http.Error(w, "draft not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(draft)
}

func (h *Handler) AmendDraft(w http.ResponseWriter, r *http.Request) {
	rootID := chi.URLParam(r, "draftID")
	var body AmendDraftRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	if err := body.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userID, _ := r.Context().Value(AuthSubKey).(string)
	draft, err := h.configService.ProposeAmendment(rootID, userID, body.Summary, body.Amount, body.Currency, body.Terms, body.Metadata)
	if err != nil {
		h.logger.Error("failed to amend draft", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(draft)
}

func (h *Handler) ApproveDraft(w http.ResponseWriter, r *http.Request) {
	rootID := chi.URLParam(r, "draftID")
	userID, _ := r.Context().Value(AuthSubKey).(string)

	if err := h.configService.AddApproval(rootID, userID); err != nil {
		h.logger.Error("failed to approve draft", zap.Error(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Check if both parties have approved to auto-ratify
	draft, err := h.configService.GetLatestDraft(rootID)
	if err == nil && len(draft.Approvals) >= 2 {
		_ = h.configService.UpdateDraftStatus(draft.ID, "RATIFIED")
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) PromoteToLedger(w http.ResponseWriter, r *http.Request) {
	draftID := chi.URLParam(r, "draftID")
	draft, err := h.configService.GetLatestDraft(draftID)
	if err != nil {
		http.Error(w, "draft not found", http.StatusNotFound)
		return
	}

	// High-Assurance: Only promoter or initiator can trigger promotion
	userID, _ := r.Context().Value(AuthSubKey).(string)

	ledgerID, err := h.escrowService.PromoteDraft(r.Context(), draft, userID)
	if err != nil {
		h.logger.Error("failed to promote draft", zap.Error(err))
		http.Error(w, "Failed to promote draft: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if err := h.configService.UpdateDraftStatus(draft.ID, "PROMOTED"); err != nil {
		h.logger.Error("failed to update draft status", zap.Error(err))
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status":   "PROMOTED",
		"ledgerId": ledgerID,
	})
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
	// High-Assurance: Resolve initiator role from context/identity
	email, _ := r.Context().Value(EmailKey).(string)
	identity, _ := h.identityService.GetOrCreateIdentity(r.Context(), userID, email, h.escrowService.GetLedgerClient())
	
	role := "Depositor"
	if identity != nil && identity.Role != "" {
		role = identity.Role
	}
	
	// High-Assurance: Authoritatively use the off-chain draft tunnel for zero-latency invitations
	terms, _ := json.Marshal(map[string]interface{}{
		"conditionDescription": req.ConditionDescription,
		"expiryDate":           req.ExpiryDate,
	})
	
	draft, err := h.configService.CreateDraft(userID, role, req.ContractType, req.InviteeEmail, req.Amount, req.Currency, terms, nil)
	if err != nil {
		h.logger.Error("failed to create institutional invitation draft", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(draft)
}

func (h *Handler) ClaimInvitation(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	userID, _ := r.Context().Value(AuthSubKey).(string)

	// High-Assurance: authoritatively bridge the off-chain invitation to the draft record.
	if err := h.configService.ClaimDraft(token, userID); err != nil {
		h.logger.Error("failed to claim institutional draft", zap.Error(err), zap.String("token", token))
		http.Error(w, "invalid or expired invitation code", http.StatusForbidden)
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
	sanitizedBeneficiary, err := ledger.SanitizeID(req.Beneficiary)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Beneficiary error: " + err.Error()})
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
	if identity.Role == "Beneficiary" {
		ledgerReq.Beneficiary = identity.DamlUserID
		ledgerReq.Depositor = sanitizedBeneficiary
	} else {
		ledgerReq.Depositor = identity.DamlUserID
		ledgerReq.Beneficiary = sanitizedBeneficiary
	}
	ledgerReq.Mediator = sanitizedMediator

	// High-Assurance: Authoritatively enforce role exclusivity and mandatory parties
	if err := ledger.ValidateTripartiteRoles(ledgerReq.Depositor, ledgerReq.Beneficiary, ledgerReq.Mediator); err != nil {
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
	id, err := h.escrowService.ProposeSettlement(r.Context(), escrowID, userID, req.DepositorReturn)
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

func (h *Handler) IngestContract(w http.ResponseWriter, r *http.Request) {
	if h.ingestService == nil {
		http.Error(w, "ingest service unavailable", http.StatusServiceUnavailable)
		return
	}

	err := r.ParseMultipartForm(10 << 20) // 10MB limit
	if err != nil {
		http.Error(w, "failed to parse multipart form", http.StatusBadRequest)
		return
	}

	files := r.MultipartForm.File["agreement"]
	if len(files) == 0 {
		http.Error(w, "missing 'agreement' files", http.StatusBadRequest)
		return
	}

	var allFileData [][]byte
	var mimeType string

	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			continue
		}
		defer func() { _ = file.Close() }()

		data, err := io.ReadAll(file)
		if err == nil {
			allFileData = append(allFileData, data)
		}
		if mimeType == "" {
			mimeType = fileHeader.Header.Get("Content-Type")
		}
	}

	if mimeType == "" {
		mimeType = "application/pdf" // Fallback
	}

	// 3. Orchestrate AI Ingest
	result, err := h.ingestService.IngestContract(r.Context(), allFileData, mimeType)

	if err != nil {
		h.logger.Error("contract ingest failed", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// High-Assurance: Enrich metadata with storage provenance
	// This will be carried through to the draft save
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(result)
}

