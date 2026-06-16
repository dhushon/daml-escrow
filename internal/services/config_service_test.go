package services

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigService_Unit(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	svc := &ConfigService{db: db}

	t.Run("GetConfig - Success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"config_value"}).
			AddRow([]byte(`{"value": 123}`))

		mock.ExpectQuery("SELECT config_value FROM configs").
			WithArgs("user1", "key1").
			WillReturnRows(rows)

		val, err := svc.GetConfig("user1", "key1")
		assert.NoError(t, err)
		assert.JSONEq(t, `{"value": 123}`, string(val))
	})

	t.Run("GetConfig - Not Found", func(t *testing.T) {
		mock.ExpectQuery("SELECT config_value FROM configs").
			WithArgs("user1", "key2").
			WillReturnRows(sqlmock.NewRows([]string{"config_value"})) // Return empty rows set

		val, err := svc.GetConfig("user1", "key2")
		assert.NoError(t, err)
		assert.Nil(t, val)
	})

	t.Run("SaveConfig - Success", func(t *testing.T) {
		payload := json.RawMessage(`{"theme": "dark"}`)
		mock.ExpectExec("INSERT INTO configs").
			WithArgs("user1", "theme", payload).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := svc.SaveConfig("user1", "theme", payload)
		assert.NoError(t, err)
	})

	t.Run("UpdateDraftStatusAndLedgerID - Success", func(t *testing.T) {
		mock.ExpectQuery("SELECT root_id, metadata FROM draft_escrows WHERE id = \\$1").
			WithArgs("draft-id-123").
			WillReturnRows(sqlmock.NewRows([]string{"root_id", "metadata"}).AddRow("root-id-456", []byte(`{"existing": "data"}`)))

		mock.ExpectExec("UPDATE draft_escrows SET status = \\$1, metadata = \\$2 WHERE id = \\$3").
			WithArgs("PROMOTED", sqlmock.AnyArg(), "draft-id-123").
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := svc.UpdateDraftStatusAndLedgerID("draft-id-123", "PROMOTED", "ledger-id-abc")
		assert.NoError(t, err)
	})

	t.Run("WithdrawDraft - Success", func(t *testing.T) {
		rows1 := sqlmock.NewRows([]string{
			"id", "root_id", "version", "proposer_id", "invitation_code", "contract_type", "initiator_id",
			"initiator_role", "depositor_id", "beneficiary_email", "beneficiary_id", "mediator_id",
			"amount", "currency", "terms", "metadata", "change_summary", "approvals", "status", "created_at",
		}).AddRow(
			"draft-id-123", "root-id-456", 1, "user1", "code", "Corporate", "user1",
			"Depositor", "user1", "user2@email.com", "user2", "mediator",
			100.0, "USD", []byte(`{}`), []byte(`{"ledgerId": "old-ledger-id"}`), "summary", []byte(`["user1"]`), "PROMOTED", time.Now(),
		)
		mock.ExpectQuery("SELECT id, root_id, version, proposer_id, invitation_code, contract_type, initiator_id, initiator_role, depositor_id, beneficiary_email, beneficiary_id, mediator_id, amount, currency, terms, metadata, change_summary, approvals, status, created_at FROM draft_escrows WHERE root_id = \\$1").
			WithArgs("root-id-456").
			WillReturnRows(rows1)

		mock.ExpectExec("UPDATE draft_escrows SET status = 'DRAFT'").
			WithArgs(sqlmock.AnyArg(), "draft-id-123").
			WillReturnResult(sqlmock.NewResult(1, 1))

		rows2 := sqlmock.NewRows([]string{
			"id", "root_id", "version", "proposer_id", "invitation_code", "contract_type", "initiator_id",
			"initiator_role", "depositor_id", "beneficiary_email", "beneficiary_id", "mediator_id",
			"amount", "currency", "terms", "metadata", "change_summary", "approvals", "status", "created_at",
		}).AddRow(
			"draft-id-123", "root-id-456", 1, "user1", "code", "Corporate", "user1",
			"Depositor", "user1", "user2@email.com", "user2", "mediator",
			100.0, "USD", []byte(`{}`), []byte(`{"previousProposalId": "old-ledger-id"}`), "summary", []byte(`[]`), "DRAFT", time.Now(),
		)
		mock.ExpectQuery("SELECT id, root_id, version, proposer_id, invitation_code, contract_type, initiator_id, initiator_role, depositor_id, beneficiary_email, beneficiary_id, mediator_id, amount, currency, terms, metadata, change_summary, approvals, status, created_at FROM draft_escrows WHERE root_id = \\$1").
			WithArgs("root-id-456").
			WillReturnRows(rows2)

		draft, err := svc.WithdrawDraft("root-id-456", "old-ledger-id")
		assert.NoError(t, err)
		assert.Equal(t, "DRAFT", draft.Status)
		assert.Len(t, draft.Approvals, 0)
	})
}
