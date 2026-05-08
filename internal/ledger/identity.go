package ledger

import (
	"errors"
	"strings"
)

// SanitizeID authoritatively converts any string into a Daml-compliant identifier.
// It returns an error if the input is empty to prevent unauthorized 'nil' operations.
func SanitizeID(id string) (string, error) {
	if strings.TrimSpace(id) == "" {
		return "", errors.New("identity identifier cannot be null or empty")
	}

	// If already valid, do not sanitize
	if IsValidID(id) {
		return id, nil
	}

	s := strings.ReplaceAll(id, "|", "-")
	s = strings.ReplaceAll(s, "@", "-")
	s = strings.ReplaceAll(s, ".", "-")
	return "u-" + s, nil
}

// IsValidID is a common validator checking if an ID matches the required Daml-compliant pattern.
// Nil/Empty is authoritatively NOT valid.
func IsValidID(id string) bool {
	if strings.TrimSpace(id) == "" {
		return false
	}
	// Must start with u- and contain no forbidden characters
	return strings.HasPrefix(id, "u-") && !strings.ContainsAny(id, "@|.")
}

// CheckValidID verifies if an ID is valid independently of roles.
// It returns an error for invalid patterns or empty strings.
func CheckValidID(id string) error {
	if strings.TrimSpace(id) == "" {
		return errors.New("identity cannot be null or empty")
	}
	if !IsValidID(id) {
		return errors.New("identity does not match the required Daml-compliant pattern (u- prefix, no special characters)")
	}
	return nil
}

// ValidateTripartiteRoles authoritatively enforces the institutional role requirements.
// Buyer and Seller are MANDATORY; Mediator is optional.
func ValidateTripartiteRoles(buyer, seller, mediator string) error {
	// First check basic validity of mandatory roles
	if err := CheckValidID(buyer); err != nil {
		return errors.New("institutional mandate violation: " + err.Error())
	}
	if err := CheckValidID(seller); err != nil {
		return errors.New("institutional mandate violation: " + err.Error())
	}

	// Roles MUST be distinct (enforced here and in DAML)
	if buyer == seller {
		return errors.New("role exclusivity violation: Buyer and Seller cannot be the same identity")
	}

	if mediator != "" {
		if err := CheckValidID(mediator); err != nil {
			return errors.New("mediator validation error: " + err.Error())
		}
		if mediator == buyer || mediator == seller {
			return errors.New("role exclusivity violation: Mediator cannot be a transacting party (Buyer or Seller)")
		}
	}

	return nil
}
