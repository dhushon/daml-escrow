package generated

import (
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/smartcontractkit/go-daml/pkg/bind"
	"github.com/smartcontractkit/go-daml/pkg/codec"
	"github.com/smartcontractkit/go-daml/pkg/model"
	"github.com/smartcontractkit/go-daml/pkg/types"
)

var (
	_ = fmt.Sprintf
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = model.Command{}
	_ bind.BoundTemplate
)

// IBase is a DAML interface
type IBase interface {

	// Archive executes the Archive choice
	Archive(contractID string) *model.ExerciseCommand
}

// IEscrow is a DAML interface
type IEscrow interface {

	// Archive executes the Archive choice
	Archive(contractID string) *model.ExerciseCommand

	// GetStatus executes the GetStatus choice
	GetStatus(contractID string, args GetStatus) *model.ExerciseCommand

	// Activate executes the Activate choice
	Activate(contractID string, args Activate) *model.ExerciseCommand
}

// IEscrowEvent is a DAML interface
type IEscrowEvent interface {

	// Archive executes the Archive choice
	Archive(contractID string) *model.ExerciseCommand
}

// IHolding is a DAML interface
type IHolding interface {

	// Archive executes the Archive choice
	Archive(contractID string) *model.ExerciseCommand

	// ArchiveHolding executes the ArchiveHolding choice
	ArchiveHolding(contractID string, args ArchiveHolding) *model.ExerciseCommand
}

// ILockable is a DAML interface
type ILockable interface {

	// Archive executes the Archive choice
	Archive(contractID string) *model.ExerciseCommand

	// Lock executes the Lock choice
	Lock(contractID string, args Lock) *model.ExerciseCommand

	// Unlock executes the Unlock choice
	Unlock(contractID string, args Unlock) *model.ExerciseCommand
}

// ITransferable is a DAML interface
type ITransferable interface {

	// Archive executes the Archive choice
	Archive(contractID string) *model.ExerciseCommand

	// Transfer executes the Transfer choice
	Transfer(contractID string, args Transfer) *model.ExerciseCommand
}

// Activate is a Record type
type Activate struct {
}

// ToMap converts Activate to a map for DAML arguments
func (t Activate) ToMap() map[string]any {
	m := make(map[string]any)
	return m
}

func (t Activate) MarshalJSON() ([]byte, error) {
	jsonCodec := codec.NewJsonCodec()
	return jsonCodec.Marshal(t)
}

func (t *Activate) UnmarshalJSON(data []byte) error {
	jsonCodec := codec.NewJsonCodec()
	return jsonCodec.Unmarshal(data, t)
}

// ArchiveHolding is a Record type
type ArchiveHolding struct {
}

// ToMap converts ArchiveHolding to a map for DAML arguments
func (t ArchiveHolding) ToMap() map[string]any {
	m := make(map[string]any)
	return m
}

func (t ArchiveHolding) MarshalJSON() ([]byte, error) {
	jsonCodec := codec.NewJsonCodec()
	return jsonCodec.Marshal(t)
}

func (t *ArchiveHolding) UnmarshalJSON(data []byte) error {
	jsonCodec := codec.NewJsonCodec()
	return jsonCodec.Unmarshal(data, t)
}

// Asset is a Record type
type Asset struct {
	AssetType  types.TEXT         `json:"assetType"`
	AssetId    types.TEXT         `json:"assetId"`
	Amount     types.NUMERIC      `json:"amount"`
	Currency   types.TEXT         `json:"currency"`
	CustodyRef *types.TEXT        `json:"custodyRef" hex:"optional"`
	HoldingCid *types.CONTRACT_ID `json:"holdingCid" hex:"optional"`
}

// ToMap converts Asset to a map for DAML arguments
func (t Asset) ToMap() map[string]any {
	m := make(map[string]any)

	m["assetType"] = string(t.AssetType)

	m["assetId"] = string(t.AssetId)

	m["amount"] = t.Amount

	m["currency"] = string(t.Currency)

	if t.CustodyRef != nil {
		m["custodyRef"] = map[string]any{
			"_type": "optional",
			"value": string(*t.CustodyRef),
		}
	} else {
		m["custodyRef"] = map[string]any{
			"_type": "optional",
		}
	}

	if t.HoldingCid != nil {
		m["holdingCid"] = map[string]any{
			"_type": "optional",
			"value": *t.HoldingCid,
		}
	} else {
		m["holdingCid"] = map[string]any{
			"_type": "optional",
		}
	}

	return m
}

