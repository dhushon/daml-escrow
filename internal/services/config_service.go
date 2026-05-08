package services

import (
	"database/sql"
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
	InitiatorID       string          `json:"initiatorId"`
	CounterpartyEmail string          `json:"counterpartyEmail"`
	MediatorID        string          `json:"mediatorId"`
	Amount            float64         `json:"amount"`
	Currency          string          `json:"currency"`
	Terms             json.RawMessage `json:"terms"`
	Metadata          json.RawMessage `json:"metadata"`
	Status            string          `json:"status"` // DRAFT, NEGOTIATION, RATIFIED
	CreatedAt         time.Time       `json:"createdAt"`
	UpdatedAt         time.Time       `json:"updatedAt"`
}

func (s *ConfigService) CreateDraft(initiatorID, counterpartyEmail string, amount float64, currency string, terms, metadata json.RawMessage) (*DraftEscrow, error) {
	var draft DraftEscrow
	query := `
		INSERT INTO draft_escrows (initiator_id, counterparty_email, amount, currency, terms, metadata)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, initiator_id, counterparty_email, amount, currency, terms, metadata, status, created_at, updated_at
	`
	err := s.db.QueryRow(query, initiatorID, counterpartyEmail, amount, currency, terms, metadata).Scan(
		&draft.ID, &draft.InitiatorID, &draft.CounterpartyEmail, &draft.Amount, &draft.Currency, &draft.Terms, &draft.Metadata, &draft.Status, &draft.CreatedAt, &draft.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create draft escrow: %w", err)
	}
	return &draft, nil
}

func (s *ConfigService) GetDraft(id string) (*DraftEscrow, error) {
	var draft DraftEscrow
	query := "SELECT id, initiator_id, counterparty_email, mediator_id, amount, currency, terms, metadata, status, created_at, updated_at FROM draft_escrows WHERE id = $1"
	var medID sql.NullString
	err := s.db.QueryRow(query, id).Scan(
		&draft.ID, &draft.InitiatorID, &draft.CounterpartyEmail, &medID, &draft.Amount, &draft.Currency, &draft.Terms, &draft.Metadata, &draft.Status, &draft.CreatedAt, &draft.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	draft.MediatorID = medID.String
	return &draft, nil
}

func (s *ConfigService) ListDraftsForUser(userID, email string) ([]*DraftEscrow, error) {
	query := "SELECT id, initiator_id, counterparty_email, mediator_id, amount, currency, terms, metadata, status, created_at, updated_at FROM draft_escrows WHERE initiator_id = $1 OR counterparty_email = $2"
	rows, err := s.db.Query(query, userID, email)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var drafts []*DraftEscrow
	for rows.Next() {
		var d DraftEscrow
		var medID sql.NullString
		if err := rows.Scan(&d.ID, &d.InitiatorID, &d.CounterpartyEmail, &medID, &d.Amount, &d.Currency, &d.Terms, &d.Metadata, &d.Status, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, err
		}
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
