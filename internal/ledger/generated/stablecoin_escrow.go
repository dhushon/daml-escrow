package generated

const (
	PackageID          = "4a1d3b2bc1a0c5141bd83bcc2228189e5398cf7484e47afc2e8ccaeafadef7a7"
	InterfacePackageID = "d209d27f09adfc9883015b5f23e89f28df6d507c31846cd09e4f2e2bb8b0726b"
)

type Milestone struct {
	Label     string  `json:"label"`
	Amount    float64 `json:"amount"`
	Completed bool    `json:"completed"`
}

type StablecoinEscrow struct {
	Issuer                string      `json:"issuer"`
	Buyer                 string      `json:"buyer"`
	Seller                string      `json:"seller"`
	Mediator              string      `json:"mediator"`
	TotalAmount           float64     `json:"totalAmount"`
	Currency              string      `json:"currency"`
	Description           string      `json:"description"`
	Milestones            []Milestone `json:"milestones"`
	CurrentMilestoneIndex int         `json:"currentMilestoneIndex"`
	Metadata              string      `json:"metadata"`
}

type StablecoinDisputedEscrow struct {
	Issuer                string      `json:"issuer"`
	Buyer                 string      `json:"buyer"`
	Seller                string      `json:"seller"`
	Mediator              string      `json:"mediator"`
	TotalAmount           float64     `json:"totalAmount"`
	Currency              string      `json:"currency"`
	Description           string      `json:"description"`
	Milestones            []Milestone `json:"milestones"`
	CurrentMilestoneIndex int         `json:"currentMilestoneIndex"`
	Metadata              string      `json:"metadata"`
}

type EscrowSettlement struct {
	Issuer    string  `json:"issuer"`
	Recipient string  `json:"recipient"`
	Amount    float64 `json:"amount"`
	Currency  string  `json:"currency"`
	Status    string  `json:"status"`
}
