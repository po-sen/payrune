package usecases

import (
	"context"
	"errors"
	"time"

	applicationoutbox "payrune/internal/application/outbox"
	outport "payrune/internal/application/ports/outbound"
	"payrune/internal/domain/entities"
	"payrune/internal/domain/events"
	"payrune/internal/domain/policies"
	"payrune/internal/domain/valueobjects"
)

type inMemoryAddressPolicyReader struct {
	ordered      []outport.AddressPolicyRecord
	issuanceByID map[valueobjects.AddressPolicyID]policies.AddressIssuancePolicy
	listErr      error
	findErr      error
}

var _ outport.AddressPolicyReader = (*inMemoryAddressPolicyReader)(nil)

func newAddressIssuancePolicy(
	addressPolicyID string,
	chain valueobjects.SupportedChain,
	network valueobjects.NetworkID,
	scheme string,
	minorUnit string,
	decimals uint8,
	accountPublicKey string,
	derivationPathPrefix string,
) policies.AddressIssuancePolicy {
	return policies.AddressIssuancePolicy{
		AddressPolicyID: valueobjects.AddressPolicyID(addressPolicyID),
		Chain:           chain,
		Network:         network,
		Scheme:          valueobjects.AddressScheme(scheme),
		MinorUnit:       minorUnit,
		Decimals:        decimals,
		IssuanceConfig: valueobjects.AddressIssuanceConfig{
			AddressSpaceRef:   accountPublicKey,
			IssuanceRefPrefix: derivationPathPrefix,
		},
	}.Normalize()
}

func newEthereumCreate2IssuancePolicy(
	addressPolicyID string,
	network valueobjects.NetworkID,
	addressSourceRef string,
	addressReferencePrefix string,
) policies.AddressIssuancePolicy {
	return newAddressIssuancePolicy(
		addressPolicyID,
		valueobjects.SupportedChainEthereum,
		network,
		"create2",
		"wei",
		18,
		addressSourceRef,
		addressReferencePrefix,
	)
}

func newInMemoryAddressPolicyReader(issuancePolicies []policies.AddressIssuancePolicy) *inMemoryAddressPolicyReader {
	ordered := make([]outport.AddressPolicyRecord, 0, len(issuancePolicies))
	issuanceByID := make(map[valueobjects.AddressPolicyID]policies.AddressIssuancePolicy, len(issuancePolicies))

	for _, policy := range issuancePolicies {
		normalized := policy.Normalize()
		if normalized.AddressPolicyID.IsZero() {
			continue
		}
		if _, exists := issuanceByID[normalized.AddressPolicyID]; exists {
			continue
		}
		ordered = append(ordered, outport.AddressPolicyRecord{
			AddressPolicyID: normalized.AddressPolicyID,
			Chain:           normalized.Chain,
			Network:         normalized.Network,
			Scheme:          normalized.Scheme,
			MinorUnit:       normalized.MinorUnit,
			Decimals:        normalized.Decimals,
			Enabled:         normalized.Enabled,
		})
		issuanceByID[normalized.AddressPolicyID] = normalized
	}

	return &inMemoryAddressPolicyReader{
		ordered:      ordered,
		issuanceByID: issuanceByID,
	}
}

func (r *inMemoryAddressPolicyReader) ListByChain(
	_ context.Context,
	chain valueobjects.SupportedChain,
) ([]outport.AddressPolicyRecord, error) {
	if r.listErr != nil {
		return nil, r.listErr
	}
	policies := make([]outport.AddressPolicyRecord, 0)
	for _, policy := range r.ordered {
		if policy.Chain != chain {
			continue
		}
		policies = append(policies, policy)
	}
	return policies, nil
}

func (r *inMemoryAddressPolicyReader) FindIssuanceByID(
	_ context.Context,
	addressPolicyID valueobjects.AddressPolicyID,
) (policies.AddressIssuancePolicy, bool, error) {
	if r.findErr != nil {
		return policies.AddressIssuancePolicy{}, false, r.findErr
	}
	policy, ok := r.issuanceByID[addressPolicyID]
	if !ok {
		return policies.AddressIssuancePolicy{}, false, nil
	}
	return policy, true, nil
}

type fakeChainAddressDeriver struct {
	supportedChains map[valueobjects.SupportedChain]bool
	output          outport.DeriveChainAddressOutput
	err             error
	lastInput       outport.DeriveChainAddressInput
	calls           int
}