func (t Asset) MarshalJSON() ([]byte, error) {
	jsonCodec := codec.NewJsonCodec()
	return jsonCodec.Marshal(t)
}

func (t *Asset) UnmarshalJSON(data []byte) error {
	jsonCodec := codec.NewJsonCodec()
	return jsonCodec.Unmarshal(data, t)
}

// BaseView is a Record type
type BaseView struct {
	Issuer   types.PARTY `json:"issuer"`
	AssetId  types.TEXT  `json:"assetId"`
	Metadata types.TEXT  `json:"metadata"`
}

// ToMap converts BaseView to a map for DAML arguments
func (t BaseView) ToMap() map[string]any {
	m := make(map[string]any)

	m["issuer"] = t.Issuer.ToMap()

	m["assetId"] = string(t.AssetId)

	m["metadata"] = string(t.Metadata)

	return m
}

func (t BaseView) MarshalJSON() ([]byte, error) {
	jsonCodec := codec.NewJsonCodec()
	return jsonCodec.Marshal(t)
}

func (t *BaseView) UnmarshalJSON(data []byte) error {
	jsonCodec := codec.NewJsonCodec()
	return jsonCodec.Unmarshal(data, t)
}

// EscrowState is an enum type
type EscrowState string

const (
	EscrowStateDRAFT EscrowState = "DRAFT"

	EscrowStateFUNDED EscrowState = "FUNDED"

	EscrowStateACTIVE EscrowState = "ACTIVE"

	EscrowStateDISPUTED EscrowState = "DISPUTED"

	EscrowStatePROPOSED EscrowState = "PROPOSED"

	EscrowStateSETTLED EscrowState = "SETTLED"

	EscrowStateEXPIRED EscrowState = "EXPIRED"

	EscrowStateCANCELLED EscrowState = "CANCELLED"
)

func (e EscrowState) GetEnumConstructor() string { return string(e) }

func (e EscrowState) GetEnumTypeID() string {
	return fmt.Sprintf("#%s:%s:%s", PackageName, "Escrow.Interface", "EscrowState")
}

// GetEnumTypeIDWithPackageID returns the enum type ID using the provided package ID instead of package name
func (e EscrowState) GetEnumTypeIDWithPackageID(packageID string) string {
	return fmt.Sprintf("#%s:%s:%s", packageID, "Escrow.Interface", "EscrowState")
}

func (e EscrowState) MarshalJSON() ([]byte, error) {
	jsonCodec := codec.NewJsonCodec()
	return jsonCodec.Marshal(e)
}

func (e *EscrowState) UnmarshalJSON(data []byte) error {
	jsonCodec := codec.NewJsonCodec()
	return jsonCodec.Unmarshal(data, e)
}

var _ types.ENUM = EscrowState("")

// EscrowTerms is a Record type
type EscrowTerms struct {
	ConditionDescription types.TEXT      `json:"conditionDescription"`
	ConditionType        types.TEXT      `json:"conditionType"`
	EvidenceRequired     types.TEXT      `json:"evidenceRequired"`
	ExpiryDate           types.TIMESTAMP `json:"expiryDate"`
	GracePeriodDays      types.INT64     `json:"gracePeriodDays"`
	DisputeWindowDays    types.INT64     `json:"disputeWindowDays"`
	PartialSchedule      []types.Tuple2  `json:"partialSchedule"`
	MinAmount            *types.NUMERIC  `json:"minAmount" hex:"optional"`
	MaxAmount            *types.NUMERIC  `json:"maxAmount" hex:"optional"`
}

// ToMap converts EscrowTerms to a map for DAML arguments
func (t EscrowTerms) ToMap() map[string]any {
	m := make(map[string]any)

	m["conditionDescription"] = string(t.ConditionDescription)

	m["conditionType"] = string(t.ConditionType)

	m["evidenceRequired"] = string(t.EvidenceRequired)

	m["expiryDate"] = t.ExpiryDate

	m["gracePeriodDays"] = int64(t.GracePeriodDays)

	m["disputeWindowDays"] = int64(t.DisputeWindowDays)

	m["partialSchedule"] = func() []any {
		res := make([]any, 0, len(t.PartialSchedule))
		for _, e := range t.PartialSchedule {
			type mapper interface{ toMap() map[string]any }
			if m, ok := any(e).(mapper); ok {
				res = append(res, m.toMap())
			} else {
				res = append(res, e)
			}
		}
		return res
	}()

	if t.MinAmount != nil {
		m["minAmount"] = map[string]any{
			"_type": "optional",
			"value": *t.MinAmount,
		}
	} else {
		m["minAmount"] = map[string]any{
			"_type": "optional",
		}
	}

	if t.MaxAmount != nil {
		m["maxAmount"] = map[string]any{
			"_type": "optional",
			"value": *t.MaxAmount,
		}
	} else {
		m["maxAmount"] = map[string]any{
			"_type": "optional",
		}
	}

	return m
}

