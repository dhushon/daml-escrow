package ledger

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIdentityUtilities(t *testing.T) {
	t.Run("SanitizeID", func(t *testing.T) {
		// Valid sanitization
		id, err := SanitizeID("joey@buyer.com")
		assert.NoError(t, err)
		assert.Equal(t, "u-joey-buyer-com", id)

		// Already sanitized
		id, err = SanitizeID("u-jimmy-seller")
		assert.NoError(t, err)
		assert.Equal(t, "u-jimmy-seller", id)

		// Character collisions
		id, err = SanitizeID("user|one@test.local")
		assert.NoError(t, err)
		assert.Equal(t, "u-user-one-test-local", id)

		// Empty/Nil check
		id, err = SanitizeID("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be null or empty")
		assert.Empty(t, id)
	})

	t.Run("IsValidID", func(t *testing.T) {
		assert.True(t, IsValidID("u-valid-id"))
		assert.False(t, IsValidID("invalid@id"))
		assert.False(t, IsValidID("u-invalid|id"))
		assert.False(t, IsValidID("u-invalid.id"))
		assert.False(t, IsValidID(""))
		assert.False(t, IsValidID("   "))
	})

	t.Run("CheckValidID", func(t *testing.T) {
		assert.NoError(t, CheckValidID("u-valid-id"))
		assert.Error(t, CheckValidID("invalid@id"))
		assert.Error(t, CheckValidID(""))
	})

	t.Run("ValidateTripartiteRoles", func(t *testing.T) {
		// Valid triplet (must use sanitized/valid IDs for this check now)
		err := ValidateTripartiteRoles("u-buyer", "u-seller", "u-mediator")
		assert.NoError(t, err)

		// Valid optional mediator
		err = ValidateTripartiteRoles("u-buyer", "u-seller", "")
		assert.NoError(t, err)

		// Missing Buyer
		err = ValidateTripartiteRoles("", "u-seller", "u-mediator")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "identity cannot be null or empty")

		// Missing Seller
		err = ValidateTripartiteRoles("u-buyer", "", "u-mediator")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "identity cannot be null or empty")

		// Role Overlap: Buyer == Seller
		err = ValidateTripartiteRoles("u-user", "u-user", "u-mediator")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Buyer and Seller cannot be the same")

		// Role Overlap: Mediator == Buyer
		err = ValidateTripartiteRoles("u-buyer", "u-seller", "u-buyer")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Mediator cannot be a transacting party")
	})
}
