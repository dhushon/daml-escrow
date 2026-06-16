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

const (
	PackageName = "stablecoin-escrow"
	PackageID   = "e40073ba574f8a8532388dae368fcd1458dd7a7bc5a2384ddbed9f097879f50b"
	SDKVersion  = "3.4.11"
)

type Template interface {
	CreateCommand() *model.CreateCommand
	GetTemplateID() string
}

func argsToMap(args any) map[string]any {
	if args == nil {
		return map[string]any{}
	}

	if m, ok := args.(map[string]any); ok {
		return m
	}

	type mapper interface {
		ToMap() map[string]any
	}
	if mapper, ok := args.(mapper); ok {
		return mapper.ToMap()
	}

	return map[string]any{"args": args}
}

// BeneficiaryAccept is a Record type
type BeneficiaryAccept struct {
}

// ToMap converts BeneficiaryAccept to a map for DAML arguments
func (t BeneficiaryAccept) ToMap() map[string]any {
	m := make(map[string]any)
	return m
}

func (t BeneficiaryAccept) MarshalJSON() ([]byte, error) {
	jsonCodec := codec.NewJsonCodec()
	return jsonCodec.Marshal(t)
}

func (t *BeneficiaryAccept) UnmarshalJSON(data []byte) error {
	jsonCodec := codec.NewJsonCodec()
	return jsonCodec.Unmarshal(data, t)
}

// BeneficiaryAcceptedProposal is a Template type
type BeneficiaryAcceptedProposal struct {
	Issuer       types.PARTY `json:"issuer"`
	Initiator    types.PARTY `json:"initiator"`
	Depositor    types.PARTY `json:"depositor"`
	Beneficiary  types.PARTY `json:"beneficiary"`
	Mediator     types.PARTY `json:"mediator"`
	ContractType types.TEXT  `json:"contractType"`
	Asset        Asset       `json:"asset"`
	Terms        EscrowTerms `json:"terms"`
	Metadata     types.TEXT  `json:"metadata"`
}

// GetTemplateID returns the template ID for this template using the package name
func (t BeneficiaryAcceptedProposal) GetTemplateID() string {
	return fmt.Sprintf("#%s:%s:%s", PackageName, "StablecoinEscrow", "BeneficiaryAcceptedProposal")
}

// GetTemplateIDWithPackageID returns the template ID using the provided package ID instead of package name
func (t BeneficiaryAcceptedProposal) GetTemplateIDWithPackageID(packageID string) string {
	return fmt.Sprintf("%s:%s:%s", packageID, "StablecoinEscrow", "BeneficiaryAcceptedProposal")
}

// CreateCommand returns a CreateCommand for this template using the package name
func (t BeneficiaryAcceptedProposal) CreateCommand() *model.CreateCommand {
	args := make(map[string]any)

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["issuer"] = t.Issuer.ToMap()

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["initiator"] = t.Initiator.ToMap()

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["depositor"] = t.Depositor.ToMap()

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["beneficiary"] = t.Beneficiary.ToMap()

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["mediator"] = t.Mediator.ToMap()

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["contractType"] = string(t.ContractType)

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["asset"] = func() any {
		type mapper interface{ toMap() map[string]any }
		if m, ok := any(t.Asset).(mapper); ok {
			return m.toMap()
		}
		return t.Asset
	}()

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["terms"] = func() any {
		type mapper interface{ toMap() map[string]any }
		if m, ok := any(t.Terms).(mapper); ok {
			return m.toMap()
		}
		return t.Terms
	}()

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["metadata"] = string(t.Metadata)

	return &model.CreateCommand{
		TemplateID: t.GetTemplateID(),
		Arguments:  args,
	}
}

// CreateCommandWithPackageID returns a CreateCommand using the provided package ID instead of package name
func (t BeneficiaryAcceptedProposal) CreateCommandWithPackageID(packageID string) *model.CreateCommand {
	args := make(map[string]any)

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["issuer"] = t.Issuer.ToMap()

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["initiator"] = t.Initiator.ToMap()

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["depositor"] = t.Depositor.ToMap()

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["beneficiary"] = t.Beneficiary.ToMap()

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["mediator"] = t.Mediator.ToMap()

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["contractType"] = string(t.ContractType)

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["asset"] = func() any {
		type mapper interface{ toMap() map[string]any }
		if m, ok := any(t.Asset).(mapper); ok {
			return m.toMap()
		}
		return t.Asset
	}()

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["terms"] = func() any {
		type mapper interface{ toMap() map[string]any }
		if m, ok := any(t.Terms).(mapper); ok {
			return m.toMap()
		}
		return t.Terms
	}()

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["metadata"] = string(t.Metadata)

	return &model.CreateCommand{
		TemplateID: t.GetTemplateIDWithPackageID(packageID),
		Arguments:  args,
	}
}

func (t BeneficiaryAcceptedProposal) MarshalJSON() ([]byte, error) {
	jsonCodec := codec.NewJsonCodec()
	return jsonCodec.Marshal(t)
}

func (t *BeneficiaryAcceptedProposal) UnmarshalJSON(data []byte) error {
	jsonCodec := codec.NewJsonCodec()
	return jsonCodec.Unmarshal(data, t)
}

// Choice methods for BeneficiaryAcceptedProposal

// Fund exercises the Fund choice on this BeneficiaryAcceptedProposal contract
// This method uses the package name in the template ID
func (t BeneficiaryAcceptedProposal) Fund(contractID string, args Fund) *model.ExerciseCommand {
	return &model.ExerciseCommand{
		TemplateID: fmt.Sprintf("#%s:%s:%s", PackageName, "StablecoinEscrow", "BeneficiaryAcceptedProposal"),
		ContractID: contractID,
		Choice:     "Fund",
		Arguments:  argsToMap(args),
	}
}

// FundWithPackageID exercises the Fund choice using the provided package ID instead of package name
func (t BeneficiaryAcceptedProposal) FundWithPackageID(contractID string, packageID string, args Fund) *model.ExerciseCommand {
	return &model.ExerciseCommand{
		TemplateID: fmt.Sprintf("#%s:%s:%s", packageID, "StablecoinEscrow", "BeneficiaryAcceptedProposal"),
		ContractID: contractID,
		Choice:     "Fund",
		Arguments:  argsToMap(args),
	}
}

// Archive exercises the Archive choice on this BeneficiaryAcceptedProposal contract
// This method uses the package name in the template ID
func (t BeneficiaryAcceptedProposal) Archive(contractID string) *model.ExerciseCommand {
	return &model.ExerciseCommand{
		TemplateID: fmt.Sprintf("#%s:%s:%s", PackageName, "StablecoinEscrow", "BeneficiaryAcceptedProposal"),
		ContractID: contractID,
		Choice:     "Archive",
		Arguments:  map[string]any{},
	}
}

// ArchiveWithPackageID exercises the Archive choice using the provided package ID instead of package name
func (t BeneficiaryAcceptedProposal) ArchiveWithPackageID(contractID string, packageID string) *model.ExerciseCommand {
	return &model.ExerciseCommand{
		TemplateID: fmt.Sprintf("#%s:%s:%s", packageID, "StablecoinEscrow", "BeneficiaryAcceptedProposal"),
		ContractID: contractID,
		Choice:     "Archive",
		Arguments:  map[string]any{},
	}
}

// GetStatus exercises the GetStatus choice on this BeneficiaryAcceptedProposal contract via the IEscrow interface
// This method uses the package name in the template ID
func (t BeneficiaryAcceptedProposal) GetStatus(contractID string, args GetStatus) *model.ExerciseCommand {
	return &model.ExerciseCommand{
		TemplateID: fmt.Sprintf("#%s:%s:%s", PackageName, "StablecoinEscrow", "Escrow"),
		ContractID: contractID,
		Choice:     "GetStatus",
		Arguments:  argsToMap(args),
	}
}

// GetStatusWithPackageID exercises the GetStatus choice using the provided package ID instead of package name
func (t BeneficiaryAcceptedProposal) GetStatusWithPackageID(contractID string, packageID string, args GetStatus) *model.ExerciseCommand {
	return &model.ExerciseCommand{
		TemplateID: fmt.Sprintf("#%s:%s:%s", packageID, "StablecoinEscrow", "Escrow"),
		ContractID: contractID,
		Choice:     "GetStatus",
		Arguments:  argsToMap(args),
	}
}

// Activate exercises the Activate choice on this BeneficiaryAcceptedProposal contract via the IEscrow interface
// This method uses the package name in the template ID
func (t BeneficiaryAcceptedProposal) Activate(contractID string, args Activate) *model.ExerciseCommand {
	return &model.ExerciseCommand{
		TemplateID: fmt.Sprintf("#%s:%s:%s", PackageName, "StablecoinEscrow", "Escrow"),
		ContractID: contractID,
		Choice:     "Activate",
		Arguments:  argsToMap(args),
	}
}