func (t EscrowTerms) MarshalJSON() ([]byte, error) {
	jsonCodec := codec.NewJsonCodec()
	return jsonCodec.Marshal(t)
}

func (t *EscrowTerms) UnmarshalJSON(data []byte) error {
	jsonCodec := codec.NewJsonCodec()
	return jsonCodec.Unmarshal(data, t)
}

// EventView is a Record type
type EventView struct {
	EscrowId  types.TEXT      `json:"escrowId"`
	EventType types.TEXT      `json:"eventType"`
	Timestamp types.TIMESTAMP `json:"timestamp"`
	Actor     types.PARTY     `json:"actor"`
	Details   types.TEXT      `json:"details"`
}

// ToMap converts EventView to a map for DAML arguments
func (t EventView) ToMap() map[string]any {
	m := make(map[string]any)

	m["escrowId"] = string(t.EscrowId)

	m["eventType"] = string(t.EventType)

	m["timestamp"] = t.Timestamp

	m["actor"] = t.Actor.ToMap()

	m["details"] = string(t.Details)

	return m
}

func (t EventView) MarshalJSON() ([]byte, error) {
	jsonCodec := codec.NewJsonCodec()
	return jsonCodec.Marshal(t)
}

func (t *EventView) UnmarshalJSON(data []byte) error {
	jsonCodec := codec.NewJsonCodec()
	return jsonCodec.Unmarshal(data, t)
}

// GetStatus is a Record type
type GetStatus struct {
}

// ToMap converts GetStatus to a map for DAML arguments
func (t GetStatus) ToMap() map[string]any {
	m := make(map[string]any)
	return m
}

func (t GetStatus) MarshalJSON() ([]byte, error) {
	jsonCodec := codec.NewJsonCodec()
	return jsonCodec.Marshal(t)
}

func (t *GetStatus) UnmarshalJSON(data []byte) error {
	jsonCodec := codec.NewJsonCodec()
	return jsonCodec.Unmarshal(data, t)
}

// HoldingView is a Record type
type HoldingView struct {
	Owner  types.PARTY   `json:"owner"`
	Amount types.NUMERIC `json:"amount"`
	Base   BaseView      `json:"base"`
}

// ToMap converts HoldingView to a map for DAML arguments
func (t HoldingView) ToMap() map[string]any {
	m := make(map[string]any)

	m["owner"] = t.Owner.ToMap()

	m["amount"] = t.Amount

	m["base"] = func() any {
		type mapper interface{ toMap() map[string]any }
		if m, ok := any(t.Base).(mapper); ok {
			return m.toMap()
		}
		return t.Base
	}()

	return m
}

func (t HoldingView) MarshalJSON() ([]byte, error) {
	jsonCodec := codec.NewJsonCodec()
	return jsonCodec.Marshal(t)
}

func (t *HoldingView) UnmarshalJSON(data []byte) error {
	jsonCodec := codec.NewJsonCodec()
	return jsonCodec.Unmarshal(data, t)
}

// Lock is a Record type
type Lock struct {
	Locker  types.PARTY `json:"locker"`
	Context types.TEXT  `json:"context"`
}

// ToMap converts Lock to a map for DAML arguments
func (t Lock) ToMap() map[string]any {
	m := make(map[string]any)

	m["locker"] = t.Locker.ToMap()

	m["context"] = string(t.Context)

	return m
}

func (t Lock) MarshalJSON() ([]byte, error) {
	jsonCodec := codec.NewJsonCodec()
	return jsonCodec.Marshal(t)
}

func (t *Lock) UnmarshalJSON(data []byte) error {
	jsonCodec := codec.NewJsonCodec()
	return jsonCodec.Unmarshal(data, t)
}

// LockView is a Record type
type LockView struct {
	Locker  types.PARTY `json:"locker"`
	Context types.TEXT  `json:"context"`
}

