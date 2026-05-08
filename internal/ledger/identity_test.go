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

		id, err = SanitizeID("   ")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be null or empty")
	})

	t.Run("ValidateTripartiteRoles", func(t *testing.T) {
		// Valid triplet
		err := ValidateTripartiteRoles("u-buyer", "u-seller", "u-mediator")
		assert.NoError(t, err)

		// Valid optional mediator
		err = ValidateTripartiteRoles("u-buyer", "u-seller", "")
		assert.NoError(t, err)

		// Missing Buyer
		err = ValidateTripartiteRoles("", "u-seller", "u-mediator")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Buyer identity cannot be null")

		// Missing Seller
		err = ValidateTripartiteRoles("u-buyer", "", "u-mediator")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Seller identity cannot be null")

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