// ActivateWithPackageID exercises the Activate choice using the provided package ID instead of package name
func (t BeneficiaryAcceptedProposal) ActivateWithPackageID(contractID string, packageID string, args Activate) *model.ExerciseCommand {
	return &model.ExerciseCommand{
		TemplateID: fmt.Sprintf("#%s:%s:%s", packageID, "StablecoinEscrow", "Escrow"),
		ContractID: contractID,
		Choice:     "Activate",
		Arguments:  argsToMap(args),
	}
}

// Verify interface implementations for BeneficiaryAcceptedProposal

var _ IEscrow = (*BeneficiaryAcceptedProposal)(nil)

// CancelProposal is a Record type
type CancelProposal struct {
}

// ToMap converts CancelProposal to a map for DAML arguments
func (t CancelProposal) ToMap() map[string]any {
	m := make(map[string]any)
	return m
}

func (t CancelProposal) MarshalJSON() ([]byte, error) {
	jsonCodec := codec.NewJsonCodec()
	return jsonCodec.Marshal(t)
}

func (t *CancelProposal) UnmarshalJSON(data []byte) error {
	jsonCodec := codec.NewJsonCodec()
	return jsonCodec.Unmarshal(data, t)
}

// ClaimInvitation is a Record type
type ClaimInvitation struct {
	ClaimingParty     types.PARTY `json:"claimingParty"`
	RegistrationToken types.TEXT  `json:"registrationToken"`
}

// ToMap converts ClaimInvitation to a map for DAML arguments
func (t ClaimInvitation) ToMap() map[string]any {
	m := make(map[string]any)

	m["claimingParty"] = t.ClaimingParty.ToMap()

	m["registrationToken"] = string(t.RegistrationToken)

	return m
}

func (t ClaimInvitation) MarshalJSON() ([]byte, error) {
	jsonCodec := codec.NewJsonCodec()
	return jsonCodec.Marshal(t)
}

func (t *ClaimInvitation) UnmarshalJSON(data []byte) error {
	jsonCodec := codec.NewJsonCodec()
	return jsonCodec.Unmarshal(data, t)
}

// ConfirmConditions is a Record type
type ConfirmConditions struct {
}

// ToMap converts ConfirmConditions to a map for DAML arguments
func (t ConfirmConditions) ToMap() map[string]any {
	m := make(map[string]any)
	return m
}

func (t ConfirmConditions) MarshalJSON() ([]byte, error) {
	jsonCodec := codec.NewJsonCodec()
	return jsonCodec.Marshal(t)
}

func (t *ConfirmConditions) UnmarshalJSON(data []byte) error {
	jsonCodec := codec.NewJsonCodec()
	return jsonCodec.Unmarshal(data, t)
}

// Disburse is a Record type
type Disburse struct {
}

// ToMap converts Disburse to a map for DAML arguments
func (t Disburse) ToMap() map[string]any {
	m := make(map[string]any)
	return m
}

func (t Disburse) MarshalJSON() ([]byte, error) {
	jsonCodec := codec.NewJsonCodec()
	return jsonCodec.Marshal(t)
}

func (t *Disburse) UnmarshalJSON(data []byte) error {
	jsonCodec := codec.NewJsonCodec()
	return jsonCodec.Unmarshal(data, t)
}

// DisbursementOrder is a Template type
type DisbursementOrder struct {
	Issuer       types.PARTY     `json:"issuer"`
	Initiator    types.PARTY     `json:"initiator"`
	Depositor    types.PARTY     `json:"depositor"`
	Beneficiary  types.PARTY     `json:"beneficiary"`
	Mediator     types.PARTY     `json:"mediator"`
	ContractType types.TEXT      `json:"contractType"`
	Asset        Asset           `json:"asset"`
	Terms        EscrowTerms     `json:"terms"`
	Metadata     types.TEXT      `json:"metadata"`
	Settlement   SettlementTerms `json:"settlement"`
}

// GetTemplateID returns the template ID for this template using the package name
func (t DisbursementOrder) GetTemplateID() string {
	return fmt.Sprintf("#%s:%s:%s", PackageName, "StablecoinEscrow", "DisbursementOrder")
}

// GetTemplateIDWithPackageID returns the template ID using the provided package ID instead of package name
func (t DisbursementOrder) GetTemplateIDWithPackageID(packageID string) string {
	return fmt.Sprintf("%s:%s:%s", packageID, "StablecoinEscrow", "DisbursementOrder")
}

// CreateCommand returns a CreateCommand for this template using the package name
func (t DisbursementOrder) CreateCommand() *model.CreateCommand {
	args := make(map[string]any)

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["issuer"] = t.Issuer.ToMap()

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["initiator"] = t.Initiator.ToMap()

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["depositor"] = t.Depositor.ToMap()

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["beneficiary"] = t.Beneficiary.ToMap()

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["mediator"] = t.Mediator.ToMap()

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["contractType"] = string(t.ContractType)

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["asset"] = func() any {
		type mapper interface{ toMap() map[string]any }
		if m, ok := any(t.Asset).(mapper); ok {
			return m.toMap()
		}
		return t.Asset
	}()

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["terms"] = func() any {
		type mapper interface{ toMap() map[string]any }
		if m, ok := any(t.Terms).(mapper); ok {
			return m.toMap()
		}
		return t.Terms
	}()

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["metadata"] = string(t.Metadata)

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["settlement"] = func() any {
		type mapper interface{ toMap() map[string]any }
		if m, ok := any(t.Settlement).(mapper); ok {
			return m.toMap()
		}
		return t.Settlement
	}()

	return &model.CreateCommand{
		TemplateID: t.GetTemplateID(),
		Arguments:  args,
	}
}

// CreateCommandWithPackageID returns a CreateCommand using the provided package ID instead of package name
func (t DisbursementOrder) CreateCommandWithPackageID(packageID string) *model.CreateCommand {
	args := make(map[string]any)

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["issuer"] = t.Issuer.ToMap()

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["initiator"] = t.Initiator.ToMap()

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["depositor"] = t.Depositor.ToMap()

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["beneficiary"] = t.Beneficiary.ToMap()

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["mediator"] = t.Mediator.ToMap()

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["contractType"] = string(t.ContractType)

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["asset"] = func() any {
		type mapper interface{ toMap() map[string]any }
		if m, ok := any(t.Asset).(mapper); ok {
			return m.toMap()
		}
		return t.Asset
	}()

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["terms"] = func() any {
		type mapper interface{ toMap() map[string]any }
		if m, ok := any(t.Terms).(mapper); ok {
			return m.toMap()
		}
		return t.Terms
	}()

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["metadata"] = string(t.Metadata)

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["settlement"] = func() any {
		type mapper interface{ toMap() map[string]any }
		if m, ok := any(t.Settlement).(mapper); ok {
			return m.toMap()
		}
		return t.Settlement
	}()

	return &model.CreateCommand{
		TemplateID: t.GetTemplateIDWithPackageID(packageID),
		Arguments:  args,
	}
}

func (t DisbursementOrder) MarshalJSON() ([]byte, error) {
	jsonCodec := codec.NewJsonCodec()
	return jsonCodec.Marshal(t)
}

func (t *DisbursementOrder) UnmarshalJSON(data []byte) error {
	jsonCodec := codec.NewJsonCodec()
	return jsonCodec.Unmarshal(data, t)
}

// Choice methods for DisbursementOrder

// Disburse exercises the Disburse choice on this DisbursementOrder contract
// This method uses the package name in the template ID
func (t DisbursementOrder) Disburse(contractID string, args Disburse) *model.ExerciseCommand {
	return &model.ExerciseCommand{
		TemplateID: fmt.Sprintf("#%s:%s:%s", PackageName, "StablecoinEscrow", "DisbursementOrder"),
		ContractID: contractID,
		Choice:     "Disburse",
		Arguments:  argsToMap(args),
	}
}

// DisburseWithPackageID exercises the Disburse choice using the provided package ID instead of package name
func (t DisbursementOrder) DisburseWithPackageID(contractID string, packageID string, args Disburse) *model.ExerciseCommand {
	return &model.ExerciseCommand{
		TemplateID: fmt.Sprintf("#%s:%s:%s", packageID, "StablecoinEscrow", "DisbursementOrder"),
		ContractID: contractID,
		Choice:     "Disburse",
		Arguments:  argsToMap(args),
	}
}

// Archive exercises the Archive choice on this DisbursementOrder contract
// This method uses the package name in the template ID
func (t DisbursementOrder) Archive(contractID string) *model.ExerciseCommand {
	return &model.ExerciseCommand{
		TemplateID: fmt.Sprintf("#%s:%s:%s", PackageName, "StablecoinEscrow", "DisbursementOrder"),
		ContractID: contractID,
		Choice:     "Archive",
		Arguments:  map[string]any{},
	}
}