// ToMap converts LockView to a map for DAML arguments
func (t LockView) ToMap() map[string]any {
	m := make(map[string]any)

	m["locker"] = t.Locker.ToMap()

	m["context"] = string(t.Context)

	return m
}

func (t LockView) MarshalJSON() ([]byte, error) {
	jsonCodec := codec.NewJsonCodec()
	return jsonCodec.Marshal(t)
}

func (t *LockView) UnmarshalJSON(data []byte) error {
	jsonCodec := codec.NewJsonCodec()
	return jsonCodec.Unmarshal(data, t)
}

// SettlementTerms is a Record type
type SettlementTerms struct {
	SettlementType     types.TEXT    `json:"settlementType"`
	DepositorReturn    types.NUMERIC `json:"depositorReturn"`
	BeneficiaryPayment types.NUMERIC `json:"beneficiaryPayment"`
	MediatorFee        types.NUMERIC `json:"mediatorFee"`
}

// ToMap converts SettlementTerms to a map for DAML arguments
func (t SettlementTerms) ToMap() map[string]any {
	m := make(map[string]any)

	m["settlementType"] = string(t.SettlementType)

	m["depositorReturn"] = t.DepositorReturn

	m["beneficiaryPayment"] = t.BeneficiaryPayment

	m["mediatorFee"] = t.MediatorFee

	return m
}

func (t SettlementTerms) MarshalJSON() ([]byte, error) {
	jsonCodec := codec.NewJsonCodec()
	return jsonCodec.Marshal(t)
}

func (t *SettlementTerms) UnmarshalJSON(data []byte) error {
	jsonCodec := codec.NewJsonCodec()
	return jsonCodec.Unmarshal(data, t)
}

// Transfer is a Record type
type Transfer struct {
	NewOwner types.PARTY `json:"newOwner"`
}

// ToMap converts Transfer to a map for DAML arguments
func (t Transfer) ToMap() map[string]any {
	m := make(map[string]any)

	m["newOwner"] = t.NewOwner.ToMap()

	return m
}

func (t Transfer) MarshalJSON() ([]byte, error) {
	jsonCodec := codec.NewJsonCodec()
	return jsonCodec.Marshal(t)
}

func (t *Transfer) UnmarshalJSON(data []byte) error {
	jsonCodec := codec.NewJsonCodec()
	return jsonCodec.Unmarshal(data, t)
}

// Unlock is a Record type
type Unlock struct {
	Actor types.PARTY `json:"actor"`
}

// ToMap converts Unlock to a map for DAML arguments
func (t Unlock) ToMap() map[string]any {
	m := make(map[string]any)

	m["actor"] = t.Actor.ToMap()

	return m
}

func (t Unlock) MarshalJSON() ([]byte, error) {
	jsonCodec := codec.NewJsonCodec()
	return jsonCodec.Marshal(t)
}

func (t *Unlock) UnmarshalJSON(data []byte) error {
	jsonCodec := codec.NewJsonCodec()
	return jsonCodec.Unmarshal(data, t)
}

// View is a Record type
type View struct {
	Issuer               types.PARTY   `json:"issuer"`
	Initiator            types.PARTY   `json:"initiator"`
	Depositors           []types.PARTY `json:"depositors"`
	DepositorThreshold   types.INT64   `json:"depositorThreshold"`
	Beneficiaries        []types.PARTY `json:"beneficiaries"`
	BeneficiaryThreshold types.INT64   `json:"beneficiaryThreshold"`
	Mediator             types.PARTY   `json:"mediator"`
	ContractType         types.TEXT    `json:"contractType"`
	Asset                Asset         `json:"asset"`
	Terms                EscrowTerms   `json:"terms"`
	State                EscrowState   `json:"state"`
	Metadata             types.TEXT    `json:"metadata"`
}