type fakeIssuedPaymentAddressDeriver struct {
	supportedChains map[valueobjects.SupportedChain]bool
	output          outport.DeriveIssuedPaymentAddressOutput
	err             error
	lastInput       outport.DeriveIssuedPaymentAddressInput
	calls           int
}

func newFakeIssuedPaymentAddressDeriver() *fakeIssuedPaymentAddressDeriver {
	return &fakeIssuedPaymentAddressDeriver{
		supportedChains: map[valueobjects.SupportedChain]bool{
			valueobjects.SupportedChainBitcoin: true,
		},
		output: outport.DeriveIssuedPaymentAddressOutput{
			Address:           "bc1qdefault",
			SweepMaterialJSON: `{"material_type":"bitcoin_hd"}`,
			IssuanceRefKind:   valueobjects.IssuanceRefKindHDPathAbsolute,
			IssuanceRef:       "0/0",
		},
	}
}

func (f *fakeIssuedPaymentAddressDeriver) SupportsChain(chain valueobjects.SupportedChain) bool {
	return f.supportedChains[chain]
}

func (f *fakeIssuedPaymentAddressDeriver) DeriveIssuedAddress(
	_ context.Context,
	input outport.DeriveIssuedPaymentAddressInput,
) (outport.DeriveIssuedPaymentAddressOutput, error) {
	f.calls++
	f.lastInput = input
	if f.err != nil {
		return outport.DeriveIssuedPaymentAddressOutput{}, f.err
	}
	return f.output, nil
}

func newFakeChainAddressDeriver() *fakeChainAddressDeriver {
	return &fakeChainAddressDeriver{
		supportedChains: map[valueobjects.SupportedChain]bool{
			valueobjects.SupportedChainBitcoin: true,
		},
		output: outport.DeriveChainAddressOutput{
			Address:             "bc1qdefault",
			IssuanceRefKind:     valueobjects.IssuanceRefKindHDPathAbsolute,
			RelativeIssuanceRef: "0/0",
		},
	}
}

func (f *fakeChainAddressDeriver) SupportsChain(chain valueobjects.SupportedChain) bool {
	return f.supportedChains[chain]
}

func (f *fakeChainAddressDeriver) DeriveAddress(
	_ context.Context,
	input outport.DeriveChainAddressInput,
) (outport.DeriveChainAddressOutput, error) {
	f.calls++
	f.lastInput = input
	if f.err != nil {
		return outport.DeriveChainAddressOutput{}, f.err
	}
	return f.output, nil
}

type fakePaymentAddressStatusFinder struct {
	record    outport.PaymentAddressStatusRecord
	found     bool
	err       error
	lastInput outport.FindPaymentAddressStatusInput
	calls     int
}

func (f *fakePaymentAddressStatusFinder) FindByID(
	_ context.Context,
	input outport.FindPaymentAddressStatusInput,
) (outport.PaymentAddressStatusRecord, bool, error) {
	f.calls++
	f.lastInput = input
	if f.err != nil {
		return outport.PaymentAddressStatusRecord{}, false, f.err
	}
	if f.found {
		return f.record, true, nil
	}
	return outport.PaymentAddressStatusRecord{}, false, nil
}

type fakePaymentAddressAllocationStore struct {
	findIssuedByIDResults []fakeFindIssuedPaymentAddressAllocationResult
	issuedByID            entities.PaymentAddressAllocation
	issuedByIDFound       bool
	issuedByIDErr         error
	reopenReservation     entities.PaymentAddressAllocation
	reopenFound           bool
	reopenErr             error
	freshReservation      entities.PaymentAddressAllocation
	reserveFreshErr       error
	completeErr           error
	markFailedErr         error
	lastFindIssuedByID    outport.FindIssuedPaymentAddressAllocationByIDInput
	lastReopenInput       outport.ReservePaymentAddressAllocationInput
	lastReserveFreshInput outport.ReservePaymentAddressAllocationInput
	lastCompleteInput     entities.PaymentAddressAllocation
	lastCompleteIssuedAt  time.Time
	lastFailedInput       entities.PaymentAddressAllocation
	findIssuedByIDCalls   int
	reopenCalls           int
	reserveFreshCalls     int
	completeCalls         int
	markFailedCalls       int
}