// ArchiveWithPackageID exercises the Archive choice using the provided package ID instead of package name
func (t DisbursementOrder) ArchiveWithPackageID(contractID string, packageID string) *model.ExerciseCommand {
	return &model.ExerciseCommand{
		TemplateID: fmt.Sprintf("#%s:%s:%s", packageID, "StablecoinEscrow", "DisbursementOrder"),
		ContractID: contractID,
		Choice:     "Archive",
		Arguments:  map[string]any{},
	}
}

// GetStatus exercises the GetStatus choice on this DisbursementOrder contract via the IEscrow interface
// This method uses the package name in the template ID
func (t DisbursementOrder) GetStatus(contractID string, args GetStatus) *model.ExerciseCommand {
	return &model.ExerciseCommand{
		TemplateID: fmt.Sprintf("#%s:%s:%s", PackageName, "StablecoinEscrow", "Escrow"),
		ContractID: contractID,
		Choice:     "GetStatus",
		Arguments:  argsToMap(args),
	}
}

// GetStatusWithPackageID exercises the GetStatus choice using the provided package ID instead of package name
func (t DisbursementOrder) GetStatusWithPackageID(contractID string, packageID string, args GetStatus) *model.ExerciseCommand {
	return &model.ExerciseCommand{
		TemplateID: fmt.Sprintf("#%s:%s:%s", packageID, "StablecoinEscrow", "Escrow"),
		ContractID: contractID,
		Choice:     "GetStatus",
		Arguments:  argsToMap(args),
	}
}

// Activate exercises the Activate choice on this DisbursementOrder contract via the IEscrow interface
// This method uses the package name in the template ID
func (t DisbursementOrder) Activate(contractID string, args Activate) *model.ExerciseCommand {
	return &model.ExerciseCommand{
		TemplateID: fmt.Sprintf("#%s:%s:%s", PackageName, "StablecoinEscrow", "Escrow"),
		ContractID: contractID,
		Choice:     "Activate",
		Arguments:  argsToMap(args),
	}
}

// ActivateWithPackageID exercises the Activate choice using the provided package ID instead of package name
func (t DisbursementOrder) ActivateWithPackageID(contractID string, packageID string, args Activate) *model.ExerciseCommand {
	return &model.ExerciseCommand{
		TemplateID: fmt.Sprintf("#%s:%s:%s", packageID, "StablecoinEscrow", "Escrow"),
		ContractID: contractID,
		Choice:     "Activate",
		Arguments:  argsToMap(args),
	}
}

// Verify interface implementations for DisbursementOrder

var _ IEscrow = (*DisbursementOrder)(nil)

// DisputeRecord is a Template type
type DisputeRecord struct {
	Issuer       types.PARTY `json:"issuer"`
	Initiator    types.PARTY `json:"initiator"`
	Depositor    types.PARTY `json:"depositor"`
	Beneficiary  types.PARTY `json:"beneficiary"`
	Mediator     types.PARTY `json:"mediator"`
	ContractType types.TEXT  `json:"contractType"`
	Asset        Asset       `json:"asset"`
	Terms        EscrowTerms `json:"terms"`
	Metadata     types.TEXT  `json:"metadata"`
	RaisingParty types.PARTY `json:"raisingParty"`
}

// GetTemplateID returns the template ID for this template using the package name
func (t DisputeRecord) GetTemplateID() string {
	return fmt.Sprintf("#%s:%s:%s", PackageName, "StablecoinEscrow", "DisputeRecord")
}

// GetTemplateIDWithPackageID returns the template ID using the provided package ID instead of package name
func (t DisputeRecord) GetTemplateIDWithPackageID(packageID string) string {
	return fmt.Sprintf("%s:%s:%s", packageID, "StablecoinEscrow", "DisputeRecord")
}

// CreateCommand returns a CreateCommand for this template using the package name
func (t DisputeRecord) CreateCommand() *model.CreateCommand {
	args := make(map[string]any)

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["issuer"] = t.Issuer.ToMap()

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["initiator"] = t.Initiator.ToMap()

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["depositor"] = t.Depositor.ToMap()

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["beneficiary"] = t.Beneficiary.ToMap()

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["mediator"] = t.Mediator.ToMap()

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["contractType"] = string(t.ContractType)

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["asset"] = func() any {
		type mapper interface{ toMap() map[string]any }
		if m, ok := any(t.Asset).(mapper); ok {
			return m.toMap()
		}
		return t.Asset
	}()

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["terms"] = func() any {
		type mapper interface{ toMap() map[string]any }
		if m, ok := any(t.Terms).(mapper); ok {
			return m.toMap()
		}
		return t.Terms
	}()

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["metadata"] = string(t.Metadata)

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["raisingParty"] = t.RaisingParty.ToMap()

	return &model.CreateCommand{
		TemplateID: t.GetTemplateID(),
		Arguments:  args,
	}
}

// CreateCommandWithPackageID returns a CreateCommand using the provided package ID instead of package name
func (t DisputeRecord) CreateCommandWithPackageID(packageID string) *model.CreateCommand {
	args := make(map[string]any)

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["issuer"] = t.Issuer.ToMap()

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["initiator"] = t.Initiator.ToMap()

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["depositor"] = t.Depositor.ToMap()

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["beneficiary"] = t.Beneficiary.ToMap()

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["mediator"] = t.Mediator.ToMap()

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["contractType"] = string(t.ContractType)

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["asset"] = func() any {
		type mapper interface{ toMap() map[string]any }
		if m, ok := any(t.Asset).(mapper); ok {
			return m.toMap()
		}
		return t.Asset
	}()

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["terms"] = func() any {
		type mapper interface{ toMap() map[string]any }
		if m, ok := any(t.Terms).(mapper); ok {
			return m.toMap()
		}
		return t.Terms
	}()

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["metadata"] = string(t.Metadata)

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["raisingParty"] = t.RaisingParty.ToMap()

	return &model.CreateCommand{
		TemplateID: t.GetTemplateIDWithPackageID(packageID),
		Arguments:  args,
	}
}

func (t DisputeRecord) MarshalJSON() ([]byte, error) {
	jsonCodec := codec.NewJsonCodec()
	return jsonCodec.Marshal(t)
}

func (t *DisputeRecord) UnmarshalJSON(data []byte) error {
	jsonCodec := codec.NewJsonCodec()
	return jsonCodec.Unmarshal(data, t)
}

// Choice methods for DisputeRecord

// ProposeSettlement exercises the ProposeSettlement choice on this DisputeRecord contract
// This method uses the package name in the template ID
func (t DisputeRecord) ProposeSettlement(contractID string, args ProposeSettlement) *model.ExerciseCommand {
	return &model.ExerciseCommand{
		TemplateID: fmt.Sprintf("#%s:%s:%s", PackageName, "StablecoinEscrow", "DisputeRecord"),
		ContractID: contractID,
		Choice:     "ProposeSettlement",
		Arguments:  argsToMap(args),
	}
}

// ProposeSettlementWithPackageID exercises the ProposeSettlement choice using the provided package ID instead of package name
func (t DisputeRecord) ProposeSettlementWithPackageID(contractID string, packageID string, args ProposeSettlement) *model.ExerciseCommand {
	return &model.ExerciseCommand{
		TemplateID: fmt.Sprintf("#%s:%s:%s", packageID, "StablecoinEscrow", "DisputeRecord"),
		ContractID: contractID,
		Choice:     "ProposeSettlement",
		Arguments:  argsToMap(args),
	}
}

// Archive exercises the Archive choice on this DisputeRecord contract
// This method uses the package name in the template ID
func (t DisputeRecord) Archive(contractID string) *model.ExerciseCommand {
	return &model.ExerciseCommand{
		TemplateID: fmt.Sprintf("#%s:%s:%s", PackageName, "StablecoinEscrow", "DisputeRecord"),
		ContractID: contractID,
		Choice:     "Archive",
		Arguments:  map[string]any{},
	}
}

// ArchiveWithPackageID exercises the Archive choice using the provided package ID instead of package name
func (t DisputeRecord) ArchiveWithPackageID(contractID string, packageID string) *model.ExerciseCommand {
	return &model.ExerciseCommand{
		TemplateID: fmt.Sprintf("#%s:%s:%s", packageID, "StablecoinEscrow", "DisputeRecord"),
		ContractID: contractID,
		Choice:     "Archive",
		Arguments:  map[string]any{},
	}
}

// GetStatus exercises the GetStatus choice on this DisputeRecord contract via the IEscrow interface
// This method uses the package name in the template ID
func (t DisputeRecord) GetStatus(contractID string, args GetStatus) *model.ExerciseCommand {
	return &model.ExerciseCommand{
		TemplateID: fmt.Sprintf("#%s:%s:%s", PackageName, "StablecoinEscrow", "Escrow"),
		ContractID: contractID,
		Choice:     "GetStatus",
		Arguments:  argsToMap(args),
	}
}

