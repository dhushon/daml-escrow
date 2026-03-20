package generated

const (
	PackageID          = "16c70b6dba04109abc3003b3877a641f38ae9ce9a6c72ecf68e2e8c0e1053756"
	InterfacePackageID = "eeada456377e4287fabfe089057b419d54159c87f98da712fd543122fc7c39f3"
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