type fakeFindIssuedPaymentAddressAllocationResult struct {
	allocation entities.PaymentAddressAllocation
	found      bool
	err        error
}

func (f *fakePaymentAddressAllocationStore) FindIssuedByID(
	_ context.Context,
	input outport.FindIssuedPaymentAddressAllocationByIDInput,
) (entities.PaymentAddressAllocation, bool, error) {
	f.findIssuedByIDCalls++
	f.lastFindIssuedByID = input
	if len(f.findIssuedByIDResults) >= f.findIssuedByIDCalls {
		result := f.findIssuedByIDResults[f.findIssuedByIDCalls-1]
		return result.allocation, result.found, result.err
	}
	if f.issuedByIDErr != nil {
		return entities.PaymentAddressAllocation{}, false, f.issuedByIDErr
	}
	if f.issuedByIDFound {
		return f.issuedByID, true, nil
	}
	return entities.PaymentAddressAllocation{}, false, nil
}

func (f *fakePaymentAddressAllocationStore) ReopenFailedReservation(
	_ context.Context,
	input outport.ReservePaymentAddressAllocationInput,
) (entities.PaymentAddressAllocation, bool, error) {
	f.reopenCalls++
	f.lastReopenInput = input
	if f.reopenErr != nil {
		return entities.PaymentAddressAllocation{}, false, f.reopenErr
	}
	if f.reopenFound {
		return f.reopenReservation, true, nil
	}
	return entities.PaymentAddressAllocation{}, false, nil
}

func (f *fakePaymentAddressAllocationStore) ReserveFresh(
	_ context.Context,
	input outport.ReservePaymentAddressAllocationInput,
) (entities.PaymentAddressAllocation, error) {
	f.reserveFreshCalls++
	f.lastReserveFreshInput = input
	if f.reserveFreshErr != nil {
		return entities.PaymentAddressAllocation{}, f.reserveFreshErr
	}
	return f.freshReservation, nil
}

func (f *fakePaymentAddressAllocationStore) Complete(
	_ context.Context,
	input entities.PaymentAddressAllocation,
	issuedAt time.Time,
) error {
	f.completeCalls++
	f.lastCompleteInput = input
	f.lastCompleteIssuedAt = issuedAt
	return f.completeErr
}

func (f *fakePaymentAddressAllocationStore) MarkDerivationFailed(
	_ context.Context,
	input entities.PaymentAddressAllocation,
) error {
	f.markFailedCalls++
	f.lastFailedInput = input
	return f.markFailedErr
}

type fakeUnitOfWork struct {
	err                  error
	calls                int
	allocationStore      outport.PaymentAddressAllocationStore
	idempotencyStore     outport.PaymentAddressIdempotencyStore
	receiptTrackingStore outport.PaymentReceiptTrackingStore
	notificationOutbox   outport.PaymentReceiptStatusNotificationOutbox
}

func newFakeUnitOfWork(store outport.PaymentAddressAllocationStore) *fakeUnitOfWork {
	return &fakeUnitOfWork{
		allocationStore:      store,
		idempotencyStore:     &fakePaymentAddressIdempotencyStore{},
		receiptTrackingStore: &fakeAllocatePaymentReceiptTrackingStore{},
		notificationOutbox:   &fakeAllocatePaymentReceiptStatusNotificationOutbox{},
	}
}

func (f *fakeUnitOfWork) WithinTransaction(
	_ context.Context,
	fn func(txScope outport.TxScope) error,
) error {
	f.calls++
	if f.err != nil {
		return f.err
	}
	if f.allocationStore == nil {
		return errors.New("payment address allocation store is not configured")
	}
	if f.idempotencyStore == nil {
		return errors.New("payment address idempotency store is not configured")
	}
	if f.receiptTrackingStore == nil {
		return errors.New("payment receipt tracking store is not configured")
	}
	return fn(outport.TxScope{
		PaymentAddressAllocation:               f.allocationStore,
		PaymentAddressIdempotency:              f.idempotencyStore,
		PaymentReceiptTracking:                 f.receiptTrackingStore,
		PaymentReceiptStatusNotificationOutbox: f.notificationOutbox,
	})
}