// GetStatusWithPackageID exercises the GetStatus choice using the provided package ID instead of package name
func (t DisputeRecord) GetStatusWithPackageID(contractID string, packageID string, args GetStatus) *model.ExerciseCommand {
	return &model.ExerciseCommand{
		TemplateID: fmt.Sprintf("#%s:%s:%s", packageID, "StablecoinEscrow", "Escrow"),
		ContractID: contractID,
		Choice:     "GetStatus",
		Arguments:  argsToMap(args),
	}
}

// Activate exercises the Activate choice on this DisputeRecord contract via the IEscrow interface
// This method uses the package name in the template ID
func (t DisputeRecord) Activate(contractID string, args Activate) *model.ExerciseCommand {
	return &model.ExerciseCommand{
		TemplateID: fmt.Sprintf("#%s:%s:%s", PackageName, "StablecoinEscrow", "Escrow"),
		ContractID: contractID,
		Choice:     "Activate",
		Arguments:  argsToMap(args),
	}
}

// ActivateWithPackageID exercises the Activate choice using the provided package ID instead of package name
func (t DisputeRecord) ActivateWithPackageID(contractID string, packageID string, args Activate) *model.ExerciseCommand {
	return &model.ExerciseCommand{
		TemplateID: fmt.Sprintf("#%s:%s:%s", packageID, "StablecoinEscrow", "Escrow"),
		ContractID: contractID,
		Choice:     "Activate",
		Arguments:  argsToMap(args),
	}
}

// Verify interface implementations for DisputeRecord

var _ IEscrow = (*DisputeRecord)(nil)

// EscrowContract is a Template type
type EscrowContract struct {
	Issuer       types.PARTY `json:"issuer"`
	Initiator    types.PARTY `json:"initiator"`
	Depositor    types.PARTY `json:"depositor"`
	Beneficiary  types.PARTY `json:"beneficiary"`
	Mediator     types.PARTY `json:"mediator"`
	ContractType types.TEXT  `json:"contractType"`
	Asset        Asset       `json:"asset"`
	Terms        EscrowTerms `json:"terms"`
	Metadata     types.TEXT  `json:"metadata"`
	State        EscrowState `json:"state"`
}

// GetTemplateID returns the template ID for this template using the package name
func (t EscrowContract) GetTemplateID() string {
	return fmt.Sprintf("#%s:%s:%s", PackageName, "StablecoinEscrow", "EscrowContract")
}

// GetTemplateIDWithPackageID returns the template ID using the provided package ID instead of package name
func (t EscrowContract) GetTemplateIDWithPackageID(packageID string) string {
	return fmt.Sprintf("%s:%s:%s", packageID, "StablecoinEscrow", "EscrowContract")
}

// CreateCommand returns a CreateCommand for this template using the package name
func (t EscrowContract) CreateCommand() *model.CreateCommand {
	args := make(map[string]any)

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["issuer"] = t.Issuer.ToMap()

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["initiator"] = t.Initiator.ToMap()

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["depositor"] = t.Depositor.ToMap()

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["beneficiary"] = t.Beneficiary.ToMap()

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["mediator"] = t.Mediator.ToMap()

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["contractType"] = string(t.ContractType)

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["asset"] = func() any {
		type mapper interface{ toMap() map[string]any }
		if m, ok := any(t.Asset).(mapper); ok {
			return m.toMap()
		}
		return t.Asset
	}()

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["terms"] = func() any {
		type mapper interface{ toMap() map[string]any }
		if m, ok := any(t.Terms).(mapper); ok {
			return m.toMap()
		}
		return t.Terms
	}()

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["metadata"] = string(t.Metadata)

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["state"] = func() any {
		type mapper interface{ toMap() map[string]any }
		if m, ok := any(t.State).(mapper); ok {
			return m.toMap()
		}
		return t.State
	}()

	return &model.CreateCommand{
		TemplateID: t.GetTemplateID(),
		Arguments:  args,
	}
}

// CreateCommandWithPackageID returns a CreateCommand using the provided package ID instead of package name
func (t EscrowContract) CreateCommandWithPackageID(packageID string) *model.CreateCommand {
	args := make(map[string]any)

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["issuer"] = t.Issuer.ToMap()

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["initiator"] = t.Initiator.ToMap()

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["depositor"] = t.Depositor.ToMap()

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["beneficiary"] = t.Beneficiary.ToMap()

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["mediator"] = t.Mediator.ToMap()

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["contractType"] = string(t.ContractType)

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["asset"] = func() any {
		type mapper interface{ toMap() map[string]any }
		if m, ok := any(t.Asset).(mapper); ok {
			return m.toMap()
		}
		return t.Asset
	}()

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["terms"] = func() any {
		type mapper interface{ toMap() map[string]any }
		if m, ok := any(t.Terms).(mapper); ok {
			return m.toMap()
		}
		return t.Terms
	}()

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["metadata"] = string(t.Metadata)

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["state"] = func() any {
		type mapper interface{ toMap() map[string]any }
		if m, ok := any(t.State).(mapper); ok {
			return m.toMap()
		}
		return t.State
	}()

	return &model.CreateCommand{
		TemplateID: t.GetTemplateIDWithPackageID(packageID),
		Arguments:  args,
	}
}

func (t EscrowContract) MarshalJSON() ([]byte, error) {
	jsonCodec := codec.NewJsonCodec()
	return jsonCodec.Marshal(t)
}

func (t *EscrowContract) UnmarshalJSON(data []byte) error {
	jsonCodec := codec.NewJsonCodec()
	return jsonCodec.Unmarshal(data, t)
}

// Choice methods for EscrowContract

// ConfirmConditions exercises the ConfirmConditions choice on this EscrowContract contract
// This method uses the package name in the template ID
func (t EscrowContract) ConfirmConditions(contractID string, args ConfirmConditions) *model.ExerciseCommand {
	return &model.ExerciseCommand{
		TemplateID: fmt.Sprintf("#%s:%s:%s", PackageName, "StablecoinEscrow", "EscrowContract"),
		ContractID: contractID,
		Choice:     "ConfirmConditions",
		Arguments:  argsToMap(args),
	}
}

// ConfirmConditionsWithPackageID exercises the ConfirmConditions choice using the provided package ID instead of package name
func (t EscrowContract) ConfirmConditionsWithPackageID(contractID string, packageID string, args ConfirmConditions) *model.ExerciseCommand {
	return &model.ExerciseCommand{
		TemplateID: fmt.Sprintf("#%s:%s:%s", packageID, "StablecoinEscrow", "EscrowContract"),
		ContractID: contractID,
		Choice:     "ConfirmConditions",
		Arguments:  argsToMap(args),
	}
}

// RaiseDispute exercises the RaiseDispute choice on this EscrowContract contract
// This method uses the package name in the template ID
func (t EscrowContract) RaiseDispute(contractID string, args RaiseDispute) *model.ExerciseCommand {
	return &model.ExerciseCommand{
		TemplateID: fmt.Sprintf("#%s:%s:%s", PackageName, "StablecoinEscrow", "EscrowContract"),
		ContractID: contractID,
		Choice:     "RaiseDispute",
		Arguments:  argsToMap(args),
	}
}

// RaiseDisputeWithPackageID exercises the RaiseDispute choice using the provided package ID instead of package name
func (t EscrowContract) RaiseDisputeWithPackageID(contractID string, packageID string, args RaiseDispute) *model.ExerciseCommand {
	return &model.ExerciseCommand{
		TemplateID: fmt.Sprintf("#%s:%s:%s", packageID, "StablecoinEscrow", "EscrowContract"),
		ContractID: contractID,
		Choice:     "RaiseDispute",
		Arguments:  argsToMap(args),
	}
}

// Archive exercises the Archive choice on this EscrowContract contract
// This method uses the package name in the template ID
func (t EscrowContract) Archive(contractID string) *model.ExerciseCommand {
	return &model.ExerciseCommand{
		TemplateID: fmt.Sprintf("#%s:%s:%s", PackageName, "StablecoinEscrow", "EscrowContract"),
		ContractID: contractID,
		Choice:     "Archive",
		Arguments:  map[string]any{},
	}
}

// ArchiveWithPackageID exercises the Archive choice using the provided package ID instead of package name
func (t EscrowContract) ArchiveWithPackageID(contractID string, packageID string) *model.ExerciseCommand {
	return &model.ExerciseCommand{
		TemplateID: fmt.Sprintf("#%s:%s:%s", packageID, "StablecoinEscrow", "EscrowContract"),
		ContractID: contractID,
		Choice:     "Archive",
		Arguments:  map[string]any{},
	}
}

