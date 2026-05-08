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

	if strings.HasPrefix(id, "u-") && !strings.ContainsAny(id, "@|.") {
		return id, nil
	}

	s := strings.ReplaceAll(id, "|", "-")
	s = strings.ReplaceAll(s, "@", "-")
	s = strings.ReplaceAll(s, ".", "-")
	return "u-" + s, nil
}

// ValidateTripartiteRoles authoritatively enforces the institutional role requirements.
// Buyer and Seller are MANDATORY; Mediator is optional.
func ValidateTripartiteRoles(buyer, seller, mediator string) error {
	if strings.TrimSpace(buyer) == "" {
		return errors.New("institutional mandate violation: Buyer identity cannot be null")
	}
	if strings.TrimSpace(seller) == "" {
		return errors.New("institutional mandate violation: Seller identity cannot be null")
	}

	// Roles MUST be distinct (enforced here and in DAML)
	if buyer == seller {
		return errors.New("role exclusivity violation: Buyer and Seller cannot be the same identity")
	}

	if mediator != "" {
		if mediator == buyer || mediator == seller {
			return errors.New("role exclusivity violation: Mediator cannot be a transacting party (Buyer or Seller)")
		}
	}

	return nil
}
