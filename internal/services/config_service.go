package services

import (
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
		return nil, err
	}
	return val, nil
}

// --- Phase 11 & 12: Versioned Draft Escrow Management ---

type DraftEscrow struct {
	ID                string          `json:"id"`
	RootID            string          `json:"rootId"`
	Version           int             `json:"version"`
	ProposerID        string          `json:"proposerId"`
	InvitationCode    string          `json:"invitationCode,omitempty"`
	InitiatorID       string          `json:"initiatorId"`
	CounterpartyEmail string          `json:"counterpartyEmail"`
	CounterpartyID    string          `json:"counterpartyId"`
	MediatorID        string          `json:"mediatorId"`
	Amount            float64         `json:"amount"`
	Currency          string          `json:"currency"`
	Terms             json.RawMessage `json:"terms"`
	Metadata          json.RawMessage `json:"metadata"`
	ChangeSummary     string          `json:"changeSummary"`
	Approvals         []string        `json:"approvals"`
	Status            string          `json:"status"` // DRAFT, CLAIMED, NEGOTIATION, RATIFIED, PROMOTED
	CreatedAt         time.Time       `json:"createdAt"`
}

func generateInvitationCode() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// CreateDraft initializes Version 1 of a negotiation.
func (s *ConfigService) CreateDraft(initiatorID, counterpartyEmail string, amount float64, currency string, terms, metadata json.RawMessage) (*DraftEscrow, error) {
	var draft DraftEscrow
	rootID := uuid.New().String()
	code := generateInvitationCode()
	
	query := `
		INSERT INTO draft_escrows (root_id, proposer_id, invitation_code, initiator_id, counterparty_email, amount, currency, terms, metadata, status)
		VALUES ($1, $2, $3, $2, $4, $5, $6, $7, $8, 'DRAFT')
		RETURNING id, root_id, version, proposer_id, invitation_code, initiator_id, counterparty_email, amount, currency, terms, metadata, change_summary, status, created_at
	`
	var changeSum sql.NullString
	err := s.db.QueryRow(query, rootID, initiatorID, code, counterpartyEmail, amount, currency, terms, metadata).Scan(
		&draft.ID, &draft.RootID, &draft.Version, &draft.ProposerID, &draft.InvitationCode, &draft.InitiatorID, &draft.CounterpartyEmail, &draft.Amount, &draft.Currency, &draft.Terms, &draft.Metadata, &changeSum, &draft.Status, &draft.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create draft escrow: %w", err)
	}
	draft.ChangeSummary = changeSum.String
	draft.Approvals = []string{}
	return &draft, nil
}

// ProposeAmendment creates a new version linked to the same root_id.
func (s *ConfigService) ProposeAmendment(rootID, proposerID, summary string, amount float64, currency string, terms, metadata json.RawMessage) (*DraftEscrow, error) {
	// 1. Get current version number
	var currentVersion int
	var initiatorID, counterpartyEmail, counterpartyID, mediatorID string
	
	lookup := `SELECT version, initiator_id, counterparty_email, counterparty_id, mediator_id FROM draft_escrows WHERE root_id = $1 ORDER BY version DESC LIMIT 1`
	var cID, mID sql.NullString
	err := s.db.QueryRow(lookup, rootID).Scan(&currentVersion, &initiatorID, &counterpartyEmail, &cID, &mID)
	if err != nil {
		return nil, fmt.Errorf("failed to locate root negotiation: %w", err)
	}
	counterpartyID = cID.String
	mediatorID = mID.String

	// 2. Insert new version
	var draft DraftEscrow
	query := `
		INSERT INTO draft_escrows (root_id, version, proposer_id, initiator_id, counterparty_email, counterparty_id, mediator_id, amount, currency, terms, metadata, change_summary, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, 'NEGOTIATION')
		RETURNING id, root_id, version, proposer_id, initiator_id, counterparty_email, counterparty_id, mediator_id, amount, currency, terms, metadata, change_summary, status, created_at
	`
	var cs sql.NullString
	err = s.db.QueryRow(query, rootID, currentVersion+1, proposerID, initiatorID, counterpartyEmail, cID, mID, amount, currency, terms, metadata, summary).Scan(
		&draft.ID, &draft.RootID, &draft.Version, &draft.ProposerID, &draft.InitiatorID, &draft.CounterpartyEmail, &cID, &mID, &draft.Amount, &draft.Currency, &draft.Terms, &draft.Metadata, &cs, &draft.Status, &draft.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to propose amendment: %w", err)
	}
	draft.CounterpartyID = cID.String
	draft.MediatorID = mID.String
	draft.ChangeSummary = cs.String
	draft.Approvals = []string{}
	return &draft, nil
}