// GetStatus exercises the GetStatus choice on this EscrowContract contract via the IEscrow interface
// This method uses the package name in the template ID
func (t EscrowContract) GetStatus(contractID string, args GetStatus) *model.ExerciseCommand {
	return &model.ExerciseCommand{
		TemplateID: fmt.Sprintf("#%s:%s:%s", PackageName, "StablecoinEscrow", "Escrow"),
		ContractID: contractID,
		Choice:     "GetStatus",
		Arguments:  argsToMap(args),
	}
}

// GetStatusWithPackageID exercises the GetStatus choice using the provided package ID instead of package name
func (t EscrowContract) GetStatusWithPackageID(contractID string, packageID string, args GetStatus) *model.ExerciseCommand {
	return &model.ExerciseCommand{
		TemplateID: fmt.Sprintf("#%s:%s:%s", packageID, "StablecoinEscrow", "Escrow"),
		ContractID: contractID,
		Choice:     "GetStatus",
		Arguments:  argsToMap(args),
	}
}

// Activate exercises the Activate choice on this EscrowContract contract via the IEscrow interface
// This method uses the package name in the template ID
func (t EscrowContract) Activate(contractID string, args Activate) *model.ExerciseCommand {
	return &model.ExerciseCommand{
		TemplateID: fmt.Sprintf("#%s:%s:%s", PackageName, "StablecoinEscrow", "Escrow"),
		ContractID: contractID,
		Choice:     "Activate",
		Arguments:  argsToMap(args),
	}
}

// ActivateWithPackageID exercises the Activate choice using the provided package ID instead of package name
func (t EscrowContract) ActivateWithPackageID(contractID string, packageID string, args Activate) *model.ExerciseCommand {
	return &model.ExerciseCommand{
		TemplateID: fmt.Sprintf("#%s:%s:%s", packageID, "StablecoinEscrow", "Escrow"),
		ContractID: contractID,
		Choice:     "Activate",
		Arguments:  argsToMap(args),
	}
}

// Verify interface implementations for EscrowContract

var _ IEscrow = (*EscrowContract)(nil)

// EscrowInvitation is a Template type
type EscrowInvitation struct {
	Issuer       types.PARTY `json:"issuer"`
	Inviter      types.PARTY `json:"inviter"`
	InviteeEmail types.TEXT  `json:"inviteeEmail"`
	InviteeRole  types.TEXT  `json:"inviteeRole"`
	InviteeType  types.TEXT  `json:"inviteeType"`
	ContractType types.TEXT  `json:"contractType"`
	TokenHash    types.TEXT  `json:"tokenHash"`
	Asset        Asset       `json:"asset"`
	Terms        EscrowTerms `json:"terms"`
	Metadata     types.TEXT  `json:"metadata"`
	Mediator     types.PARTY `json:"mediator"`
}

// GetTemplateID returns the template ID for this template using the package name
func (t EscrowInvitation) GetTemplateID() string {
	return fmt.Sprintf("#%s:%s:%s", PackageName, "StablecoinEscrow", "EscrowInvitation")
}

// GetTemplateIDWithPackageID returns the template ID using the provided package ID instead of package name
func (t EscrowInvitation) GetTemplateIDWithPackageID(packageID string) string {
	return fmt.Sprintf("%s:%s:%s", packageID, "StablecoinEscrow", "EscrowInvitation")
}

// CreateCommand returns a CreateCommand for this template using the package name
func (t EscrowInvitation) CreateCommand() *model.CreateCommand {
	args := make(map[string]any)

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["issuer"] = t.Issuer.ToMap()

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["inviter"] = t.Inviter.ToMap()

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["inviteeEmail"] = string(t.InviteeEmail)

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["inviteeRole"] = string(t.InviteeRole)

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["inviteeType"] = string(t.InviteeType)

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["contractType"] = string(t.ContractType)

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["tokenHash"] = string(t.TokenHash)

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["asset"] = func() any {
		type mapper interface{ toMap() map[string]any }
		if m, ok := any(t.Asset).(mapper); ok {
			return m.toMap()
		}
		return t.Asset
	}()

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["terms"] = func() any {
		type mapper interface{ toMap() map[string]any }
		if m, ok := any(t.Terms).(mapper); ok {
			return m.toMap()
		}
		return t.Terms
	}()

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["metadata"] = string(t.Metadata)

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["mediator"] = t.Mediator.ToMap()

	return &model.CreateCommand{
		TemplateID: t.GetTemplateID(),
		Arguments:  args,
	}
}

// CreateCommandWithPackageID returns a CreateCommand using the provided package ID instead of package name
func (t EscrowInvitation) CreateCommandWithPackageID(packageID string) *model.CreateCommand {
	args := make(map[string]any)

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["issuer"] = t.Issuer.ToMap()

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["inviter"] = t.Inviter.ToMap()

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["inviteeEmail"] = string(t.InviteeEmail)

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["inviteeRole"] = string(t.InviteeRole)

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["inviteeType"] = string(t.InviteeType)

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["contractType"] = string(t.ContractType)

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["tokenHash"] = string(t.TokenHash)

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["asset"] = func() any {
		type mapper interface{ toMap() map[string]any }
		if m, ok := any(t.Asset).(mapper); ok {
			return m.toMap()
		}
		return t.Asset
	}()

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["terms"] = func() any {
		type mapper interface{ toMap() map[string]any }
		if m, ok := any(t.Terms).(mapper); ok {
			return m.toMap()
		}
		return t.Terms
	}()

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["metadata"] = string(t.Metadata)

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["mediator"] = t.Mediator.ToMap()

	return &model.CreateCommand{
		TemplateID: t.GetTemplateIDWithPackageID(packageID),
		Arguments:  args,
	}
}

func (t EscrowInvitation) MarshalJSON() ([]byte, error) {
	jsonCodec := codec.NewJsonCodec()
	return jsonCodec.Marshal(t)
}

func (t *EscrowInvitation) UnmarshalJSON(data []byte) error {
	jsonCodec := codec.NewJsonCodec()
	return jsonCodec.Unmarshal(data, t)
}

// Choice methods for EscrowInvitation

// Archive exercises the Archive choice on this EscrowInvitation contract
// This method uses the package name in the template ID
func (t EscrowInvitation) Archive(contractID string) *model.ExerciseCommand {
	return &model.ExerciseCommand{
		TemplateID: fmt.Sprintf("#%s:%s:%s", PackageName, "StablecoinEscrow", "EscrowInvitation"),
		ContractID: contractID,
		Choice:     "Archive",
		Arguments:  map[string]any{},
	}
}

// ArchiveWithPackageID exercises the Archive choice using the provided package ID instead of package name
func (t EscrowInvitation) ArchiveWithPackageID(contractID string, packageID string) *model.ExerciseCommand {
	return &model.ExerciseCommand{
		TemplateID: fmt.Sprintf("#%s:%s:%s", packageID, "StablecoinEscrow", "EscrowInvitation"),
		ContractID: contractID,
		Choice:     "Archive",
		Arguments:  map[string]any{},
	}
}

// ClaimInvitation exercises the ClaimInvitation choice on this EscrowInvitation contract
// This method uses the package name in the template ID
func (t EscrowInvitation) ClaimInvitation(contractID string, args ClaimInvitation) *model.ExerciseCommand {
	return &model.ExerciseCommand{
		TemplateID: fmt.Sprintf("#%s:%s:%s", PackageName, "StablecoinEscrow", "EscrowInvitation"),
		ContractID: contractID,
		Choice:     "ClaimInvitation",
		Arguments:  argsToMap(args),
	}
}

// ClaimInvitationWithPackageID exercises the ClaimInvitation choice using the provided package ID instead of package name
func (t EscrowInvitation) ClaimInvitationWithPackageID(contractID string, packageID string, args ClaimInvitation) *model.ExerciseCommand {
	return &model.ExerciseCommand{
		TemplateID: fmt.Sprintf("#%s:%s:%s", packageID, "StablecoinEscrow", "EscrowInvitation"),
		ContractID: contractID,
		Choice:     "ClaimInvitation",
		Arguments:  argsToMap(args),
	}
}

// EscrowProposal is a Template type
type EscrowProposal struct {
	Issuer       types.PARTY `json:"issuer"`
	Initiator    types.PARTY `json:"initiator"`
	Depositor    types.PARTY `json:"depositor"`
	Beneficiary  types.PARTY `json:"beneficiary"`
	Mediator     types.PARTY `json:"mediator"`
	ContractType types.TEXT  `json:"contractType"`
	Asset        Asset       `json:"asset"`
	Terms        EscrowTerms `json:"terms"`
	Metadata     types.TEXT  `json:"metadata"`
}

// GetTemplateID returns the template ID for this template using the package name
func (t EscrowProposal) GetTemplateID() string {
	return fmt.Sprintf("#%s:%s:%s", PackageName, "StablecoinEscrow", "EscrowProposal")
}

// GetTemplateIDWithPackageID returns the template ID using the provided package ID instead of package name
func (t EscrowProposal) GetTemplateIDWithPackageID(packageID string) string {
	return fmt.Sprintf("%s:%s:%s", packageID, "StablecoinEscrow", "EscrowProposal")
}

