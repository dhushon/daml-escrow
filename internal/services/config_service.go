package services

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

type ConfigService struct {
	db *sql.DB
}

func NewConfigService(dsn string) (*ConfigService, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to user_config db: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping user_config db: %w", err)
	}

	return &ConfigService{db: db}, nil
}

func (s *ConfigService) Close() error {
	return s.db.Close()
}

func (s *ConfigService) GetDB() *sql.DB {
	return s.db
}

// --- User Configuration CRUD ---

func (s *ConfigService) SaveConfig(userID, key string, val json.RawMessage) error {
	query := `
		INSERT INTO configs (user_id, config_key, config_value, updated_at)
		VALUES ($1, $2, $3, CURRENT_TIMESTAMP)
		ON CONFLICT (user_id, config_key) DO UPDATE
		SET config_value = $3, updated_at = CURRENT_TIMESTAMP
	`
	_, err := s.db.Exec(query, userID, key, val)
	return err
}

func (s *ConfigService) GetConfig(userID, key string) (json.RawMessage, error) {
	var val json.RawMessage
	query := "SELECT config_value FROM configs WHERE user_id = $1 AND config_key = $2"
	err := s.db.QueryRow(query, userID, key).Scan(&val)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return val, nil
}

func (s *ConfigService) ListConfigs(userID string) (map[string]json.RawMessage, error) {
	query := "SELECT config_key, config_value FROM configs WHERE user_id = $1"
	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	configs := make(map[string]json.RawMessage)
	for rows.Next() {
		var key string
		var val json.RawMessage
		if err := rows.Scan(&key, &val); err != nil {
			return nil, err
		}
		configs[key] = val
	}
	return configs, nil
}

// --- Auth Nonce Management (Phase 17.1) ---

func (s *ConfigService) CreateNonce(ctx context.Context, nonce string) error {
	query := `
		INSERT INTO auth_nonces (nonce, created_at)
		VALUES ($1, CURRENT_TIMESTAMP)
	`
	_, err := s.db.ExecContext(ctx, query, nonce)
	return err
}

func (s *ConfigService) VerifyAndConsumeNonce(ctx context.Context, nonce string) (bool, error) {
	// High-Assurance: Single-use delete on retrieve with strict 5-minute expiration
	query := `
		DELETE FROM auth_nonces
		WHERE nonce = $1 AND created_at >= CURRENT_TIMESTAMP - INTERVAL '5 minutes'
		RETURNING nonce
	`
	var found string
	err := s.db.QueryRowContext(ctx, query, nonce).Scan(&found)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return found != "", nil
}


// --- Phase 11 & 12: Versioned Draft Escrow Management ---

type DraftEscrow struct {
	ID               string          `json:"id"`
	RootID           string          `json:"rootId"`
	Version          int             `json:"version"`
	ProposerID       string          `json:"proposerId"`
	InvitationCode   string          `json:"invitationCode,omitempty"`
	ContractType     string          `json:"contractType"` // ImportExport, RealEstate, Grants, Corporate
	InitiatorID      string          `json:"initiatorId"`
	InitiatorRole    string          `json:"initiatorRole"`
	DepositorID      string          `json:"depositorId"`
	BeneficiaryEmail string          `json:"beneficiaryEmail"`
	BeneficiaryID    string          `json:"beneficiaryId"`
	MediatorID       string          `json:"mediatorId"`
	Amount           float64         `json:"amount"`
	Currency         string          `json:"currency"`
	Terms            json.RawMessage `json:"terms" swaggertype:"object"`
	Metadata         json.RawMessage `json:"metadata" swaggertype:"object"`
	ChangeSummary    string          `json:"changeSummary"`
	Approvals        []string        `json:"approvals"`
	Status           string          `json:"status"` // DRAFT, CLAIMED, NEGOTIATION, RATIFIED, PROMOTED
	CreatedAt        time.Time       `json:"createdAt"`
}

