package services

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

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

// User Configuration CRUD

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

// --- Phase 11: Draft Escrow Management ---

type DraftEscrow struct {
	ID                string          `json:"id"`
	InvitationCode    string          `json:"invitationCode"`
	InitiatorID       string          `json:"initiatorId"`
	CounterpartyEmail string          `json:"counterpartyEmail"`
	CounterpartyID    string          `json:"counterpartyId"`
	MediatorID        string          `json:"mediatorId"`
	Amount            float64         `json:"amount"`
	Currency          string          `json:"currency"`
	Terms             json.RawMessage `json:"terms"`
	Metadata          json.RawMessage `json:"metadata"`
	Status            string          `json:"status"` // DRAFT, CLAIMED, NEGOTIATION, RATIFIED
	CreatedAt         time.Time       `json:"createdAt"`
	UpdatedAt         time.Time       `json:"updatedAt"`
}

func generateInvitationCode() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func (s *ConfigService) CreateDraft(initiatorID, counterpartyEmail string, amount float64, currency string, terms, metadata json.RawMessage) (*DraftEscrow, error) {
	var draft DraftEscrow
	code := generateInvitationCode()
	query := `
		INSERT INTO draft_escrows (invitation_code, initiator_id, counterparty_email, amount, currency, terms, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, invitation_code, initiator_id, counterparty_email, amount, currency, terms, metadata, status, created_at, updated_at
	`
	err := s.db.QueryRow(query, code, initiatorID, counterpartyEmail, amount, currency, terms, metadata).Scan(
		&draft.ID, &draft.InvitationCode, &draft.InitiatorID, &draft.CounterpartyEmail, &draft.Amount, &draft.Currency, &draft.Terms, &draft.Metadata, &draft.Status, &draft.CreatedAt, &draft.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create draft escrow: %w", err)
	}
	return &draft, nil
}

func (s *ConfigService) GetDraft(id string) (*DraftEscrow, error) {
	var draft DraftEscrow
	query := "SELECT id, invitation_code, initiator_id, counterparty_email, counterparty_id, mediator_id, amount, currency, terms, metadata, status, created_at, updated_at FROM draft_escrows WHERE id = $1"
	var medID, counterID, inviteCode sql.NullString
	err := s.db.QueryRow(query, id).Scan(
		&draft.ID, &inviteCode, &draft.InitiatorID, &draft.CounterpartyEmail, &counterID, &medID, &draft.Amount, &draft.Currency, &draft.Terms, &draft.Metadata, &draft.Status, &draft.CreatedAt, &draft.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	draft.InvitationCode = inviteCode.String
	draft.CounterpartyID = counterID.String
	draft.MediatorID = medID.String
	return &draft, nil
}

func (s *ConfigService) ClaimDraft(code, counterpartyID string) error {
	query := `
		UPDATE draft_escrows 
		SET counterparty_id = $1, status = 'CLAIMED', updated_at = CURRENT_TIMESTAMP 
		WHERE invitation_code = $2 AND status = 'DRAFT'
	`
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
	query := "SELECT id, invitation_code, initiator_id, counterparty_email, counterparty_id, mediator_id, amount, currency, terms, metadata, status, created_at, updated_at FROM draft_escrows WHERE initiator_id = $1 OR counterparty_email = $2 OR counterparty_id = $1"
	rows, err := s.db.Query(query, userID, email)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var drafts []*DraftEscrow
	for rows.Next() {
		var d DraftEscrow
		var medID, counterID, inviteCode sql.NullString
		if err := rows.Scan(&d.ID, &inviteCode, &d.InitiatorID, &d.CounterpartyEmail, &counterID, &medID, &d.Amount, &d.Currency, &d.Terms, &d.Metadata, &d.Status, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, err
		}
		d.InvitationCode = inviteCode.String
		d.CounterpartyID = counterID.String
		d.MediatorID = medID.String
		drafts = append(drafts, &d)
	}
	return drafts, nil
}

func (s *ConfigService) UpdateDraftStatus(id, status string) error {
	query := "UPDATE draft_escrows SET status = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2"
	_, err := s.db.Exec(query, status, id)
	return err
}