// CreateCommand returns a CreateCommand for this template using the package name
func (t EscrowProposal) CreateCommand() *model.CreateCommand {
	args := make(map[string]any)

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["issuer"] = t.Issuer.ToMap()

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["initiator"] = t.Initiator.ToMap()

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["depositor"] = t.Depositor.ToMap()

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["beneficiary"] = t.Beneficiary.ToMap()

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["mediator"] = t.Mediator.ToMap()

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["contractType"] = string(t.ContractType)

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["asset"] = func() any {
		type mapper interface{ toMap() map[string]any }
		if m, ok := any(t.Asset).(mapper); ok {
			return m.toMap()
		}
		return t.Asset
	}()

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["terms"] = func() any {
		type mapper interface{ toMap() map[string]any }
		if m, ok := any(t.Terms).(mapper); ok {
			return m.toMap()
		}
		return t.Terms
	}()

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["metadata"] = string(t.Metadata)

	return &model.CreateCommand{
		TemplateID: t.GetTemplateID(),
		Arguments:  args,
	}
}

// CreateCommandWithPackageID returns a CreateCommand using the provided package ID instead of package name
func (t EscrowProposal) CreateCommandWithPackageID(packageID string) *model.CreateCommand {
	args := make(map[string]any)

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["issuer"] = t.Issuer.ToMap()

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["initiator"] = t.Initiator.ToMap()

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["depositor"] = t.Depositor.ToMap()

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["beneficiary"] = t.Beneficiary.ToMap()

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["mediator"] = t.Mediator.ToMap()

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["contractType"] = string(t.ContractType)

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["asset"] = func() any {
		type mapper interface{ toMap() map[string]any }
		if m, ok := any(t.Asset).(mapper); ok {
			return m.toMap()
		}
		return t.Asset
	}()

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["terms"] = func() any {
		type mapper interface{ toMap() map[string]any }
		if m, ok := any(t.Terms).(mapper); ok {
			return m.toMap()
		}
		return t.Terms
	}()

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["metadata"] = string(t.Metadata)

	return &model.CreateCommand{
		TemplateID: t.GetTemplateIDWithPackageID(packageID),
		Arguments:  args,
	}
}

func (t EscrowProposal) MarshalJSON() ([]byte, error) {
	jsonCodec := codec.NewJsonCodec()
	return jsonCodec.Marshal(t)
}

func (t *EscrowProposal) UnmarshalJSON(data []byte) error {
	jsonCodec := codec.NewJsonCodec()
	return jsonCodec.Unmarshal(data, t)
}

// Choice methods for EscrowProposal

// BeneficiaryAccept exercises the BeneficiaryAccept choice on this EscrowProposal contract
// This method uses the package name in the template ID
func (t EscrowProposal) BeneficiaryAccept(contractID string, args BeneficiaryAccept) *model.ExerciseCommand {
	return &model.ExerciseCommand{
		TemplateID: fmt.Sprintf("#%s:%s:%s", PackageName, "StablecoinEscrow", "EscrowProposal"),
		ContractID: contractID,
		Choice:     "BeneficiaryAccept",
		Arguments:  argsToMap(args),
	}
}

// BeneficiaryAcceptWithPackageID exercises the BeneficiaryAccept choice using the provided package ID instead of package name
func (t EscrowProposal) BeneficiaryAcceptWithPackageID(contractID string, packageID string, args BeneficiaryAccept) *model.ExerciseCommand {
	return &model.ExerciseCommand{
		TemplateID: fmt.Sprintf("#%s:%s:%s", packageID, "StablecoinEscrow", "EscrowProposal"),
		ContractID: contractID,
		Choice:     "BeneficiaryAccept",
		Arguments:  argsToMap(args),
	}
}

// RequestChanges exercises the RequestChanges choice on this EscrowProposal contract
// This method uses the package name in the template ID
func (t EscrowProposal) RequestChanges(contractID string, args RequestChanges) *model.ExerciseCommand {
	return &model.ExerciseCommand{
		TemplateID: fmt.Sprintf("#%s:%s:%s", PackageName, "StablecoinEscrow", "EscrowProposal"),
		ContractID: contractID,
		Choice:     "RequestChanges",
		Arguments:  argsToMap(args),
	}
}

// RequestChangesWithPackageID exercises the RequestChanges choice using the provided package ID instead of package name
func (t EscrowProposal) RequestChangesWithPackageID(contractID string, packageID string, args RequestChanges) *model.ExerciseCommand {
	return &model.ExerciseCommand{
		TemplateID: fmt.Sprintf("#%s:%s:%s", packageID, "StablecoinEscrow", "EscrowProposal"),
		ContractID: contractID,
		Choice:     "RequestChanges",
		Arguments:  argsToMap(args),
	}
}

// CancelProposal exercises the CancelProposal choice on this EscrowProposal contract
// This method uses the package name in the template ID
func (t EscrowProposal) CancelProposal(contractID string, args CancelProposal) *model.ExerciseCommand {
	return &model.ExerciseCommand{
		TemplateID: fmt.Sprintf("#%s:%s:%s", PackageName, "StablecoinEscrow", "EscrowProposal"),
		ContractID: contractID,
		Choice:     "CancelProposal",
		Arguments:  argsToMap(args),
	}
}

// CancelProposalWithPackageID exercises the CancelProposal choice using the provided package ID instead of package name
func (t EscrowProposal) CancelProposalWithPackageID(contractID string, packageID string, args CancelProposal) *model.ExerciseCommand {
	return &model.ExerciseCommand{
		TemplateID: fmt.Sprintf("#%s:%s:%s", packageID, "StablecoinEscrow", "EscrowProposal"),
		ContractID: contractID,
		Choice:     "CancelProposal",
		Arguments:  argsToMap(args),
	}
}

// WithdrawProposal exercises the WithdrawProposal choice on this EscrowProposal contract
// This method uses the package name in the template ID
func (t EscrowProposal) WithdrawProposal(contractID string, args WithdrawProposal) *model.ExerciseCommand {
	return &model.ExerciseCommand{
		TemplateID: fmt.Sprintf("#%s:%s:%s", PackageName, "StablecoinEscrow", "EscrowProposal"),
		ContractID: contractID,
		Choice:     "WithdrawProposal",
		Arguments:  argsToMap(args),
	}
}

// WithdrawProposalWithPackageID exercises the WithdrawProposal choice using the provided package ID instead of package name
func (t EscrowProposal) WithdrawProposalWithPackageID(contractID string, packageID string, args WithdrawProposal) *model.ExerciseCommand {
	return &model.ExerciseCommand{
		TemplateID: fmt.Sprintf("#%s:%s:%s", packageID, "StablecoinEscrow", "EscrowProposal"),
		ContractID: contractID,
		Choice:     "WithdrawProposal",
		Arguments:  argsToMap(args),
	}
}

// Archive exercises the Archive choice on this EscrowProposal contract
// This method uses the package name in the template ID
func (t EscrowProposal) Archive(contractID string) *model.ExerciseCommand {
	return &model.ExerciseCommand{
		TemplateID: fmt.Sprintf("#%s:%s:%s", PackageName, "StablecoinEscrow", "EscrowProposal"),
		ContractID: contractID,
		Choice:     "Archive",
		Arguments:  map[string]any{},
	}
}

// ArchiveWithPackageID exercises the Archive choice using the provided package ID instead of package name
func (t EscrowProposal) ArchiveWithPackageID(contractID string, packageID string) *model.ExerciseCommand {
	return &model.ExerciseCommand{
		TemplateID: fmt.Sprintf("#%s:%s:%s", packageID, "StablecoinEscrow", "EscrowProposal"),
		ContractID: contractID,
		Choice:     "Archive",
		Arguments:  map[string]any{},
	}
}

// GetStatus exercises the GetStatus choice on this EscrowProposal contract via the IEscrow interface
// This method uses the package name in the template ID
func (t EscrowProposal) GetStatus(contractID string, args GetStatus) *model.ExerciseCommand {
	return &model.ExerciseCommand{
		TemplateID: fmt.Sprintf("#%s:%s:%s", PackageName, "StablecoinEscrow", "Escrow"),
		ContractID: contractID,
		Choice:     "GetStatus",
		Arguments:  argsToMap(args),
	}
}

// GetStatusWithPackageID exercises the GetStatus choice using the provided package ID instead of package name
func (t EscrowProposal) GetStatusWithPackageID(contractID string, packageID string, args GetStatus) *model.ExerciseCommand {
	return &model.ExerciseCommand{
		TemplateID: fmt.Sprintf("#%s:%s:%s", packageID, "StablecoinEscrow", "Escrow"),
		ContractID: contractID,
		Choice:     "GetStatus",
		Arguments:  argsToMap(args),
	}
}