type fakePaymentAddressIdempotencyStore struct {
	findByKeyResults []fakeFindPaymentAddressIdempotencyResult
	record           outport.PaymentAddressIdempotencyRecord
	found            bool
	findErr          error
	claimErr         error
	releaseErr       error
	completeErr      error
	lastFindByKey    outport.FindPaymentAddressIdempotencyInput
	lastClaim        outport.ClaimPaymentAddressIdempotencyInput
	lastRelease      outport.ReleasePaymentAddressIdempotencyInput
	lastComplete     outport.CompletePaymentAddressIdempotencyInput
	findCalls        int
	claimCalls       int
	releaseCalls     int
	completeCalls    int
}

type fakeFindPaymentAddressIdempotencyResult struct {
	record outport.PaymentAddressIdempotencyRecord
	found  bool
	err    error
}

func (f *fakePaymentAddressIdempotencyStore) FindByKey(
	_ context.Context,
	input outport.FindPaymentAddressIdempotencyInput,
) (outport.PaymentAddressIdempotencyRecord, bool, error) {
	f.findCalls++
	f.lastFindByKey = input
	if len(f.findByKeyResults) >= f.findCalls {
		result := f.findByKeyResults[f.findCalls-1]
		return result.record, result.found, result.err
	}
	if f.findErr != nil {
		return outport.PaymentAddressIdempotencyRecord{}, false, f.findErr
	}
	if f.found {
		return f.record, true, nil
	}
	return outport.PaymentAddressIdempotencyRecord{}, false, nil
}

func (f *fakePaymentAddressIdempotencyStore) Claim(
	_ context.Context,
	input outport.ClaimPaymentAddressIdempotencyInput,
) (outport.PaymentAddressIdempotencyRecord, error) {
	f.claimCalls++
	f.lastClaim = input
	if f.claimErr != nil {
		return outport.PaymentAddressIdempotencyRecord{}, f.claimErr
	}
	return outport.PaymentAddressIdempotencyRecord{
		Chain:               input.Chain,
		IdempotencyKey:      input.IdempotencyKey,
		AddressPolicyID:     input.AddressPolicyID,
		ExpectedAmountMinor: input.ExpectedAmountMinor,
		CustomerReference:   input.CustomerReference,
	}, nil
}

func (f *fakePaymentAddressIdempotencyStore) Complete(
	_ context.Context,
	input outport.CompletePaymentAddressIdempotencyInput,
) error {
	f.completeCalls++
	f.lastComplete = input
	return f.completeErr
}

func (f *fakePaymentAddressIdempotencyStore) Release(
	_ context.Context,
	input outport.ReleasePaymentAddressIdempotencyInput,
) error {
	f.releaseCalls++
	f.lastRelease = input
	return f.releaseErr
}

type fakeAllocatePaymentReceiptTrackingStore struct {
	createErr          error
	createCalls        int
	lastCreateTracking entities.PaymentReceiptTracking
	lastCreateNextPoll time.Time
}

func (f *fakeAllocatePaymentReceiptTrackingStore) Create(
	_ context.Context,
	tracking entities.PaymentReceiptTracking,
	nextPollAt time.Time,
) error {
	f.createCalls++
	f.lastCreateTracking = tracking
	f.lastCreateNextPoll = nextPollAt
	return f.createErr
}

func (f *fakeAllocatePaymentReceiptTrackingStore) ClaimDue(
	_ context.Context,
	_ outport.ClaimPaymentReceiptTrackingsInput,
) ([]entities.PaymentReceiptTracking, error) {
	return nil, nil
}

func (f *fakeAllocatePaymentReceiptTrackingStore) Save(
	_ context.Context,
	_ entities.PaymentReceiptTracking,
	_ time.Time,
	_ time.Time,
) error {
	return nil
}

type fakeAllocatePaymentReceiptStatusNotificationOutbox struct{}

func (f *fakeAllocatePaymentReceiptStatusNotificationOutbox) EnqueueStatusChanged(
	_ context.Context,
	_ events.PaymentReceiptStatusChanged,
) error {
	return nil
}

func (f *fakeAllocatePaymentReceiptStatusNotificationOutbox) ClaimPending(
	_ context.Context,
	_ outport.ClaimPaymentReceiptStatusNotificationsInput,
) ([]applicationoutbox.PaymentReceiptStatusNotificationOutboxMessage, error) {
	return nil, nil
}

func (f *fakeAllocatePaymentReceiptStatusNotificationOutbox) SaveDeliveryResult(
	_ context.Context,
	_ applicationoutbox.PaymentReceiptStatusNotificationDeliveryResult,
) error {
	return nil
}
