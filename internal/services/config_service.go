package services

import (
	"database/sql"
	"encoding/json"
	"fmt"

	_ "github.com/lib/pq"
)

type ConfigService struct {
	db *sql.DB
}

func NewConfigService(dsn string) (*ConfigService, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open user_config db: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping user_config db: %w", err)
	}

	return &ConfigService{db: db}, nil
}

func (s *ConfigService) GetConfig(userID, key string) (json.RawMessage, error) {
	var val json.RawMessage
	err := s.db.QueryRow(
		"SELECT config_value FROM configs WHERE user_id = $1 AND config_key = $2",
		userID, key,
	).Scan(&val)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	return val, err
}

func (s *ConfigService) SaveConfig(userID, key string, value json.RawMessage) error {
	_, err := s.db.Exec(`
		INSERT INTO configs (user_id, config_key, config_value, updated_at)
		VALUES ($1, $2, $3, CURRENT_TIMESTAMP)
		ON CONFLICT (user_id, config_key) 
		DO UPDATE SET config_value = $3, updated_at = CURRENT_TIMESTAMP
	`, userID, key, value)
	return err
}

func (s *ConfigService) Close() error {
	return s.db.Close()
}