// Activate exercises the Activate choice on this EscrowProposal contract via the IEscrow interface
// This method uses the package name in the template ID
func (t EscrowProposal) Activate(contractID string, args Activate) *model.ExerciseCommand {
	return &model.ExerciseCommand{
		TemplateID: fmt.Sprintf("#%s:%s:%s", PackageName, "StablecoinEscrow", "Escrow"),
		ContractID: contractID,
		Choice:     "Activate",
		Arguments:  argsToMap(args),
	}
}

// ActivateWithPackageID exercises the Activate choice using the provided package ID instead of package name
func (t EscrowProposal) ActivateWithPackageID(contractID string, packageID string, args Activate) *model.ExerciseCommand {
	return &model.ExerciseCommand{
		TemplateID: fmt.Sprintf("#%s:%s:%s", packageID, "StablecoinEscrow", "Escrow"),
		ContractID: contractID,
		Choice:     "Activate",
		Arguments:  argsToMap(args),
	}
}

// Verify interface implementations for EscrowProposal

var _ IEscrow = (*EscrowProposal)(nil)

// FinalizeSettlement is a Record type
type FinalizeSettlement struct {
}

// ToMap converts FinalizeSettlement to a map for DAML arguments
func (t FinalizeSettlement) ToMap() map[string]any {
	m := make(map[string]any)
	return m
}

func (t FinalizeSettlement) MarshalJSON() ([]byte, error) {
	jsonCodec := codec.NewJsonCodec()
	return jsonCodec.Marshal(t)
}

func (t *FinalizeSettlement) UnmarshalJSON(data []byte) error {
	jsonCodec := codec.NewJsonCodec()
	return jsonCodec.Unmarshal(data, t)
}

// Fund is a Record type
type Fund struct {
	CustodyRef types.TEXT        `json:"custodyRef"`
	HoldingCid types.CONTRACT_ID `json:"holdingCid"`
}

// ToMap converts Fund to a map for DAML arguments
func (t Fund) ToMap() map[string]any {
	m := make(map[string]any)

	m["custodyRef"] = string(t.CustodyRef)

	m["holdingCid"] = func() any {
		type mapper interface{ toMap() map[string]any }
		if m, ok := any(t.HoldingCid).(mapper); ok {
			return m.toMap()
		}
		return t.HoldingCid
	}()

	return m
}

func (t Fund) MarshalJSON() ([]byte, error) {
	jsonCodec := codec.NewJsonCodec()
	return jsonCodec.Marshal(t)
}

func (t *Fund) UnmarshalJSON(data []byte) error {
	jsonCodec := codec.NewJsonCodec()
	return jsonCodec.Unmarshal(data, t)
}

// ProposeSettlement is a Record type
type ProposeSettlement struct {
	Proposal SettlementTerms `json:"proposal"`
}

// ToMap converts ProposeSettlement to a map for DAML arguments
func (t ProposeSettlement) ToMap() map[string]any {
	m := make(map[string]any)

	m["proposal"] = func() any {
		type mapper interface{ toMap() map[string]any }
		if m, ok := any(t.Proposal).(mapper); ok {
			return m.toMap()
		}
		return t.Proposal
	}()

	return m
}

func (t ProposeSettlement) MarshalJSON() ([]byte, error) {
	jsonCodec := codec.NewJsonCodec()
	return jsonCodec.Marshal(t)
}

func (t *ProposeSettlement) UnmarshalJSON(data []byte) error {
	jsonCodec := codec.NewJsonCodec()
	return jsonCodec.Unmarshal(data, t)
}

// RaiseDispute is a Record type
type RaiseDispute struct {
}

// ToMap converts RaiseDispute to a map for DAML arguments
func (t RaiseDispute) ToMap() map[string]any {
	m := make(map[string]any)
	return m
}

func (t RaiseDispute) MarshalJSON() ([]byte, error) {
	jsonCodec := codec.NewJsonCodec()
	return jsonCodec.Marshal(t)
}

func (t *RaiseDispute) UnmarshalJSON(data []byte) error {
	jsonCodec := codec.NewJsonCodec()
	return jsonCodec.Unmarshal(data, t)
}

// RatifySettlement is a Record type
type RatifySettlement struct {
	Actor types.PARTY `json:"actor"`
}

// ToMap converts RatifySettlement to a map for DAML arguments
func (t RatifySettlement) ToMap() map[string]any {
	m := make(map[string]any)

	m["actor"] = t.Actor.ToMap()

	return m
}

func (t RatifySettlement) MarshalJSON() ([]byte, error) {
	jsonCodec := codec.NewJsonCodec()
	return jsonCodec.Marshal(t)
}

func (t *RatifySettlement) UnmarshalJSON(data []byte) error {
	jsonCodec := codec.NewJsonCodec()
	return jsonCodec.Unmarshal(data, t)
}

// RejectSettlement is a Record type
type RejectSettlement struct {
}

// ToMap converts RejectSettlement to a map for DAML arguments
func (t RejectSettlement) ToMap() map[string]any {
	m := make(map[string]any)
	return m
}

func (t RejectSettlement) MarshalJSON() ([]byte, error) {
	jsonCodec := codec.NewJsonCodec()
	return jsonCodec.Marshal(t)
}

func (t *RejectSettlement) UnmarshalJSON(data []byte) error {
	jsonCodec := codec.NewJsonCodec()
	return jsonCodec.Unmarshal(data, t)
}

// RequestChanges is a Record type
type RequestChanges struct {
	NewTerms       EscrowTerms `json:"newTerms"`
	ProposingParty types.PARTY `json:"proposingParty"`
}

// ToMap converts RequestChanges to a map for DAML arguments
func (t RequestChanges) ToMap() map[string]any {
	m := make(map[string]any)

	m["newTerms"] = func() any {
		type mapper interface{ toMap() map[string]any }
		if m, ok := any(t.NewTerms).(mapper); ok {
			return m.toMap()
		}
		return t.NewTerms
	}()

	m["proposingParty"] = t.ProposingParty.ToMap()

	return m
}

func (t RequestChanges) MarshalJSON() ([]byte, error) {
	jsonCodec := codec.NewJsonCodec()
	return jsonCodec.Marshal(t)
}

func (t *RequestChanges) UnmarshalJSON(data []byte) error {
	jsonCodec := codec.NewJsonCodec()
	return jsonCodec.Unmarshal(data, t)
}

// SettlementRecord is a Template type
type SettlementRecord struct {
	Issuer              types.PARTY     `json:"issuer"`
	Initiator           types.PARTY     `json:"initiator"`
	Depositor           types.PARTY     `json:"depositor"`
	Beneficiary         types.PARTY     `json:"beneficiary"`
	Mediator            types.PARTY     `json:"mediator"`
	ContractType        types.TEXT      `json:"contractType"`
	Asset               Asset           `json:"asset"`
	Terms               EscrowTerms     `json:"terms"`
	Metadata            types.TEXT      `json:"metadata"`
	Settlement          SettlementTerms `json:"settlement"`
	DepositorAccepted   types.BOOL      `json:"depositorAccepted"`
	BeneficiaryAccepted types.BOOL      `json:"beneficiaryAccepted"`
}

// GetTemplateID returns the template ID for this template using the package name
func (t SettlementRecord) GetTemplateID() string {
	return fmt.Sprintf("#%s:%s:%s", PackageName, "StablecoinEscrow", "SettlementRecord")
}

// GetTemplateIDWithPackageID returns the template ID using the provided package ID instead of package name
func (t SettlementRecord) GetTemplateIDWithPackageID(packageID string) string {
	return fmt.Sprintf("%s:%s:%s", packageID, "StablecoinEscrow", "SettlementRecord")
}

// CreateCommand returns a CreateCommand for this template using the package name
func (t SettlementRecord) CreateCommand() *model.CreateCommand {
	args := make(map[string]any)

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["issuer"] = t.Issuer.ToMap()

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["initiator"] = t.Initiator.ToMap()

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["depositor"] = t.Depositor.ToMap()

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["beneficiary"] = t.Beneficiary.ToMap()

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["mediator"] = t.Mediator.ToMap()

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["contractType"] = string(t.ContractType)

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["asset"] = func() any {
		type mapper interface{ toMap() map[string]any }
		if m, ok := any(t.Asset).(mapper); ok {
			return m.toMap()
		}
		return t.Asset
	}()

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["terms"] = func() any {
		type mapper interface{ toMap() map[string]any }
		if m, ok := any(t.Terms).(mapper); ok {
			return m.toMap()
		}
		return t.Terms
	}()

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["metadata"] = string(t.Metadata)

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["settlement"] = func() any {
		type mapper interface{ toMap() map[string]any }
		if m, ok := any(t.Settlement).(mapper); ok {
			return m.toMap()
		}
		return t.Settlement
	}()

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["depositorAccepted"] = bool(t.DepositorAccepted)

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["beneficiaryAccepted"] = bool(t.BeneficiaryAccepted)

	return &model.CreateCommand{
		TemplateID: t.GetTemplateID(),
		Arguments:  args,
	}
}

