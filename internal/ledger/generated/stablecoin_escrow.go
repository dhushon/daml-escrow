package generated

import (
	"fmt"

	"github.com/smartcontractkit/go-daml/pkg/codec"
	. "github.com/smartcontractkit/go-daml/pkg/types"
)

const (
	PackageID = "ec35fce924adbefbae43d1f546879c29fdc42b9efac531f4de8eaeb39a5693c1"
)

// Milestone is a Record type
type Milestone struct {
	Label     TEXT    `json:"label"`
	Amount    NUMERIC `json:"amount"`
	Completed BOOL    `json:"completed"`
}

// toMap converts Milestone to a map for DAML arguments
func (m Milestone) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"label":     m.Label,
		"amount":    m.Amount,
		"completed": m.Completed,
	}
}

// StablecoinEscrow is a Template type
type StablecoinEscrow struct {
	Issuer                PARTY       `json:"issuer"`
	Buyer                 PARTY       `json:"buyer"`
	Seller                PARTY       `json:"seller"`
	Mediator              PARTY       `json:"mediator"`
	TotalAmount           NUMERIC     `json:"totalAmount"`
	Currency              TEXT        `json:"currency"`
	Description           TEXT        `json:"description"`
	Milestones            []Milestone `json:"milestones"`
	CurrentMilestoneIndex INT64       `json:"currentMilestoneIndex"`
}

func (t StablecoinEscrow) GetTemplateID() string {
	return fmt.Sprintf("%s:%s:%s", PackageID, "StablecoinEscrow", "StablecoinEscrow")
}

func (t StablecoinEscrow) MarshalJSON() ([]byte, error) {
	return codec.NewJsonCodec().Marshal(t)
}

func (t *StablecoinEscrow) UnmarshalJSON(data []byte) error {
	return codec.NewJsonCodec().Unmarshal(data, t)
}

// EscrowSettlement is a Template type
type EscrowSettlement struct {
	Issuer    PARTY   `json:"issuer"`
	Recipient PARTY   `json:"recipient"`
	Amount    NUMERIC `json:"amount"`
	Currency  TEXT    `json:"currency"`
	Status    TEXT    `json:"status"`
}

func (t EscrowSettlement) GetTemplateID() string {
	return fmt.Sprintf("%s:%s:%s", PackageID, "StablecoinEscrow", "EscrowSettlement")
}

// DisputedEscrow is a Template type
type DisputedEscrow struct {
	Escrow StablecoinEscrow `json:"escrow"`
}

func (t DisputedEscrow) GetTemplateID() string {
	return fmt.Sprintf("%s:%s:%s", PackageID, "StablecoinEscrow", "DisputedEscrow")
}

// Choice argument types

type ResolveDispute struct {
	PayoutToBuyer  NUMERIC `json:"payoutToBuyer"`
	PayoutToSeller NUMERIC `json:"payoutToSeller"`
}

func (r ResolveDispute) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"payoutToBuyer":  string(r.PayoutToBuyer),
		"payoutToSeller": string(r.PayoutToSeller),
	}
}