func generateInvitationCode() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// CreateDraft initializes Version 1 of a negotiation.
func (s *ConfigService) CreateDraft(initiatorID, initiatorRole, contractType, beneficiaryEmail string, amount float64, currency string, terms, metadata json.RawMessage) (*DraftEscrow, error) {
	var draft DraftEscrow
	rootID := uuid.New().String()
	code := generateInvitationCode()

	depositorID := ""
	if initiatorRole == "Depositor" {
		depositorID = initiatorID
	}

	query := `
		INSERT INTO draft_escrows (root_id, proposer_id, invitation_code, contract_type, initiator_id, initiator_role, depositor_id, beneficiary_email, amount, currency, terms, metadata, status)
		VALUES ($1, $2, $3, $4, $2, $5, $6, $7, $8, $9, $10, $11, 'DRAFT')
		RETURNING id, root_id, version, proposer_id, invitation_code, contract_type, initiator_id, initiator_role, depositor_id, beneficiary_email, amount, currency, terms, metadata, change_summary, status, created_at
	`
	var changeSum, dID sql.NullString
	err := s.db.QueryRow(query, rootID, initiatorID, code, contractType, initiatorRole, depositorID, beneficiaryEmail, amount, currency, terms, metadata).Scan(
		&draft.ID, &draft.RootID, &draft.Version, &draft.ProposerID, &draft.InvitationCode, &draft.ContractType, &draft.InitiatorID, &draft.InitiatorRole, &dID, &draft.BeneficiaryEmail, &draft.Amount, &draft.Currency, &draft.Terms, &draft.Metadata, &changeSum, &draft.Status, &draft.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create draft escrow: %w", err)
	}
	draft.DepositorID = dID.String
	draft.ChangeSummary = changeSum.String
	draft.Approvals = []string{}
	return &draft, nil
}

// ProposeAmendment creates a new version linked to the same root_id.
func (s *ConfigService) ProposeAmendment(rootID, proposerID, summary string, amount float64, currency string, terms, metadata json.RawMessage) (*DraftEscrow, error) {
	// 1. Get current version info
	var currentVersion int
	var initiatorID, initiatorRole, beneficiaryEmail, contractType string
	
	lookup := `SELECT version, initiator_id, initiator_role, beneficiary_email, contract_type FROM draft_escrows WHERE root_id = $1 ORDER BY version DESC LIMIT 1`
	err := s.db.QueryRow(lookup, rootID).Scan(&currentVersion, &initiatorID, &initiatorRole, &beneficiaryEmail, &contractType)
	if err != nil {
		return nil, fmt.Errorf("failed to locate root negotiation: %w", err)
	}

	// 2. Insert new version
	var draft DraftEscrow
	query := `
		INSERT INTO draft_escrows (root_id, version, proposer_id, contract_type, initiator_id, initiator_role, beneficiary_email, amount, currency, terms, metadata, change_summary, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, 'NEGOTIATION')
		RETURNING id, root_id, version, proposer_id, contract_type, initiator_id, initiator_role, beneficiary_email, amount, currency, terms, metadata, change_summary, status, created_at
	`
	var cs sql.NullString
	err = s.db.QueryRow(query, rootID, currentVersion+1, proposerID, contractType, initiatorID, initiatorRole, beneficiaryEmail, amount, currency, terms, metadata, summary).Scan(
		&draft.ID, &draft.RootID, &draft.Version, &draft.ProposerID, &draft.ContractType, &draft.InitiatorID, &draft.InitiatorRole, &draft.BeneficiaryEmail, &draft.Amount, &draft.Currency, &draft.Terms, &draft.Metadata, &cs, &draft.Status, &draft.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to propose amendment: %w", err)
	}
	draft.ChangeSummary = cs.String
	draft.Approvals = []string{}
	return &draft, nil
}

// AddApproval appends an identity to the approvals array of the LATEST version.
func (s *ConfigService) AddApproval(rootID, approverID string) error {
	// Fetch latest version of the draft first
	draft, err := s.GetLatestDraft(rootID)
	if err != nil {
		return fmt.Errorf("negotiation draft not found: %w", err)
	}

	// Idempotency: If the user has already approved this version, succeed silently
	for _, a := range draft.Approvals {
		if a == approverID {
			return nil
		}
	}

	query := `
		UPDATE draft_escrows 
		SET approvals = approvals || jsonb_build_array($1::text)
		WHERE id = $2
	`
	_, err = s.db.Exec(query, approverID, draft.ID)
	if err != nil {
		return fmt.Errorf("failed to append approval: %w", err)
	}

	return nil
}