// CreateCommandWithPackageID returns a CreateCommand using the provided package ID instead of package name
func (t SettlementRecord) CreateCommandWithPackageID(packageID string) *model.CreateCommand {
	args := make(map[string]any)

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["issuer"] = t.Issuer.ToMap()

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["initiator"] = t.Initiator.ToMap()

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["depositor"] = t.Depositor.ToMap()

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["beneficiary"] = t.Beneficiary.ToMap()

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["mediator"] = t.Mediator.ToMap()

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["contractType"] = string(t.ContractType)

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["asset"] = func() any {
		type mapper interface{ toMap() map[string]any }
		if m, ok := any(t.Asset).(mapper); ok {
			return m.toMap()
		}
		return t.Asset
	}()

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["terms"] = func() any {
		type mapper interface{ toMap() map[string]any }
		if m, ok := any(t.Terms).(mapper); ok {
			return m.toMap()
		}
		return t.Terms
	}()

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["metadata"] = string(t.Metadata)

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["settlement"] = func() any {
		type mapper interface{ toMap() map[string]any }
		if m, ok := any(t.Settlement).(mapper); ok {
			return m.toMap()
		}
		return t.Settlement
	}()

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["depositorAccepted"] = bool(t.DepositorAccepted)

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["beneficiaryAccepted"] = bool(t.BeneficiaryAccepted)

	return &model.CreateCommand{
		TemplateID: t.GetTemplateIDWithPackageID(packageID),
		Arguments:  args,
	}
}

func (t SettlementRecord) MarshalJSON() ([]byte, error) {
	jsonCodec := codec.NewJsonCodec()
	return jsonCodec.Marshal(t)
}

func (t *SettlementRecord) UnmarshalJSON(data []byte) error {
	jsonCodec := codec.NewJsonCodec()
	return jsonCodec.Unmarshal(data, t)
}

// Choice methods for SettlementRecord

// RatifySettlement exercises the RatifySettlement choice on this SettlementRecord contract
// This method uses the package name in the template ID
func (t SettlementRecord) RatifySettlement(contractID string, args RatifySettlement) *model.ExerciseCommand {
	return &model.ExerciseCommand{
		TemplateID: fmt.Sprintf("#%s:%s:%s", PackageName, "StablecoinEscrow", "SettlementRecord"),
		ContractID: contractID,
		Choice:     "RatifySettlement",
		Arguments:  argsToMap(args),
	}
}

// RatifySettlementWithPackageID exercises the RatifySettlement choice using the provided package ID instead of package name
func (t SettlementRecord) RatifySettlementWithPackageID(contractID string, packageID string, args RatifySettlement) *model.ExerciseCommand {
	return &model.ExerciseCommand{
		TemplateID: fmt.Sprintf("#%s:%s:%s", packageID, "StablecoinEscrow", "SettlementRecord"),
		ContractID: contractID,
		Choice:     "RatifySettlement",
		Arguments:  argsToMap(args),
	}
}

// FinalizeSettlement exercises the FinalizeSettlement choice on this SettlementRecord contract
// This method uses the package name in the template ID
func (t SettlementRecord) FinalizeSettlement(contractID string, args FinalizeSettlement) *model.ExerciseCommand {
	return &model.ExerciseCommand{
		TemplateID: fmt.Sprintf("#%s:%s:%s", PackageName, "StablecoinEscrow", "SettlementRecord"),
		ContractID: contractID,
		Choice:     "FinalizeSettlement",
		Arguments:  argsToMap(args),
	}
}

// FinalizeSettlementWithPackageID exercises the FinalizeSettlement choice using the provided package ID instead of package name
func (t SettlementRecord) FinalizeSettlementWithPackageID(contractID string, packageID string, args FinalizeSettlement) *model.ExerciseCommand {
	return &model.ExerciseCommand{
		TemplateID: fmt.Sprintf("#%s:%s:%s", packageID, "StablecoinEscrow", "SettlementRecord"),
		ContractID: contractID,
		Choice:     "FinalizeSettlement",
		Arguments:  argsToMap(args),
	}
}

// RejectSettlement exercises the RejectSettlement choice on this SettlementRecord contract
// This method uses the package name in the template ID
func (t SettlementRecord) RejectSettlement(contractID string, args RejectSettlement) *model.ExerciseCommand {
	return &model.ExerciseCommand{
		TemplateID: fmt.Sprintf("#%s:%s:%s", PackageName, "StablecoinEscrow", "SettlementRecord"),
		ContractID: contractID,
		Choice:     "RejectSettlement",
		Arguments:  argsToMap(args),
	}
}

// RejectSettlementWithPackageID exercises the RejectSettlement choice using the provided package ID instead of package name
func (t SettlementRecord) RejectSettlementWithPackageID(contractID string, packageID string, args RejectSettlement) *model.ExerciseCommand {
	return &model.ExerciseCommand{
		TemplateID: fmt.Sprintf("#%s:%s:%s", packageID, "StablecoinEscrow", "SettlementRecord"),
		ContractID: contractID,
		Choice:     "RejectSettlement",
		Arguments:  argsToMap(args),
	}
}

// Archive exercises the Archive choice on this SettlementRecord contract
// This method uses the package name in the template ID
func (t SettlementRecord) Archive(contractID string) *model.ExerciseCommand {
	return &model.ExerciseCommand{
		TemplateID: fmt.Sprintf("#%s:%s:%s", PackageName, "StablecoinEscrow", "SettlementRecord"),
		ContractID: contractID,
		Choice:     "Archive",
		Arguments:  map[string]any{},
	}
}

// ArchiveWithPackageID exercises the Archive choice using the provided package ID instead of package name
func (t SettlementRecord) ArchiveWithPackageID(contractID string, packageID string) *model.ExerciseCommand {
	return &model.ExerciseCommand{
		TemplateID: fmt.Sprintf("#%s:%s:%s", packageID, "StablecoinEscrow", "SettlementRecord"),
		ContractID: contractID,
		Choice:     "Archive",
		Arguments:  map[string]any{},
	}
}

// GetStatus exercises the GetStatus choice on this SettlementRecord contract via the IEscrow interface
// This method uses the package name in the template ID
func (t SettlementRecord) GetStatus(contractID string, args GetStatus) *model.ExerciseCommand {
	return &model.ExerciseCommand{
		TemplateID: fmt.Sprintf("#%s:%s:%s", PackageName, "StablecoinEscrow", "Escrow"),
		ContractID: contractID,
		Choice:     "GetStatus",
		Arguments:  argsToMap(args),
	}
}

// GetStatusWithPackageID exercises the GetStatus choice using the provided package ID instead of package name
func (t SettlementRecord) GetStatusWithPackageID(contractID string, packageID string, args GetStatus) *model.ExerciseCommand {
	return &model.ExerciseCommand{
		TemplateID: fmt.Sprintf("#%s:%s:%s", packageID, "StablecoinEscrow", "Escrow"),
		ContractID: contractID,
		Choice:     "GetStatus",
		Arguments:  argsToMap(args),
	}
}

// Activate exercises the Activate choice on this SettlementRecord contract via the IEscrow interface
// This method uses the package name in the template ID
func (t SettlementRecord) Activate(contractID string, args Activate) *model.ExerciseCommand {
	return &model.ExerciseCommand{
		TemplateID: fmt.Sprintf("#%s:%s:%s", PackageName, "StablecoinEscrow", "Escrow"),
		ContractID: contractID,
		Choice:     "Activate",
		Arguments:  argsToMap(args),
	}
}

// ActivateWithPackageID exercises the Activate choice using the provided package ID instead of package name
func (t SettlementRecord) ActivateWithPackageID(contractID string, packageID string, args Activate) *model.ExerciseCommand {
	return &model.ExerciseCommand{
		TemplateID: fmt.Sprintf("#%s:%s:%s", packageID, "StablecoinEscrow", "Escrow"),
		ContractID: contractID,
		Choice:     "Activate",
		Arguments:  argsToMap(args),
	}
}

// Verify interface implementations for SettlementRecord

var _ IEscrow = (*SettlementRecord)(nil)

// WithdrawProposal is a Record type
type WithdrawProposal struct {
}

// ToMap converts WithdrawProposal to a map for DAML arguments
func (t WithdrawProposal) ToMap() map[string]any {
	m := make(map[string]any)
	return m
}

func (t WithdrawProposal) MarshalJSON() ([]byte, error) {
	jsonCodec := codec.NewJsonCodec()
	return jsonCodec.Marshal(t)
}

func (t *WithdrawProposal) UnmarshalJSON(data []byte) error {
	jsonCodec := codec.NewJsonCodec()
	return jsonCodec.Unmarshal(data, t)
}
