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

type DamlCommand struct {
	CommandType string                 `json:"commandType"` // "create" or "exercise"
	TemplateID  string                 `json:"templateId"`
	ContractID  string                 `json:"contractId,omitempty"`
	Choice      string                 `json:"choice,omitempty"`
	Argument    map[string]interface{} `json:"argument"`
}

type DryRunResponse struct {
	IsDryRun bool          `json:"isDryRun"`
	Commands []DamlCommand `json:"commands"`
}

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

func (h *Handler) isDryRun(r *http.Request) bool {
	authMethod, _ := r.Context().Value(AuthMethodKey).(string)
	return authMethod == "wallet"
}

func (h *Handler) writeDryRun(w http.ResponseWriter, commands []DamlCommand) {
	payload := DryRunResponse{
		IsDryRun: true,
		Commands: commands,
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(payload)
}


// GetHealth godoc
// @Summary      Get system health and uptime
// @Description  Check uptime, CPU/Memory stats and status of internal services
// @Tags         system
// @Produce      json
// @Success      200  {object}  ledger.HealthResponse
// @Router       /health [get]
func (h *Handler) GetHealth(w http.ResponseWriter, r *http.Request) {
	h.logger.Debug("health check requested")
	health := h.metricsService.GetHealth(h.configService, h.escrowService.GetLedgerClient(), h.escrowService.GetOracleSecret())
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(health)
}

func (h *Handler) getIdentity(r *http.Request) (*ledger.UserIdentity, error) {
	userID, ok := r.Context().Value(AuthSubKey).(string)
	if !ok {
		return nil, fmt.Errorf("unauthorized")
	}
	email, _ := r.Context().Value(EmailKey).(string)
	identity, err := h.identityService.GetOrCreateIdentity(r.Context(), userID, email, h.escrowService.GetLedgerClient())
	if err != nil {
		return nil, err
	}
	assumedRole, _ := r.Context().Value(AssumedRoleKey).(string)
	if assumedRole != "" {
		identity.Role = assumedRole
	}
	return identity, nil
}

// GetIdentity godoc
// @Summary      Get current user identity
// @Description  Retrieve the authenticated user's profile and resolved Canton Party ID
// @Tags         auth
// @Produce      json
// @Success      200  {object}  ledger.UserIdentity
// @Failure      401  {object}  map[string]string
// @Router       /auth/me [get]
func (h *Handler) GetIdentity(w http.ResponseWriter, r *http.Request) {
	identity, err := h.getIdentity(r)
	if err != nil {
		h.logger.Error("failed to resolve identity", zap.Error(err))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(identity)
}

// DiscoverAuth godoc
// @Summary      Discover auth settings for an email
// @Description  Return the configured Identity Provider (IdP) or bypass settings for a given email address
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      DiscoverAuthRequest  true  "Discover Auth Request"
// @Success      200      {object}  map[string]interface{}
// @Failure      400      {string}  string  "invalid request"
// @Failure      404      {string}  string  "provider not found"
// @Router       /auth/discover [post]
func (h *Handler) DiscoverAuth(w http.ResponseWriter, r *http.Request) {
	var body DiscoverAuthRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := body.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	config, err := h.identityService.GetIdPConfigForEmail(body.Email)
	if err != nil {
		http.Error(w, "provider not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(config)
}

// GetNonce godoc
// @Summary      Get cryptographic challenge nonce
// @Description  Generate and persist a single-use UUID challenge string for wallet authentication handshakes (replays prevented)
// @Tags         auth
// @Produce      json
// @Success      200  {object}  map[string]string
// @Failure      500  {string}  string  "internal server error"
// @Router       /auth/nonce [get]
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
	AssumedRole string `json:"assumedRole"`
}

// VerifyWallet godoc
// @Summary      Verify wallet signature and obtain token
// @Description  Authoritatively validates the cryptographic signature of the nonce and returns a secure JWT platform session token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      VerifyWalletRequest  true  "Verify Wallet Request"
// @Success      200      {object}  map[string]interface{}
// @Failure      400      {string}  string  "invalid request body"
// @Failure      401      {string}  string  "invalid signature or nonce"
// @Failure      500      {string}  string  "token generation failed"
// @Router       /auth/wallet/verify [post]
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

	// Override role if explicitly assumed during wallet verify
	if req.AssumedRole != "" {
		identity.Role = req.AssumedRole
	}

	// 5. Generate secure Platform JWT Session Token
	jwtSecret := []byte("platform-jwt-signing-secret-key-32-bytes!") // Default fallback
	
	claims := jwt.MapClaims{
		"sub":           oktaSub,
		"email":         email,
		"scp":           []string{ScopeEscrowRead, ScopeEscrowWrite, ScopeEscrowAccept},
		"role":          identity.Role,
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

// ListIdentities godoc
// @Summary      List enriched participant identities
// @Description  Get a directory of all Canton identities enriched with displays names, affiliation, titles, and emails from Postgres database
// @Tags         identities
// @Produce      json
// @Success      200  {array}   ledger.UserIdentity
// @Failure      500  {string}  string  "internal server error"
// @Router       /identities [get]
func (h *Handler) ListIdentities(w http.ResponseWriter, r *http.Request) {
	identities, err := h.identityService.ListEnrichedIdentities(r.Context(), h.escrowService.GetLedgerClient())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(identities)
}

// GetConfig godoc
// @Summary      Get application config settings
// @Description  Retrieve all persistent database configurations for the authenticated user context
// @Tags         config
// @Produce      json
// @Success      200  {object}  map[string]string
// @Failure      500  {string}  string  "internal error"
// @Router       /config [get]
func (h *Handler) GetConfig(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(AuthSubKey).(string)
	cfgs, err := h.configService.ListConfigs(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(cfgs)
}

// SaveConfig godoc
// @Summary      Save config settings
// @Description  Set or update a persistent database configuration key-value pair for the current authenticated user context
// @Tags         config
// @Accept       json
// @Param        request  body  SaveConfigRequest  true  "Save Config Request"
// @Success      204      "No Content"
// @Failure      400      {string}  string  "invalid request"
// @Failure      500      {string}  string  "internal error"
// @Router       /config [post]
func (h *Handler) SaveConfig(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(AuthSubKey).(string)
	var body SaveConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if err := body.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := h.configService.SaveConfig(userID, body.Key, body.Value); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- Phase 11: Draft & Negotiation Handlers ---

// SaveDraft godoc
// @Summary      Save new draft agreement
// @Description  Create a new persistent contract proposal draft in the Postgres database, initialized as DRAFT state
// @Tags         drafts
// @Accept       json
// @Produce      json
// @Param        request  body      SaveDraftRequest  true  "Save Draft Request"
// @Success      200      {object}  services.DraftEscrow
// @Failure      400      {string}  string  "invalid request"
// @Failure      401      {string}  string  "unauthorized"
// @Failure      500      {string}  string  "internal error"
// @Router       /drafts [post]
func (h *Handler) SaveDraft(w http.ResponseWriter, r *http.Request) {
	var body SaveDraftRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	if err := body.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if body.ContractType == "" {
		body.ContractType = "Corporate"
	}

	identity, err := h.getIdentity(r)
	if err != nil {
		h.logger.Error("failed to resolve identity for draft", zap.Error(err))
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	
	role := "Depositor"
	if identity != nil && identity.Role != "" {
		role = identity.Role
	}

	draft, err := h.configService.CreateDraft(identity.DamlPartyID, role, body.ContractType, body.BeneficiaryEmail, body.Amount, body.Currency, body.Terms, body.Metadata)
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

// ListDrafts godoc
// @Summary      List drafts for user
// @Description  Get all active or completed draft versions where the user is an initiator or invitee counterparty
// @Tags         drafts
// @Produce      json
// @Success      200  {array}   services.DraftEscrow
// @Failure      401  {string}  string  "unauthorized"
// @Failure      500  {string}  string  "internal error"
// @Router       /drafts [get]
func (h *Handler) ListDrafts(w http.ResponseWriter, r *http.Request) {
	identity, err := h.getIdentity(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	drafts, err := h.configService.ListDraftsForUser(identity.DamlPartyID, identity.Email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(drafts)
}

// GetDraft godoc
// @Summary      Get draft by ID
// @Description  Retrieve the latest version details of a specific contract draft by its ID
// @Tags         drafts
// @Produce      json
// @Param        draftID  path      string  true  "Draft ID"
// @Success      200      {object}  services.DraftEscrow
// @Failure      404      {string}  string  "draft not found"
// @Router       /drafts/{draftID} [get]
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

// AmendDraft godoc
// @Summary      Propose draft amendment
// @Description  Propose a modified version of a contract draft, creating a new history item in Postgres and bumping the version
// @Tags         drafts
// @Accept       json
// @Produce      json
// @Param        draftID  path      string             true  "Draft ID"
// @Param        request  body      AmendDraftRequest  true  "Amend Draft Request"
// @Success      200      {object}  services.DraftEscrow
// @Failure      400      {string}  string  "invalid request"
// @Failure      401      {string}  string  "unauthorized"
// @Failure      500      {string}  string  "internal error"
// @Router       /drafts/{draftID}/amend [post]
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

	identity, err := h.getIdentity(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	proposerName := identity.DisplayName
	if proposerName == "" {
		proposerName = identity.DamlUserID
	}

	draft, err := h.configService.ProposeAmendment(rootID, proposerName, body.Summary, body.Amount, body.Currency, body.Terms, body.Metadata)
	if err != nil {
		h.logger.Error("failed to amend draft", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(draft)
}

// ApproveDraft godoc
// @Summary      Approve contract draft
// @Description  Approve the current latest version of a contract draft. If both counterparty roles approve, state shifts to RATIFIED.
// @Tags         drafts
// @Param        draftID  path      string  true  "Draft ID"
// @Success      204      "No Content"
// @Failure      400      {string}  string  "failed to approve"
// @Failure      401      {string}  string  "unauthorized"
// @Router       /drafts/{draftID}/approve [post]
func (h *Handler) ApproveDraft(w http.ResponseWriter, r *http.Request) {
	rootID := chi.URLParam(r, "draftID")
	identity, err := h.getIdentity(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	approverName := identity.DisplayName
	if approverName == "" {
		approverName = identity.DamlUserID
	}

	if err := h.configService.AddApproval(rootID, approverName); err != nil {
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

// PromoteToLedger godoc
// @Summary      Promote draft to active ledger escrow
// @Description  Bilateral promote a RATIFIED contract draft onto the Canton ledger as a proposed or active Escrow contract
// @Tags         drafts
// @Produce      json
// @Param        draftID  path      string  true  "Draft ID"
// @Success      200      {object}  map[string]string  "returns status and ledgerId"
// @Failure      401      {string}  string  "unauthorized"
// @Failure      404      {string}  string  "draft not found"
// @Failure      500      {string}  string  "promotion failed"
// @Router       /drafts/{draftID}/promote [post]
func (h *Handler) PromoteToLedger(w http.ResponseWriter, r *http.Request) {
	draftID := chi.URLParam(r, "draftID")
	draft, err := h.configService.GetLatestDraft(draftID)
	if err != nil {
		http.Error(w, "draft not found", http.StatusNotFound)
		return
	}

	// High-Assurance: Only promoter or initiator can trigger promotion
	identity, err := h.getIdentity(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	if draft.BeneficiaryID == "" && draft.BeneficiaryEmail != "" {
		resolvedPartyID, err := h.identityService.FindPartyIDByEmail(r.Context(), draft.BeneficiaryEmail)
		if err == nil && resolvedPartyID != "" {
			draft.BeneficiaryID = resolvedPartyID
			_ = h.configService.UpdateDraftBeneficiary(draft.ID, resolvedPartyID)
		}
	}

	ledgerID, err := h.escrowService.PromoteDraft(r.Context(), draft, identity.DamlPartyID)
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

// ListInvitations godoc
// @Summary      List ledger invitations
// @Description  Fetch a list of active invitations awaiting claimant countersignatures from the Canton ledger
// @Tags         invitations
// @Produce      json
// @Success      200  {array}   ledger.EscrowInvitation
// @Failure      401  {string}  string  "unauthorized"
// @Failure      500  {string}  string  "internal error"
// @Router       /invites [get]
func (h *Handler) ListInvitations(w http.ResponseWriter, r *http.Request) {
	identity, err := h.getIdentity(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	invites, err := h.escrowService.GetLedgerClient().ListInvitations(r.Context(), identity.DamlUserID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(invites)
}

// GetInvitationByToken godoc
// @Summary      Get invitation by token
// @Description  Resolve a specific Canton ledger invitation details by token hash
// @Tags         invitations
// @Produce      json
// @Param        token  path      string  true  "Invitation Token Hash"
// @Success      200    {object}  ledger.EscrowInvitation
// @Failure      404    {string}  string  "invitation not found"
// @Router       /invites/token/{token} [get]
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

// CreateInvitation godoc
// @Summary      Create off-chain invitation
// @Description  Propose an off-chain contract draft for a counterparty email invite, establishing a zero-latency initiation tunnel
// @Tags         invitations
// @Accept       json
// @Produce      json
// @Param        request  body      CreateInvitationRequest  true  "Create Invitation Request"
// @Success      200      {object}  services.DraftEscrow
// @Failure      400      {string}  string  "invalid request"
// @Failure      401      {string}  string  "unauthorized"
// @Failure      500      {string}  string  "internal error"
// @Router       /invites [post]
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
	identity, err := h.getIdentity(r)
	if err != nil {
		h.logger.Error("failed to resolve identity for invitation", zap.Error(err))
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	
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

// ClaimInvitation godoc
// @Summary      Claim invitation token
// @Description  Claim an invitation token hash, linking the off-chain draft with the claimant user and initiating verification
// @Tags         invitations
// @Param        token  path      string  true  "Invitation Token Hash"
// @Success      204    "No Content"
// @Failure      403    {string}  string  "invalid or expired token"
// @Router       /invites/token/{token}/claim [post]
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

// ProposeEscrow godoc
// @Summary      Propose a new escrow agreement
// @Description  Create a draft agreement on the Canton ledger awaiting counterparty acceptance
// @Tags         escrows
// @Accept       json
// @Produce      json
// @Param        request  body      ProposeEscrowRequest  true  "Propose Escrow Request"
// @Success      200      {object}  ledger.EscrowProposal
// @Failure      400      {string}  string  "invalid request"
// @Failure      401      {string}  string  "unauthorized"
// @Failure      500      {string}  string  "proposal failed"
// @Router       /escrows [post]
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

	identity, err := h.getIdentity(r)
	if err != nil {
		h.logger.Error("failed to resolve identity for proposal", zap.Error(err))
		http.Error(w, "unauthorized", http.StatusUnauthorized)
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

	if h.isDryRun(r) {
		h.writeDryRun(w, []DamlCommand{
			{
				CommandType: "create",
				TemplateID:  "StablecoinEscrow:EscrowProposal",
				Argument: map[string]interface{}{
					"depositor":   ledgerReq.Depositor,
					"beneficiary": ledgerReq.Beneficiary,
					"mediator":    ledgerReq.Mediator,
					"amount":      ledgerReq.Asset.Amount,
					"currency":    ledgerReq.Asset.Currency,
				},
			},
		})
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

// Fund godoc
// @Summary      Fund active escrow
// @Description  Approve and authorize transfer of stablecoins from depositor custody vault holding into the escrow contract
// @Tags         escrows
// @Accept       json
// @Param        escrowID  path      string             true  "Escrow ID"
// @Param        request   body      FundEscrowRequest  true  "Fund Escrow Request"
// @Success      204       "No Content"
// @Failure      400       {string}  string  "invalid request"
// @Failure      401       {string}  string  "unauthorized"
// @Failure      500       {string}  string  "funding failed"
// @Router       /escrows/{escrowID}/fund [post]
func (h *Handler) Fund(w http.ResponseWriter, r *http.Request) {
	escrowID := chi.URLParam(r, "escrowID")
	var req FundEscrowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if h.isDryRun(r) {
		h.writeDryRun(w, []DamlCommand{
			{
				CommandType: "exercise",
				TemplateID:  "StablecoinEscrow:Escrow",
				ContractID:  escrowID,
				Choice:      "fund",
				Argument: map[string]interface{}{
					"holdingCid": req.HoldingCid,
				},
			},
		})
		return
	}

	identity, err := h.getIdentity(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	err = h.escrowService.FundEscrow(r.Context(), escrowID, identity.DamlUserID, req.HoldingCid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Activate godoc
// @Summary      Activate funded escrow
// @Description  Acknowledge funding and activate the escrow contract for condition validation
// @Tags         escrows
// @Param        escrowID  path      string  true  "Escrow ID"
// @Success      200       {string}  string  "activated contract ID"
// @Failure      401       {string}  string  "unauthorized"
// @Failure      500       {string}  string  "activation failed"
// @Router       /escrows/{escrowID}/activate [post]
func (h *Handler) Activate(w http.ResponseWriter, r *http.Request) {
	escrowID := chi.URLParam(r, "escrowID")
	
	if h.isDryRun(r) {
		h.writeDryRun(w, []DamlCommand{
			{
				CommandType: "exercise",
				TemplateID:  "StablecoinEscrow:Escrow",
				ContractID:  escrowID,
				Choice:      "activate",
				Argument:    map[string]interface{}{},
			},
		})
		return
	}

	identity, err := h.getIdentity(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	id, err := h.escrowService.ActivateEscrow(r.Context(), escrowID, identity.DamlUserID, []string{identity.DamlUserID})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, _ = w.Write([]byte(id))
}

// ConfirmConditions godoc
// @Summary      Confirm escrow conditions
// @Description  Bilateral or mediator counter-signing confirming conditions of the active escrow are met
// @Tags         escrows
// @Param        escrowID  path      string  true  "Escrow ID"
// @Success      204       "No Content"
// @Failure      401       {string}  string  "unauthorized"
// @Failure      500       {string}  string  "confirmation failed"
// @Router       /escrows/{escrowID}/confirm [post]
func (h *Handler) ConfirmConditions(w http.ResponseWriter, r *http.Request) {
	escrowID := chi.URLParam(r, "escrowID")

	if h.isDryRun(r) {
		h.writeDryRun(w, []DamlCommand{
			{
				CommandType: "exercise",
				TemplateID:  "StablecoinEscrow:Escrow",
				ContractID:  escrowID,
				Choice:      "confirmConditions",
				Argument:    map[string]interface{}{},
			},
		})
		return
	}

	identity, err := h.getIdentity(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	err = h.escrowService.GetLedgerClient().ConfirmConditions(r.Context(), escrowID, identity.DamlUserID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// RaiseDispute godoc
// @Summary      Raise escrow dispute
// @Description  Initiates dispute state on the active escrow, halting normal milestone disbursement and routing to the mediator
// @Tags         escrows
// @Param        escrowID  path      string  true  "Escrow ID"
// @Success      204       "No Content"
// @Failure      401       {string}  string  "unauthorized"
// @Failure      500       {string}  string  "dispute failed"
// @Router       /escrows/{escrowID}/dispute [post]
func (h *Handler) RaiseDispute(w http.ResponseWriter, r *http.Request) {
	escrowID := chi.URLParam(r, "escrowID")

	if h.isDryRun(r) {
		h.writeDryRun(w, []DamlCommand{
			{
				CommandType: "exercise",
				TemplateID:  "StablecoinEscrow:Escrow",
				ContractID:  escrowID,
				Choice:      "raiseDispute",
				Argument:    map[string]interface{}{},
			},
		})
		return
	}

	identity, err := h.getIdentity(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	err = h.escrowService.RaiseDispute(r.Context(), escrowID, identity.DamlUserID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ProposeSettlement godoc
// @Summary      Propose dispute settlement split
// @Description  Create a settlement payout distribution proposal (buyer refund amount vs seller disburse amount)
// @Tags         escrows
// @Accept       json
// @Param        escrowID  path      string                    true  "Escrow ID"
// @Param        request   body      ProposeSettlementRequest  true  "Propose Settlement Request"
// @Success      200       {string}  string                    "proposed settlement ID"
// @Failure      400       {string}  string                    "invalid request"
// @Failure      401       {string}  string                    "unauthorized"
// @Failure      500       {string}  string                    "proposal failed"
// @Router       /escrows/{escrowID}/propose-settlement [post]
func (h *Handler) ProposeSettlement(w http.ResponseWriter, r *http.Request) {
	escrowID := chi.URLParam(r, "escrowID")
	var req ProposeSettlementRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if h.isDryRun(r) {
		h.writeDryRun(w, []DamlCommand{
			{
				CommandType: "exercise",
				TemplateID:  "StablecoinEscrow:Escrow",
				ContractID:  escrowID,
				Choice:      "proposeSettlement",
				Argument: map[string]interface{}{
					"depositorReturn": req.DepositorReturn,
				},
			},
		})
		return
	}

	identity, err := h.getIdentity(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	id, err := h.escrowService.ProposeSettlement(r.Context(), escrowID, identity.DamlUserID, req.DepositorReturn)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, _ = w.Write([]byte(id))
}

// RatifySettlement godoc
// @Summary      Ratify dispute settlement split
// @Description  Bilateral ratify the proposed dispute payout split
// @Tags         escrows
// @Param        escrowID  path      string  true  "Escrow ID"
// @Success      200       {string}  string  "ratified contract ID"
// @Failure      401       {string}  string  "unauthorized"
// @Failure      500       {string}  string  "ratification failed"
// @Router       /escrows/{escrowID}/ratify [post]
func (h *Handler) RatifySettlement(w http.ResponseWriter, r *http.Request) {
	escrowID := chi.URLParam(r, "escrowID")

	if h.isDryRun(r) {
		h.writeDryRun(w, []DamlCommand{
			{
				CommandType: "exercise",
				TemplateID:  "StablecoinEscrow:Escrow",
				ContractID:  escrowID,
				Choice:      "ratifySettlement",
				Argument:    map[string]interface{}{},
			},
		})
		return
	}

	identity, err := h.getIdentity(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	id, err := h.escrowService.GetLedgerClient().RatifySettlement(r.Context(), escrowID, identity.DamlUserID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, _ = w.Write([]byte(id))
}

// FinalizeSettlement godoc
// @Summary      Finalize dispute settlement
// @Description  Finalize a ratified dispute settlement split on the ledger
// @Tags         escrows
// @Param        escrowID  path      string  true  "Escrow ID"
// @Success      200       {string}  string  "finalized contract ID"
// @Failure      401       {string}  string  "unauthorized"
// @Failure      500       {string}  string  "finalization failed"
// @Router       /escrows/{escrowID}/finalize [post]
func (h *Handler) FinalizeSettlement(w http.ResponseWriter, r *http.Request) {
	escrowID := chi.URLParam(r, "escrowID")

	if h.isDryRun(r) {
		h.writeDryRun(w, []DamlCommand{
			{
				CommandType: "exercise",
				TemplateID:  "StablecoinEscrow:Escrow",
				ContractID:  escrowID,
				Choice:      "finalizeSettlement",
				Argument:    map[string]interface{}{},
			},
		})
		return
	}

	identity, err := h.getIdentity(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	id, err := h.escrowService.GetLedgerClient().FinalizeSettlement(r.Context(), escrowID, identity.DamlUserID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, _ = w.Write([]byte(id))
}

// Disburse godoc
// @Summary      Disburse active milestone funds
// @Description  Release funds for the active milestone to the beneficiary upon successful condition confirmation
// @Tags         escrows
// @Param        escrowID  path      string  true  "Escrow ID"
// @Success      204       "No Content"
// @Failure      401       {string}  string  "unauthorized"
// @Failure      500       {string}  string  "disbursement failed"
// @Router       /escrows/{escrowID}/disburse [post]
func (h *Handler) Disburse(w http.ResponseWriter, r *http.Request) {
	escrowID := chi.URLParam(r, "escrowID")

	if h.isDryRun(r) {
		h.writeDryRun(w, []DamlCommand{
			{
				CommandType: "exercise",
				TemplateID:  "StablecoinEscrow:Escrow",
				ContractID:  escrowID,
				Choice:      "disburse",
				Argument:    map[string]interface{}{},
			},
		})
		return
	}

	identity, err := h.getIdentity(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	err = h.escrowService.DisburseEscrow(r.Context(), escrowID, identity.DamlUserID, []string{identity.DamlUserID})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ListEscrows godoc
// @Summary      List active escrows
// @Description  Get a list of active Escrow contracts from the Canton ledger associated with the user's role/party
// @Tags         escrows
// @Produce      json
// @Success      200  {array}   ledger.EscrowContract
// @Failure      401  {string}  string  "unauthorized"
// @Failure      500  {string}  string  "internal error"
// @Router       /escrows [get]
func (h *Handler) ListEscrows(w http.ResponseWriter, r *http.Request) {
	identity, err := h.getIdentity(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	party := identity.DamlPartyID
	h.logger.Info("listing escrows for user", 
		zap.String("userID", identity.DamlUserID), 
		zap.String("resolvedParty", party))

	escrows, err := h.escrowService.ListEscrows(r.Context(), identity.DamlPartyID)
	if err != nil {
		h.logger.Error("failed to list escrows", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.Info("retrieved escrows from ledger", zap.Int("count", len(escrows)))
	for _, e := range escrows {
		h.logger.Info("escrow contract item", 
			zap.String("id", e.ID), 
			zap.String("depositor", e.Depositor), 
			zap.String("beneficiary", e.Beneficiary), 
			zap.String("state", e.State))
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(escrows)
}

// GetEscrow godoc
// @Summary      Get escrow details
// @Description  Retrieve the full contract state details of a specific escrow by its ledger ID
// @Tags         escrows
// @Produce      json
// @Param        id   path      string  true  "Escrow ID"
// @Success      200  {object}  ledger.EscrowContract
// @Failure      401  {string}  string  "unauthorized"
// @Failure      404  {string}  string  "escrow not found"
// @Router       /escrows/{id} [get]
func (h *Handler) GetEscrow(w http.ResponseWriter, r *http.Request) {
	escrowID := chi.URLParam(r, "escrowID")
	identity, err := h.getIdentity(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	escrow, err := h.escrowService.GetEscrow(r.Context(), escrowID, identity.DamlUserID)
	if err != nil {
		http.Error(w, "escrow not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(escrow)
}

// GetEscrowLifecycle godoc
// @Summary      Get escrow event history
// @Description  Retrieve structured audit log events (creation, funding, activate, confirm, etc.) for a specific escrow contract
// @Tags         escrows
// @Produce      json
// @Param        id   path      string  true  "Escrow ID"
// @Success      200  {object}  services.EscrowLifecycleMetadata
// @Failure      401  {string}  string  "unauthorized"
// @Failure      500  {string}  string  "internal error"
// @Router       /escrows/{id}/lifecycle [get]
func (h *Handler) GetEscrowLifecycle(w http.ResponseWriter, r *http.Request) {
	escrowID := chi.URLParam(r, "escrowID")
	identity, err := h.getIdentity(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	lifecycle, err := h.analyticsService.GetEscrowLifecycle(r.Context(), escrowID, identity.DamlUserID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(lifecycle)
}

// OracleMilestoneTrigger godoc
// @Summary      Oracle webhook callback
// @Description  External webhook receiver for oracle providers to sign off milestone target conditions
// @Tags         system
// @Accept       json
// @Param        request  body  OracleWebhookRequest  true  "Oracle Webhook Request"
// @Success      200      "OK"
// @Failure      400      {string}  string  "invalid request body"
// @Failure      403      {string}  string  "trigger rejected"
// @Router       /webhooks/milestone [post]
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

// GetMetrics godoc
// @Summary      Get system performance metrics
// @Description  Retrieve real-time metrics including total volume in escrow, active counts, API latency, and CPU/Memory usage
// @Tags         system
// @Produce      json
// @Success      200  {object}  ledger.LedgerMetrics
// @Failure      401  {string}  string  "unauthorized"
// @Failure      500  {string}  string  "internal error"
// @Router       /metrics [get]
func (h *Handler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	identity, err := h.getIdentity(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	metrics, err := h.escrowService.GetMetrics(r.Context(), identity.DamlUserID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(metrics)
}

// ListSettlements godoc
// @Summary      List pending settlements
// @Description  Retrieve all active settlement payment instructions awaiting Central Bank stablecoin release execution
// @Tags         settlements
// @Produce      json
// @Success      200  {array}   ledger.EscrowSettlement
// @Failure      500  {string}  string  "internal error"
// @Router       /settlements [get]
func (h *Handler) ListSettlements(w http.ResponseWriter, r *http.Request) {
	settlements, err := h.escrowService.GetLedgerClient().ListSettlements(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(settlements)
}

// SettlePayment godoc
// @Summary      Settle pending payment
// @Description  Central Bank settles the payout instruction, releasing stablecoins to the target recipient and marking it completed
// @Tags         settlements
// @Param        settlementID  path      string  true  "Settlement ID"
// @Success      204           "No Content"
// @Failure      500           {string}  string  "settlement failed"
// @Router       /settlements/{settlementID}/settle [post]
func (h *Handler) SettlePayment(w http.ResponseWriter, r *http.Request) {
	settlementID := chi.URLParam(r, "settlementID")
	err := h.escrowService.SettleEscrow(r.Context(), settlementID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ListWallets godoc
// @Summary      List simulated wallets
// @Description  Retrieve mock/simulated wallet balances for Central Bank, Depositor, Beneficiary, and Mediator in dev perimeters
// @Tags         wallets
// @Produce      json
// @Success      200  {array}   ledger.Wallet
// @Failure      401  {string}  string  "unauthorized"
// @Failure      500  {string}  string  "internal error"
// @Router       /wallets [get]
func (h *Handler) ListWallets(w http.ResponseWriter, r *http.Request) {
	identity, err := h.getIdentity(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	wallets, err := h.escrowService.GetLedgerClient().ListWallets(r.Context(), identity.DamlUserID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(wallets)
}

// IngestContract godoc
// @Summary      Ingest agreement documents via AI
// @Description  Accept uploaded PDF/images, perform AI extraction of escrow terms and amount, and return structured metadata
// @Tags         system
// @Accept       multipart/form-data
// @Produce      json
// @Param        agreement  formData  file  true  "Agreement contract PDF or image files"
// @Success      200        {object}  services.IngestResult
// @Failure      400        {string}  string  "missing or invalid files"
// @Failure      500        {string}  string  "ingestion failed"
// @Failure      503        {string}  string  "ingest service unavailable"
// @Router       /ingest [post]
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