func (s *ConfigService) GetLatestDraft(rootID string) (*DraftEscrow, error) {
	var draft DraftEscrow
	query := "SELECT id, root_id, version, proposer_id, invitation_code, contract_type, initiator_id, initiator_role, depositor_id, beneficiary_email, beneficiary_id, mediator_id, amount, currency, terms, metadata, change_summary, approvals, status, created_at FROM draft_escrows WHERE root_id = $1 ORDER BY version DESC LIMIT 1"
	
	var medID, beneficiaryID, inviteCode, changeSum, dID sql.NullString
	var approvalsJSON []byte
	err := s.db.QueryRow(query, rootID).Scan(
		&draft.ID, &draft.RootID, &draft.Version, &draft.ProposerID, &inviteCode, &draft.ContractType, &draft.InitiatorID, &draft.InitiatorRole, &dID, &draft.BeneficiaryEmail, &beneficiaryID, &medID, &draft.Amount, &draft.Currency, &draft.Terms, &draft.Metadata, &changeSum, &approvalsJSON, &draft.Status, &draft.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	draft.InvitationCode = inviteCode.String
	draft.DepositorID = dID.String
	draft.BeneficiaryID = beneficiaryID.String
	draft.MediatorID = medID.String
	draft.ChangeSummary = changeSum.String
	_ = json.Unmarshal(approvalsJSON, &draft.Approvals)
	return &draft, nil
}

func (s *ConfigService) ClaimDraft(code, beneficiaryID string) error {
	query := `
		UPDATE draft_escrows 
		SET beneficiary_id = $1, status = 'CLAIMED', updated_at = CURRENT_TIMESTAMP 
		WHERE invitation_code = $2 AND status = 'DRAFT'
	`
	// Since invitation_code is only on version 1, this works
	res, err := s.db.Exec(query, beneficiaryID, code)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("invalid or already claimed invitation code")
	}
	return nil
}

func (s *ConfigService) ListDraftsForUser(userID, email string) ([]*DraftEscrow, error) {
	// Return latest version for each root_id where user is involved
	query := `
		SELECT id, root_id, version, proposer_id, contract_type, initiator_id, initiator_role, depositor_id, beneficiary_email, beneficiary_id, mediator_id, amount, currency, terms, metadata, status, created_at
		FROM (
			SELECT DISTINCT ON (root_id) 
				id, root_id, version, proposer_id, contract_type, initiator_id, initiator_role, depositor_id, beneficiary_email, beneficiary_id, mediator_id, amount, currency, terms, metadata, status, created_at
			FROM draft_escrows 
			WHERE initiator_id = $1 OR beneficiary_email = $2 OR beneficiary_id = $1 OR mediator_id = $1 OR depositor_id = $1
			ORDER BY root_id, version DESC
		) t
		ORDER BY created_at DESC
	`
	rows, err := s.db.Query(query, userID, email)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var drafts []*DraftEscrow
	for rows.Next() {
		var d DraftEscrow
		var medID, beneficiaryID, dID sql.NullString
		if err := rows.Scan(&d.ID, &d.RootID, &d.Version, &d.ProposerID, &d.ContractType, &d.InitiatorID, &d.InitiatorRole, &dID, &d.BeneficiaryEmail, &beneficiaryID, &medID, &d.Amount, &d.Currency, &d.Terms, &d.Metadata, &d.Status, &d.CreatedAt); err != nil {
			return nil, err
		}
		d.DepositorID = dID.String
		d.BeneficiaryID = beneficiaryID.String
		d.MediatorID = medID.String
		drafts = append(drafts, &d)
	}
	return drafts, nil
}

func (s *ConfigService) UpdateDraftStatus(id, status string) error {
	query := "UPDATE draft_escrows SET status = $1 WHERE id = $2"
	_, err := s.db.Exec(query, status, id)
	return err
}

func (s *ConfigService) UpdateDraftBeneficiary(id, beneficiaryID string) error {
	query := "UPDATE draft_escrows SET beneficiary_id = $1 WHERE id = $2"
	_, err := s.db.Exec(query, beneficiaryID, id)
	return err
}


func NewMockConfigService(db *sql.DB) *ConfigService {
	return &ConfigService{db: db}
}