// AddApproval appends an identity to the approvals array of the LATEST version.
func (s *ConfigService) AddApproval(rootID, approverID string) error {
	query := `
		UPDATE draft_escrows 
		SET approvals = approvals || jsonb_build_array($1::text)
		WHERE id = (SELECT id FROM draft_escrows WHERE root_id = $2 ORDER BY version DESC LIMIT 1)
		AND NOT (approvals ? $1) -- Prevent double signature on same version
	`
	res, err := s.db.Exec(query, approverID, rootID)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("version already approved or negotiation not found")
	}

	// Logic for transition to RATIFIED could go here
	return nil
}

func (s *ConfigService) GetLatestDraft(rootID string) (*DraftEscrow, error) {
	var draft DraftEscrow
	query := "SELECT id, root_id, version, proposer_id, invitation_code, initiator_id, counterparty_email, counterparty_id, mediator_id, amount, currency, terms, metadata, change_summary, approvals, status, created_at FROM draft_escrows WHERE root_id = $1 ORDER BY version DESC LIMIT 1"
	
	var medID, counterID, inviteCode, changeSum sql.NullString
	var approvalsJSON []byte
	err := s.db.QueryRow(query, rootID).Scan(
		&draft.ID, &draft.RootID, &draft.Version, &draft.ProposerID, &inviteCode, &draft.InitiatorID, &draft.CounterpartyEmail, &counterID, &medID, &draft.Amount, &draft.Currency, &draft.Terms, &draft.Metadata, &changeSum, &approvalsJSON, &draft.Status, &draft.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	draft.InvitationCode = inviteCode.String
	draft.CounterpartyID = counterID.String
	draft.MediatorID = medID.String
	draft.ChangeSummary = changeSum.String
	_ = json.Unmarshal(approvalsJSON, &draft.Approvals)
	return &draft, nil
}

func (s *ConfigService) ClaimDraft(code, counterpartyID string) error {
	query := `
		UPDATE draft_escrows 
		SET counterparty_id = $1, status = 'CLAIMED', updated_at = CURRENT_TIMESTAMP 
		WHERE invitation_code = $2 AND status = 'DRAFT'
	`
	// Since invitation_code is only on version 1, this works
	res, err := s.db.Exec(query, counterpartyID, code)
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
		SELECT DISTINCT ON (root_id) 
			id, root_id, version, proposer_id, initiator_id, counterparty_email, counterparty_id, mediator_id, amount, currency, terms, metadata, status, created_at
		FROM draft_escrows 
		WHERE initiator_id = $1 OR counterparty_email = $2 OR counterparty_id = $1 OR mediator_id = $1
		ORDER BY root_id, version DESC
	`
	rows, err := s.db.Query(query, userID, email)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var drafts []*DraftEscrow
	for rows.Next() {
		var d DraftEscrow
		var medID, counterID sql.NullString
		if err := rows.Scan(&d.ID, &d.RootID, &d.Version, &d.ProposerID, &d.InitiatorID, &d.CounterpartyEmail, &counterID, &medID, &d.Amount, &d.Currency, &d.Terms, &d.Metadata, &d.Status, &d.CreatedAt); err != nil {
			return nil, err
		}
		d.CounterpartyID = counterID.String
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