// ToMap converts View to a map for DAML arguments
func (t View) ToMap() map[string]any {
	m := make(map[string]any)

	m["issuer"] = t.Issuer.ToMap()

	m["initiator"] = t.Initiator.ToMap()

	m["depositors"] = func() []any {
		res := make([]any, 0, len(t.Depositors))
		for _, e := range t.Depositors {
			res = append(res, e.ToMap())
		}
		return res
	}()

	m["depositorThreshold"] = int64(t.DepositorThreshold)

	m["beneficiaries"] = func() []any {
		res := make([]any, 0, len(t.Beneficiaries))
		for _, e := range t.Beneficiaries {
			res = append(res, e.ToMap())
		}
		return res
	}()

	m["beneficiaryThreshold"] = int64(t.BeneficiaryThreshold)

	m["mediator"] = t.Mediator.ToMap()

	m["contractType"] = string(t.ContractType)

	m["asset"] = func() any {
		type mapper interface{ toMap() map[string]any }
		if m, ok := any(t.Asset).(mapper); ok {
			return m.toMap()
		}
		return t.Asset
	}()

	m["terms"] = func() any {
		type mapper interface{ toMap() map[string]any }
		if m, ok := any(t.Terms).(mapper); ok {
			return m.toMap()
		}
		return t.Terms
	}()

	m["state"] = func() any {
		type mapper interface{ toMap() map[string]any }
		if m, ok := any(t.State).(mapper); ok {
			return m.toMap()
		}
		return t.State
	}()

	m["metadata"] = string(t.Metadata)

	return m
}

func (t View) MarshalJSON() ([]byte, error) {
	jsonCodec := codec.NewJsonCodec()
	return jsonCodec.Marshal(t)
}

func (t *View) UnmarshalJSON(data []byte) error {
	jsonCodec := codec.NewJsonCodec()
	return jsonCodec.Unmarshal(data, t)
}

// IBaseInterfaceID returns the interface ID for the IBase interface using the package name
func IBaseInterfaceID() string {
	return fmt.Sprintf("#%s:%s:%s", PackageName, "Token.CIP56", "Base")
}

// IBaseInterfaceIDWithPackageID returns the interface ID using the provided package ID instead of package name
func IBaseInterfaceIDWithPackageID(packageID string) string {
	return fmt.Sprintf("%s:%s:%s", packageID, "Token.CIP56", "Base")
}

// IEscrowInterfaceID returns the interface ID for the IEscrow interface using the package name
func IEscrowInterfaceID() string {
	return fmt.Sprintf("#%s:%s:%s", PackageName, "Escrow.Interface", "Escrow")
}

// IEscrowInterfaceIDWithPackageID returns the interface ID using the provided package ID instead of package name
func IEscrowInterfaceIDWithPackageID(packageID string) string {
	return fmt.Sprintf("%s:%s:%s", packageID, "Escrow.Interface", "Escrow")
}

// IEscrowEventInterfaceID returns the interface ID for the IEscrowEvent interface using the package name
func IEscrowEventInterfaceID() string {
	return fmt.Sprintf("#%s:%s:%s", PackageName, "Escrow.Interface", "EscrowEvent")
}

// IEscrowEventInterfaceIDWithPackageID returns the interface ID using the provided package ID instead of package name
func IEscrowEventInterfaceIDWithPackageID(packageID string) string {
	return fmt.Sprintf("%s:%s:%s", packageID, "Escrow.Interface", "EscrowEvent")
}

// IHoldingInterfaceID returns the interface ID for the IHolding interface using the package name
func IHoldingInterfaceID() string {
	return fmt.Sprintf("#%s:%s:%s", PackageName, "Token.CIP56", "Holding")
}

// IHoldingInterfaceIDWithPackageID returns the interface ID using the provided package ID instead of package name
func IHoldingInterfaceIDWithPackageID(packageID string) string {
	return fmt.Sprintf("%s:%s:%s", packageID, "Token.CIP56", "Holding")
}

// ILockableInterfaceID returns the interface ID for the ILockable interface using the package name
func ILockableInterfaceID() string {
	return fmt.Sprintf("#%s:%s:%s", PackageName, "Token.CIP56", "Lockable")
}

// ILockableInterfaceIDWithPackageID returns the interface ID using the provided package ID instead of package name
func ILockableInterfaceIDWithPackageID(packageID string) string {
	return fmt.Sprintf("%s:%s:%s", packageID, "Token.CIP56", "Lockable")
}

// ITransferableInterfaceID returns the interface ID for the ITransferable interface using the package name
func ITransferableInterfaceID() string {
	return fmt.Sprintf("#%s:%s:%s", PackageName, "Token.CIP56", "Transferable")
}

// ITransferableInterfaceIDWithPackageID returns the interface ID using the provided package ID instead of package name
func ITransferableInterfaceIDWithPackageID(packageID string) string {
	return fmt.Sprintf("%s:%s:%s", packageID, "Token.CIP56", "Transferable")
}
